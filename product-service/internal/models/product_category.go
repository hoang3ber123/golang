package models

import "github.com/google/uuid"

type ProductCategory struct {
	ProductID  uuid.UUID `gorm:"type:char(36);not null;primaryKey"`
	CategoryID uuid.UUID `gorm:"type:char(36);not null;primaryKey"`
}
