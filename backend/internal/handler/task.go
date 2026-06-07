package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
	if t.ExecutorID != nil && *t.ExecutorID != 0 {
		go h.Telegram.NotifyKey(*t.ExecutorID, "tg.task_assigned", t.Title, t.Description)
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
	if !middleware.IsAdmin(c) && t.AuthorID != nil && *t.AuthorID != uid {
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

	updates := map[string]any{}
	if in.Title != "" {
		updates["title"] = in.Title
	}
	if in.Description != "" {
		updates["description"] = in.Description
	}
	if in.ExecutorID != nil {
		updates["executor_id"] = in.ExecutorID
	}
	if in.Status != "" {
		updates["status"] = in.Status
	}
	if file, err := c.FormFile("photo"); err == nil {
		rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "tasks")
		if err == nil {
			updates["photo"] = rel
		}
	}
	if err := h.DB.Model(&t).Updates(updates).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	_ = h.DB.First(&t, id)

	// Notify on executor change or status change
	if t.ExecutorID != nil && *t.ExecutorID != 0 {
		newAssigned := prevExecutor == nil || *prevExecutor != *t.ExecutorID
		statusChanged := prevStatus != t.Status
		if newAssigned {
			go h.Telegram.NotifyKey(*t.ExecutorID, "tg.task_assigned", t.Title, t.Description)
		} else if statusChanged {
			go h.Telegram.NotifyKey(*t.ExecutorID, "tg.task_updated", t.Title, t.Status)
		}
	}
	utils.OK(c, gin.H{"success": true, "task": t})
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
