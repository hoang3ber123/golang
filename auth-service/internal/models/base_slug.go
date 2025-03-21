package models

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/gosimple/slug"
	"gorm.io/gorm"
)

// Define base model with slug
type BaseSlug struct {
	BaseModel
	Title string `gorm:"not null"`
	Slug  string `gorm:"type:varchar(255);uniqueIndex;not null"`
}

// Interface để các model cụ thể implement
type Slugable interface {
	GetTableName() string
}

// Hàm tạo slug duy nhất
func (b *BaseSlug) generateUniqueSlug(db *gorm.DB) error {
	// Lấy model từ tx.Statement.Dest
	dest := db.Statement.Dest

	// Dùng reflection để xử lý double pointer
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Ptr {
		dest = v.Elem().Interface() // Giải tham chiếu
	}

	// Kiểm tra xem model có implement Slugable không
	_, ok := dest.(Slugable)
	if !ok {
		return errors.New("model must implement Slugable interface")
	}

	// Tạo slug cơ bản từ title và timestamp
	slugBase := slug.MakeLang(b.Title, "en")

	// Lấy timestamp theo định dạng YYYYMMDDHHMMSSmmm
	now := time.Now()
	timestamp := fmt.Sprintf("%s%03d", now.Format("20060102-150405-"), now.Nanosecond()/1e6)
	// Kết hợp slugBase với timestamp
	b.Slug = fmt.Sprintf("%s-%s", slugBase, timestamp)

	return nil
}

// Hook GORM: Trước khi tạo hoặc cập nhật
func (b *BaseSlug) BeforeCreate(tx *gorm.DB) error {
	b.BaseModel.BeforeCreate(tx)
	// generate slug
	if err := b.generateUniqueSlug(tx); err != nil {
		return err
	}
	return nil
}

func (b *BaseSlug) BeforeUpdate(tx *gorm.DB) error {
	b.generateUniqueSlug(tx)
	return nil
}
