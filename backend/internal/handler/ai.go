package handler

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/instaagrammeta/somon-crm/backend/internal/middleware"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/utils"
)

// AIHandler exposes the AI-powered features of v-2. Every endpoint requires
// auth and degrades with 503 when AI_API_KEY is not configured. We keep the
// caller-facing surface tiny on purpose — most of the smarts live in the
// service layer (service/ai.go) so adding new providers later is just a
// config swap.
type AIHandler struct{ *App }

func NewAIHandler(a *App) *AIHandler { return &AIHandler{a} }

func (h *AIHandler) guard(c *gin.Context) bool {
	if h.AI == nil || !h.AI.Enabled() {
		utils.ErrorResp(c, http.StatusServiceUnavailable, "ai.disabled")
		return false
	}
	return true
}

// POST /api/ai/assistant
//
//	{ "question": "Top 5 leads of this week?", "context": "<optional>" }
//
// If `context` is empty the handler builds a small CRM context (counts of
// open leads/tasks per status, top sources) so the model has something
// concrete to ground its answer in.
func (h *AIHandler) Assistant(c *gin.Context) {
	if !h.guard(c) {
		return
	}
	var body struct {
		Question string `json:"question"`
		Context  string `json:"context"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Question) == "" {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	uidPtr := &uid

	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	crmCtx := body.Context
	if crmCtx == "" {
		crmCtx = h.buildAssistantContext()
	}
	locale := c.GetString("locale")
	if locale == "" {
		locale = h.Cfg.Locale.Default
	}

	out, err := h.AI.Assistant(ctx, locale, body.Question, crmCtx, uidPtr)
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadGateway, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"answer": out})
}

// buildAssistantContext gathers a concise CRM snapshot that the assistant can
// use to ground its answers (avoids hallucinations). We deliberately keep
// the size small — these queries should never become a bottleneck.
func (h *AIHandler) buildAssistantContext() string {
	var b strings.Builder

	type cnt struct {
		Status string
		N      int64
	}
	var leads []cnt
	h.DB.Raw(`SELECT COALESCE(status,'new') AS status, COUNT(*) AS n FROM lids GROUP BY 1`).Scan(&leads)
	b.WriteString("LEADS BY STATUS:\n")
	for _, r := range leads {
		b.WriteString("- " + r.Status + ": " + strconv.FormatInt(r.N, 10) + "\n")
	}

	var tasks []cnt
	h.DB.Raw(`SELECT status, COUNT(*) AS n FROM tasks GROUP BY 1`).Scan(&tasks)
	b.WriteString("\nTASKS BY STATUS:\n")
	for _, r := range tasks {
		b.WriteString("- " + r.Status + ": " + strconv.FormatInt(r.N, 10) + "\n")
	}

	var src []struct {
		Source string
		N      int64
	}
	h.DB.Raw(`SELECT source, COUNT(*) AS n FROM website_leads
		WHERE created_at >= NOW() - INTERVAL '30 days' GROUP BY 1 ORDER BY n DESC LIMIT 10`).Scan(&src)
	if len(src) > 0 {
		b.WriteString("\nWEBSITE LEAD SOURCES (30d):\n")
		for _, r := range src {
			b.WriteString("- " + r.Source + ": " + strconv.FormatInt(r.N, 10) + "\n")
		}
	}
	return b.String()
}

// POST /api/ai/score-lead/:id   — re-score a single website lead.
func (h *AIHandler) ScoreLead(c *gin.Context) {
	if !h.guard(c) {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var lead models.WebsiteLead
	if err := h.DB.First(&lead, id).Error; err != nil {
		utils.ErrorResp(c, http.StatusNotFound, "error.not_found")
		return
	}
	uid := middleware.CurrentUserID(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()

	score, err := h.AI.ScoreLead(ctx, lead, &uid)
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadGateway, err.Error())
		return
	}
	h.DB.Model(&lead).Updates(map[string]any{
		"ai_score":   score.Score,
		"ai_summary": score.Verdict + " — " + score.Reason + " | " + score.NextStep,
	})
	c.JSON(http.StatusOK, score)
}

// POST /api/ai/summarize  { "text": "..." }
func (h *AIHandler) Summarize(c *gin.Context) {
	if !h.guard(c) {
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Text == "" {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	locale := c.GetString("locale")
	if locale == "" {
		locale = h.Cfg.Locale.Default
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	out, err := h.AI.Summarize(ctx, locale, body.Text, &uid)
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadGateway, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"summary": out})
}

// POST /api/ai/categorize { "text": "..." }
func (h *AIHandler) Categorize(c *gin.Context) {
	if !h.guard(c) {
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Text == "" {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	uid := middleware.CurrentUserID(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
	defer cancel()
	tags, err := h.AI.Categorize(ctx, body.Text, &uid)
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadGateway, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// POST /api/ai/ocr — multipart/form-data with file=image/jpeg.
// Returns the JSON the model produced (typically {"fields": {...}}).
func (h *AIHandler) OCR(c *gin.Context) {
	if !h.guard(c) {
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		utils.ErrorResp(c, http.StatusBadRequest, "error.bad_request")
		return
	}
	f, err := fh.Open()
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadRequest, err.Error())
		return
	}
	defer f.Close()
	buf, err := io.ReadAll(f)
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadRequest, err.Error())
		return
	}
	mime := fh.Header.Get("Content-Type")
	hint := c.PostForm("hint") // e.g. "Tajik passport"
	uid := middleware.CurrentUserID(c)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 90*time.Second)
	defer cancel()

	out, err := h.AI.OCRDocument(ctx, buf, mime, hint, &uid)
	if err != nil {
		utils.ErrorRaw(c, http.StatusBadGateway, err.Error())
		return
	}
	// out is JSON-shaped; pass it through as a string so the UI can show the
	// raw provider response and parse it client-side as needed.
	c.JSON(http.StatusOK, gin.H{"raw": out})
}

// GET /api/ai/logs?feature=&limit= — recent AI calls (admin).
func (h *AIHandler) Logs(c *gin.Context) {
	if !middleware.IsAdmin(c) {
		utils.ErrorResp(c, http.StatusForbidden, "error.forbidden")
		return
	}
	q := h.DB.Model(&models.AILog{}).Order("created_at DESC")
	if v := c.Query("feature"); v != "" {
		q = q.Where("feature = ?", v)
	}
	limit := 100
	if v, _ := strconv.Atoi(c.Query("limit")); v > 0 && v <= 500 {
		limit = v
	}
	var rows []models.AILog
	if err := q.Limit(limit).Find(&rows).Error; err != nil {
		utils.ErrorRaw(c, http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, rows)
}
