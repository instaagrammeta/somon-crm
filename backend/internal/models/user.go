package models

import "time"

// User roles
const (
	RoleAdmin    = "admin"
	RoleEmployee = "employee"
	RoleManager  = "manager"
)

// User mirrors the legacy `users` table plus Telegram integration fields.
type User struct {
	BaseModel
	FullName       string      `gorm:"size:255" json:"full_name"`
	Age            int         `json:"age"`
	PersonalPhones StringSlice `gorm:"type:jsonb;default:'[]'" json:"personal_phones"`
	WorkPhones     StringSlice `gorm:"type:jsonb;default:'[]'" json:"work_phones"`
	Login          string      `gorm:"size:64;uniqueIndex" json:"login"`
	Password       string      `gorm:"size:255" json:"-"`
	Photo          string      `gorm:"size:512" json:"photo"`
	Category       string      `gorm:"size:128" json:"category"`
	Role           string      `gorm:"size:32;default:'employee';index" json:"role"`

	// Telegram integration
	TelegramUsername string     `gorm:"size:64" json:"telegram_username"`
	TelegramChatID   int64      `gorm:"index" json:"telegram_chat_id"`
	TelegramLinkedAt *time.Time `json:"telegram_linked_at"`
	NotifyTelegram   bool       `gorm:"default:true" json:"notify_telegram"`

	IsActive  bool       `gorm:"default:true;index" json:"is_active"`
	LastLogin *time.Time `json:"last_login"`
}

// PublicUser — DTO without sensitive fields.
type PublicUser struct {
	ID               uint        `json:"id"`
	FullName         string      `json:"full_name"`
	Age              int         `json:"age"`
	PersonalPhones   StringSlice `json:"personal_phones"`
	WorkPhones       StringSlice `json:"work_phones"`
	Login            string      `json:"login"`
	Photo            string      `json:"photo"`
	Category         string      `json:"category"`
	Role             string      `json:"role"`
	TelegramUsername string      `json:"telegram_username"`
	HasTelegram      bool        `json:"has_telegram"`
	IsActive         bool        `json:"is_active"`
	CreatedAt        time.Time   `json:"created_at"`
}

func (u *User) ToPublic() PublicUser {
	return PublicUser{
		ID:               u.ID,
		FullName:         u.FullName,
		Age:              u.Age,
		PersonalPhones:   u.PersonalPhones,
		WorkPhones:       u.WorkPhones,
		Login:            u.Login,
		Photo:            u.Photo,
		Category:         u.Category,
		Role:             u.Role,
		TelegramUsername: u.TelegramUsername,
		HasTelegram:      u.TelegramChatID != 0,
		IsActive:         u.IsActive,
		CreatedAt:        u.CreatedAt,
	}
}

// TelegramLinkCode — short-lived one-time code for binding chat to user.
// Stored only in Redis, but we expose its struct here for the service.
type TelegramLinkCode struct {
	UserID uint   `json:"user_id"`
	Code   string `json:"code"`
}
