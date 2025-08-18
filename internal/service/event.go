package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/h22k/analyzify/ent"
	"github.com/h22k/analyzify/internal/dto"
)

var (
	EventCreationErr = errors.New("failed to create event")
	JsonUnmarshalErr = errors.New("failed to unmarshal metadata from JSON")
)

type EventRepository interface {
	CreateEvent(ctx context.Context, dto dto.CreateEventDTO) (*ent.Event, error)
	GetEventCountByEventType(ctx context.Context, eventType *string, from, to time.Time) ([]dto.EventCountDTO, error)
	GetEventsByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.Event, error)
}

type Event struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) Event {
	return Event{
		repo: repo,
	}
}

func (e Event) CreateEvent(ctx context.Context, createEventDto dto.CreateEventDTO) (*dto.EventDTO, error) {
	event, err := e.repo.CreateEvent(ctx, createEventDto)

	if err != nil || event == nil {
		return nil, errors.Join(err, EventCreationErr)
	}

	var metadataMap map[string]any
	err = json.Unmarshal([]byte(event.Metadata), &metadataMap)

	if err != nil {
		return nil, errors.Join(err, JsonUnmarshalErr)
	}

	return dto.ToEventDTO(event, metadataMap), nil
}

func (e Event) GetEventCountByEventType(ctx context.Context, eventType *string, from, to time.Time) ([]dto.EventCountDTO, error) {
	return e.repo.GetEventCountByEventType(ctx, eventType, from, to)
}

func (e Event) GetEventsByUserID(ctx context.Context, userID uuid.UUID) ([]*dto.EventDTO, error) {
	events, err := e.repo.GetEventsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var eventDTOs []*dto.EventDTO
	for _, event := range events {
		var metadataMap map[string]any
		err = json.Unmarshal([]byte(event.Metadata), &metadataMap)
		if err != nil {
			return nil, errors.Join(err, JsonUnmarshalErr)
		}

		eventDTOs = append(eventDTOs, dto.ToEventDTO(event, metadataMap))
	}

	return eventDTOs, nil
}
