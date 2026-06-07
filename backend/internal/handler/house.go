package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

type HouseHandler struct{ *App }

func NewHouseHandler(a *App) *HouseHandler { return &HouseHandler{a} }

// GET /api/houses
func (h *HouseHandler) List(c *gin.Context) {
	var rows []models.House
	q := h.DB.Order("created_at DESC")
	if dist := c.Query("district"); dist != "" {
		q = q.Where("district = ?", dist)
	}
	if ctype := c.Query("construction_type"); ctype != "" {
		q = q.Where("construction_type = ?", ctype)
	}
	if rooms := c.Query("rooms"); rooms != "" {
		q = q.Where("rooms = ?", rooms)
	}
	if err := q.Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// POST /api/houses
func (h *HouseHandler) Create(c *gin.Context) {
	var in models.House
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	in.AuthorID = &uid

	// Allow optional file uploads
	if mp, _ := c.MultipartForm(); mp != nil {
		var files []string
		for _, fh := range mp.File["files"] {
			rel, _, err := utils.SaveUpload(fh, h.Cfg.Upload.Dir, "houses")
			if err == nil {
				files = append(files, rel)
			}
		}
		if len(files) > 0 {
			b, _ := jsonMarshal(files)
			in.Files = models.JSONB(b)
		}
	}
	if err := h.DB.Create(&in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "id": in.ID})
}

// PUT /api/houses/:id
func (h *HouseHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var house models.House
	if err := h.DB.First(&house, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "house.not_found")
		return
	}
	var in models.House
	if err := c.ShouldBind(&in); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	in.ID = house.ID
	in.AuthorID = house.AuthorID
	if err := h.DB.Model(&house).Updates(in).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// DELETE /api/houses/:id
func (h *HouseHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.House{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, gin.H{"success": true})
}

// GET /api/houses/export
func (h *HouseHandler) Export(c *gin.Context) {
	var rows []models.House
	h.DB.Order("created_at DESC").Find(&rows)
	headers := []string{"ID", "Унвон", "Намуд", "Ноҳия", "Суроға", "Майдон", "Ҳуҷра", "Тиреза", "Ошёна",
		"Ҷамъи ошёна", "Нархи 1м²", "Нархи умумӣ", "Сохтор", "Тел.", "Сана"}
	body := make([][]any, len(rows))
	for i, r := range rows {
		body[i] = []any{
			r.ID, r.Title, r.ConstructionType, r.District, r.Address,
			r.Area, r.Rooms, r.Windows, r.Floor, r.TotalFloors,
			r.PricePerM2, r.TotalPrice, r.Developer, r.ContactPhone, r.CreatedAt,
		}
	}
	b, err := utils.BuildExcel("Houses", headers, body)
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.Header("Content-Disposition", "attachment; filename=houses.xlsx")
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", b)
}
