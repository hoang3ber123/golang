package models

import (
	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// Define base model with slug
type BaseSlugUnique struct {
	BaseModel
	Title string `gorm:"not null;unique" json:"title"`
	Slug  string `gorm:"type:varchar(255);unique;not null" json:"slug"`
}

// Hook GORM: Trước khi tạo hoặc cập nhật
func (b *BaseSlugUnique) BeforeCreate(tx *gorm.DB) error {
	b.BaseModel.BeforeCreate(tx)
	// generate slug
	b.Slug = slug.MakeLang(b.Title, "en")
	return nil
}

func (b *BaseSlugUnique) BeforeUpdate(tx *gorm.DB) error {
	b.Slug = slug.MakeLang(b.Title, "en")
	return nil
}
