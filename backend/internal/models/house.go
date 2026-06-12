package models

// House — single-row real-estate listing (legacy "houses" table).
type House struct {
	BaseModel
	Title               string  `gorm:"size:255" json:"title"`
	ConstructionType    string  `gorm:"size:64" json:"construction_type"`
	District            string  `gorm:"size:128" json:"district"`
	Address             string  `gorm:"size:512" json:"address"`
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
	Files               JSONB   `gorm:"type:jsonb;default:'[]'" json:"files"`
	AuthorID            *uint   `gorm:"index" json:"author_id"`

	// v-2 public-website fields, embedded so they live as columns on `houses`.
	HousePublic
}
