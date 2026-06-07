package models

type Bank struct {
	BaseModel
	Name        string `gorm:"size:255" json:"name"`
	Slug        string `gorm:"size:128;uniqueIndex" json:"slug"`
	Logo        string `gorm:"size:512" json:"logo"`
	Description string `gorm:"type:text" json:"description"`
	OrderIndex  int    `gorm:"default:0;index" json:"order_index"`
	IsActive    bool   `gorm:"default:true;index" json:"is_active"`
}

type MortgageCondition struct {
	BaseModel
	BankID             uint    `gorm:"index" json:"bank_id"`
	Currency           string  `gorm:"size:8;default:'TJS'" json:"currency"`
	InterestYearly     float64 `gorm:"default:0" json:"interest_yearly"`
	InterestMonthly    float64 `gorm:"default:0" json:"interest_monthly"`
	MinAmount          float64 `gorm:"default:0" json:"min_amount"`
	MaxAmount          float64 `gorm:"default:0" json:"max_amount"`
	MinMonths          int     `gorm:"default:12" json:"min_months"`
	MaxMonths          int     `gorm:"default:240" json:"max_months"`
	DownPaymentPercent float64 `gorm:"default:30" json:"down_payment_percent"`
	GuarantorRequired  bool    `gorm:"default:false" json:"guarantor_required"`
	CollateralRequired bool    `gorm:"default:false" json:"collateral_required"`
	ExtraConditions    string  `gorm:"type:text" json:"extra_conditions"`
}

type InstallmentObject struct {
	BaseModel
	Name        string `gorm:"size:255" json:"name"`
	Slug        string `gorm:"size:128;uniqueIndex" json:"slug"`
	Image       string `gorm:"size:512" json:"image"`
	Description string `gorm:"type:text" json:"description"`
	Address     string `gorm:"size:512" json:"address"`
	Developer   string `gorm:"size:255" json:"developer"`
	OrderIndex  int    `gorm:"default:0;index" json:"order_index"`
	IsActive    bool   `gorm:"default:true;index" json:"is_active"`
}

type InstallmentCondition struct {
	BaseModel
	ObjectID           uint    `gorm:"index" json:"object_id"`
	Title              string  `gorm:"size:255" json:"title"`
	Currency           string  `gorm:"size:8;default:'TJS'" json:"currency"`
	MinPrice           float64 `gorm:"default:0" json:"min_price"`
	MaxPrice           float64 `gorm:"default:0" json:"max_price"`
	DownPaymentPercent float64 `gorm:"default:30" json:"down_payment_percent"`
	Months             int     `gorm:"default:12" json:"months"`
	MonthlyPaymentRule string  `gorm:"type:text" json:"monthly_payment_rule"`
	InterestPercent    float64 `gorm:"default:0" json:"interest_percent"`
	ExtraConditions    string  `gorm:"type:text" json:"extra_conditions"`
}

func (Bank) TableName() string                 { return "banks" }
func (MortgageCondition) TableName() string    { return "mortgage_conditions" }
func (InstallmentObject) TableName() string    { return "installment_objects" }
func (InstallmentCondition) TableName() string { return "installment_conditions" }
