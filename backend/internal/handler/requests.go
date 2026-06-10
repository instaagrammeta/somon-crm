package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/i18n"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
	"gorm.io/gorm"
)

type RequestsHandler struct{ *App }

func NewRequestsHandler(a *App) *RequestsHandler { return &RequestsHandler{a} }

// ========== BOARDS ==========

// GET /api/requests-boards
func (h *RequestsHandler) ListBoards(c *gin.Context) {
	uid := middleware.CurrentUserID(c)
	q := h.DB.Model(&models.RequestsBoard{}).Order("id DESC")
	if !middleware.IsAdmin(c) {
		q = q.Where("is_public = TRUE OR author_id = ?", uid)
	}
	var boards []models.RequestsBoard
	if err := q.Find(&boards).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, boards)
}

// POST /api/requests-boards
func (h *RequestsHandler) CreateBoard(c *gin.Context) {
	var in struct {
		Title    string `json:"title" binding:"required"`
		Color    string `json:"color"`
		IsPublic bool   `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	board := models.RequestsBoard{
		Title:    in.Title,
		Color:    ifEmpty(in.Color, "#0f172a"),
		IsPublic: in.IsPublic,
		AuthorID: &uid,
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&board).Error; err != nil {
			return err
		}
		// Default columns
		defaults := []struct {
			Title string
			Color string
		}{
			{"Нав", "#3b82f6"},
			{"Дар кор", "#f59e0b"},
			{"Анҷом", "#22c55e"},
			{"Бекор", "#ef4444"},
		}
		for i, d := range defaults {
			col := models.RequestsColumn{BoardID: board.ID, Title: d.Title, Color: d.Color, OrderIndex: i}
			if err := tx.Create(&col).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": board.ID})
}

// PUT /api/requests-boards/:id
func (h *RequestsHandler) UpdateBoard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var board models.RequestsBoard
	if err := h.DB.First(&board, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "board.not_found")
		return
	}
	if !h.canEditBoard(c, &board) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	var in struct {
		Title    string `json:"title"`
		Color    string `json:"color"`
		IsPublic *bool  `json:"is_public"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	updates := map[string]any{}
	if in.Title != "" {
		updates["title"] = in.Title
	}
	if in.Color != "" {
		updates["color"] = in.Color
	}
	if in.IsPublic != nil {
		updates["is_public"] = *in.IsPublic
	}
	if err := h.DB.Model(&board).Updates(updates).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/requests-boards/:id
func (h *RequestsHandler) DeleteBoard(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var board models.RequestsBoard
	if err := h.DB.First(&board, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "board.not_found")
		return
	}
	if !h.canEditBoard(c, &board) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("board_id = ?", board.ID).Delete(&models.RequestItem{}).Error; err != nil {
			return err
		}
		if err := tx.Where("board_id = ?", board.ID).Delete(&models.RequestsColumn{}).Error; err != nil {
			return err
		}
		return tx.Delete(&board).Error
	})
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

func (h *RequestsHandler) canEditBoard(c *gin.Context, b *models.RequestsBoard) bool {
	if middleware.IsAdmin(c) {
		return true
	}
	uid := middleware.CurrentUserID(c)
	return b.AuthorID != nil && *b.AuthorID == uid
}

// ========== COLUMNS ==========

// GET /api/requests-boards/:id/columns
func (h *RequestsHandler) ListColumns(c *gin.Context) {
	boardID, _ := strconv.Atoi(c.Param("id"))
	var cols []models.RequestsColumn
	if err := h.DB.Where("board_id = ?", boardID).Order("order_index ASC").Find(&cols).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, cols)
}

