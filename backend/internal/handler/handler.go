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
type App struct {
	DB       *gorm.DB
	Redis    *redis.Client
	Cfg      *config.Config
	Auth     *service.AuthService
	Telegram *service.TelegramService
}

func NewApp(db *gorm.DB, rdb *redis.Client, cfg *config.Config, auth *service.AuthService, tg *service.TelegramService) *App {
	return &App{DB: db, Redis: rdb, Cfg: cfg, Auth: auth, Telegram: tg}
}
