package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
	"gorm.io/gorm"
)

// ObjektHandler implements "Шахматка" — apartment matrix per project.
type ObjektHandler struct{ *App }

func NewObjektHandler(a *App) *ObjektHandler { return &ObjektHandler{a} }

// ===== Projects =====

// GET /api/objekt/projects
func (h *ObjektHandler) ListProjects(c *gin.Context) {
	var rows []models.ObjektProject
	if err := h.DB.Order("created_at DESC").Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// POST /api/objekt/projects
func (h *ObjektHandler) CreateProject(c *gin.Context) {
	var in models.ObjektProject
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/objekt/projects/:id
func (h *ObjektHandler) UpdateProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p models.ObjektProject
	if err := h.DB.First(&p, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "objekt.not_found")
		return
	}
	var in models.ObjektProject
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = p.ID
	if err := h.DB.Model(&p).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/objekt/projects/:id
func (h *ObjektHandler) DeleteProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		var blocks []models.ObjektBlock
		tx.Where("project_id = ?", id).Find(&blocks)
		for _, b := range blocks {
			if err := tx.Where("block_id = ?", b.ID).Delete(&models.ObjektApartment{}).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("project_id = ?", id).Delete(&models.ObjektBlock{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.ObjektProject{}, id).Error
	})
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// ===== Blocks (with auto apartment generation) =====

// GET /api/objekt/projects/:id/blocks
func (h *ObjektHandler) ListBlocks(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Param("id"))
	var rows []models.ObjektBlock
	h.DB.Where("project_id = ?", pid).Order("order_index ASC, id ASC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

type blockInput struct {
	ProjectID            uint    `form:"project_id" json:"project_id"`
	Name                 string  `form:"name" json:"name"`
	FloorFrom            int     `form:"floor_from" json:"floor_from"`
	FloorTo              int     `form:"floor_to" json:"floor_to"`
	DefaultArea          float64 `form:"default_area" json:"default_area"`
	DefaultRooms         int     `form:"default_rooms" json:"default_rooms"`
	DefaultWindows       int     `form:"default_windows" json:"default_windows"`
	DefaultPricePerM2    float64 `form:"default_price_per_m2" json:"default_price_per_m2"`
	DefaultBalcony       bool    `form:"default_balcony" json:"default_balcony"`
	DefaultBathroomType  string  `form:"default_bathroom_type" json:"default_bathroom_type"`
	DefaultBathroomCount int     `form:"default_bathroom_count" json:"default_bathroom_count"`
	Description          string  `form:"description" json:"description"`
	OrderIndex           int     `form:"order_index" json:"order_index"`
}

// POST /api/objekt/blocks  (multipart for plan_image)
func (h *ObjektHandler) CreateBlock(c *gin.Context) {
	var in blockInput
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	planImage := ""
	if file, err := c.FormFile("plan_image"); err == nil {
		rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "objekt_plans")
		if err == nil {
			planImage = rel
		}
	}
	block := models.ObjektBlock{
		ProjectID:            in.ProjectID,
		Name:                 in.Name,
		FloorFrom:            in.FloorFrom,
		FloorTo:              in.FloorTo,
		DefaultArea:          in.DefaultArea,
		DefaultRooms:         in.DefaultRooms,
		DefaultWindows:       in.DefaultWindows,
		DefaultPricePerM2:    in.DefaultPricePerM2,
		DefaultBalcony:       in.DefaultBalcony,
		DefaultBathroomType:  ifEmpty(in.DefaultBathroomType, "combined"),
		DefaultBathroomCount: in.DefaultBathroomCount,
		DefaultPlanImage:     planImage,
		Description:          in.Description,
		OrderIndex:           in.OrderIndex,
	}
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&block).Error; err != nil {
			return err
		}
		// Auto-generate apartments for each floor
		for f := block.FloorFrom; f <= block.FloorTo; f++ {
			ap := models.ObjektApartment{
				BlockID:       block.ID,
				Floor:         f,
				Area:          block.DefaultArea,
				Rooms:         block.DefaultRooms,
				Windows:       block.DefaultWindows,
				PricePerM2:    block.DefaultPricePerM2,
				TotalPrice:    block.DefaultArea * block.DefaultPricePerM2,
				Status:        models.ObjektStatusFree,
				Balcony:       block.DefaultBalcony,
				BathroomType:  block.DefaultBathroomType,
				BathroomCount: block.DefaultBathroomCount,
				PlanImage:     planImage,
				Description:   block.Description,
			}
			if err := tx.Create(&ap).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": block.ID})
}

// PUT /api/objekt/blocks/:id  (multipart)
func (h *ObjektHandler) UpdateBlock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var b models.ObjektBlock
	if err := h.DB.First(&b, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "objekt.not_found")
		return
	}
	var in blockInput
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	planImage := b.DefaultPlanImage
	if file, err := c.FormFile("plan_image"); err == nil {
		if rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "objekt_plans"); err == nil {
			planImage = rel
		}
	}

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		updates := map[string]any{
			"name":                   in.Name,
			"floor_from":             in.FloorFrom,
			"floor_to":               in.FloorTo,
			"default_area":           in.DefaultArea,
			"default_rooms":          in.DefaultRooms,
			"default_windows":        in.DefaultWindows,
			"default_price_per_m2":   in.DefaultPricePerM2,
			"default_balcony":        in.DefaultBalcony,
			"default_bathroom_type":  ifEmpty(in.DefaultBathroomType, "combined"),
			"default_bathroom_count": in.DefaultBathroomCount,
			"default_plan_image":     planImage,
			"description":            in.Description,
		}
		if err := tx.Model(&b).Updates(updates).Error; err != nil {
			return err
		}
		// Sync apartments: ensure rows for each floor in [floor_from..floor_to]
		for f := in.FloorFrom; f <= in.FloorTo; f++ {
			ap := models.ObjektApartment{
				BlockID: b.ID, Floor: f,
				Area: in.DefaultArea, Rooms: in.DefaultRooms, Windows: in.DefaultWindows,
				PricePerM2: in.DefaultPricePerM2, TotalPrice: in.DefaultArea * in.DefaultPricePerM2,
				Status: models.ObjektStatusFree, Balcony: in.DefaultBalcony,
				BathroomType: ifEmpty(in.DefaultBathroomType, "combined"),
				BathroomCount: in.DefaultBathroomCount, PlanImage: planImage,
				Description: in.Description,
			}
			tx.Where("block_id = ? AND floor = ?", b.ID, f).
				Assign(map[string]any{
					"area":           ap.Area,
					"rooms":          ap.Rooms,
					"windows":        ap.Windows,
					"price_per_m2":   ap.PricePerM2,
					"total_price":    ap.TotalPrice,
					"balcony":        ap.Balcony,
					"bathroom_type":  ap.BathroomType,
					"bathroom_count": ap.BathroomCount,
					"plan_image":     ap.PlanImage,
					"description":    ap.Description,
				}).
				FirstOrCreate(&ap)
		}
		// Drop apartments outside new range
		return tx.Where("block_id = ? AND (floor < ? OR floor > ?)", b.ID, in.FloorFrom, in.FloorTo).
			Delete(&models.ObjektApartment{}).Error
	})
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/objekt/blocks/:id
func (h *ObjektHandler) DeleteBlock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("block_id = ?", id).Delete(&models.ObjektApartment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&models.ObjektBlock{}, id).Error
	})
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// POST /api/objekt/blocks/order  body: [{id, order_index}]
func (h *ObjektHandler) UpdateBlocksOrder(c *gin.Context) {
	var in []struct {
		ID         uint `json:"id"`
		OrderIndex int  `json:"order_index"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	for _, b := range in {
		h.DB.Model(&models.ObjektBlock{}).Where("id = ?", b.ID).Update("order_index", b.OrderIndex)
	}
	utils.OK(c, gin.H{"success": true})
}

// ===== Apartments =====

// GET /api/objekt/projects/:id/apartments
func (h *ObjektHandler) ListApartments(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Param("id"))
	type apRow struct {
		models.ObjektApartment
		BlockName string `json:"block_name"`
	}
	var rows []apRow
	h.DB.Table("objekt_apartments a").
		Select("a.*, b.name AS block_name").
		Joins("JOIN objekt_blocks b ON a.block_id = b.id").
		Where("b.project_id = ?", pid).
		Order("b.order_index ASC, b.id ASC, a.floor ASC").
		Scan(&rows)
	c.JSON(http.StatusOK, rows)
}

// PUT /api/objekt/apartments/:id
func (h *ObjektHandler) UpdateApartment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ap models.ObjektApartment
	if err := h.DB.First(&ap, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "objekt.not_found")
		return
	}
	var in models.ObjektApartment
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = ap.ID
	in.BlockID = ap.BlockID
	in.Floor = ap.Floor
	if in.Area > 0 && in.PricePerM2 > 0 {
		in.TotalPrice = in.Area * in.PricePerM2
	}
	if err := h.DB.Model(&ap).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// PATCH /api/objekt/apartments/:id/status
func (h *ObjektHandler) UpdateApartmentStatus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var in struct {
		Status      string `json:"status" binding:"required"`
		ClientName  string `json:"client_name"`
		ClientPhone string `json:"client_phone"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	updates := map[string]any{"status": in.Status}
	if in.Status == models.ObjektStatusReserved || in.Status == models.ObjektStatusSold {
		updates["client_name"] = in.ClientName
		updates["client_phone"] = in.ClientPhone
	} else {
		updates["client_name"] = ""
		updates["client_phone"] = ""
	}
	if err := h.DB.Model(&models.ObjektApartment{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// POST /api/objekt/upload-image
func (h *ObjektHandler) UploadImage(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	_ = uid
	rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "objekt")
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "path": rel})
}

// GET /api/objekt/export/excel?project_id=
func (h *ObjektHandler) ExportExcel(c *gin.Context) {
	pid, _ := strconv.Atoi(c.Query("project_id"))
	if pid == 0 {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	var blocks []models.ObjektBlock
	h.DB.Where("project_id = ?", pid).Order("order_index ASC, id ASC").Find(&blocks)
	var rows []models.ObjektApartment
	h.DB.Joins("JOIN objekt_blocks b ON b.id = objekt_apartments.block_id").
		Where("b.project_id = ?", pid).
		Order("b.order_index ASC, b.id ASC, objekt_apartments.floor ASC").
		Find(&rows)

	headers := []string{"ID", "Блок", "Ошёна", "Майдон", "Ҳуҷра", "Тиреза",
		"Нархи 1м²", "Нархи умумӣ", "Ҳолат", "Балкон", "Намуди ваннахона",
		"Ҳаҷми ваннахона", "Мизоҷ", "Тел.", "Сана"}
	body := make([][]any, len(rows))
	blockMap := map[uint]string{}
	for _, b := range blocks {
		blockMap[b.ID] = b.Name
	}
	for i, r := range rows {
		body[i] = []any{r.ID, blockMap[r.BlockID], r.Floor, r.Area, r.Rooms, r.Windows,
			r.PricePerM2, r.TotalPrice, r.Status, r.Balcony, r.BathroomType,
			r.BathroomCount, r.ClientName, r.ClientPhone, r.CreatedAt}
	}
	bytes, err := utils.BuildExcel("Шахматка", headers, body)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=shahmatka_"+strconv.Itoa(pid)+".xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", bytes)
}

// GET /api/objekt/export/pdf/:id
// PDF export omitted from initial scope — we return Excel as fallback.
// (gofpdf doesn't support Cyrillic out-of-box without TTF. A future task.)
func (h *ObjektHandler) ExportPDF(c *gin.Context) {
	c.Request.URL.RawQuery = "project_id=" + c.Param("id")
	h.ExportExcel(c)
}
