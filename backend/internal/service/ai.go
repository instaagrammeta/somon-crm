package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"

	"github.com/instaagrammeta/somon-crm/backend/internal/config"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"gorm.io/gorm"
)

// AIService is a thin client around any OpenAI-compatible chat-completions
// endpoint. It is configured once at startup with DeepSeek (default) but the
// same code works against OpenAI, Groq, Together, etc. Every call is also
// recorded to ai_logs for observability.
type AIService struct {
	cfg    config.AIConfig
	db     *gorm.DB
	client *http.Client
}

func NewAIService(cfg config.AIConfig, db *gorm.DB) *AIService {
	timeout := time.Duration(cfg.HTTPTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &AIService{cfg: cfg, db: db, client: &http.Client{Timeout: timeout}}
}

// Enabled reports whether an API key is configured. Handlers should refuse the
// request when this returns false instead of leaking 500s.
func (s *AIService) Enabled() bool { return s != nil && s.cfg.APIKey != "" }

// ===== OpenAI-compatible request/response shapes =====

type ChatMessage struct {
	Role    string `json:"role"`              // system|user|assistant
	Content any    `json:"content,omitempty"` // string OR []ContentPart for vision
	Name    string `json:"name,omitempty"`
}

// ContentPart is used by vision-capable models to mix text + image_url.
type ContentPart struct {
	Type     string           `json:"type"` // text | image_url
	Text     string           `json:"text,omitempty"`
	ImageURL *ContentImageURL `json:"image_url,omitempty"`
}

type ContentImageURL struct {
	URL    string `json:"url"`              // either http(s):// or data:image/...;base64,...
	Detail string `json:"detail,omitempty"` // low|high|auto
}

type chatRequest struct {
	Model          string        `json:"model"`
	Messages       []ChatMessage `json:"messages"`
	Temperature    float32       `json:"temperature,omitempty"`
	MaxTokens      int           `json:"max_tokens,omitempty"`
	Stream         bool          `json:"stream"`
	JSONMode       bool          `json:"-"`
	ResponseFormat *struct {
		Type string `json:"type"`
	} `json:"response_format,omitempty"`
}

type chatResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Index   int         `json:"index"`
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// ChatOptions tweak a single chat() call.
type ChatOptions struct {
	Model       string
	Temperature float32
	MaxTokens   int
	JSONMode    bool   // ask for response_format=json_object
	Feature     string // recorded in ai_logs
	UserID      *uint
}

// Chat is the low-level call. Most callers should use one of the higher-level
// helpers (Assistant, ScoreLead, …) instead.
func (s *AIService) Chat(ctx context.Context, msgs []ChatMessage, opt ChatOptions) (string, error) {
	if !s.Enabled() {
		return "", errors.New("ai_disabled")
	}
	model := opt.Model
	if model == "" {
		model = s.cfg.ChatModel
	}
	body := chatRequest{
		Model:       model,
		Messages:    msgs,
		Temperature: opt.Temperature,
		MaxTokens:   opt.MaxTokens,
	}
	if opt.JSONMode {
		body.ResponseFormat = &struct {
			Type string `json:"type"`
		}{Type: "json_object"}
	}
	raw, _ := json.Marshal(body)

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.BaseURL+"/chat/completions", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.cfg.APIKey)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logCall(opt, model, briefMsgs(msgs), "", 0, 0, 0, time.Since(start), err)
		return "", err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	var out chatResponse
	if jerr := json.Unmarshal(respBytes, &out); jerr != nil {
		s.logCall(opt, model, briefMsgs(msgs), string(respBytes), 0, 0, 0, time.Since(start), jerr)
		return "", fmt.Errorf("ai: invalid json (status=%d)", resp.StatusCode)
	}
	if resp.StatusCode >= 400 || out.Error != nil {
		msg := "ai: status " + resp.Status
		if out.Error != nil {
			msg = out.Error.Message
		}
		s.logCall(opt, model, briefMsgs(msgs), msg, 0, 0, 0, time.Since(start), errors.New(msg))
		return "", errors.New(msg)
	}
	if len(out.Choices) == 0 {
		return "", errors.New("ai: empty response")
	}
	text := stringContent(out.Choices[0].Message.Content)
	s.logCall(opt, model, briefMsgs(msgs), text,
		out.Usage.PromptTokens, out.Usage.CompletionTokens, out.Usage.TotalTokens,
		time.Since(start), nil)
	return text, nil
}

// ===== High-level features =====

