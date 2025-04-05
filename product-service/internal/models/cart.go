package models

import "github.com/google/uuid"

type Cart struct {
	BaseModel
	UserID uuid.UUID  `gorm:"type:char(36);uniqueIndex;not null"`
	Items  []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"`
}
