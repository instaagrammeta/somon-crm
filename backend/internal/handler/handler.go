// Package handler contains HTTP handlers grouped by domain. Each handler accepts
// the shared App container so we don't have to wire every dependency one-by-one.
package handler

import (
	"github.com/instaagrammeta/somon-crm/backend/internal/config"
	"github.com/instaagrammeta/somon-crm/backend/internal/service"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// App bundles all infrastructure dependencies that handlers may need.
// New v-2 services (Hub for realtime, AI for DeepSeek, WhatsApp/Email/TwoFA/
// Workflow) are added here so existing handlers don't need to change.
type App struct {
	DB       *gorm.DB
	Redis    *redis.Client
	Cfg      *config.Config
	Auth     *service.AuthService
	Telegram *service.TelegramService
	Notif    *service.NotificationService

	// v-2
	Hub      *service.Hub
	AI       *service.AIService
	WhatsApp *service.WhatsAppService
	Email    *service.EmailService
	TwoFA    *service.TwoFactorService
	Workflow *service.WorkflowEngine
}

// NewApp constructs the App. Optional v-2 services may be nil — handlers that
// require them must check before use and return a 503 when absent.
func NewApp(
	db *gorm.DB,
	rdb *redis.Client,
	cfg *config.Config,
	auth *service.AuthService,
	tg *service.TelegramService,
) *App {
	hub := service.NewHub(db)
	notif := service.NewNotification(db, tg)
	notif.AttachHub(hub)
	app := &App{
		DB:       db,
		Redis:    rdb,
		Cfg:      cfg,
		Auth:     auth,
		Telegram: tg,
		Notif:    notif,
		Hub:      hub,
		AI:       service.NewAIService(cfg.AI, db),
		WhatsApp: service.NewWhatsAppService(cfg.WhatsApp),
		Email:    service.NewEmailService(cfg.Email),
		TwoFA:    service.NewTwoFactorService(db, cfg.Site.BrandName),
	}
	app.Workflow = service.NewWorkflowEngine(db, app.Notif, app.Telegram, app.WhatsApp, app.Email)
	return app
}
