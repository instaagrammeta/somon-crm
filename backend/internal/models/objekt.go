package models

// Objekt = "Шахматка" (apartment matrix per floor).

type ObjektProject struct {
	BaseModel
	Name        string `gorm:"size:255" json:"name"`
	Address     string `gorm:"size:512" json:"address"`
	Developer   string `gorm:"size:255" json:"developer"`
	Description string `gorm:"type:text" json:"description"`
}

type ObjektBlock struct {
	BaseModel
	ProjectID            uint    `gorm:"index" json:"project_id"`
	Name                 string  `gorm:"size:255" json:"name"`
	FloorFrom            int     `json:"floor_from"`
	FloorTo              int     `json:"floor_to"`
	DefaultArea          float64 `json:"default_area"`
	DefaultRooms         int     `json:"default_rooms"`
	DefaultWindows       int     `json:"default_windows"`
	DefaultPricePerM2    float64 `json:"default_price_per_m2"`
	DefaultBalcony       bool    `gorm:"default:false" json:"default_balcony"`
	DefaultBathroomType  string  `gorm:"size:32;default:'combined'" json:"default_bathroom_type"`
	DefaultBathroomCount int     `gorm:"default:1" json:"default_bathroom_count"`
	DefaultPlanImage     string  `gorm:"size:512" json:"default_plan_image"`
	Description          string  `gorm:"type:text" json:"description"`
	OrderIndex           int     `gorm:"default:0;index" json:"order_index"`
}

const (
	ObjektStatusFree     = "free"
	ObjektStatusReserved = "reserved"
	ObjektStatusSold     = "sold"
)

type ObjektApartment struct {
	BaseModel
	BlockID        uint    `gorm:"index;uniqueIndex:ux_block_floor" json:"block_id"`
	Floor          int     `gorm:"uniqueIndex:ux_block_floor" json:"floor"`
	Area           float64 `json:"area"`
	Rooms          int     `json:"rooms"`
	Windows        int     `json:"windows"`
	PricePerM2     float64 `json:"price_per_m2"`
	TotalPrice     float64 `json:"total_price"`
	Status         string  `gorm:"size:16;default:'free';index" json:"status"`
	Balcony        bool    `gorm:"default:false" json:"balcony"`
	BathroomType   string  `gorm:"size:32;default:'combined'" json:"bathroom_type"`
	BathroomCount  int     `gorm:"default:1" json:"bathroom_count"`
	PlanImage      string  `gorm:"size:512" json:"plan_image"`
	Description    string  `gorm:"type:text" json:"description"`
	ClientName     string  `gorm:"size:255" json:"client_name"`
	ClientPhone    string  `gorm:"size:64" json:"client_phone"`
}

func (ObjektProject) TableName() string   { return "objekt_projects" }
func (ObjektBlock) TableName() string     { return "objekt_blocks" }
func (ObjektApartment) TableName() string { return "objekt_apartments" }
