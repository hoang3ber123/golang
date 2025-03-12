package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	MediaStatusUsing         = "using"
	MediaStatusUpdated       = "updated"
	MediaStatusDeleted       = "deleted"
	MediaStatusDeleteCascade = "delete_cascade" // trạng thái bị xóa khi mà object bị xóa
)

type Media struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	File        string    `gorm:"type:varchar(255);not null"`
	FileType    string    `gorm:"type:varchar(20);not null"`
	RelatedID   uuid.UUID `gorm:"type:char(36);not null;index:idx_related"`    // Chỉ lưu user_id, không định nghĩa relation
	RelatedType string    `gorm:"type:varchar(50);not null;index:idx_related"` // Chỉ lưu user_id, không định nghĩa relation
	Status      string    `gorm:"type:varchar(20);not null;default:'using'"`
}
