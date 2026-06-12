package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/instaagrammeta/somon-crm/backend/internal/models"
)

// WhatsAppWebhookHandler accepts WhatsApp Cloud API webhooks. The GET request
// is Meta's verification handshake; the POST request carries the actual
// message events.
type WhatsAppWebhookHandler struct{ *App }

func NewWhatsAppWebhookHandler(a *App) *WhatsAppWebhookHandler { return &WhatsAppWebhookHandler{a} }

// GET /api/webhooks/whatsapp  — Meta hits this once when you save the URL.
//
//	?hub.mode=subscribe&hub.verify_token=...&hub.challenge=...
func (h *WhatsAppWebhookHandler) Verify(c *gin.Context) {
	if h.WhatsApp == nil {
		c.Status(http.StatusServiceUnavailable)
		return
	}
	mode := c.Query("hub.mode")
	token := c.Query("hub.verify_token")
	challenge := c.Query("hub.challenge")
	if reply, ok := h.WhatsApp.VerifyWebhook(mode, token, challenge); ok {
		c.String(http.StatusOK, reply)
		return
	}
	c.Status(http.StatusForbidden)
}

// POST /api/webhooks/whatsapp — every inbound text/image/audio message arrives
// here. We turn them into a generic "website_lead" so the operator gets
// notified inside the CRM and via Telegram, and store the original WhatsApp
// number for later reply.
func (h *WhatsAppWebhookHandler) Receive(c *gin.Context) {
	if h.WhatsApp == nil {
		c.Status(http.StatusOK) // ack so Meta does not retry
		return
	}
	body, _ := io.ReadAll(c.Request.Body)
	msgs, err := h.WhatsApp.ParseWebhook(body)
	if err != nil {
		c.Status(http.StatusOK)
		return
	}
	for _, m := range msgs {
		// One row per inbound user message → mirrors the website_lead flow so
		// admins see WhatsApp leads next to popup leads in the same dashboard.
		row := models.WebsiteLead{
			Name:    m.Name,
			Phone:   m.From,
			Message: m.Text,
			Source:  "whatsapp",
			Status:  "new",
		}
		_ = h.DB.Create(&row).Error

		// Notify all admins.
		var admins []models.User
		h.DB.Where("role = ? AND is_active = ?", models.RoleAdmin, true).Find(&admins)
		body := m.Name + " (" + m.From + ")"
		if m.Text != "" {
			body += "\n" + m.Text
		}
		for _, u := range admins {
			h.Notif.Push(u.ID, "website_lead", "WhatsApp: паёми нав", body, "/admin/website-leads", "")
		}
	}
	c.Status(http.StatusOK)
}
