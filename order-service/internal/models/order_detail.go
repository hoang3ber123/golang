package models

import "github.com/google/uuid"

type OrderDetail struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	OrderID     uuid.UUID `gorm:"type:char(36);not null;index"`
	RelatedID   string    `gorm:"type:varchar(36);not null"`
	RelatedType string    `gorm:"type:varchar(25);not null"`
	TotalPrice  float64   `gorm:"type:float;not null"`
	Order       Order     `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
}
