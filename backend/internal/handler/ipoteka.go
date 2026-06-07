package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// IpotekaHandler covers /api/banks/* and /api/installment-objects/*.
type IpotekaHandler struct{ *App }

func NewIpotekaHandler(a *App) *IpotekaHandler { return &IpotekaHandler{a} }

// ===== Banks =====

// GET /api/banks
func (h *IpotekaHandler) ListBanks(c *gin.Context) {
	var rows []models.Bank
	h.DB.Where("is_active = TRUE").Order("order_index ASC, id ASC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// GET /api/banks/:id
func (h *IpotekaHandler) GetBank(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var b models.Bank
	if err := h.DB.First(&b, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "bank.not_found")
		return
	}
	c.JSON(http.StatusOK, b)
}

// GET /api/banks/by-slug/:slug
func (h *IpotekaHandler) GetBankBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var b models.Bank
	if err := h.DB.Where("slug = ? AND is_active = TRUE", slug).First(&b).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "bank.not_found")
		return
	}
	c.JSON(http.StatusOK, b)
}

// POST /api/banks (admin, multipart for logo)
func (h *IpotekaHandler) CreateBank(c *gin.Context) {
	var in struct {
		Name        string `form:"name" binding:"required"`
		Description string `form:"description"`
		OrderIndex  int    `form:"order_index"`
	}
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	bank := models.Bank{
		Name:        in.Name,
		Description: in.Description,
		OrderIndex:  in.OrderIndex,
		IsActive:    true,
		Slug: utils.EnsureUniqueSlug(utils.MakeSlug(in.Name), func(s string) bool {
			var n int64
			h.DB.Model(&models.Bank{}).Where("slug = ?", s).Count(&n)
			return n == 0
		}),
	}
	if file, err := c.FormFile("logo"); err == nil {
		if rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "banks"); err == nil {
			bank.Logo = rel
		}
	}
	if err := h.DB.Create(&bank).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, bank)
}

// PUT /api/banks/:id (multipart)
func (h *IpotekaHandler) UpdateBank(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var b models.Bank
	if err := h.DB.First(&b, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "bank.not_found")
		return
	}
	updates := map[string]any{}
	if v := c.PostForm("name"); v != "" {
		updates["name"] = v
		updates["slug"] = utils.EnsureUniqueSlug(utils.MakeSlug(v), func(s string) bool {
			var n int64
			h.DB.Model(&models.Bank{}).Where("slug = ? AND id <> ?", s, b.ID).Count(&n)
			return n == 0
		})
	}
	if v := c.PostForm("description"); v != "" {
		updates["description"] = v
	}
	if v := c.PostForm("order_index"); v != "" {
		i, _ := strconv.Atoi(v)
		updates["order_index"] = i
	}
	if v := c.PostForm("is_active"); v != "" {
		updates["is_active"] = v == "true"
	}
	if file, err := c.FormFile("logo"); err == nil {
		if rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "banks"); err == nil {
			updates["logo"] = rel
		}
	}
	if err := h.DB.Model(&b).Updates(updates).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/banks/:id
