package models

// Lid — simple legacy lead row.
type Lid struct {
	BaseModel
	ClientName string `gorm:"size:255" json:"client_name"`
	Phone      string `gorm:"size:64;index" json:"phone"`
	Topic      string `gorm:"size:255" json:"topic"`
	Comment    string `gorm:"type:text" json:"comment"`
	Source     string `gorm:"size:128" json:"source"`
	Mortgage   bool   `gorm:"default:false" json:"mortgage"`
	Box        bool   `gorm:"default:false" json:"box"`
	AuthorID   *uint  `gorm:"index" json:"author_id"`
}

// KanbanBoard / KanbanColumn / KanbanLead — modern kanban for leads.
type KanbanBoard struct {
	BaseModel
	Title      string `gorm:"size:255" json:"title"`
	Color      string `gorm:"size:32;default:'#0079bf'" json:"color"`
	IsPublic   bool   `gorm:"default:true" json:"is_public"`
	AuthorID   *uint  `gorm:"index" json:"author_id"`
	IsArchived bool   `gorm:"default:false;index" json:"is_archived"`
}

type KanbanBoardMember struct {
	BaseModel
	BoardID uint  `gorm:"index" json:"board_id"`
	UserID  uint  `gorm:"index" json:"user_id"`
}

type KanbanColumn struct {
	BaseModel
	BoardID    uint   `gorm:"index" json:"board_id"`
	Title      string `gorm:"size:255" json:"title"`
	Color      string `gorm:"size:32;default:'#0079bf'" json:"color"`
	OrderIndex int    `gorm:"default:0;index" json:"order_index"`
	IsArchived bool   `gorm:"default:false" json:"is_archived"`
}

type KanbanLead struct {
	BaseModel
	ColumnID   uint   `gorm:"index" json:"column_id"`
	BoardID    uint   `gorm:"index" json:"board_id"`
	ClientName string `gorm:"size:255" json:"client_name"`
	Phone      string `gorm:"size:64;index" json:"phone"`
	Topic      string `gorm:"size:255" json:"topic"`
	Comment    string `gorm:"type:text" json:"comment"`
	Source     string `gorm:"size:128" json:"source"`
	Mortgage   bool   `gorm:"default:false" json:"mortgage"`
	Box        bool   `gorm:"default:false" json:"box"`
	AuthorID   *uint  `gorm:"index" json:"author_id"`
	AuthorName string `gorm:"size:255" json:"author_name"`
	OrderIndex int    `gorm:"default:0;index" json:"order_index"`
}

type KanbanLeadInteraction struct {
	BaseModel
	LeadID      uint   `gorm:"index" json:"lead_id"`
	Phone       string `gorm:"size:64" json:"phone"`
	Topic       string `gorm:"size:255" json:"topic"`
	Source      string `gorm:"size:128" json:"source"`
	ContactType string `gorm:"size:64" json:"contact_type"`
	Comment     string `gorm:"type:text" json:"comment"`
	Mortgage    bool   `gorm:"default:false" json:"mortgage"`
	Box         bool   `gorm:"default:false" json:"box"`
}

func (KanbanBoard) TableName() string           { return "kanban_boards" }
func (KanbanBoardMember) TableName() string     { return "kanban_board_members" }
func (KanbanColumn) TableName() string          { return "kanban_columns" }
func (KanbanLead) TableName() string            { return "kanban_leads" }
func (KanbanLeadInteraction) TableName() string { return "kanban_lead_interactions" }
