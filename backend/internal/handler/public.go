package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// PublicHandler serves the website-facing API. EVERY route here is unauth-
// enticated and only exposes data marked `public_visible = true`.
//
// The frontend uses these endpoints to render the public marketing site
// (houses, virtual tours, contact form, lead-capture popup) and to download
// the Android APK.
type PublicHandler struct{ *App }

func NewPublicHandler(a *App) *PublicHandler { return &PublicHandler{a} }

// publicHouseRow is what we send to the website. It hides CRM-only fields
// (developer phone, ownership notes, etc.) and only carries fields the
// marketing site renders.
type publicHouseRow struct {
	ID              uint     `json:"id"`
	Title           string   `json:"title"`
	District        string   `json:"district"`
	Address         string   `json:"address"`
	Area            float64  `json:"area"`
	Rooms           int      `json:"rooms"`
	Floor           int      `json:"floor"`
	TotalFloors     int      `json:"total_floors"`
	PublicPriceFrom int64    `json:"price_from"`
	ShortDesc       string   `json:"short_desc"`
	FullDesc        string   `json:"full_desc"`
	Gallery         []string `json:"gallery"`
	Features        []string `json:"features"`
	Lat             float64  `json:"lat"`
	Lng             float64  `json:"lng"`
	HasTour         bool     `json:"has_tour"`
	TourSlug        string   `json:"tour_slug,omitempty"`
}

// GET /api/public/site/config — branding data the SPA reads on first paint.
func (h *PublicHandler) Config(c *gin.Context) {
	s := h.Cfg.Site
	c.JSON(http.StatusOK, gin.H{
		"brand_name":   s.BrandName,
		"phone":        s.Phone,
		"email":        s.Email,
		"address":      s.Address,
		"apk_url":      s.APKDownloadURL,
		"apk_version":  s.APKVersion,
		"locales":      []string{"tg", "ru"},
		"chat_enabled": h.Cfg.WhatsApp.Enabled,
	})
}

// GET /api/public/houses?district=&min_price=&max_price=&rooms=&q=&limit=
func (h *PublicHandler) Houses(c *gin.Context) {
	q := h.DB.Model(&models.House{}).
		Where("public_visible = ?", true).
		Order("created_at DESC")

	if v := strings.TrimSpace(c.Query("district")); v != "" {
		q = q.Where("district ILIKE ?", "%"+v+"%")
	}
	if v := c.Query("rooms"); v != "" {
		q = q.Where("rooms = ?", v)
	}
	if v, _ := strconv.ParseInt(c.Query("min_price"), 10, 64); v > 0 {
		q = q.Where("public_price_from >= ?", v)
	}
	if v, _ := strconv.ParseInt(c.Query("max_price"), 10, 64); v > 0 {
		q = q.Where("public_price_from <= ?", v)
	}
	if v := strings.TrimSpace(c.Query("q")); v != "" {
		like := "%" + v + "%"
		q = q.Where("title ILIKE ? OR address ILIKE ? OR public_short_desc ILIKE ?", like, like, like)
	}

	limit := 24
	if n, _ := strconv.Atoi(c.Query("limit")); n > 0 && n <= 100 {
		limit = n
	}
	offset, _ := strconv.Atoi(c.Query("offset"))

	var houses []models.House
	if err := q.Limit(limit).Offset(offset).Find(&houses).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]publicHouseRow, 0, len(houses))
	for _, ho := range houses {
		row := toPublicRow(ho)
		// Lookup tour (best effort — keep it cheap).
		var tour models.Tour
		if err := h.DB.Where("house_id = ? AND public_visible = ?", ho.ID, true).First(&tour).Error; err == nil {
			row.HasTour = true
			row.TourSlug = tour.Slug
		}
		out = append(out, row)
	}
	c.JSON(http.StatusOK, out)
}

// GET /api/public/houses/:id — full detail.
func (h *PublicHandler) HouseDetail(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var ho models.House
	if err := h.DB.Where("id = ? AND public_visible = ?", id, true).First(&ho).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	row := toPublicRow(ho)
	var tour models.Tour
	if err := h.DB.Where("house_id = ? AND public_visible = ?", ho.ID, true).First(&tour).Error; err == nil {
		row.HasTour = true
		row.TourSlug = tour.Slug
	}
	c.JSON(http.StatusOK, row)
}

// GET /api/public/tours/:slug — full virtual tour with panoramas + hotspots,
// ready for Marzipano on the frontend.
func (h *PublicHandler) Tour(c *gin.Context) {
	slug := c.Param("slug")
	var tour models.Tour
	q := h.DB.Where("public_visible = ?", true)
	// Accept either the slug or the numeric ID for convenience.
	if id, err := strconv.Atoi(slug); err == nil && id > 0 {
		q = q.Where("id = ?", id)
	} else {
		q = q.Where("slug = ?", slug)
	}
	if err := q.First(&tour).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	var panos []models.Panorama
	h.DB.Where("tour_id = ?", tour.ID).Order("sort_order ASC, id ASC").Find(&panos)
	for i := range panos {
		var hs []models.PanoramaHotspot
		h.DB.Where("panorama_id = ?", panos[i].ID).Find(&hs)
		panos[i].Hotspots = hs
	}
	tour.Panoramas = panos
	c.JSON(http.StatusOK, tour)
}

