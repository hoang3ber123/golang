package models

import (
	"github.com/google/uuid"
)

type Product struct {
	BaseSlug
	Description string     `gorm:"type:text"`
	Link        *string    `gorm:"type:varchar(255)"` // Dùng pointer để hỗ trợ nullable
	Categories  []Category `gorm:"many2many:product_categories;constraint:OnDelete:CASCADE"`
	UserID      uuid.UUID  `gorm:"type:char(36);not null"` // Chỉ lưu user_id, không định nghĩa relation
	Price       *float64   `gorm:"type:decimal(10,2)"`     // Dùng pointer để hỗ trợ nullable
}

func (*Product) GetRelatedType() string {
	return "product"
}
func (*Product) GetDirectory() string {
	return "product"
}
func (*Product) GetTableName() string {
	return "products"
}
