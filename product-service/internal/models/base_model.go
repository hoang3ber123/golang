package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Define Base Model
type BaseModel struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// Hook tự động gán UUID trước khi lưu vào database
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	b.ID = uuid.New()
	return
}
