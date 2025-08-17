package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"

	"github.com/h22k/analyzify/ent"
	"github.com/h22k/analyzify/internal/dto"
)

var (
	EventCreationErr = errors.New("failed to create event")
)

type EventRepository interface {
	CreateEvent(ctx context.Context, dto dto.CreateEventDTO) (*ent.Event, error)
}

type Event struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) Event {
	return Event{
		repo: repo,
	}
}

func (e Event) CreateEvent(ctx context.Context, createEventDto dto.CreateEventDTO) (*dto.CreateEventDTO, error) {
	event, err := e.repo.CreateEvent(ctx, createEventDto)

	if err != nil || event == nil {
		return nil, errors.Join(err, EventCreationErr)
	}

	var metadataMap map[string]any
	err = json.Unmarshal([]byte(event.Metadata), &metadataMap)

	if err != nil {
		log.Fatalf("failed to unmarshal metadata: %v", err)
	}

	return &dto.CreateEventDTO{
		EventID:   event.ID,
		UserID:    event.UserID,
		EventType: event.EventType,
		Timestamp: event.Timestamp,
		Metadata:  metadataMap,
	}, nil
}
