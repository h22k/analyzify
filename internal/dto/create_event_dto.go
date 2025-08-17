package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateEventDTO struct {
	EventID   uuid.UUID
	UserID    uuid.UUID
	EventType string
	Timestamp time.Time
	Metadata  map[string]any
}