// Assistant — free-form Q&A about CRM data. The caller passes already-formatted
// CRM context (e.g. "Top 5 leads from this week: …") so this layer is content-
// agnostic. Replies in the requested locale (tg|ru|en).
func (s *AIService) Assistant(ctx context.Context, locale, userQuestion, crmContext string, userID *uint) (string, error) {
	sysLang := map[string]string{
		"tg": "тоҷикӣ",
		"ru": "русском",
		"en": "English",
	}[locale]
	if sysLang == "" {
		sysLang = "тоҷикӣ"
	}
	system := "Ту ёрдамчии CRM-и Somon Real Estate ҳастӣ. " +
		"Ҷавоб бо забони " + sysLang + ". Кӯтоҳ, аниқ, бо рӯйхатҳо ва рақамҳо. " +
		"Танҳо бар асоси контексти додашуда ҷавоб деҳ. Чизе ки намедонӣ — рости гӯй."
	msgs := []ChatMessage{
		{Role: "system", Content: system},
		{Role: "system", Content: "CRM CONTEXT:\n" + crmContext},
		{Role: "user", Content: userQuestion},
	}
	return s.Chat(ctx, msgs, ChatOptions{Feature: "assistant", UserID: userID, Temperature: 0.3, MaxTokens: 800})
}

// LeadScore is the structured result returned by ScoreLead.
type LeadScore struct {
	Score    int    `json:"score"`     // 0..100
	Verdict  string `json:"verdict"`   // hot|warm|cold|spam
	Reason   string `json:"reason"`    // 1-2 sentences
	NextStep string `json:"next_step"` // recommended action
}

// ScoreLead asks the model to rate a website lead 0-100 + give a one-line
// verdict and recommendation. Returns a structured result + raw JSON.
func (s *AIService) ScoreLead(ctx context.Context, lead any, userID *uint) (*LeadScore, error) {
	leadJSON, _ := json.MarshalIndent(lead, "", "  ")
	prompt := "Lid-и зерин (CRM-и амволи ғайриманқул) бо тарзи зерин баҳо деҳ:\n" +
		string(leadJSON) +
		"\n\nҶавоб танҳо JSON бо майдонҳо: score (0..100), verdict (hot|warm|cold|spam), reason, next_step."
	out, err := s.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "You are a B2C real-estate sales analyst. Output ONLY valid JSON, no prose."},
		{Role: "user", Content: prompt},
	}, ChatOptions{Feature: "score", UserID: userID, Temperature: 0.1, JSONMode: true, MaxTokens: 300})
	if err != nil {
		return nil, err
	}
	var ls LeadScore
	if err := json.Unmarshal([]byte(extractJSON(out)), &ls); err != nil {
		return nil, fmt.Errorf("ai: lead-score parse: %w", err)
	}
	if ls.Score < 0 {
		ls.Score = 0
	}
	if ls.Score > 100 {
		ls.Score = 100
	}
	return &ls, nil
}

// Summarize collapses a long chat or note thread into ~3 sentences.
func (s *AIService) Summarize(ctx context.Context, locale, text string, userID *uint) (string, error) {
	system := "Дар 2-3 ҷумла хулосаи матн диҳ. Забон: " + locale + ". Танҳо хулоса, бе мукаддима."
	return s.Chat(ctx, []ChatMessage{
		{Role: "system", Content: system},
		{Role: "user", Content: text},
	}, ChatOptions{Feature: "summarize", UserID: userID, Temperature: 0.2, MaxTokens: 250})
}

// Categorize returns a short tag list (3-6 lowercase keywords) describing the
// given task / request body.
func (s *AIService) Categorize(ctx context.Context, text string, userID *uint) ([]string, error) {
	out, err := s.Chat(ctx, []ChatMessage{
		{Role: "system", Content: "Output ONLY a JSON array of 3-6 short, lowercase, snake_case category tags. No prose."},
		{Role: "user", Content: text},
	}, ChatOptions{Feature: "categorize", UserID: userID, Temperature: 0.1, JSONMode: true, MaxTokens: 120})
	if err != nil {
		return nil, err
	}
	var tags []string
	if err := json.Unmarshal([]byte(extractJSON(out)), &tags); err != nil {
		// Some models return {"tags":[…]} — try that too.
		var wrap struct {
			Tags []string `json:"tags"`
		}
		if err2 := json.Unmarshal([]byte(extractJSON(out)), &wrap); err2 == nil {
			tags = wrap.Tags
		} else {
			return nil, err
		}
	}
	return tags, nil
}

