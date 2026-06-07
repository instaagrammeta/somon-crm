package models

// RealtyObject — building with map coordinates.
type RealtyObject struct {
	BaseModel
	Name                string  `gorm:"size:255" json:"name"`
	Description         string  `gorm:"type:text" json:"description"`
	Address             string  `gorm:"size:512" json:"address"`
	Lat                 float64 `gorm:"default:38.5598" json:"lat"`
	Lng                 float64 `gorm:"default:68.7870" json:"lng"`
	ConstructionType    string  `gorm:"size:64;default:'новостройка'" json:"construction_type"`
	District            string  `gorm:"size:128;default:'н.Сино'" json:"district"`
	Area                float64 `json:"area"`
	Rooms               int     `json:"rooms"`
	Windows             int     `json:"windows"`
	Floor               int     `json:"floor"`
	TotalFloors         int     `json:"total_floors"`
	PricePerM2          float64 `json:"price_per_m2"`
	TotalPrice          float64 `json:"total_price"`
	Developer           string  `gorm:"size:255" json:"developer"`
	ContactPhone        string  `gorm:"size:64" json:"contact_phone"`
	HasTechPassport     string  `gorm:"size:8;default:'нет'" json:"has_tech_passport"`
	HasRenovationPermit string  `gorm:"size:8;default:'нет'" json:"has_renovation_permit"`
	AuthorID            *uint   `gorm:"index" json:"author_id"`
}

type RealtyBlock struct {
	BaseModel
	ObjectID   uint   `gorm:"index" json:"object_id"`
	Name       string `gorm:"size:255" json:"name"`
	Code       string `gorm:"size:64" json:"code"`
	OrderIndex int    `gorm:"default:0;index" json:"order_index"`
}

type RealtyPricing struct {
	BaseModel
	ObjectID        uint    `gorm:"index" json:"object_id"`
	BlockID         *uint   `gorm:"index" json:"block_id"`
	FloorNumber     int     `json:"floor_number"`
	FloorRangeStart int     `json:"floor_range_start"`
	FloorRangeEnd   int     `json:"floor_range_end"`
	PercentValue    int     `json:"percent_value"`
	PriceUSD        float64 `json:"price_usd"`
	PriceTJS        float64 `json:"price_tjs"`
	Currency        string  `gorm:"size:8;default:'USD'" json:"currency"`
}

type RealtyLayout struct {
	BaseModel
	ObjectID     uint    `gorm:"index" json:"object_id"`
	RoomType     string  `gorm:"size:64" json:"room_type"`
	WindowsCount int     `json:"windows_count"`
	Area         float64 `json:"area"`
	PriceUSD     float64 `json:"price_usd"`
	PriceTJS     float64 `json:"price_tjs"`
	Description  string  `gorm:"type:text" json:"description"`
}
