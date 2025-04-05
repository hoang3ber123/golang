package models

import (
	"github.com/google/uuid"
)

type Product struct {
	BaseSlug               // ID,CreatedAt,UpdatedAt,Slug,Title
	Description string     `gorm:"type:text"`
	Link        *string    `gorm:"type:varchar(255)"` // Dùng pointer để hỗ trợ nullable
	Categories  []Category `gorm:"many2many:product_categories;constraint:OnDelete:CASCADE"`
	UserID      uuid.UUID  `gorm:"type:char(36);not null;index"` // Chỉ lưu user_id, không định nghĩa relation
	Price       *float64   `gorm:"type:decimal(10,2)"`           // Dùng pointer để hỗ trợ nullable
}

func (*Product) GetRelatedType() string {
	return "products"
}
func (*Product) GetDirectory() string {
	return "products"
}
func (*Product) GetTableName() string {
	return "products"
}
