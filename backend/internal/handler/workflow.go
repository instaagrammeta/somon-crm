package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// WorkflowHandler — admin CRUD for the no-code automation builder. Only admins
// may create / edit / delete workflows; everyone else may only list them
// (read-only) so the UI can show "what is automated".
type WorkflowHandler struct{ *App }

func NewWorkflowHandler(a *App) *WorkflowHandler { return &WorkflowHandler{a} }

// GET /api/workflows
func (h *WorkflowHandler) List(c *gin.Context) {
	var rows []models.Workflow
	if err := h.DB.Order("created_at DESC").Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// POST /api/workflows
func (h *WorkflowHandler) Create(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	var body models.Workflow
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Create(&body).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, body)
}

// PUT /api/workflows/:id
func (h *WorkflowHandler) Update(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var row models.Workflow
	if err := h.DB.First(&row, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	var body models.Workflow
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	row.Name = body.Name
	row.Description = body.Description
	row.Enabled = body.Enabled
	row.Trigger = body.Trigger
	row.Config = body.Config
	row.Steps = body.Steps
	if err := h.DB.Save(&row).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, row)
}

// DELETE /api/workflows/:id
func (h *WorkflowHandler) Delete(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Workflow{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}

// GET /api/workflows/:id/runs — last 50 runs of a workflow.
func (h *WorkflowHandler) Runs(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var rows []models.WorkflowRun
	if err := h.DB.Where("workflow_id = ?", id).Order("started_at DESC").Limit(50).Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// GET /api/audit-logs?entity=&user_id=&limit=
func (h *WorkflowHandler) AuditList(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	q := h.DB.Model(&models.AuditLog{}).Order("created_at DESC")
	if v := c.Query("entity"); v != "" {
		q = q.Where("entity = ?", v)
	}
	if v := c.Query("user_id"); v != "" {
		q = q.Where("user_id = ?", v)
	}
	limit := 100
	if v, _ := strconv.Atoi(c.Query("limit")); v > 0 && v <= 500 {
		limit = v
	}
	var rows []models.AuditLog
	if err := q.Limit(limit).Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}
