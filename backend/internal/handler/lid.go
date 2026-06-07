package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// LidHandler covers BOTH the legacy `lids` table and the modern Kanban (`kanban_*`) flow.
type LidHandler struct{ *App }

func NewLidHandler(a *App) *LidHandler { return &LidHandler{a} }

// ===== legacy lids =====

// GET /api/lids
func (h *LidHandler) List(c *gin.Context) {
	var rows []models.Lid
	if err := h.DB.Order("created_at DESC").Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// POST /api/lids
func (h *LidHandler) Create(c *gin.Context) {
	var in models.Lid
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	in.AuthorID = &uid
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/lids/:id
func (h *LidHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var l models.Lid
	if err := h.DB.First(&l, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "lead.not_found")
		return
	}
	var in models.Lid
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = l.ID
	if err := h.DB.Model(&l).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/lids/:id
func (h *LidHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Lid{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/lids/export
func (h *LidHandler) Export(c *gin.Context) {
	var rows []models.Lid
	h.DB.Order("created_at DESC").Find(&rows)
	headers := []string{"ID", "Мизоҷ", "Телефон", "Мавзуъ", "Сарчашма", "Ипотека", "Каропка", "Изоҳ", "Сана"}
	body := make([][]any, len(rows))
	for i, r := range rows {
		body[i] = []any{r.ID, r.ClientName, r.Phone, r.Topic, r.Source, r.Mortgage, r.Box, r.Comment, r.CreatedAt}
	}
	b, err := utils.BuildExcel("Lids", headers, body)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=lids.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", b)
}

// ===== Kanban (modern) =====

// Boards =====
// GET /api/kanban/boards
func (h *LidHandler) KanbanBoards(c *gin.Context) {
	q := h.DB.Model(&models.KanbanBoard{}).Where("is_archived = FALSE").Order("id DESC")
	uid := middleware.CurrentUserID(c)
	if !middleware.IsAdmin(c) {
		q = q.Where("is_public = TRUE OR author_id = ?", uid)
	}
	var rows []models.KanbanBoard
	if err := q.Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// POST /api/kanban/boards
func (h *LidHandler) KanbanCreateBoard(c *gin.Context) {
	var in models.KanbanBoard
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	in.AuthorID = &uid
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/kanban/boards/:id
func (h *LidHandler) KanbanUpdateBoard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var b models.KanbanBoard
	if err := h.DB.First(&b, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "board.not_found")
		return
	}
	var in models.KanbanBoard
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = b.ID
	if err := h.DB.Model(&b).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/kanban/boards/:id
func (h *LidHandler) KanbanDeleteBoard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.KanbanBoard{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// PUT /api/kanban/boards/:id/archive
func (h *LidHandler) KanbanArchiveBoard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Model(&models.KanbanBoard{}).Where("id = ?", id).Update("is_archived", true).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// Columns =====
// GET /api/kanban/boards/:id/columns
func (h *LidHandler) KanbanColumns(c *gin.Context) {
	bid, _ := strconv.Atoi(c.Param("id"))
	var rows []models.KanbanColumn
	h.DB.Where("board_id = ? AND is_archived = FALSE", bid).Order("order_index ASC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/kanban/columns
func (h *LidHandler) KanbanCreateColumn(c *gin.Context) {
	var in models.KanbanColumn
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	var maxOrder int
	h.DB.Model(&models.KanbanColumn{}).Where("board_id = ?", in.BoardID).
		Select("COALESCE(MAX(order_index), -1)").Row().Scan(&maxOrder)
	in.OrderIndex = maxOrder + 1
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/kanban/columns/:id
func (h *LidHandler) KanbanUpdateColumn(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var col models.KanbanColumn
	if err := h.DB.First(&col, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "column.not_found")
		return
	}
	var in models.KanbanColumn
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = col.ID
	if err := h.DB.Model(&col).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/kanban/columns/:id
func (h *LidHandler) KanbanDeleteColumn(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.KanbanColumn{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// PUT /api/kanban/columns/:id/order
func (h *LidHandler) KanbanColumnOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var in struct {
		OrderIndex int `json:"order_index"`
	}
	c.ShouldBindJSON(&in)
	h.DB.Model(&models.KanbanColumn{}).Where("id = ?", id).Update("order_index", in.OrderIndex)
	utils.OK(c, gin.H{"success": true})
}

// Leads =====
// GET /api/kanban/leads?board_id=
func (h *LidHandler) KanbanLeads(c *gin.Context) {
	var rows []models.KanbanLead
	q := h.DB.Order("column_id ASC, order_index ASC")
	if bid := c.Query("board_id"); bid != "" {
		q = q.Where("board_id = ?", bid)
	}
	if cid := c.Query("column_id"); cid != "" {
		q = q.Where("column_id = ?", cid)
	}
	q.Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/kanban/leads
func (h *LidHandler) KanbanCreateLead(c *gin.Context) {
	var in models.KanbanLead
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	in.AuthorID = &uid
	in.AuthorName = middleware.CurrentUserName(c)
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/kanban/leads/:id
func (h *LidHandler) KanbanUpdateLead(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var l models.KanbanLead
	if err := h.DB.First(&l, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "lead.not_found")
		return
	}
	var in models.KanbanLead
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = l.ID
	if err := h.DB.Model(&l).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/kanban/leads/:id
func (h *LidHandler) KanbanDeleteLead(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.KanbanLead{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// POST /api/kanban/leads/move
func (h *LidHandler) KanbanMoveLead(c *gin.Context) {
	var in struct {
		LeadID     uint `json:"lead_id" binding:"required"`
		ColumnID   uint `json:"column_id" binding:"required"`
		OrderIndex int  `json:"order_index"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	var lead models.KanbanLead
	if err := h.DB.First(&lead, in.LeadID).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "lead.not_found")
		return
	}
	var oldColTitle, newColTitle string
	h.DB.Model(&models.KanbanColumn{}).Where("id = ?", lead.ColumnID).Select("title").Row().Scan(&oldColTitle)
	h.DB.Model(&models.KanbanColumn{}).Where("id = ?", in.ColumnID).Select("title").Row().Scan(&newColTitle)

	if err := h.DB.Model(&lead).Updates(map[string]any{
		"column_id":   in.ColumnID,
		"order_index": in.OrderIndex,
	}).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	if lead.AuthorID != nil && *lead.AuthorID != 0 && oldColTitle != newColTitle {
		go h.Telegram.NotifyKey(*lead.AuthorID, "tg.lead_moved", oldColTitle, newColTitle)
	}
	utils.OK(c, gin.H{"success": true})
}

// POST /api/kanban/actions/track
func (h *LidHandler) KanbanTrack(c *gin.Context) {
	var in models.KanbanLeadInteraction
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true, "id": in.ID})
}

// GET /api/kanban/boards/:id/export/excel
func (h *LidHandler) KanbanExport(c *gin.Context) {
	bid, _ := strconv.Atoi(c.Param("id"))
	var rows []map[string]any
	h.DB.Table("kanban_leads l").
		Select("l.id, l.client_name, l.phone, l.topic, l.source, l.mortgage, l.box, l.comment, c.title AS column_title, l.created_at").
		Joins("LEFT JOIN kanban_columns c ON l.column_id = c.id").
		Where("l.board_id = ?", bid).
		Order("c.order_index ASC, l.order_index ASC").Find(&rows)
	headers := []string{"ID", "Колонка", "Мизоҷ", "Телефон", "Мавзуъ", "Сарчашма", "Ипотека", "Каропка", "Изоҳ", "Сана"}
	body := make([][]any, len(rows))
	for i, r := range rows {
		body[i] = []any{r["id"], r["column_title"], r["client_name"], r["phone"], r["topic"], r["source"], r["mortgage"], r["box"], r["comment"], r["created_at"]}
	}
	b, err := utils.BuildExcel("Kanban", headers, body)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=kanban_board_"+strconv.Itoa(bid)+".xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", b)
}
