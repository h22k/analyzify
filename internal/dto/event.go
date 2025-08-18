package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/h22k/analyzify/ent"
	"github.com/h22k/analyzify/graph/model"
)

type CreateEventDTO struct {
	EventID   uuid.UUID
	UserID    uuid.UUID
	EventType string
	Timestamp time.Time
	Metadata  map[string]any
}

type EventCountDTO struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
}

type EventDTO struct {
	EventID   uuid.UUID
	UserID    uuid.UUID
	EventType string
	Timestamp time.Time
	Metadata  map[string]any
}

func (e EventDTO) ToGraphqlEventModel() *model.Event {
	return &model.Event{
		EventID:   e.EventID,
		UserID:    e.UserID,
		EventType: e.EventType,
		Timestamp: e.Timestamp,
		Metadata:  e.Metadata,
	}
}

func ToEventDTO(event *ent.Event, metadataMap map[string]any) *EventDTO {
	return &EventDTO{
		EventID:   event.ID,
		UserID:    event.UserID,
		EventType: event.EventType,
		Timestamp: event.Timestamp,
		Metadata:  metadataMap,
	}
}
