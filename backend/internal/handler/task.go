package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/i18n"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type TaskHandler struct{ *App }

func NewTaskHandler(a *App) *TaskHandler { return &TaskHandler{a} }

// GET /api/tasks
func (h *TaskHandler) List(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	role := middleware.CurrentRole(c)
	q := h.DB.Model(&models.Task{}).Order("created_at DESC")
	if role != models.RoleAdmin {
		q = q.Where("author_id = ? OR executor_id = ?", uid, uid)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	type taskRow struct {
		models.Task
		AuthorName   string `json:"author_name"`
		ExecutorName string `json:"executor_name"`
	}
	var rows []taskRow
	if err := q.Select(`tasks.*,
		(SELECT full_name FROM users WHERE id = tasks.author_id)   AS author_name,
		(SELECT full_name FROM users WHERE id = tasks.executor_id) AS executor_name`).
		Scan(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

type taskInput struct {
	Title       string `form:"title" json:"title" binding:"required"`
	Description string `form:"description" json:"description"`
	ExecutorID  *uint  `form:"executor_id" json:"executor_id"`
	Status      string `form:"status" json:"status"`
}

// POST /api/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	var in taskInput
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	t := models.Task{
		Title:       in.Title,
		Description: in.Description,
		AuthorID:    &uid,
		ExecutorID:  in.ExecutorID,
		Status:      ifEmpty(in.Status, models.TaskStatusNew),
	}
	if file, err := c.FormFile("photo"); err == nil {
		rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "tasks")
		if err == nil {
			t.Photo = rel
		}
	}
	if err := h.DB.Create(&t).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	// Notify the assignee (in-app + Telegram), unless self-assigned.
	if t.ExecutorID != nil && *t.ExecutorID != 0 && *t.ExecutorID != uid {
		go h.Notif.Push(*t.ExecutorID, models.NotifyTaskAssigned,
			i18n.Translate(i18n.LocaleTG, "notify.task_assigned"), t.Title, "/zadacha",
			"tg.task_assigned", t.Title, t.Description)
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
	isExecutor := t.ExecutorID != nil && *t.ExecutorID == uid
	// Author, executor and admin may all touch the task. The executor is limited
	// to changing the status (see below); author/admin may edit everything.
	if !isAdmin && !isAuthor && !isExecutor {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	var in taskInput
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	prevExecutor := t.ExecutorID
	prevStatus := t.Status
	canEditAll := isAdmin || isAuthor

	updates := map[string]any{}
	// Status can be changed by author, executor and admin.
	if in.Status != "" {
		updates["status"] = in.Status
	}
	// Everything else is author/admin only.
	if canEditAll {
		if in.Title != "" {
			updates["title"] = in.Title
		}
		if in.Description != "" {
			updates["description"] = in.Description
		}
		if in.ExecutorID != nil {
			updates["executor_id"] = in.ExecutorID
		}
		if file, err := c.FormFile("photo"); err == nil {
			rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "tasks")
			if err == nil {
				updates["photo"] = rel
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

	h.notifyTaskChange(uid, &t, prevExecutor, prevStatus)
	utils.OK(c, gin.H{"success": true, "task": t})
}

// notifyTaskChange fires in-app + Telegram notifications when a task is updated.
//   - a (re)assigned executor is told the task is theirs;
//   - on a status change, the *other* party (author/executor) is informed.
func (h *TaskHandler) notifyTaskChange(actorID uint, t *models.Task, prevExecutor *uint, prevStatus string) {
	assignedTo := uint(0)
	if t.ExecutorID != nil && *t.ExecutorID != 0 {
		newAssigned := prevExecutor == nil || *prevExecutor != *t.ExecutorID
		if newAssigned && *t.ExecutorID != actorID {
			assignedTo = *t.ExecutorID
			go h.Notif.Push(*t.ExecutorID, models.NotifyTaskAssigned,
				i18n.Translate(i18n.LocaleTG, "notify.task_assigned"), t.Title, "/zadacha",
				"tg.task_assigned", t.Title, t.Description)
		}
	}

	if prevStatus == t.Status {
		return
	}
	statusLabel := i18n.Translate(i18n.LocaleTG, "task.status_"+t.Status)
	body := t.Title + " → " + statusLabel
	recipients := []uint{}
	if t.AuthorID != nil {
		recipients = append(recipients, *t.AuthorID)
	}
	if t.ExecutorID != nil {
		recipients = append(recipients, *t.ExecutorID)
	}
	for _, r := range recipients {
		if r == 0 || r == actorID || r == assignedTo {
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

func ifEmpty(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
