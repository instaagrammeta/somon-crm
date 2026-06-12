package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// WebsiteLeadHandler is the admin-side dashboard for leads captured by the
// public website + WhatsApp inbound. CRM operators triage, contact, and
// "promote" website leads into proper CRM leads (`lids`) when they convert.
type WebsiteLeadHandler struct{ *App }

func NewWebsiteLeadHandler(a *App) *WebsiteLeadHandler { return &WebsiteLeadHandler{a} }

// GET /api/website-leads?status=&source=&q=&limit=&offset=
func (h *WebsiteLeadHandler) List(c *gin.Context) {
	q := h.DB.Model(&models.WebsiteLead{}).Order("created_at DESC")
	if v := c.Query("status"); v != "" {
		q = q.Where("status = ?", v)
	}
	if v := c.Query("source"); v != "" {
		q = q.Where("source = ?", v)
	}
	if v := strings.TrimSpace(c.Query("q")); v != "" {
		like := "%" + v + "%"
		q = q.Where("name ILIKE ? OR phone ILIKE ? OR message ILIKE ?", like, like, like)
	}
	limit := 50
	if n, _ := strconv.Atoi(c.Query("limit")); n > 0 && n <= 200 {
		limit = n
	}
	offset, _ := strconv.Atoi(c.Query("offset"))
	var rows []models.WebsiteLead
	if err := q.Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// PUT /api/website-leads/:id — update status / notes / promotion.
func (h *WebsiteLeadHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var row models.WebsiteLead
	if err := h.DB.First(&row, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	var body struct {
		Status *string `json:"status"`
		Notes  *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if body.Status != nil {
		row.Status = *body.Status
	}
	if body.Notes != nil {
		row.Notes = *body.Notes
	}
	if err := h.DB.Save(&row).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, row)
}

// POST /api/website-leads/:id/promote — convert to a real CRM lead.
//
// Creates a Lid row (default first kanban column when configured) and
// links the source website_lead row to it for traceability.
func (h *WebsiteLeadHandler) Promote(c *gin.Context) {
	if !middleware.IsAdmin(c) && middleware.CurrentRole(c) != models.RoleManager {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var src models.WebsiteLead
	if err := h.DB.First(&src, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	if src.PromotedLeadID != nil && *src.PromotedLeadID > 0 {
		utils.ErrorResp(c, http.StatusConflict, "site.already_promoted")
		return
	}
	lid := models.Lid{
		ClientName: src.Name,
		Phone:      src.Phone,
		Comment:    src.Message,
		Source:     src.Source,
	}
	if err := h.DB.Create(&lid).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	src.PromotedLeadID = &lid.ID
	src.Status = "promoted"
	h.DB.Save(&src)
	c.JSON(http.StatusOK, gin.H{"lid_id": lid.ID, "website_lead": src})
}

// DELETE /api/website-leads/:id (admin)
func (h *WebsiteLeadHandler) Delete(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.WebsiteLead{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}
