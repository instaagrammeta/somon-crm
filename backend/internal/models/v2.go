package models

import "time"

// AuditLog records every mutating action by every authenticated user. The
// payload field stores a small redacted snapshot of the request body so that
// sensitive fields (passwords, tokens) never reach this table.
type AuditLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time `gorm:"index;default:CURRENT_TIMESTAMP" json:"created_at"`
	UserID    *uint     `gorm:"index" json:"user_id"`
	UserLogin string    `gorm:"size:64" json:"user_login"`
	Method    string    `gorm:"size:10" json:"method"`
	Path      string    `gorm:"size:255" json:"path"`
	Entity    string    `gorm:"size:64;index" json:"entity"`
	EntityID  string    `gorm:"size:64" json:"entity_id"`
	Action    string    `gorm:"size:32" json:"action"`
	IP        string    `gorm:"size:64" json:"ip"`
	UserAgent string    `gorm:"size:255" json:"user_agent"`
	Status    int       `json:"status"`
	Payload   JSONB     `gorm:"type:jsonb;default:'{}'" json:"payload"`
}

func (AuditLog) TableName() string { return "audit_logs" }

// UserTwoFactor stores the TOTP secret of a user. `Enabled` flips on only
// after the user has successfully verified the first 6-digit code; until then
// the row is "pending".
type UserTwoFactor struct {
	BaseModel
	UserID      uint        `gorm:"uniqueIndex;not null" json:"user_id"`
	Secret      string      `gorm:"size:64;not null" json:"-"` // never serialized
	Enabled     bool        `gorm:"not null;default:false" json:"enabled"`
	Recovery    StringSlice `gorm:"type:jsonb;default:'[]'" json:"-"` // one-time recovery codes
	ConfirmedAt *time.Time  `json:"confirmed_at"`
}

func (UserTwoFactor) TableName() string { return "user_two_factors" }

// WebsiteLead — lead captured by the public website auto-popup. Kept in a
// separate table from the CRM `lids` so the public form cannot pollute the
// pipeline. Operators promote a website_lead to a proper CRM lead manually.
type WebsiteLead struct {
	BaseModel
	Name           string `gorm:"size:128" json:"name"`
	Phone          string `gorm:"size:32;not null;index" json:"phone"`
	Message        string `gorm:"type:text" json:"message"`
	Source         string `gorm:"size:64" json:"source"` // popup|contact_form|house_card|tour
	HouseID        *uint  `gorm:"index" json:"house_id,omitempty"`
	ObjectID       *uint  `gorm:"index" json:"object_id,omitempty"`
	Page           string `gorm:"size:255" json:"page"`
	Referrer       string `gorm:"size:255" json:"referrer"`
	UserAgent      string `gorm:"size:255" json:"user_agent"`
	IP             string `gorm:"size:64" json:"ip"`
	UTMSource      string `gorm:"size:64" json:"utm_source,omitempty"`
	UTMMedium      string `gorm:"size:64" json:"utm_medium,omitempty"`
	UTMCampaign    string `gorm:"size:128" json:"utm_campaign,omitempty"`
	Status         string `gorm:"size:32;not null;default:'new';index" json:"status"`
	PromotedLeadID *uint  `json:"promoted_lead_id,omitempty"`
	Notes          string `gorm:"type:text" json:"notes"`
	AIScore        int    `gorm:"default:0" json:"ai_score"`
	AISummary      string `gorm:"type:text" json:"ai_summary"`
}

func (WebsiteLead) TableName() string { return "website_leads" }

// Tour — virtual 360° tour, attached to a house or stand-alone.
type Tour struct {
	BaseModel
	Title             string `gorm:"size:255;not null" json:"title"`
	Slug              string `gorm:"size:128;uniqueIndex" json:"slug"`
	Description       string `gorm:"type:text" json:"description"`
	HouseID           *uint  `gorm:"index" json:"house_id,omitempty"`
	ObjectID          *uint  `gorm:"index" json:"object_id,omitempty"`
	CoverImage        string `gorm:"size:512" json:"cover_image"`
	PublicVisible     bool   `gorm:"not null;default:true" json:"public_visible"`
	InitialPanoramaID *uint  `json:"initial_panorama_id,omitempty"`

	Panoramas []Panorama `gorm:"foreignKey:TourID" json:"panoramas,omitempty"`
}

// Panorama — one 360° image of a room (Kitchen, Living, Bedroom, …) inside a
// virtual tour. Marzipano on the frontend reads ImageURL and renders it.
type Panorama struct {
	BaseModel
	TourID       uint    `gorm:"index;not null" json:"tour_id"`
	Title        string  `gorm:"size:128;not null" json:"title"`
	ImageURL     string  `gorm:"size:512;not null" json:"image_url"`
	ImageType    string  `gorm:"size:32;not null;default:'equirect'" json:"image_type"`
	InitialYaw   float64 `gorm:"default:0" json:"initial_yaw"`
	InitialPitch float64 `gorm:"default:0" json:"initial_pitch"`
	InitialZoom  float64 `gorm:"default:90" json:"initial_zoom"`
	SortOrder    int     `gorm:"default:0" json:"sort_order"`

	Hotspots []PanoramaHotspot `gorm:"foreignKey:PanoramaID" json:"hotspots,omitempty"`
}

