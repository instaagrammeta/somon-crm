package models

import "gorm.io/gorm"

// AllModels returns the full list of GORM models for AutoMigrate.
// Used as a fallback when SQL migrations aren't available.
func AllModels() []any {
	return []any{
		&User{},
		&Task{},
		&LegacyRequest{},
		&RequestsBoard{},
		&RequestsColumn{},
		&RequestItem{},
		&Post{},
		&Message{},
		&Lid{},
		&KanbanBoard{},
		&KanbanBoardMember{},
		&KanbanColumn{},
		&KanbanLead{},
		&KanbanLeadInteraction{},
		&Folder{},
		&FolderFile{},
		&House{},
		&RealtyObject{},
		&RealtyBlock{},
		&RealtyPricing{},
		&RealtyLayout{},
		&ObjektProject{},
		&ObjektBlock{},
		&ObjektApartment{},
		&CompanyPhone{},
		&SimCard{},
		&SimTariff{},
		&TariffPayment{},
		&Bank{},
		&MortgageCondition{},
		&InstallmentObject{},
		&InstallmentCondition{},
	}
}

// AutoMigrateAll runs GORM AutoMigrate for every registered model.
func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(AllModels()...)
}
