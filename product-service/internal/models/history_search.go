package models

import "github.com/google/uuid"

type HistorySearch struct {
	BaseModel
	UserID     uuid.UUID `gorm:"type:char(36);index;not null"`
	Categories string    `gorm:"type:text;not null"`
}