// PanoramaHotspot — clickable point on a panorama that links to another
// panorama (kind="link") or shows an info bubble (kind="info").
type PanoramaHotspot struct {
	BaseModel
	PanoramaID       uint    `gorm:"index;not null" json:"panorama_id"`
	TargetPanoramaID *uint   `json:"target_panorama_id,omitempty"`
	Kind             string  `gorm:"size:16;not null;default:'link'" json:"kind"`
	Label            string  `gorm:"size:128" json:"label"`
	Yaw              float64 `gorm:"not null" json:"yaw"`
	Pitch            float64 `gorm:"not null" json:"pitch"`
	InfoText         string  `gorm:"type:text" json:"info_text,omitempty"`
	InfoURL          string  `gorm:"size:512" json:"info_url,omitempty"`
}

func (PanoramaHotspot) TableName() string { return "panorama_hotspots" }

// VoiceNote — audio recording attached to a task / lead / request / chat.
// Once uploaded, the transcription worker calls Whisper-compatible API and
// fills `Transcript`.
type VoiceNote struct {
	BaseModel
	UserID         *uint  `gorm:"index" json:"user_id"`
	Entity         string `gorm:"size:32;not null" json:"entity"` // task|lead|request|chat
	EntityID       uint   `gorm:"not null;index" json:"entity_id"`
	AudioURL       string `gorm:"size:512;not null" json:"audio_url"`
	DurationSec    int    `gorm:"default:0" json:"duration_sec"`
	Transcript     string `gorm:"type:text" json:"transcript"`
	TranscriptLang string `gorm:"size:8" json:"transcript_lang"`
	Status         string `gorm:"size:16;not null;default:'pending'" json:"status"`
}

func (VoiceNote) TableName() string { return "voice_notes" }

// AILog — every call to the AI provider, used for accountability + UI history.
type AILog struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt        time.Time `gorm:"index;default:CURRENT_TIMESTAMP" json:"created_at"`
	UserID           *uint     `gorm:"index" json:"user_id"`
	Feature          string    `gorm:"size:32;not null" json:"feature"` // assistant|score|summarize|ocr|whisper|categorize
	Model            string    `gorm:"size:64" json:"model"`
	PromptTokens     int       `json:"prompt_tokens"`
	CompletionTokens int       `json:"completion_tokens"`
	TotalTokens      int       `json:"total_tokens"`
	DurationMS       int       `json:"duration_ms"`
	RequestBrief     string    `gorm:"type:text" json:"request_brief"`
	ResponseBrief    string    `gorm:"type:text" json:"response_brief"`
	Error            string    `gorm:"type:text" json:"error"`
}

func (AILog) TableName() string { return "ai_logs" }

// Workflow — automation rule. Trigger fires → steps run in order. Steps stored
// as JSON for flexibility; the engine validates them at run time.
type Workflow struct {
	BaseModel
	Name        string     `gorm:"size:128;not null" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	Enabled     bool       `gorm:"not null;default:true" json:"enabled"`
	Trigger     string     `gorm:"size:64;not null;index" json:"trigger"`
	Config      JSONB      `gorm:"type:jsonb;default:'{}'" json:"config"`
	Steps       JSONB      `gorm:"type:jsonb;default:'[]'" json:"steps"`
	LastRunAt   *time.Time `json:"last_run_at"`
	LastStatus  string     `gorm:"size:16" json:"last_status"`
	Runs        int        `gorm:"not null;default:0" json:"runs"`
}

// WorkflowRun — single execution of a workflow.
type WorkflowRun struct {
	ID         uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	WorkflowID uint       `gorm:"index;not null" json:"workflow_id"`
	StartedAt  time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"started_at"`
	FinishedAt *time.Time `json:"finished_at"`
	Status     string     `gorm:"size:16;not null;default:'running'" json:"status"`
	Payload    JSONB      `gorm:"type:jsonb;default:'{}'" json:"payload"`
	Log        string     `gorm:"type:text" json:"log"`
}

// IntegrationChannel — credentials + settings for an outbound/inbound channel
// (WhatsApp Business API, Email IMAP/SMTP, Telegram Bot).
type IntegrationChannel struct {
	BaseModel
	Type    string `gorm:"size:32;not null" json:"type"` // whatsapp|email|telegram_bot
	Name    string `gorm:"size:128" json:"name"`
	Enabled bool   `gorm:"not null;default:true" json:"enabled"`
	Config  JSONB  `gorm:"type:jsonb;default:'{}'" json:"config"`
}

func (IntegrationChannel) TableName() string { return "integration_channels" }

// UserPresence — last-seen tracker (online indicator).
type UserPresence struct {
	UserID   uint      `gorm:"primaryKey" json:"user_id"`
	LastSeen time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"last_seen"`
	Online   bool      `gorm:"not null;default:false" json:"online"`
}

func (UserPresence) TableName() string { return "user_presence" }

// House public-website fields appended in the v-2 migration.
type HousePublic struct {
	PublicVisible   bool        `gorm:"not null;default:false" json:"public_visible"`
	PublicShortDesc string      `gorm:"size:500" json:"public_short_desc"`
	PublicFullDesc  string      `gorm:"type:text" json:"public_full_desc"`
	PublicPriceFrom int64       `json:"public_price_from"`
	PublicGallery   StringSlice `gorm:"type:jsonb;default:'[]'" json:"public_gallery"`
	PublicFeatures  StringSlice `gorm:"type:jsonb;default:'[]'" json:"public_features"`
	Lat             float64     `json:"lat"`
	Lng             float64     `json:"lng"`
}
