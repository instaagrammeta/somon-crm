package models

import "time"

type Post struct {
	BaseModel
	UserID      *uint      `gorm:"index" json:"user_id"`
	UserName    string     `gorm:"size:255" json:"user_name"`
	Title       string     `gorm:"size:255" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	Category    string     `gorm:"size:128;index" json:"category"`
	ContentType string     `gorm:"size:64" json:"content_type"`
	Project     string     `gorm:"size:255" json:"project"`
	MediaPath   string     `gorm:"size:512" json:"media_path"`
	MediaType   string     `gorm:"size:32" json:"media_type"`
	Link        string     `gorm:"size:512" json:"link"`
	PostDate    *time.Time `json:"post_date"`
	Likes       int        `gorm:"default:0" json:"likes"`
	Comments    int        `gorm:"default:0" json:"comments"`
	Shares      int        `gorm:"default:0" json:"shares"`
	Views       int        `gorm:"default:0" json:"views"`
	Reach       int        `gorm:"default:0" json:"reach"`
	IsPublished bool       `gorm:"default:true" json:"is_published"`
	PublishedAt *time.Time `json:"published_at"`
}

type Message struct {
	BaseModel
	UserID   *uint  `gorm:"index" json:"user_id"`
	UserName string `gorm:"size:255" json:"user_name"`
	Message  string `gorm:"type:text" json:"message"`
	FilePath string `gorm:"size:512" json:"file_path"`
	FileName string `gorm:"size:255" json:"file_name"`
	FileType string `gorm:"size:64" json:"file_type"`
}
