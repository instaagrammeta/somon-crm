package models

// Notification types (stable identifiers used by the frontend for icons).
const (
	NotifyTaskAssigned   = "task_assigned"
	NotifyTaskStatus     = "task_status"
	NotifyRequestAssign  = "request_assigned"
	NotifyLeadMoved      = "lead_moved"
	NotifyTariffExpiring = "tariff_expiring"
)

// Notification is an in-app notification shown in the CRM "Огоҳиҳо" page and the
// topbar bell. The same events are also delivered via Telegram (best effort).
type Notification struct {
	BaseModel
	UserID uint   `gorm:"index" json:"user_id"` // recipient
	Type   string `gorm:"size:48;index" json:"type"`
	Title  string `gorm:"size:255" json:"title"`
	Body   string `gorm:"type:text" json:"body"`
	Link   string `gorm:"size:255" json:"link"`
	IsRead bool   `gorm:"default:false;index" json:"is_read"`
}
