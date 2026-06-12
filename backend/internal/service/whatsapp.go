package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/instaagrammeta/somon-crm/backend/internal/config"
)

// WhatsAppService talks to the WhatsApp Business Cloud API (Meta Graph API).
//
// Outbound:  SendText / SendTemplate (already approved by Meta)
// Inbound:   ParseWebhook (called by /api/webhooks/whatsapp)
// Verify:    VerifyWebhook for the GET handshake on first connect.
//
// We deliberately do NOT pull a 3rd-party Meta SDK — the surface we need is
// only ~120 lines of HTTP.
type WhatsAppService struct {
	cfg    config.WhatsAppConfig
	client *http.Client
}

func NewWhatsAppService(cfg config.WhatsAppConfig) *WhatsAppService {
	return &WhatsAppService{cfg: cfg, client: &http.Client{Timeout: 20 * time.Second}}
}

// Enabled reports whether the integration has all required credentials. Always
// guard outbound calls with this check so a missing token only logs/skip s.
func (s *WhatsAppService) Enabled() bool {
	return s != nil && s.cfg.Enabled && s.cfg.AccessToken != "" && s.cfg.PhoneNumberID != ""
}

// SendText sends a free-form text message. Note: outside the 24-hour
// "customer service window" Meta only allows pre-approved templates — see
// SendTemplate below for that flow.
func (s *WhatsAppService) SendText(to, body string) error {
	if !s.Enabled() {
		return errors.New("whatsapp.disabled")
	}
	to = normalizePhone(to)
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text":              map[string]any{"body": body, "preview_url": false},
	}
	return s.post("/messages", payload)
}

// SendTemplate sends a pre-approved template message (e.g. "lead_followup_v1")
// with positional body parameters.
func (s *WhatsAppService) SendTemplate(to, templateName, langCode string, bodyParams ...string) error {
	if !s.Enabled() {
		return errors.New("whatsapp.disabled")
	}
	if langCode == "" {
		langCode = "ru"
	}
	params := make([]map[string]any, 0, len(bodyParams))
	for _, p := range bodyParams {
		params = append(params, map[string]any{"type": "text", "text": p})
	}
	payload := map[string]any{
		"messaging_product": "whatsapp",
		"to":                normalizePhone(to),
		"type":              "template",
		"template": map[string]any{
			"name":     templateName,
			"language": map[string]string{"code": langCode},
			"components": []map[string]any{
				{"type": "body", "parameters": params},
			},
		},
	}
	return s.post("/messages", payload)
}

// VerifyWebhook implements Meta's GET-handshake on the webhook URL. Return the
// challenge unchanged when verify_token matches.
func (s *WhatsAppService) VerifyWebhook(mode, token, challenge string) (string, bool) {
	if mode == "subscribe" && token == s.cfg.VerifyToken {
		return challenge, true
	}
	return "", false
}

// IncomingMessage is the trimmed-down shape we hand back to the rest of the
// app. The real Meta payload has dozens of fields; we only surface what the
// CRM cares about (sender, text, optional media, etc.).
type IncomingMessage struct {
	From      string    `json:"from"`       // E.164
	Name      string    `json:"name"`       // sender's WhatsApp profile name
	Text      string    `json:"text"`       // plain message body
	MediaID   string    `json:"media_id"`   // for image/audio/document
	MediaType string    `json:"media_type"` // image|audio|document|video
	Caption   string    `json:"caption"`    // optional
	Timestamp time.Time `json:"timestamp"`
	WAID      string    `json:"wa_id"` // message id
}

// ParseWebhook decodes Meta's POST payload into a flat slice of incoming
// messages. Status callbacks (sent/read/delivered) are ignored at this layer;
// add a small extension later if read receipts are needed.
func (s *WhatsAppService) ParseWebhook(body []byte) ([]IncomingMessage, error) {
	var p struct {
		Entry []struct {
			Changes []struct {
				Value struct {
					Contacts []struct {
						WAID    string `json:"wa_id"`
						Profile struct {
							Name string `json:"name"`
						} `json:"profile"`
					} `json:"contacts"`
					Messages []struct {
						From      string `json:"from"`
						ID        string `json:"id"`
						Timestamp string `json:"timestamp"`
						Type      string `json:"type"`
						Text      struct {
							Body string `json:"body"`
						} `json:"text"`
						Image *struct {
							ID      string `json:"id"`
							Caption string `json:"caption"`
						} `json:"image,omitempty"`
						Audio *struct {
							ID string `json:"id"`
						} `json:"audio,omitempty"`
						Document *struct {
							ID       string `json:"id"`
							Filename string `json:"filename"`
						} `json:"document,omitempty"`
					} `json:"messages"`
				} `json:"value"`
			} `json:"changes"`
		} `json:"entry"`
	}
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	out := []IncomingMessage{}
	for _, e := range p.Entry {
		for _, ch := range e.Changes {
			names := map[string]string{}
			for _, ct := range ch.Value.Contacts {
				names[ct.WAID] = ct.Profile.Name
			}
			for _, m := range ch.Value.Messages {
				ts, _ := time.Parse(time.RFC3339, m.Timestamp)
				if ts.IsZero() {
					ts = time.Now()
				}
				msg := IncomingMessage{
					From:      m.From,
					Name:      names[m.From],
					Text:      m.Text.Body,
					Timestamp: ts,
					WAID:      m.ID,
				}
				switch m.Type {
				case "image":
					if m.Image != nil {
						msg.MediaID = m.Image.ID
						msg.MediaType = "image"
						msg.Caption = m.Image.Caption
					}
				case "audio":
					if m.Audio != nil {
						msg.MediaID = m.Audio.ID
						msg.MediaType = "audio"
					}
				case "document":
					if m.Document != nil {
						msg.MediaID = m.Document.ID
						msg.MediaType = "document"
						msg.Caption = m.Document.Filename
					}
				}
				out = append(out, msg)
			}
		}
	}
	return out, nil
}

// post is the generic Graph API helper.
func (s *WhatsAppService) post(path string, payload any) error {
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/%s%s", strings.TrimRight(s.cfg.BaseURL, "/"), s.cfg.PhoneNumberID, path)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		buf, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp: %s: %s", resp.Status, string(buf))
	}
	return nil
}

// normalizePhone strips spaces, parens, dashes and a leading '+' so the result
// is the digits-only E.164 form Meta expects.
func normalizePhone(p string) string {
	rep := strings.NewReplacer(" ", "", "(", "", ")", "", "-", "", "+", "")
	return rep.Replace(strings.TrimSpace(p))
}
