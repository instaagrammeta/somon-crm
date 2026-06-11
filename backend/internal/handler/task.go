package handler

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/i18n"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type TaskHandler struct{ *App }

func NewTaskHandler(a *App) *TaskHandler { return &TaskHandler{a} }

// taskRow is the JSON shape returned to the frontend: the task plus resolved
// author/executor display names. The embedded Task already carries the
// executor_ids / photos arrays (jsonb) used by the multi-select UI.
type taskRow struct {
	models.Task
	AuthorName    string   `json:"author_name"`
	ExecutorName  string   `json:"executor_name"`
	ExecutorNames []string `json:"executor_names"`
}

// GET /api/tasks
func (h *TaskHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	role := middleware.CurrentRole(c)
	q := h.DB.Model(&models.Task{}).Order("created_at DESC")
	if role != models.RoleAdmin {
		// A non-admin sees a task if they authored it, are the primary executor,
		// or appear in the executor_ids array.
		q = q.Where("author_id = ? OR executor_id = ? OR executor_ids @> ?::jsonb",
			uid, uid, uintJSONArray(uid))
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	var rows []taskRow
	if err := q.Select(`tasks.*,
		(SELECT full_name FROM users WHERE id = tasks.author_id)   AS author_name,
		(SELECT full_name FROM users WHERE id = tasks.executor_id) AS executor_name`).
		Scan(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	h.fillExecutorNames(rows)
	c.JSON(http.StatusOK, rows)
}

// fillExecutorNames backfills the array fields from the legacy single-value
// columns (so old rows keep working) and resolves the executor display names.
func (h *TaskHandler) fillExecutorNames(rows []taskRow) {
	idSet := map[uint]bool{}
	for i := range rows {
		r := &rows[i]
		if len(r.ExecutorIDs) == 0 && r.ExecutorID != nil && *r.ExecutorID != 0 {
			r.ExecutorIDs = models.UintSlice{*r.ExecutorID}
		}
		if len(r.Photos) == 0 && r.Photo != "" {
			r.Photos = models.StringSlice{r.Photo}
		}
		for _, id := range r.ExecutorIDs {
			if id != 0 {
				idSet[id] = true
			}
		}
	}
	if len(idSet) == 0 {
		for i := range rows {
			rows[i].ExecutorNames = []string{}
		}
		return
	}
	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}
	var users []struct {
		ID       uint
		FullName string
	}
	h.DB.Model(&models.User{}).Select("id, full_name").Where("id IN ?", ids).Scan(&users)
	nameMap := make(map[uint]string, len(users))
	for _, u := range users {
		nameMap[u.ID] = u.FullName
	}
	for i := range rows {
		names := make([]string, 0, len(rows[i].ExecutorIDs))
		for _, id := range rows[i].ExecutorIDs {
			if n, ok := nameMap[id]; ok {
				names = append(names, n)
			}
		}
		rows[i].ExecutorNames = names
	}
}

// POST /api/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	execIDs := parseExecutorIDs(c)
	photos := h.saveTaskPhotos(c)

	t := models.Task{
		Title:       title,
		Description: c.PostForm("description"),
		AuthorID:    &uid,
		Status:      ifEmpty(c.PostForm("status"), models.TaskStatusNew),
		ExecutorIDs: models.UintSlice(execIDs),
		Photos:      models.StringSlice(photos),
	}
	if len(execIDs) > 0 {
		primary := execIDs[0]
		t.ExecutorID = &primary
	}
	if len(photos) > 0 {
		t.Photo = photos[0]
	}
	if err := h.DB.Create(&t).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	// Notify every assignee (in-app + Telegram), skipping the author themselves.
	for _, eid := range execIDs {
		if eid != 0 && eid != uid {
			h.notifyAssigned(eid, &t)
		}
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "task": t})
}

