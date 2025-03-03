package models

import (
	"github.com/google/uuid"
)

type Product struct {
	BaseSlug
	Title       string    `gorm:"type:varchar(255);not null"`
	Description string    `gorm:"type:text"`
	Link        *string   `gorm:"type:varchar(255)"` // Dùng pointer để hỗ trợ nullable
	CategoryID  uuid.UUID `gorm:"type:uuid;not null"`
	Category    Category  `gorm:"foreignKey:CategoryID"` // Quan hệ với Category
	UserID      uuid.UUID `gorm:"type:uuid;not null"`    // Chỉ lưu user_id, không định nghĩa relation
	Price       *float64  `gorm:"type:decimal(10,2)"`    // Dùng pointer để hỗ trợ nullable
}
