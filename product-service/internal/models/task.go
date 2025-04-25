package models

import "time"

const (
	TASK_MODE_INFINITY    = "infinity"
	TASK_MODE_MINUTES     = "minutes"
	TASK_STATUS_RUNNING   = "running"
	TASK_STATUS_STOPPED   = "stopped"
	TASK_STATUS_COMPLETED = "COMPLETED"
)

type Task struct {
	BaseSlugUnique            // ID, CreatedAt, UpdatedAt, Title, Slug
	Description    string     `json:"description"` // mô tả task
	Mode           string     `json:"mode"`        // "infinity" hoặc "minutes"
	Frequency      int        `json:"frequency"`   // Số phút hoặc khoảng cách thời gian chạy định kỳ
	StartedAt      time.Time  `json:"started_at"`
	StoppedAt      *time.Time `json:"stopped_at"`
	Status         string     `gorm:"default:'stopped'" json:"status"` // Mặc định là "stopped"
}