// OCRDocument runs an OCR-style request on an image (passport, contract, etc.)
// using a vision-capable chat model. Returns plain text + key/value extraction
// when the model can find it.
func (s *AIService) OCRDocument(ctx context.Context, imageBytes []byte, mime, hint string, userID *uint) (string, error) {
	if !s.Enabled() {
		return "", errors.New("ai_disabled")
	}
	if mime == "" {
		mime = "image/jpeg"
	}
	dataURL := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(imageBytes)
	prompt := "Extract every visible text from this document. " +
		"Then output a JSON object with a 'fields' map of detected key→value pairs " +
		"(e.g. full_name, passport_number, birth_date, address). Reply with JSON only."
	if hint != "" {
		prompt = hint + "\n\n" + prompt
	}
	parts := []ContentPart{
		{Type: "text", Text: prompt},
		{Type: "image_url", ImageURL: &ContentImageURL{URL: dataURL, Detail: "high"}},
	}
	return s.Chat(ctx, []ChatMessage{
		{Role: "user", Content: parts},
	}, ChatOptions{Model: s.cfg.VisionModel, Feature: "ocr", UserID: userID, Temperature: 0.0, MaxTokens: 1500, JSONMode: true})
}

// Transcribe sends an audio buffer to a Whisper-compatible /audio/transcriptions
// endpoint and returns plain text. Falls back to the chat BaseURL/APIKey if
// AI_WHISPER_BASE_URL / AI_WHISPER_API_KEY are not set.
func (s *AIService) Transcribe(ctx context.Context, audio []byte, filename, lang string, userID *uint) (string, error) {
	if !s.Enabled() && s.cfg.WhisperAPIKey == "" {
		return "", errors.New("ai_disabled")
	}
	baseURL := s.cfg.WhisperBaseURL
	if baseURL == "" {
		baseURL = s.cfg.BaseURL
	}
	apiKey := s.cfg.WhisperAPIKey
	if apiKey == "" {
		apiKey = s.cfg.APIKey
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	if filename == "" {
		filename = "voice.webm"
	}
	fileWriter, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := fileWriter.Write(audio); err != nil {
		return "", err
	}
	_ = w.WriteField("model", s.cfg.WhisperModel)
	if lang != "" {
		_ = w.WriteField("language", lang)
	}
	_ = w.WriteField("response_format", "json")
	w.Close()

	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/audio/transcriptions", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := s.client.Do(req)
	if err != nil {
		s.logCall(ChatOptions{Feature: "whisper", UserID: userID}, s.cfg.WhisperModel, "audio:"+filename, "", 0, 0, 0, time.Since(start), err)
		return "", err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		err := fmt.Errorf("whisper: %s: %s", resp.Status, string(body))
		s.logCall(ChatOptions{Feature: "whisper", UserID: userID}, s.cfg.WhisperModel, "audio:"+filename, string(body), 0, 0, 0, time.Since(start), err)
		return "", err
	}
	var out struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	s.logCall(ChatOptions{Feature: "whisper", UserID: userID}, s.cfg.WhisperModel, "audio:"+filename, out.Text, 0, 0, 0, time.Since(start), nil)
	return out.Text, nil
}

// ===== Internals =====

func (s *AIService) logCall(opt ChatOptions, model, reqBrief, respBrief string,
	pt, ct, tt int, dur time.Duration, err error) {
	if s.db == nil {
		return
	}
	row := models.AILog{
		UserID:           opt.UserID,
		Feature:          opt.Feature,
		Model:            model,
		PromptTokens:     pt,
		CompletionTokens: ct,
		TotalTokens:      tt,
		DurationMS:       int(dur / time.Millisecond),
		RequestBrief:     truncate(reqBrief, 800),
		ResponseBrief:    truncate(respBrief, 800),
	}
	if err != nil {
		row.Error = truncate(err.Error(), 800)
	}
	// Best-effort: never block on the audit log.
	go func() { _ = s.db.Create(&row).Error }()
}

// stringContent collapses ChatMessage.Content (which may be string or []ContentPart)
// into a plain string, used when reading model responses.
func stringContent(c any) string {
	switch v := c.(type) {
	case string:
		return v
	case []any:
		var b strings.Builder
		for _, p := range v {
			if m, ok := p.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					b.WriteString(t)
				}
			}
		}
		return b.String()
	}
	return ""
}

// briefMsgs flattens chat messages to a single short string for the audit log.
// Long contents (e.g. base64 images) are stripped.
func briefMsgs(msgs []ChatMessage) string {
	var b strings.Builder
	for _, m := range msgs {
		txt := stringContent(m.Content)
		if len(txt) == 0 {
			txt = "<binary or non-text>"
		}
		b.WriteString(m.Role)
		b.WriteString(": ")
		b.WriteString(truncate(txt, 200))
		b.WriteString("\n")
	}
	return b.String()
}

// extractJSON pulls the first {...} or [...] block out of an LLM reply,
// because some providers wrap JSON in ```json fences even with response_format=json.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "```"); i >= 0 {
		// strip a fenced code block
		s = s[i+3:]
		if nl := strings.IndexByte(s, '\n'); nl >= 0 {
			s = s[nl+1:]
		}
		if j := strings.LastIndex(s, "```"); j >= 0 {
			s = s[:j]
		}
		s = strings.TrimSpace(s)
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
