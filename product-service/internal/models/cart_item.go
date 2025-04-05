package models

import "github.com/google/uuid"

type CartItem struct {
	BaseModel
	CartID      string    `gorm:"type:char(36),index;not null"`
	RelatedID   uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_related"`
	RelatedType string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_related"`
}
