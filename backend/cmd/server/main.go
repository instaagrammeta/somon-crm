// Package main starts the Somon CRM HTTP server.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/instaagrammeta/somon-crm/backend/internal/config"
	"github.com/instaagrammeta/somon-crm/backend/internal/database"
	"github.com/instaagrammeta/somon-crm/backend/internal/handler"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/instaagrammeta/somon-crm/backend/internal/router"
	"github.com/instaagrammeta/somon-crm/backend/internal/service"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := database.NewPostgres(cfg.DB, cfg.App.Env)
	if err != nil {
		log.Fatalf("postgres: %v", err)
	}

	// Auto-migrate as a safety net (idempotent). SQL migrations remain authoritative.
	if err := models.AutoMigrateAll(db); err != nil {
		log.Printf("automigrate warning: %v", err)
	}

	authSvc := service.NewAuth(cfg.JWT, db, nil)
	if err := seedAdmin(db, authSvc, cfg); err != nil {
		log.Printf("seed admin warning: %v", err)
	}

	rdb, err := database.NewRedis(cfg.Redis)
	if err != nil {
		log.Printf("redis warning: %v — continuing without redis", err)
		rdb = nil
	} else {
		authSvc = service.NewAuth(cfg.JWT, db, rdb)
	}

	tgSvc, err := service.NewTelegram(cfg.Telegram.Token, cfg.Telegram.BotUsername, cfg.Telegram.NotificationsEnabled, db, rdb)
	if err != nil {
		log.Printf("telegram warning: %v — continuing without bot", err)
		tgSvc, _ = service.NewTelegram("", "", false, db, rdb)
	}

	app := handler.NewApp(db, rdb, cfg, authSvc, tgSvc)

	r := router.Build(app, cfg)
	srv := &http.Server{
		Addr:              cfg.App.Host + ":" + cfg.App.Port,
		Handler:           r,
		ReadHeaderTimeout: 15 * time.Second,
	}

	tgCtx, tgCancel := context.WithCancel(context.Background())
	defer tgCancel()
	tgSvc.StartPolling(tgCtx)

	go func() {
		log.Printf("[server] listening on %s (env=%s)", srv.Addr, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Println("[server] shutting down…")

	tgSvc.Stop()
	tgCancel()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
	if rdb != nil {
		_ = rdb.Close()
	}
	log.Println("[server] goodbye")
}

// seedAdmin creates the default admin on first launch.
// If RESET_ADMIN_PASSWORD=true (env), it ALSO re-hashes the password from cfg
// on every startup — useful when ADMIN_PASSWORD in .env was changed after the
// initial container build.
//
// In addition: if the existing admin row is not active or not in the admin role,
// we force-fix that on every startup (cheap and safe).
func seedAdmin(db *gorm.DB, auth *service.AuthService, cfg *config.Config) error {
	resetPwd := strings.EqualFold(strings.TrimSpace(os.Getenv("RESET_ADMIN_PASSWORD")), "true")

	login := strings.TrimSpace(cfg.Admin.Login)
	password := cfg.Admin.Password

	if login == "" || password == "" {
		log.Printf("[seed] ADMIN_LOGIN or ADMIN_PASSWORD is empty — skipping seed")
		return nil
	}

	var existing models.User
	err := db.Where("login = ?", login).First(&existing).Error
	if err == nil {
		// User exists. Always make sure it is admin + active.
		needsFix := !existing.IsActive || existing.Role != models.RoleAdmin
		updates := map[string]any{}
		if needsFix {
			updates["role"] = models.RoleAdmin
			updates["is_active"] = true
		}
		if resetPwd {
			hashed, herr := auth.HashPassword(password)
			if herr != nil {
				return herr
			}
			updates["password"] = hashed
		}
		if len(updates) > 0 {
			if uerr := db.Model(&existing).Updates(updates).Error; uerr != nil {
				return uerr
			}
			if resetPwd {
				log.Printf("[seed] admin password reset + role/active fixed (login=%s)", login)
			} else {
				log.Printf("[seed] admin role/active fixed (login=%s)", login)
			}
		} else {
			log.Printf("[seed] admin already exists (login=%s) — set RESET_ADMIN_PASSWORD=true to rotate password", login)
		}
		return nil
	}

	// Create fresh admin.
	hashed, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	admin := models.User{
		FullName:       cfg.Admin.FullName,
		Login:          login,
		Password:       hashed,
		Role:           models.RoleAdmin,
		Category:       "Руководство",
		IsActive:       true,
		NotifyTelegram: true,
		PersonalPhones: models.StringSlice{},
		WorkPhones:     models.StringSlice{},
	}
	if err := db.Create(&admin).Error; err != nil {
		return err
	}
	log.Printf("[seed] admin user created (login=%s)", login)
	return nil
}