// POST /api/requests-columns
func (h *RequestsHandler) CreateColumn(c *gin.Context) {
	var in struct {
		BoardID uint   `json:"board_id" binding:"required"`
		Title   string `json:"title" binding:"required"`
		Color   string `json:"color"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	var maxOrder int
	h.DB.Model(&models.RequestsColumn{}).
		Where("board_id = ?", in.BoardID).
		Select("COALESCE(MAX(order_index), -1)").
		Row().Scan(&maxOrder)
	col := models.RequestsColumn{
		BoardID:    in.BoardID,
		Title:      in.Title,
		Color:      ifEmpty(in.Color, "#3b82f6"),
		OrderIndex: maxOrder + 1,
	}
	if err := h.DB.Create(&col).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": col.ID})
}

// PUT /api/requests-columns/:id
func (h *RequestsHandler) UpdateColumn(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var col models.RequestsColumn
	if err := h.DB.First(&col, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "column.not_found")
		return
	}
	var in struct {
		Title      string `json:"title"`
		Color      string `json:"color"`
		OrderIndex *int   `json:"order_index"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	updates := map[string]any{}
	if in.Title != "" {
		updates["title"] = in.Title
	}
	if in.Color != "" {
		updates["color"] = in.Color
	}
	if in.OrderIndex != nil {
		updates["order_index"] = *in.OrderIndex
	}
	if err := h.DB.Model(&col).Updates(updates).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/requests-columns/:id
func (h *RequestsHandler) DeleteColumn(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("column_id = ?", id).Delete(&models.RequestItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.RequestsColumn{}, id).Error
	}); err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// PUT /api/requests-columns/:id/order  (swap with target)
func (h *RequestsHandler) SwapColumnOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var in struct {
		TargetID uint `json:"target_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var a, b models.RequestsColumn
		if err := tx.First(&a, id).Error; err != nil {
			return err
		}
		if err := tx.First(&b, in.TargetID).Error; err != nil {
			return err
		}
		ao, bo := a.OrderIndex, b.OrderIndex
		if err := tx.Model(&a).Update("order_index", bo).Error; err != nil {
			return err
		}
		if err := tx.Model(&b).Update("order_index", ao).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// ========== ITEMS ==========

type requestItemInput struct {
	ColumnID     uint    `json:"column_id" binding:"required"`
	BoardID      uint    `json:"board_id" binding:"required"`
	PropertyType string  `json:"property_type"`
	Address      string  `json:"address"`
	Area         float64 `json:"area"`
	Rooms        int     `json:"rooms"`
	Windows      int     `json:"windows"`
	Floor        int     `json:"floor"`
	TotalFloors  int     `json:"total_floors"`
	TotalPrice   float64 `json:"total_price"`
	PricePerM2   float64 `json:"price_per_m2"`
	Phone        string  `json:"phone"`
	ClientName   string  `json:"client_name"`
	Comment      string  `json:"comment"`
	ExecutorID   *uint   `json:"executor_id"`
	Files        any     `json:"files"`
}

// GET /api/requests-board/:id/requests
func (h *RequestsHandler) ListItems(c *gin.Context) {
	boardID, _ := strconv.Atoi(c.Param("id"))
	var rows []map[string]any
	err := h.DB.Table("requests_items r").
		Select(`r.*, u1.full_name AS author_name, u2.full_name AS executor_name`).
		Joins("LEFT JOIN users u1 ON r.author_id = u1.id").
		Joins("LEFT JOIN users u2 ON r.executor_id = u2.id").
		Where("r.board_id = ?", boardID).
		Order("r.column_id ASC, r.order_index ASC").
		Find(&rows).Error
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// POST /api/requests-new
func (h *RequestsHandler) CreateItem(c *gin.Context) {
	var in requestItemInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	uname := middleware.CurrentUserName(c)

	executorName := ""
	if in.ExecutorID != nil && *in.ExecutorID > 0 {
		var u models.User
		if err := h.DB.Select("full_name").First(&u, *in.ExecutorID).Error; err == nil {
			executorName = u.FullName
		}
	}
	var maxOrder int
	h.DB.Model(&models.RequestItem{}).
		Where("column_id = ?", in.ColumnID).
		Select("COALESCE(MAX(order_index), -1)").
		Row().Scan(&maxOrder)

	item := models.RequestItem{
		ColumnID:     in.ColumnID,
		BoardID:      in.BoardID,
		PropertyType: in.PropertyType,
		Address:      in.Address,
		Area:         in.Area,
		Rooms:        in.Rooms,
		Windows:      in.Windows,
		Floor:        in.Floor,
		TotalFloors:  in.TotalFloors,
		TotalPrice:   in.TotalPrice,
		PricePerM2:   in.PricePerM2,
		Phone:        in.Phone,
		ClientName:   in.ClientName,
		Comment:      in.Comment,
		AuthorID:     &uid,
		AuthorName:   uname,
		ExecutorID:   in.ExecutorID,
		ExecutorName: executorName,
		Files:        marshalAnyJSON(in.Files),
		OrderIndex:   maxOrder + 1,
	}
	if err := h.DB.Create(&item).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	if in.ExecutorID != nil && *in.ExecutorID != 0 && *in.ExecutorID != uid {
		go h.Notif.Push(*in.ExecutorID, models.NotifyRequestAssign,
			i18n.Translate(i18n.LocaleTG, "notify.request_assigned"),
			in.ClientName+" · "+in.Phone, "/zayavka",
			"tg.request_assigned", in.ClientName, in.Phone, in.Address)
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": item.ID})
}

// PUT /api/requests-update/:id
func (h *RequestsHandler) UpdateItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.RequestItem
	if err := h.DB.First(&item, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "request.not_found")
		return
	}
	uid := middleware.CurrentUserID(c)
	if !middleware.IsAdmin(c) && item.AuthorID != nil && *item.AuthorID != uid {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	var in requestItemInput
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	executorName := ""
	if in.ExecutorID != nil && *in.ExecutorID > 0 {
		var u models.User
		if err := h.DB.Select("full_name").First(&u, *in.ExecutorID).Error; err == nil {
			executorName = u.FullName
		}
	}
	updates := map[string]any{
		"property_type": in.PropertyType,
		"address":       in.Address,
		"area":          in.Area,
		"rooms":         in.Rooms,
		"windows":       in.Windows,
		"floor":         in.Floor,
		"total_floors":  in.TotalFloors,
		"total_price":   in.TotalPrice,
		"price_per_m2":  in.PricePerM2,
		"phone":         in.Phone,
		"client_name":   in.ClientName,
		"comment":       in.Comment,
		"executor_id":   in.ExecutorID,
		"executor_name": executorName,
		"files":         marshalAnyJSON(in.Files),
	}
	if err := h.DB.Model(&item).Updates(updates).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/requests-delete/:id
func (h *RequestsHandler) DeleteItem(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var item models.RequestItem
	if err := h.DB.First(&item, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "request.not_found")
		return
	}
	uid := middleware.CurrentUserID(c)
	if !middleware.IsAdmin(c) && item.AuthorID != nil && *item.AuthorID != uid {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	if err := h.DB.Delete(&item).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// POST /api/requests-move
func (h *RequestsHandler) MoveItem(c *gin.Context) {
	var in struct {
		RequestID  uint `json:"request_id" binding:"required"`
		ColumnID   uint `json:"column_id" binding:"required"`
		OrderIndex *int `json:"order_index"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var item models.RequestItem
		if err := tx.First(&item, in.RequestID).Error; err != nil {
			return err
		}
		oldCol := item.ColumnID
		oldIdx := item.OrderIndex
		newIdx := 0
		if in.OrderIndex != nil {
			newIdx = *in.OrderIndex
		}
		if oldCol == in.ColumnID {
			if in.OrderIndex == nil || newIdx == oldIdx {
				return nil
			}
			if newIdx < oldIdx {
				if err := tx.Model(&models.RequestItem{}).
					Where("column_id = ? AND order_index >= ? AND order_index < ?", in.ColumnID, newIdx, oldIdx).
					UpdateColumn("order_index", gorm.Expr("order_index + 1")).Error; err != nil {
					return err
				}
			} else {
				if err := tx.Model(&models.RequestItem{}).
					Where("column_id = ? AND order_index > ? AND order_index <= ?", in.ColumnID, oldIdx, newIdx).
					UpdateColumn("order_index", gorm.Expr("order_index - 1")).Error; err != nil {
					return err
				}
			}
			return tx.Model(&item).Update("order_index", newIdx).Error
		}
		// Different column
		if err := tx.Model(&models.RequestItem{}).
			Where("column_id = ? AND order_index > ?", oldCol, oldIdx).
			UpdateColumn("order_index", gorm.Expr("order_index - 1")).Error; err != nil {
			return err
		}
		if in.OrderIndex != nil {
			if err := tx.Model(&models.RequestItem{}).
				Where("column_id = ? AND order_index >= ?", in.ColumnID, newIdx).
				UpdateColumn("order_index", gorm.Expr("order_index + 1")).Error; err != nil {
				return err
			}
		}
		return tx.Model(&item).Updates(map[string]any{
			"column_id":   in.ColumnID,
			"order_index": newIdx,
		}).Error
	})
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// POST /api/upload-request-file
func (h *RequestsHandler) UploadFile(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	rel, ftype, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "requests")
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"name":    file.Filename,
		"path":    rel,
		"type":    ftype,
	})
}