// PUT /api/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var t models.Task
	if err := h.DB.First(&t, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "task.not_found")
		return
	}
	uid := middleware.CurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)
	isAuthor := t.AuthorID != nil && *t.AuthorID == uid
	isExecutor := taskHasExecutor(&t, uid)
	// Author, executor and admin may all touch the task. The executor is limited
	// to changing the status (see below); author/admin may edit everything.
	if !isAdmin && !isAuthor && !isExecutor {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}

	prevExecutors := taskExecutorSet(&t)
	prevStatus := t.Status
	canEditAll := isAdmin || isAuthor

	updates := map[string]any{}
	// Status can be changed by author, executor and admin.
	if s := c.PostForm("status"); s != "" {
		updates["status"] = s
	}

	if canEditAll {
		if title := strings.TrimSpace(c.PostForm("title")); title != "" {
			updates["title"] = title
		}
		if _, ok := c.GetPostForm("description"); ok {
			updates["description"] = c.PostForm("description")
		}
		// Executors are only rewritten when the edit form explicitly manages
		// them (so the quick status-change action never wipes the assignees).
		if c.PostForm("manage_executors") == "1" {
			execIDs := parseExecutorIDs(c)
			updates["executor_ids"] = models.UintSlice(execIDs)
			if len(execIDs) > 0 {
				updates["executor_id"] = execIDs[0]
			} else {
				updates["executor_id"] = nil
			}
		}
		// Photos are likewise only rewritten on an explicit manage request. The
		// final ordered list = kept existing photos (reordered/deleted on the
		// client) followed by any newly uploaded files.
		if c.PostForm("manage_photos") == "1" {
			final := h.keptPhotos(c, &t)
			final = append(final, h.saveTaskPhotos(c)...)
			updates["photos"] = models.StringSlice(final)
			if len(final) > 0 {
				updates["photo"] = final[0]
			} else {
				updates["photo"] = ""
			}
		}
	}

	if len(updates) > 0 {
		if err := h.DB.Model(&t).Updates(updates).Error; err != nil {
			utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	_ = h.DB.First(&t, id)

	h.notifyTaskChange(uid, &t, prevExecutors, prevStatus)
	utils.OK(c, gin.H{"success": true, "task": t})
}

// notifyAssigned tells a single user that the task is theirs.
func (h *TaskHandler) notifyAssigned(userID uint, t *models.Task) {
	go h.Notif.Push(userID, models.NotifyTaskAssigned,
		i18n.Translate(i18n.LocaleTG, "notify.task_assigned"), t.Title, "/zadacha",
		"tg.task_assigned", t.Title, t.Description)
}

// notifyTaskChange fires in-app + Telegram notifications when a task is updated.
//   - every *newly added* executor is told the task is theirs;
//   - on a status change, the other involved parties are informed.
func (h *TaskHandler) notifyTaskChange(actorID uint, t *models.Task, prevExecutors map[uint]bool, prevStatus string) {
	current := taskExecutorSet(t)
	notified := map[uint]bool{}
	for eid := range current {
		if eid == 0 || eid == actorID || prevExecutors[eid] {
			continue // unchanged assignee or the actor themselves
		}
		h.notifyAssigned(eid, t)
		notified[eid] = true
	}

	if prevStatus == t.Status {
		return
	}
	statusLabel := i18n.Translate(i18n.LocaleTG, "task.status_"+t.Status)
	body := t.Title + " → " + statusLabel
	recipients := map[uint]bool{}
	if t.AuthorID != nil {
		recipients[*t.AuthorID] = true
	}
	for eid := range current {
		recipients[eid] = true
	}
	for r := range recipients {
		if r == 0 || r == actorID || notified[r] {
			continue // skip the actor and anyone already told via "assigned"
		}
		go h.Notif.Push(r, models.NotifyTaskStatus,
			i18n.Translate(i18n.LocaleTG, "notify.task_status_changed"), body, "/zadacha",
			"tg.task_updated", t.Title, statusLabel)
	}
}

// DELETE /api/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var t models.Task
	if err := h.DB.First(&t, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "task.not_found")
		return
	}
	uid := middleware.CurrentUserID(c)
	if !middleware.IsAdmin(c) && t.AuthorID != nil && *t.AuthorID != uid {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	if err := h.DB.Delete(&t).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}

// ===== helpers =====

// saveTaskPhotos persists every uploaded image under the "photos" (and legacy
// "photo") multipart fields, returning their relative URLs in upload order.
func (h *TaskHandler) saveTaskPhotos(c *gin.Context) []string {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return nil
	}
	files := append([]*multipart.FileHeader{}, form.File["photos"]...)
	files = append(files, form.File["photo"]...)
	out := make([]string, 0, len(files))
	for _, f := range files {
		rel, _, err := utils.SaveUpload(f, h.Cfg.Upload.Dir, "tasks")
		if err == nil {
			out = append(out, rel)
		}
	}
	return out
}

// keptPhotos returns the "existing_photos" the client asked to keep, filtered to
// those actually attached to the task (so a request can't inject arbitrary
// paths) and de-duplicated while preserving the client-provided order.
func (h *TaskHandler) keptPhotos(c *gin.Context, t *models.Task) []string {
	allowed := map[string]bool{}
	for _, p := range t.Photos {
		allowed[p] = true
	}
	if t.Photo != "" {
		allowed[t.Photo] = true
	}
	seen := map[string]bool{}
	out := []string{}
	for _, p := range c.PostFormArray("existing_photos") {
		if allowed[p] && !seen[p] {
			out = append(out, p)
			seen[p] = true
		}
	}
	return out
}

// parseExecutorIDs collects assignee user IDs from the request. It accepts the
// repeated "executor_ids" field, a comma-separated value, and the legacy single
// "executor_id" field. Zeros are dropped and duplicates removed (order kept).
func parseExecutorIDs(c *gin.Context) []uint {
	raw := c.PostFormArray("executor_ids")
	if single := c.PostForm("executor_id"); single != "" {
		raw = append(raw, single)
	}
	seen := map[uint]bool{}
	out := []uint{}
	for _, chunk := range raw {
		for _, part := range strings.Split(chunk, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			n, err := strconv.ParseUint(part, 10, 64)
			if err != nil || n == 0 {
				continue
			}
			id := uint(n)
			if !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	return out
}

// taskExecutorSet returns the set of executor IDs, falling back to the legacy
// single executor_id when the array is empty.
func taskExecutorSet(t *models.Task) map[uint]bool {
	set := map[uint]bool{}
	for _, id := range t.ExecutorIDs {
		if id != 0 {
			set[id] = true
		}
	}
	if len(set) == 0 && t.ExecutorID != nil && *t.ExecutorID != 0 {
		set[*t.ExecutorID] = true
	}
	return set
}

func taskHasExecutor(t *models.Task, uid uint) bool {
	return taskExecutorSet(t)[uid]
}

// uintJSONArray builds the JSON text used to test jsonb containment (@>).
func uintJSONArray(id uint) string {
	return "[" + strconv.FormatUint(uint64(id), 10) + "]"
}

func ifEmpty(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
