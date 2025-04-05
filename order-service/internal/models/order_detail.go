package models

import "github.com/google/uuid"

type OrderDetail struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	OrderID     uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_related"`
	RelatedID   string    `gorm:"type:char(36);not null;uniqueIndex:idx_related"`
	RelatedType string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_related"`
	TotalPrice  float64   `gorm:"type:float;not null"`
	Order       Order     `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE"`
}
