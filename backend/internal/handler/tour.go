package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// TourHandler is the admin-side CRUD for virtual tours, panoramas and
// hotspots. The website-facing read endpoint lives in handler/public.go.
type TourHandler struct{ *App }

func NewTourHandler(a *App) *TourHandler { return &TourHandler{a} }

// GET /api/tours
func (h *TourHandler) List(c *gin.Context) {
	var rows []models.Tour
	if err := h.DB.Order("created_at DESC").Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}

// GET /api/tours/:id — full payload (tour + panoramas + hotspots).
func (h *TourHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var t models.Tour
	if err := h.DB.First(&t, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	var panos []models.Panorama
	h.DB.Where("tour_id = ?", t.ID).Order("sort_order ASC, id ASC").Find(&panos)
	for i := range panos {
		var hs []models.PanoramaHotspot
		h.DB.Where("panorama_id = ?", panos[i].ID).Find(&hs)
		panos[i].Hotspots = hs
	}
	t.Panoramas = panos
	c.JSON(http.StatusOK, t)
}

// POST /api/tours  (admin)
func (h *TourHandler) Create(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	var body models.Tour
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	if body.Slug == "" {
		body.Slug = utils.MakeSlug(body.Title)
	}
	if err := h.DB.Create(&body).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, body)
}

// PUT /api/tours/:id  (admin)
func (h *TourHandler) Update(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var t models.Tour
	if err := h.DB.First(&t, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	var body models.Tour
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	t.Title = body.Title
	t.Slug = body.Slug
	t.Description = body.Description
	t.HouseID = body.HouseID
	t.ObjectID = body.ObjectID
	t.CoverImage = body.CoverImage
	t.PublicVisible = body.PublicVisible
	t.InitialPanoramaID = body.InitialPanoramaID
	if err := h.DB.Save(&t).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, t)
}

// DELETE /api/tours/:id  (admin) — cascades to panoramas + hotspots.
func (h *TourHandler) Delete(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Tour{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}

// POST /api/tours/:id/panoramas  (admin) — multipart with file=image
//
// We treat each panorama upload as one row that links to the tour. The image
// itself is saved under uploads/tours/<tourID>/.
func (h *TourHandler) UploadPanorama(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	tid, _ := strconv.Atoi(c.Param("id"))
	var t models.Tour
	if err := h.DB.First(&t, tid).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	rel, _, err := utils.SaveUpload(fh, h.Cfg.Upload.Dir, "tours/"+strconv.FormatUint(uint64(t.ID), 10))
	if err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	title := c.PostForm("title")
	if title == "" {
		title = fh.Filename
	}
	row := models.Panorama{
		TourID:    t.ID,
		Title:     title,
		ImageURL:  rel,
		ImageType: "equirect",
	}
	if v := c.PostForm("image_type"); v != "" {
		row.ImageType = v
	}
	if v, _ := strconv.Atoi(c.PostForm("sort_order")); v > 0 {
		row.SortOrder = v
	}
	if err := h.DB.Create(&row).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	// First panorama becomes the default starting scene.
	if t.InitialPanoramaID == nil {
		h.DB.Model(&t).Update("initial_panorama_id", row.ID)
	}
	c.JSON(http.StatusCreated, row)
}

// PUT /api/panoramas/:id  (admin) — update title/yaw/pitch/zoom/sort.
func (h *TourHandler) UpdatePanorama(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var p models.Panorama
	if err := h.DB.First(&p, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	var body struct {
		Title        *string  `json:"title"`
		ImageURL     *string  `json:"image_url"`
		ImageType    *string  `json:"image_type"`
		InitialYaw   *float64 `json:"initial_yaw"`
		InitialPitch *float64 `json:"initial_pitch"`
		InitialZoom  *float64 `json:"initial_zoom"`
		SortOrder    *int     `json:"sort_order"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.Title != nil {
		p.Title = *body.Title
	}
	if body.ImageURL != nil {
		p.ImageURL = *body.ImageURL
	}
	if body.ImageType != nil {
		p.ImageType = *body.ImageType
	}
	if body.InitialYaw != nil {
		p.InitialYaw = *body.InitialYaw
	}
	if body.InitialPitch != nil {
		p.InitialPitch = *body.InitialPitch
	}
	if body.InitialZoom != nil {
		p.InitialZoom = *body.InitialZoom
	}
	if body.SortOrder != nil {
		p.SortOrder = *body.SortOrder
	}
	if err := h.DB.Save(&p).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, p)
}

// DELETE /api/panoramas/:id  (admin)
func (h *TourHandler) DeletePanorama(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.Panorama{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}

// POST /api/panoramas/:id/hotspots  (admin)
func (h *TourHandler) CreateHotspot(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	pid, _ := strconv.Atoi(c.Param("id"))
	var p models.Panorama
	if err := h.DB.First(&p, pid).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	var body models.PanoramaHotspot
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	body.PanoramaID = p.ID
	if body.Kind == "" {
		body.Kind = "link"
	}
	if err := h.DB.Create(&body).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, body)
}

// DELETE /api/hotspots/:id  (admin)
func (h *TourHandler) DeleteHotspot(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.DB.Delete(&models.PanoramaHotspot{}, id).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.OK(c, nil)
}
