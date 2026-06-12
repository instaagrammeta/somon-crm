package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Config holds all application configuration loaded from environment.
type Config struct {
	App      AppConfig
	DB       DBConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Admin    AdminSeedConfig
	Telegram TelegramConfig
	Upload   UploadConfig
	CORS     CORSConfig
	Locale   LocaleConfig

	// v-2 additions
	AI       AIConfig
	WhatsApp WhatsAppConfig
	Email    EmailConfig
	Site     SiteConfig
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
	Token                string
	BotUsername          string
	NotificationsEnabled bool
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

// AIConfig — DeepSeek (OpenAI-compatible). The same client is reused for chat
// completions, lead scoring, summarization, OCR (vision) and Whisper-style
// audio transcription. To use a different provider just point BaseURL at it
// — the wire protocol is OpenAI v1 in all known cases.
type AIConfig struct {
	Provider           string // "deepseek" (default) | "openai" | "custom"
	APIKey             string
	BaseURL            string // https://api.deepseek.com/v1
	ChatModel          string // deepseek-chat
	ReasoningModel     string // deepseek-reasoner
	VisionModel        string // model used for OCR (deepseek-vl or gpt-4o)
	WhisperModel       string // whisper-1 / whisper-large-v3
	WhisperBaseURL     string // optional: separate endpoint for Whisper (Groq, OpenAI). Falls back to BaseURL.
	WhisperAPIKey      string // optional: separate key for Whisper. Falls back to APIKey.
	HTTPTimeoutSeconds int
	MaxLeadsPerMinute  int // soft per-user rate limit on AI features
}

func (a AIConfig) Enabled() bool { return a.APIKey != "" }

type WhatsAppConfig struct {
	PhoneNumberID string
	AccessToken   string
	VerifyToken   string // for the webhook handshake
	BaseURL       string // https://graph.facebook.com/v20.0
	Enabled       bool
}

type EmailConfig struct {
	IMAPHost string
	IMAPPort int
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
	UseTLS   bool
	Enabled  bool
}

// SiteConfig — values exposed by the public website (about, contact, etc.).
type SiteConfig struct {
	BrandName      string
	Phone          string
	Email          string
	Address        string
	APKDownloadURL string // direct link served from /static/apk/<filename> or external URL
	APKVersion     string
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

	// AI (DeepSeek by default — fully OpenAI-compatible).
	v.SetDefault("AI_PROVIDER", "deepseek")
	v.SetDefault("AI_API_KEY", "")
	v.SetDefault("AI_BASE_URL", "https://api.deepseek.com/v1")
	v.SetDefault("AI_CHAT_MODEL", "deepseek-chat")
	v.SetDefault("AI_REASONING_MODEL", "deepseek-reasoner")
	v.SetDefault("AI_VISION_MODEL", "deepseek-vl2") // any OpenAI-compatible vision model
	v.SetDefault("AI_WHISPER_MODEL", "whisper-1")
	v.SetDefault("AI_WHISPER_BASE_URL", "")
	v.SetDefault("AI_WHISPER_API_KEY", "")
	v.SetDefault("AI_HTTP_TIMEOUT_SECONDS", 60)
	v.SetDefault("AI_MAX_PER_MINUTE", 30)

	// WhatsApp Business Cloud API (Meta).
	v.SetDefault("WHATSAPP_PHONE_NUMBER_ID", "")
	v.SetDefault("WHATSAPP_ACCESS_TOKEN", "")
	v.SetDefault("WHATSAPP_VERIFY_TOKEN", "somon-verify")
	v.SetDefault("WHATSAPP_BASE_URL", "https://graph.facebook.com/v20.0")
	v.SetDefault("WHATSAPP_ENABLED", false)

	// Email IMAP/SMTP.
	v.SetDefault("EMAIL_IMAP_HOST", "")
	v.SetDefault("EMAIL_IMAP_PORT", 993)
	v.SetDefault("EMAIL_SMTP_HOST", "")
	v.SetDefault("EMAIL_SMTP_PORT", 587)
	v.SetDefault("EMAIL_USERNAME", "")
	v.SetDefault("EMAIL_PASSWORD", "")
	v.SetDefault("EMAIL_FROM", "")
	v.SetDefault("EMAIL_TLS", true)
	v.SetDefault("EMAIL_ENABLED", false)

	// Public website branding.
	v.SetDefault("SITE_BRAND_NAME", "Somon Real Estate")
	v.SetDefault("SITE_PHONE", "")
	v.SetDefault("SITE_EMAIL", "")
	v.SetDefault("SITE_ADDRESS", "")
	v.SetDefault("SITE_APK_URL", "/static/apk/somon-crm.apk")
	v.SetDefault("SITE_APK_VERSION", "1.0.0")

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
		AI: AIConfig{
			Provider:           v.GetString("AI_PROVIDER"),
			APIKey:             v.GetString("AI_API_KEY"),
			BaseURL:            strings.TrimRight(v.GetString("AI_BASE_URL"), "/"),
			ChatModel:          v.GetString("AI_CHAT_MODEL"),
			ReasoningModel:     v.GetString("AI_REASONING_MODEL"),
			VisionModel:        v.GetString("AI_VISION_MODEL"),
			WhisperModel:       v.GetString("AI_WHISPER_MODEL"),
			WhisperBaseURL:     strings.TrimRight(v.GetString("AI_WHISPER_BASE_URL"), "/"),
			WhisperAPIKey:      v.GetString("AI_WHISPER_API_KEY"),
			HTTPTimeoutSeconds: v.GetInt("AI_HTTP_TIMEOUT_SECONDS"),
			MaxLeadsPerMinute:  v.GetInt("AI_MAX_PER_MINUTE"),
		},
		WhatsApp: WhatsAppConfig{
			PhoneNumberID: v.GetString("WHATSAPP_PHONE_NUMBER_ID"),
			AccessToken:   v.GetString("WHATSAPP_ACCESS_TOKEN"),
			VerifyToken:   v.GetString("WHATSAPP_VERIFY_TOKEN"),
			BaseURL:       strings.TrimRight(v.GetString("WHATSAPP_BASE_URL"), "/"),
			Enabled:       v.GetBool("WHATSAPP_ENABLED"),
		},
		Email: EmailConfig{
			IMAPHost: v.GetString("EMAIL_IMAP_HOST"),
			IMAPPort: v.GetInt("EMAIL_IMAP_PORT"),
			SMTPHost: v.GetString("EMAIL_SMTP_HOST"),
			SMTPPort: v.GetInt("EMAIL_SMTP_PORT"),
			Username: v.GetString("EMAIL_USERNAME"),
			Password: v.GetString("EMAIL_PASSWORD"),
			From:     v.GetString("EMAIL_FROM"),
			UseTLS:   v.GetBool("EMAIL_TLS"),
			Enabled:  v.GetBool("EMAIL_ENABLED"),
		},
		Site: SiteConfig{
			BrandName:      v.GetString("SITE_BRAND_NAME"),
			Phone:          v.GetString("SITE_PHONE"),
			Email:          v.GetString("SITE_EMAIL"),
			Address:        v.GetString("SITE_ADDRESS"),
			APKDownloadURL: v.GetString("SITE_APK_URL"),
			APKVersion:     v.GetString("SITE_APK_VERSION"),
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