// POST /api/public/leads — body: {name, phone, message?, source?, house_id?,
// page?, utm_*}. The lead is saved immediately, then (best effort) AI
// scores it and the result is patched into the same row.
func (h *PublicHandler) CaptureLead(c *gin.Context) {
	var body struct {
		Name        string `json:"name"`
		Phone       string `json:"phone"`
		Message     string `json:"message"`
		Source      string `json:"source"`
		HouseID     *uint  `json:"house_id"`
		ObjectID    *uint  `json:"object_id"`
		Page        string `json:"page"`
		Referrer    string `json:"referrer"`
		UTMSource   string `json:"utm_source"`
		UTMMedium   string `json:"utm_medium"`
		UTMCampaign string `json:"utm_campaign"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	body.Phone = strings.TrimSpace(body.Phone)
	if len(body.Phone) < 6 {
		utils.ErrorResp(c, http.StatusBadRequest, "site.phone_required")
		return
	}
	row := models.WebsiteLead{
		Name:        strings.TrimSpace(body.Name),
		Phone:       body.Phone,
		Message:     body.Message,
		Source:      ifEmptyStr(body.Source, "popup"),
		HouseID:     body.HouseID,
		ObjectID:    body.ObjectID,
		Page:        body.Page,
		Referrer:    body.Referrer,
		UserAgent:   c.Request.UserAgent(),
		IP:          c.ClientIP(),
		UTMSource:   body.UTMSource,
		UTMMedium:   body.UTMMedium,
		UTMCampaign: body.UTMCampaign,
		Status:      "new",
	}
	if err := h.DB.Create(&row).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Notify all admins inside the CRM (in-app + Telegram + WS).
	go h.notifyAdminsAboutWebsiteLead(&row)

	// Fire-and-forget AI scoring; don't block the user-facing response.
	if h.AI != nil && h.AI.Enabled() {
		go h.scoreWebsiteLead(row.ID)
	}

	// Trigger the workflow engine.
	if h.Workflow != nil {
		h.Workflow.Trigger(c.Request.Context(), "website_lead.created", map[string]any{
			"id":     row.ID,
			"phone":  row.Phone,
			"name":   row.Name,
			"source": row.Source,
		})
	}

	utils.OK(c, gin.H{"id": row.ID})
}

func (h *PublicHandler) notifyAdminsAboutWebsiteLead(lead *models.WebsiteLead) {
	var admins []models.User
	h.DB.Where("role = ? AND is_active = ?", models.RoleAdmin, true).Find(&admins)
	title := "Лиди нави вебсайт"
	body := lead.Name + " — " + lead.Phone
	if lead.Message != "" {
		body += "\n" + lead.Message
	}
	for _, u := range admins {
		h.Notif.Push(u.ID, "website_lead", title, body, "/admin/website-leads", "")
	}
}

// scoreWebsiteLead runs in a goroutine and patches ai_score / ai_summary into
// the website_leads row once the AI provider answers.
func (h *PublicHandler) scoreWebsiteLead(id uint) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	var lead models.WebsiteLead
	if err := h.DB.First(&lead, id).Error; err != nil {
		return
	}
	score, err := h.AI.ScoreLead(ctx, map[string]any{
		"name":     lead.Name,
		"phone":    lead.Phone,
		"message":  lead.Message,
		"source":   lead.Source,
		"page":     lead.Page,
		"referrer": lead.Referrer,
		"utm":      []string{lead.UTMSource, lead.UTMMedium, lead.UTMCampaign},
		"house_id": lead.HouseID,
	}, nil)
	if err != nil || score == nil {
		return
	}
	h.DB.Model(&lead).Updates(map[string]any{
		"ai_score":   score.Score,
		"ai_summary": score.Verdict + " — " + score.Reason + " | " + score.NextStep,
	})
}

// GET /api/public/apk — small redirect that points at the configured APK URL.
// Lets the website link be /api/public/apk regardless of where the binary
// actually lives.
func (h *PublicHandler) APK(c *gin.Context) {
	url := h.Cfg.Site.APKDownloadURL
	if url == "" {
		utils.ErrorResp(c, http.StatusNotFound, "site.apk_unavailable")
		return
	}
	c.Redirect(http.StatusFound, url)
}

// ===== helpers =====

func toPublicRow(h models.House) publicHouseRow {
	return publicHouseRow{
		ID:              h.ID,
		Title:           h.Title,
		District:        h.District,
		Address:         h.Address,
		Area:            h.Area,
		Rooms:           h.Rooms,
		Floor:           h.Floor,
		TotalFloors:     h.TotalFloors,
		PublicPriceFrom: h.PublicPriceFrom,
		ShortDesc:       h.PublicShortDesc,
		FullDesc:        h.PublicFullDesc,
		Gallery:         []string(h.PublicGallery),
		Features:        []string(h.PublicFeatures),
		Lat:             h.Lat,
		Lng:             h.Lng,
	}
}

func ifEmptyStr(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