// POST /api/requests-bulk-paste
func (h *RequestsHandler) BulkPaste(c *gin.Context) {
	var in struct {
		ColumnID uint     `json:"column_id" binding:"required"`
		BoardID  uint     `json:"board_id" binding:"required"`
		Lines    []string `json:"lines" binding:"required"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	uname := middleware.CurrentUserName(c)
	created := 0
	for _, raw := range in.Lines {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		// Format: client | phone | address | area | rooms | total_price (tab- or |-separated)
		parts := splitFlexible(raw)
		item := models.RequestItem{
			ColumnID:   in.ColumnID,
			BoardID:    in.BoardID,
			AuthorID:   &uid,
			AuthorName: uname,
		}
		if len(parts) > 0 {
			item.ClientName = parts[0]
		}
		if len(parts) > 1 {
			item.Phone = parts[1]
		}
		if len(parts) > 2 {
			item.Address = parts[2]
		}
		if len(parts) > 3 {
			item.Area, _ = strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		}
		if len(parts) > 4 {
			item.Rooms, _ = strconv.Atoi(strings.TrimSpace(parts[4]))
		}
		if len(parts) > 5 {
			item.TotalPrice, _ = strconv.ParseFloat(strings.TrimSpace(parts[5]), 64)
		}
		if err := h.DB.Create(&item).Error; err == nil {
			created++
		}
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "created": created})
}

// GET /api/requests-export/excel
func (h *RequestsHandler) ExportBoardExcel(c *gin.Context) {
	boardID, _ := strconv.Atoi(c.Query("board_id"))
	if boardID == 0 {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	var board models.RequestsBoard
	if err := h.DB.First(&board, boardID).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "board.not_found")
		return
	}
	var rows []map[string]any
	h.DB.Table("requests_items r").
		Select("r.id, r.client_name, r.phone, r.property_type, r.address, r.area, r.rooms, r.windows, r.floor, r.total_floors, r.total_price, r.price_per_m2, r.author_name, r.executor_name, r.created_at, c.title AS column_title").
		Joins("LEFT JOIN requests_columns c ON r.column_id = c.id").
		Where("r.board_id = ?", boardID).
		Order("c.order_index ASC, r.order_index ASC").
		Find(&rows)
	headers := []string{"ID", "Колонка", "Мизоҷ", "Телефон", "Намуди объект", "Суроға",
		"Майдон", "Ҳуҷра", "Тиреза", "Ошёна", "Ҷамъи ошёна",
		"Нархи умумӣ", "Нархи 1м²", "Муаллиф", "Иҷрокунанда", "Сана"}
	body := make([][]any, len(rows))
	for i, r := range rows {
		body[i] = []any{
			r["id"], r["column_title"], r["client_name"], r["phone"], r["property_type"],
			r["address"], r["area"], r["rooms"], r["windows"], r["floor"], r["total_floors"],
			r["total_price"], r["price_per_m2"], r["author_name"], r["executor_name"], r["created_at"],
		}
	}
	bytes, err := utils.BuildExcel("Заявки", headers, body)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	fname := "requests_board_" + strconv.Itoa(boardID) + ".xlsx"
	c.Header("Content-Disposition", "attachment; filename="+fname)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", bytes)
}

// GET /api/requests-column-export/:id
func (h *RequestsHandler) ExportColumnExcel(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var col models.RequestsColumn
	if err := h.DB.First(&col, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "column.not_found")
		return
	}
	var rows []models.RequestItem
	h.DB.Where("column_id = ?", id).Order("order_index ASC").Find(&rows)
	headers := []string{"ID", "Мизоҷ", "Телефон", "Суроға", "Майдон", "Ҳуҷра", "Нархи умумӣ", "Сана"}
	body := make([][]any, len(rows))
	for i, r := range rows {
		body[i] = []any{r.ID, r.ClientName, r.Phone, r.Address, r.Area, r.Rooms, r.TotalPrice, r.CreatedAt}
	}
	bytes, err := utils.BuildExcel(col.Title, headers, body)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=column_"+strconv.Itoa(id)+".xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", bytes)
}

// ===== helpers =====

func splitFlexible(s string) []string {
	for _, sep := range []string{"\t", "|", ";"} {
		if strings.Contains(s, sep) {
			parts := strings.Split(s, sep)
			for i := range parts {
				parts[i] = strings.TrimSpace(parts[i])
			}
			return parts
		}
	}
	// fallback: split on 2+ spaces
	return strings.Fields(s)
}

func marshalAnyJSON(v any) models.JSONB {
	if v == nil {
		return models.JSONB("[]")
	}
	b, err := jsonMarshal(v)
	if err != nil {
		return models.JSONB("[]")
	}
	return models.JSONB(b)
}
