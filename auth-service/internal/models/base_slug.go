package models

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// Define base model with slug
type BaseSlug struct {
	BaseModel
	Title string `gorm:"not null"`
	Slug  string `gorm:"type:varchar(255);uniqueIndex;not null"`
}

// Hàm tạo slug duy nhất bằng cách sử dụng LIKE 'slug%'
func (b *BaseSlug) generateUniqueSlug(db *gorm.DB) {
	slug := slug.MakeLang(b.Title, "en")
	uniqueSlug := slug
	var existingSlugs []string

	// Tìm tất cả slug có dạng 'slug%'
	db.Model(&BaseSlug{}).
		Where("slug LIKE ?", slug+"%").
		Pluck("slug", &existingSlugs)

	if len(existingSlugs) == 0 {
		b.Slug = uniqueSlug
		return
	}

	// Tìm số lớn nhất trong các slug có dạng 'slug-1', 'slug-2'
	maxNumber := 0
	slugPattern := regexp.MustCompile(`^` + regexp.QuoteMeta(slug) + `-(\d+)$`)

	for _, existingSlug := range existingSlugs {
		matches := slugPattern.FindStringSubmatch(existingSlug)
		if len(matches) == 2 {
			num, err := strconv.Atoi(matches[1])
			if err == nil && num > maxNumber {
				maxNumber = num
			}
		}
	}

	// Gán slug mới với số lớn nhất + 1
	b.Slug = fmt.Sprintf("%s-%d", slug, maxNumber+1)
}

// Hook GORM: Trước khi tạo hoặc cập nhật
func (b *BaseSlug) BeforeCreate(tx *gorm.DB) error {
	b.BaseModel.BeforeCreate(tx)
	// generate slug
	b.generateUniqueSlug(tx)
	return nil
}

func (b *BaseSlug) BeforeUpdate(tx *gorm.DB) error {
	b.generateUniqueSlug(tx)
	return nil
}
