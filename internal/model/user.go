package model

type User struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	UserID   string `gorm:"uniqueIndex;not null"`
	Username string `gorm:"not null"`
	IsActive bool   `gorm:"default:true"`

	TeamID uint
	// Позволяет удалить всех юзеров вместе с Team объектом
	Team Team `gorm:"constraint:OnDelete:CASCADE"`
}
