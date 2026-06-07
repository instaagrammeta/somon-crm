package models

// RequestsBoard — board for kanban-style request management.
type RequestsBoard struct {
	BaseModel
	Title    string `gorm:"size:255" json:"title"`
	Color    string `gorm:"size:32;default:'#0f172a'" json:"color"`
	IsPublic bool   `gorm:"default:true" json:"is_public"`
	AuthorID *uint  `gorm:"index" json:"author_id"`
}

type RequestsColumn struct {
	BaseModel
	BoardID    uint   `gorm:"index" json:"board_id"`
	Title      string `gorm:"size:255" json:"title"`
	Color      string `gorm:"size:32;default:'#3b82f6'" json:"color"`
	OrderIndex int    `gorm:"default:0;index" json:"order_index"`
}

// RequestItem — request inside a column.
type RequestItem struct {
	BaseModel
	ColumnID     uint    `gorm:"index" json:"column_id"`
	BoardID      uint    `gorm:"index" json:"board_id"`
	PropertyType string  `gorm:"size:64" json:"property_type"`
	Address      string  `gorm:"size:512" json:"address"`
	Area         float64 `json:"area"`
	Rooms        int     `json:"rooms"`
	Windows      int     `json:"windows"`
	Floor        int     `json:"floor"`
	TotalFloors  int     `json:"total_floors"`
	TotalPrice   float64 `json:"total_price"`
	PricePerM2   float64 `json:"price_per_m2"`
	Phone        string  `gorm:"size:64" json:"phone"`
	ClientName   string  `gorm:"size:255" json:"client_name"`
	Comment      string  `gorm:"type:text" json:"comment"`

	AuthorID     *uint  `gorm:"index" json:"author_id"`
	AuthorName   string `gorm:"size:255" json:"author_name"`
	ExecutorID   *uint  `gorm:"index" json:"executor_id"`
	ExecutorName string `gorm:"size:255" json:"executor_name"`

	Files      JSONB `gorm:"type:jsonb;default:'[]'" json:"files"`
	OrderIndex int   `gorm:"default:0;index" json:"order_index"`
}

// LegacyRequest — kept for compatibility with old /api/requests endpoints.
type LegacyRequest struct {
	BaseModel
	PropertyType string  `gorm:"size:64" json:"property_type"`
	Address      string  `gorm:"size:512" json:"address"`
	Area         float64 `json:"area"`
	Rooms        int     `json:"rooms"`
	Windows      int     `json:"windows"`
	Floor        int     `json:"floor"`
	TotalFloors  int     `json:"total_floors"`
	Documents    string  `gorm:"size:255" json:"documents"`
	TotalPrice   float64 `json:"total_price"`
	PricePerM2   float64 `json:"price_per_m2"`
	Phone        string  `gorm:"size:64" json:"phone"`
	ClientName   string  `gorm:"size:255" json:"client_name"`
	Manager      string  `gorm:"size:255" json:"manager"`
	SMM          string  `gorm:"size:255" json:"smm"`
	Comment      string  `gorm:"type:text" json:"comment"`
	AuthorID     *uint   `gorm:"index" json:"author_id"`
	ExecutorID   *uint   `gorm:"index" json:"executor_id"`
	Files        JSONB   `gorm:"type:jsonb;default:'[]'" json:"files"`
	Status       string  `gorm:"size:32;default:'new';index" json:"status"`
}

func (LegacyRequest) TableName() string { return "legacy_requests" }
func (RequestItem) TableName() string   { return "requests_items" }
func (RequestsBoard) TableName() string { return "requests_boards" }
func (RequestsColumn) TableName() string { return "requests_columns" }