func (h *IpotekaHandler) DeleteBank(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Bank{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/banks/:id/conditions
func (h *IpotekaHandler) ListBankConditions(c *gin.Context) {
	bankID, _ := strconv.Atoi(c.Param("id"))
	var rows []models.MortgageCondition
	h.DB.Where("bank_id = ?", bankID).Order("currency ASC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/banks/:id/conditions
func (h *IpotekaHandler) CreateBankCondition(c *gin.Context) {
	bankID, _ := strconv.Atoi(c.Param("id"))
	var in models.MortgageCondition
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.BankID = uint(bankID)
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, in)
}

// PUT /api/conditions/:id
func (h *IpotekaHandler) UpdateBankCondition(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var cond models.MortgageCondition
	if err := h.DB.First(&cond, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "bank.not_found")
		return
	}
	var in models.MortgageCondition
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = cond.ID
	in.BankID = cond.BankID
	if err := h.DB.Model(&cond).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/conditions/:id
func (h *IpotekaHandler) DeleteBankCondition(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.MortgageCondition{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// ===== Installment objects (Rasrochka) =====

// GET /api/installment-objects
func (h *IpotekaHandler) ListInstallments(c *gin.Context) {
	var rows []models.InstallmentObject
	h.DB.Where("is_active = TRUE").Order("order_index ASC, id ASC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// GET /api/installment-objects/:id
func (h *IpotekaHandler) GetInstallment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var o models.InstallmentObject
	if err := h.DB.First(&o, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "installment.not_found")
		return
	}
	c.JSON(http.StatusOK, o)
}

// GET /api/installment-objects/by-slug/:slug
func (h *IpotekaHandler) GetInstallmentBySlug(c *gin.Context) {
	slug := c.Param("slug")
	var o models.InstallmentObject
	if err := h.DB.Where("slug = ? AND is_active = TRUE", slug).First(&o).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "installment.not_found")
		return
	}
	c.JSON(http.StatusOK, o)
}

// POST /api/installment-objects (multipart for image)
func (h *IpotekaHandler) CreateInstallment(c *gin.Context) {
	var in struct {
		Name        string `form:"name" binding:"required"`
		Description string `form:"description"`
		Address     string `form:"address"`
		Developer   string `form:"developer"`
		OrderIndex  int    `form:"order_index"`
	}
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	obj := models.InstallmentObject{
		Name: in.Name, Description: in.Description, Address: in.Address,
		Developer: in.Developer, OrderIndex: in.OrderIndex, IsActive: true,
		Slug: utils.EnsureUniqueSlug(utils.MakeSlug(in.Name), func(s string) bool {
			var n int64
			h.DB.Model(&models.InstallmentObject{}).Where("slug = ?", s).Count(&n)
			return n == 0
		}),
	}
	if file, err := c.FormFile("image"); err == nil {
		if rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "installments"); err == nil {
			obj.Image = rel
		}
	}
	if err := h.DB.Create(&obj).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, obj)
}

// PUT /api/installment-objects/:id (multipart)
func (h *IpotekaHandler) UpdateInstallment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var o models.InstallmentObject
	if err := h.DB.First(&o, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "installment.not_found")
		return
	}
	updates := map[string]any{}
	if v := c.PostForm("name"); v != "" {
		updates["name"] = v
		updates["slug"] = utils.EnsureUniqueSlug(utils.MakeSlug(v), func(s string) bool {
			var n int64
			h.DB.Model(&models.InstallmentObject{}).Where("slug = ? AND id <> ?", s, o.ID).Count(&n)
			return n == 0
		})
	}
	for _, k := range []string{"description", "address", "developer"} {
		if v := c.PostForm(k); v != "" {
			updates[k] = v
		}
	}
	if v := c.PostForm("order_index"); v != "" {
		i, _ := strconv.Atoi(v)
		updates["order_index"] = i
	}
	if v := c.PostForm("is_active"); v != "" {
		updates["is_active"] = v == "true"
	}
	if file, err := c.FormFile("image"); err == nil {
		if rel, _, err := utils.SaveUpload(file, h.Cfg.Upload.Dir, "installments"); err == nil {
			updates["image"] = rel
		}
	}
	if err := h.DB.Model(&o).Updates(updates).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/installment-objects/:id
func (h *IpotekaHandler) DeleteInstallment(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.InstallmentObject{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/installment-objects/:id/conditions
func (h *IpotekaHandler) ListInstallmentConditions(c *gin.Context) {
	objID, _ := strconv.Atoi(c.Param("id"))
	var rows []models.InstallmentCondition
	h.DB.Where("object_id = ?", objID).Order("months ASC").Find(&rows)
	c.JSON(http.StatusOK, rows)
}

// POST /api/installment-objects/:id/conditions
func (h *IpotekaHandler) CreateInstallmentCondition(c *gin.Context) {
	objID, _ := strconv.Atoi(c.Param("id"))
	var in models.InstallmentCondition
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ObjectID = uint(objID)
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, in)
}

// PUT /api/installment-conditions/:id
func (h *IpotekaHandler) UpdateInstallmentCondition(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var cond models.InstallmentCondition
	if err := h.DB.First(&cond, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "installment.not_found")
		return
	}
	var in models.InstallmentCondition
	if err := c.ShouldBindJSON(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = cond.ID
	in.ObjectID = cond.ObjectID
	if err := h.DB.Model(&cond).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/installment-conditions/:id
func (h *IpotekaHandler) DeleteInstallmentCondition(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.InstallmentCondition{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}
