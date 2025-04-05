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
	RelatedID   uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_related"`
	RelatedType string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_related"`
	Status      string    `gorm:"type:varchar(20);not null;default:'using'"`
}
