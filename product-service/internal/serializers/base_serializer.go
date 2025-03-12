package serializers

import (
	"time"

	"github.com/google/uuid"
)

// BaseResponseSerializer chứa thông tin chung
type BaseResponseSerializer struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
