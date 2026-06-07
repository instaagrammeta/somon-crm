package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type RealtyHandler struct{ *App }

func NewRealtyHandler(a *App) *RealtyHandler { return &RealtyHandler{a} }

// ========== OBJECTS ==========

// GET /api/realty/objects
func (h *RealtyHandler) ListObjects(c *gin.Context) {
	var rows []models.RealtyObject
	if err := h.DB.Order("created_at DESC").Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// POST /api/realty/objects
func (h *RealtyHandler) CreateObject(c *gin.Context) {
	var in models.RealtyObject
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

// PUT /api/realty/objects/:id
func (h *RealtyHandler) UpdateObject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var obj models.RealtyObject
	if err := h.DB.First(&obj, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "realty.not_found")
		return
	}
	var in models.RealtyObject
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = obj.ID
	if err := h.DB.Model(&obj).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/realty/objects/:id
func (h *RealtyHandler) DeleteObject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.RealtyObject{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// ========== BLOCKS ==========

// GET /api/realty/objects/:id/blocks
func (h *RealtyHandler) ListBlocks(c *gin.Context) {
	objID, _ := strconv.Atoi(c.Param("id"))
	var rows []models.RealtyBlock
	h.DB.Where("object_id = ?", objID).Order("order_index ASC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/realty/blocks
func (h *RealtyHandler) CreateBlock(c *gin.Context) {
	var in models.RealtyBlock
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

// PUT /api/realty/blocks/:id
func (h *RealtyHandler) UpdateBlock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var b models.RealtyBlock
	if err := h.DB.First(&b, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "realty.not_found")
		return
	}
	var in models.RealtyBlock
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

// DELETE /api/realty/blocks/:id
func (h *RealtyHandler) DeleteBlock(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.RealtyBlock{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// ========== PRICING ==========

// GET /api/realty/pricing?object_id=
func (h *RealtyHandler) ListPricing(c *gin.Context) {
	objID := c.Query("object_id")
	var rows []models.RealtyPricing
	q := h.DB.Order("floor_number ASC")
	if objID != "" {
		q = q.Where("object_id = ?", objID)
	}
	q.Find(&rows)
	c.JSON(http.StatusOK, rows)
}

func (h *RealtyHandler) CreatePricing(c *gin.Context) {
	var in models.RealtyPricing
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

func (h *RealtyHandler) UpdatePricing(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var p models.RealtyPricing
	if err := h.DB.First(&p, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "realty.not_found")
		return
	}
	var in models.RealtyPricing
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

func (h *RealtyHandler) DeletePricing(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.RealtyPricing{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// ========== LAYOUTS ==========

func (h *RealtyHandler) ListLayouts(c *gin.Context) {
	objID := c.Query("object_id")
	var rows []models.RealtyLayout
	q := h.DB.Order("id ASC")
	if objID != "" {
		q = q.Where("object_id = ?", objID)
	}
	q.Find(&rows)
	c.JSON(http.StatusOK, rows)
}

func (h *RealtyHandler) CreateLayout(c *gin.Context) {
	var in models.RealtyLayout
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

func (h *RealtyHandler) UpdateLayout(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var l models.RealtyLayout
	if err := h.DB.First(&l, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "realty.not_found")
		return
	}
	var in models.RealtyLayout
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

func (h *RealtyHandler) DeleteLayout(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.RealtyLayout{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}
