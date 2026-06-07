package config

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all application configuration loaded from environment.
type Config struct {
	App       AppConfig
	DB        DBConfig
	Redis     RedisConfig
	JWT       JWTConfig
	Admin     AdminSeedConfig
	Telegram  TelegramConfig
	Upload    UploadConfig
	CORS      CORSConfig
	Locale    LocaleConfig
}

type AppConfig struct {
	Env     string
	Host    string
	Port    string
	BaseURL string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode)
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

type JWTConfig struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type AdminSeedConfig struct {
	Login    string
	Password string
	FullName string
}

type TelegramConfig struct {
	Token                  string
	BotUsername            string
	NotificationsEnabled   bool
}

type UploadConfig struct {
	Dir   string
	MaxMB int64
}

type CORSConfig struct {
	AllowedOrigins []string
}

type LocaleConfig struct {
	Default string
}

// Load reads .env (if present) and environment variables.
func Load() (*Config, error) {
	_ = godotenv.Load()

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_HOST", "0.0.0.0")
	v.SetDefault("APP_PORT", "8080")
	v.SetDefault("APP_BASE_URL", "http://localhost:8080")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "5432")
	v.SetDefault("DB_USER", "somon")
	v.SetDefault("DB_PASSWORD", "somon")
	v.SetDefault("DB_NAME", "somon_crm")
	v.SetDefault("DB_SSLMODE", "disable")
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", "6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("JWT_SECRET", "dev-only-change-me")
	v.SetDefault("JWT_ACCESS_TTL_MINUTES", 60)
	v.SetDefault("JWT_REFRESH_TTL_HOURS", 168)
	v.SetDefault("ADMIN_LOGIN", "admin")
	v.SetDefault("ADMIN_PASSWORD", "nav-xona@2026")
	v.SetDefault("ADMIN_FULL_NAME", "Главный Администратор")
	v.SetDefault("TELEGRAM_BOT_TOKEN", "")
	v.SetDefault("TELEGRAM_BOT_USERNAME", "somon_crm_bot")
	v.SetDefault("TELEGRAM_NOTIFICATIONS_ENABLED", true)
	v.SetDefault("UPLOAD_DIR", "./uploads")
	v.SetDefault("UPLOAD_MAX_MB", 50)
	v.SetDefault("CORS_ALLOWED_ORIGINS", "*")
	v.SetDefault("DEFAULT_LOCALE", "tg")

	cfg := &Config{
		App: AppConfig{
			Env:     v.GetString("APP_ENV"),
			Host:    v.GetString("APP_HOST"),
			Port:    v.GetString("APP_PORT"),
			BaseURL: v.GetString("APP_BASE_URL"),
		},
		DB: DBConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetString("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			Name:     v.GetString("DB_NAME"),
			SSLMode:  v.GetString("DB_SSLMODE"),
		},
		Redis: RedisConfig{
			Host:     v.GetString("REDIS_HOST"),
			Port:     v.GetString("REDIS_PORT"),
			Password: v.GetString("REDIS_PASSWORD"),
			DB:       v.GetInt("REDIS_DB"),
		},
		JWT: JWTConfig{
			Secret:     v.GetString("JWT_SECRET"),
			AccessTTL:  time.Duration(v.GetInt("JWT_ACCESS_TTL_MINUTES")) * time.Minute,
			RefreshTTL: time.Duration(v.GetInt("JWT_REFRESH_TTL_HOURS")) * time.Hour,
		},
		Admin: AdminSeedConfig{
			Login:    v.GetString("ADMIN_LOGIN"),
			Password: v.GetString("ADMIN_PASSWORD"),
			FullName: v.GetString("ADMIN_FULL_NAME"),
		},
		Telegram: TelegramConfig{
			Token:                v.GetString("TELEGRAM_BOT_TOKEN"),
			BotUsername:          v.GetString("TELEGRAM_BOT_USERNAME"),
			NotificationsEnabled: v.GetBool("TELEGRAM_NOTIFICATIONS_ENABLED"),
		},
		Upload: UploadConfig{
			Dir:   v.GetString("UPLOAD_DIR"),
			MaxMB: int64(v.GetInt("UPLOAD_MAX_MB")),
		},
		CORS: CORSConfig{
			AllowedOrigins: parseCSV(v.GetString("CORS_ALLOWED_ORIGINS")),
		},
		Locale: LocaleConfig{
			Default: v.GetString("DEFAULT_LOCALE"),
		},
	}

	return cfg, nil
}

func parseCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Helper to safely parse int from env string.
func atoi(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
