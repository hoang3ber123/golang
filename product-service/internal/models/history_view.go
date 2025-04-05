package models

import "github.com/google/uuid"

type HistoryView struct {
	BaseModel              // ID, CreatedAt, UpdatedAt
	UserID       uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_history_view"`
	RelatedID    uuid.UUID `gorm:"type:char(36);not null;uniqueIndex:idx_history_view"`
	RealatedType string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_history_view"`
	ClickTime    int       `gorm:"type:int;default:0"`
}
