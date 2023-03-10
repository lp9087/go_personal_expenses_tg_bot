package models

type CategoryName struct {
	ID    uint
	Name  string
	Owner string
}

type Expenses struct {
	ID         uint
	CategoryID int
	Category   CategoryName `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Amount     int
}
