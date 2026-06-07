package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/instaagrammeta/somon-crm/backend/internal/i18n"
	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// TelegramService wraps the bot API + linking flow + outbound notifications.
type TelegramService struct {
	bot         *tgbotapi.BotAPI
	db          *gorm.DB
	rdb         *redis.Client
	enabled     bool
	botUsername string

	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewTelegram(token, username string, enabled bool, db *gorm.DB, rdb *redis.Client) (*TelegramService, error) {
	t := &TelegramService{db: db, rdb: rdb, enabled: enabled, botUsername: username, stopCh: make(chan struct{})}
	if !enabled || token == "" {
		log.Println("[telegram] disabled (token empty or feature-flag off)")
		return t, nil
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	t.bot = bot
	if username == "" {
		t.botUsername = bot.Self.UserName
	}
	log.Printf("[telegram] bot connected as @%s", t.botUsername)
	return t, nil
}

// StartPolling launches a background goroutine that handles updates.
func (s *TelegramService) StartPolling(ctx context.Context) {
	if s.bot == nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 30
		updates := s.bot.GetUpdatesChan(u)
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopCh:
				return
			case upd, ok := <-updates:
				if !ok {
					return
				}
				if upd.Message == nil {
					continue
				}
				s.handleUpdate(ctx, upd.Message)
			}
		}
	}()
}

func (s *TelegramService) Stop() {
	if s.bot == nil {
		return
	}
	close(s.stopCh)
	s.bot.StopReceivingUpdates()
	s.wg.Wait()
}

func (s *TelegramService) BotUsername() string { return s.botUsername }
func (s *TelegramService) Enabled() bool       { return s.bot != nil }

// ===== Linking flow =====

// GenerateLinkCode produces a 6-byte hex code valid for 15 minutes per user.
func (s *TelegramService) GenerateLinkCode(ctx context.Context, userID uint) (string, error) {
	if s.rdb == nil {
		return "", errors.New("redis not configured")
	}
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := strings.ToUpper(hex.EncodeToString(buf))
	key := "tg:link:" + code
	if err := s.rdb.Set(ctx, key, userID, 15*time.Minute).Err(); err != nil {
		return "", err
	}
	return code, nil
}

func (s *TelegramService) consumeLinkCode(ctx context.Context, code string) (uint, error) {
	if s.rdb == nil {
		return 0, errors.New("redis not configured")
	}
	key := "tg:link:" + strings.ToUpper(strings.TrimSpace(code))
	v, err := s.rdb.GetDel(ctx, key).Uint64()
	if err == redis.Nil {
		return 0, errors.New("invalid or expired code")
	}
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func (s *TelegramService) handleUpdate(ctx context.Context, m *tgbotapi.Message) {
	text := strings.TrimSpace(m.Text)
	chatID := m.Chat.ID
	username := m.From.UserName

	switch {
	case strings.HasPrefix(text, "/start"):
		// /start CODE — auto-link if a code is provided as deep-link argument
		args := strings.Fields(text)
		if len(args) == 2 {
			s.tryLink(ctx, chatID, username, args[1])
			return
		}
		s.send(chatID, i18n.Translate(i18n.LocaleTG, "tg.start_prompt"))
	case strings.HasPrefix(text, "/link"):
		args := strings.Fields(text)
		if len(args) < 2 {
			s.send(chatID, i18n.Translate(i18n.LocaleTG, "tg.start_prompt"))
			return
		}
		s.tryLink(ctx, chatID, username, args[1])
	case strings.HasPrefix(text, "/help"):
		s.send(chatID, i18n.Translate(i18n.LocaleTG, "tg.start_prompt"))
	}
}

func (s *TelegramService) tryLink(ctx context.Context, chatID int64, username, code string) {
	uid, err := s.consumeLinkCode(ctx, code)
	if err != nil || uid == 0 {
		s.send(chatID, i18n.Translate(i18n.LocaleTG, "tg.link_invalid"))
		return
	}
	now := time.Now().UTC()
	if err := s.db.Model(&models.User{}).Where("id = ?", uid).Updates(map[string]any{
		"telegram_chat_id":   chatID,
		"telegram_username":  username,
		"telegram_linked_at": now,
	}).Error; err != nil {
		log.Printf("[telegram] link db error: %v", err)
		s.send(chatID, i18n.Translate(i18n.LocaleTG, "error.internal"))
		return
	}
	var u models.User
	_ = s.db.First(&u, uid).Error
	s.send(chatID, i18n.Translate(i18n.LocaleTG, "tg.linked", u.FullName))
}

// ===== Outbound notifications =====

// Notify sends a markdown message to a user (no-op if user not linked).
func (s *TelegramService) Notify(userID uint, text string) {
	if !s.Enabled() {
		return
	}
	var u models.User
	if err := s.db.Select("telegram_chat_id, notify_telegram").
		First(&u, userID).Error; err != nil {
		return
	}
	if u.TelegramChatID == 0 || !u.NotifyTelegram {
		return
	}
	s.sendMarkdown(u.TelegramChatID, text)
}

// NotifyKey is shorthand for sending a translated message by i18n key.
func (s *TelegramService) NotifyKey(userID uint, key string, args ...any) {
	if !s.Enabled() {
		return
	}
	// Each user has same default-locale (Tajik); future: use user.preferred_locale.
	s.Notify(userID, i18n.Translate(i18n.LocaleTG, key, args...))
}

func (s *TelegramService) send(chatID int64, text string) {
	if s.bot == nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := s.bot.Send(msg); err != nil {
		log.Printf("[telegram] send: %v", err)
	}
}

func (s *TelegramService) sendMarkdown(chatID int64, text string) {
	if s.bot == nil {
		return
	}
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := s.bot.Send(msg); err != nil {
		// Retry without markdown if formatting was the cause
		log.Printf("[telegram] markdown failed: %v — retrying as plain", err)
		plain := tgbotapi.NewMessage(chatID, text)
		_, _ = s.bot.Send(plain)
	}
}
