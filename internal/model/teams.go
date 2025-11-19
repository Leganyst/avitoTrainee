package model

type Team struct {
	ID    uint   `gorm:"primaryKey;autoIncrement"`
	Name  string `gorm:"uniqueIndex;not null"`
	Users []User `gorm:"foreignKey:TeamID"`
}
