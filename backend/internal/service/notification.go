package service

import (
	"log"

	"github.com/instaagrammeta/somon-crm/backend/internal/models"
	"gorm.io/gorm"
)

// NotificationService creates in-app notifications AND mirrors them to Telegram,
// so every alert is visible both inside the CRM and via the bot.
type NotificationService struct {
	db *gorm.DB
	tg *TelegramService
}

func NewNotification(db *gorm.DB, tg *TelegramService) *NotificationService {
	return &NotificationService{db: db, tg: tg}
}

// Push stores an in-app notification for the recipient and sends a Telegram
// message (best effort, only if the user is linked and has notifications on).
//
//   - title/body are stored verbatim and rendered in the CRM notifications page.
//   - tgKey, when non-empty, is an i18n key used to format the Telegram message
//     (preserving the existing rich/markdown formatting). When empty, the plain
//     "title\nbody" text is sent instead.
func (s *NotificationService) Push(userID uint, ntype, title, body, link, tgKey string, tgArgs ...any) {
	if s == nil || userID == 0 {
		return
	}
	if s.db != nil {
		n := models.Notification{
			UserID: userID,
			Type:   ntype,
			Title:  title,
			Body:   body,
			Link:   link,
		}
		if err := s.db.Create(&n).Error; err != nil {
			log.Printf("[notify] db create failed: %v", err)
		}
	}
	if s.tg == nil {
		return
	}
	if tgKey != "" {
		s.tg.NotifyKey(userID, tgKey, tgArgs...)
		return
	}
	msg := title
	if body != "" {
		if msg != "" {
			msg += "\n"
		}
		msg += body
	}
	if msg != "" {
		s.tg.Notify(userID, msg)
	}
}
