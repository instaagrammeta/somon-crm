package models

import "time"

type CompanyPhone struct {
	BaseModel
	Model       string `gorm:"size:128" json:"model"`
	PhoneID     string `gorm:"size:64;uniqueIndex" json:"phone_id"`
	AssignedTo  *uint  `gorm:"index" json:"assigned_to"`
	Description string `gorm:"type:text" json:"description"`
	Status      string `gorm:"size:32;default:'free';index" json:"status"`
}

type SimCard struct {
	BaseModel
	PhoneNumber string `gorm:"size:32;uniqueIndex" json:"phone_number"`
	Operator    string `gorm:"size:64" json:"operator"`
	AssignedTo  *uint  `gorm:"index" json:"assigned_to"`
	PhoneID     *uint  `gorm:"index" json:"phone_id"`
	Description string `gorm:"type:text" json:"description"`
	Status      string `gorm:"size:32;default:'active';index" json:"status"`
}

type SimTariff struct {
	BaseModel
	SimID     uint       `gorm:"index" json:"sim_id"`
	Minutes   int        `gorm:"default:0" json:"minutes"`
	GB        float64    `gorm:"default:0" json:"gb"`
	SMS       int        `gorm:"default:0" json:"sms"`
	Cost      float64    `gorm:"default:0" json:"cost"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `gorm:"index" json:"end_date"`
	Status    string     `gorm:"size:32;default:'active';index" json:"status"`
}

type TariffPayment struct {
	BaseModel
	SimID       uint       `gorm:"index" json:"sim_id"`
	TariffID    *uint      `gorm:"index" json:"tariff_id"`
	Amount      float64    `gorm:"default:0" json:"amount"`
	PaymentDate *time.Time `json:"payment_date"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Status      string     `gorm:"size:32;default:'paid'" json:"status"`
}
