package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/h22k/analyzify/ent"
	"github.com/h22k/analyzify/internal/dto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockEventRepo struct{ mock.Mock }

func (m *mockEventRepo) CreateEvent(ctx context.Context, d dto.CreateEventDTO) (*ent.Event, error) {
	args := m.Called(ctx, d)
	var ev *ent.Event
	if v := args.Get(0); v != nil {
		ev = v.(*ent.Event)
	}
	return ev, args.Error(1)
}

func (m *mockEventRepo) GetEventCountByEventType(ctx context.Context, eventType *string, from, to time.Time) ([]dto.EventCountDTO, error) {
	args := m.Called(ctx, eventType, from, to)
	var res []dto.EventCountDTO
	if v := args.Get(0); v != nil {
		res = v.([]dto.EventCountDTO)
	}
	return res, args.Error(1)
}

func (m *mockEventRepo) GetEventsByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.Event, error) {
	args := m.Called(ctx, userID)
	var res []*ent.Event
	if v := args.Get(0); v != nil {
		res = v.([]*ent.Event)
	}
	return res, args.Error(1)
}

func TestEvent_CreateEvent_Success(t *testing.T) {
	ctx := context.Background()
	createDTO := dto.CreateEventDTO{
		EventID:   uuid.New(),
		UserID:    uuid.New(),
		EventType: "click",
		Timestamp: time.Now().UTC().Truncate(time.Second),
		Metadata:  map[string]any{"path": "/home", "count": 3},
	}
	metadataJSONBytes, _ := json.Marshal(createDTO.Metadata)

	expectedEnt := &ent.Event{
		ID:        createDTO.EventID,
		UserID:    createDTO.UserID,
		EventType: createDTO.EventType,
		Timestamp: createDTO.Timestamp,
		Metadata:  string(metadataJSONBytes),
	}

	repo := new(mockEventRepo)
	repo.On("CreateEvent", ctx, createDTO).Return(expectedEnt, nil).Once()

	service := NewEventService(repo)
	result, err := service.CreateEvent(ctx, createDTO)
	require.NoError(t, err)
	repo.AssertExpectations(t)
	require.NotNil(t, result)
	assert.Equal(t, createDTO.EventID, result.EventID)
	assert.Equal(t, createDTO.UserID, result.UserID)
	assert.Equal(t, createDTO.EventType, result.EventType)
	assert.True(t, createDTO.Timestamp.Equal(result.Timestamp))

	var expectedMeta map[string]any
	require.NoError(t, json.Unmarshal(metadataJSONBytes, &expectedMeta))
	assert.Equal(t, expectedMeta, result.Metadata)
}

func TestEvent_CreateEvent_ErrorFromRepo(t *testing.T) {
	ctx := context.Background()
	repo := new(mockEventRepo)
	repo.On("CreateEvent", ctx, mock.AnythingOfType("dto.CreateEventDTO")).Return((*ent.Event)(nil), errors.New("db down")).Once()

	service := NewEventService(repo)
	res, err := service.CreateEvent(ctx, dto.CreateEventDTO{})
	repo.AssertExpectations(t)
	require.Nil(t, res)
	assert.Error(t, err)
	assert.ErrorIs(t, err, EventCreationErr)
}

func TestEvent_CreateEvent_NilEventNoError(t *testing.T) {
	ctx := context.Background()
	repo := new(mockEventRepo)
	repo.On("CreateEvent", ctx, mock.AnythingOfType("dto.CreateEventDTO")).Return((*ent.Event)(nil), nil).Once()

	service := NewEventService(repo)
	res, err := service.CreateEvent(ctx, dto.CreateEventDTO{})
	repo.AssertExpectations(t)
	require.Nil(t, res)
	assert.Error(t, err)
	assert.ErrorIs(t, err, EventCreationErr)
}

func TestEvent_CreateEvent_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	bad := &ent.Event{ID: uuid.New(), UserID: uuid.New(), EventType: "x", Timestamp: time.Now(), Metadata: "{not-json"}
	repo := new(mockEventRepo)
	repo.On("CreateEvent", ctx, mock.AnythingOfType("dto.CreateEventDTO")).Return(bad, nil).Once()

	service := NewEventService(repo)
	res, err := service.CreateEvent(ctx, dto.CreateEventDTO{})
	repo.AssertExpectations(t)
	require.Nil(t, res)
	assert.Error(t, err)
	assert.ErrorIs(t, err, JsonUnmarshalErr)
}

func TestEvent_GetEventCountByEventType_Success(t *testing.T) {
	ctx := context.Background()
	evtType := "click"
	from := time.Now().Add(-24 * time.Hour).UTC().Truncate(time.Second)
	to := time.Now().UTC().Truncate(time.Second)
	expected := []dto.EventCountDTO{{EventType: "click", Count: 5}, {EventType: "view", Count: 2}}

	repo := new(mockEventRepo)
	repo.On("GetEventCountByEventType", ctx, &evtType, from, to).Return(expected, nil).Once()

	service := NewEventService(repo)
	counts, err := service.GetEventCountByEventType(ctx, &evtType, from, to)
	repo.AssertExpectations(t)
	require.NoError(t, err)
	assert.Equal(t, expected, counts)
}

func TestEvent_GetEventCountByEventType_Error(t *testing.T) {
	ctx := context.Background()
	repo := new(mockEventRepo)
	repo.On("GetEventCountByEventType", ctx, (*string)(nil), mock.AnythingOfType("time.Time"), mock.AnythingOfType("time.Time")).Return(([]dto.EventCountDTO)(nil), errors.New("query failed")).Once()

	service := NewEventService(repo)
	counts, err := service.GetEventCountByEventType(ctx, nil, time.Now(), time.Now())
	repo.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, counts)
}

func TestEvent_GetEventsByUserID_Success(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	metadata1 := map[string]any{"foo": "bar", "n": 1}
	metadata2 := map[string]any{"ok": true}
	b1, _ := json.Marshal(metadata1)
	b2, _ := json.Marshal(metadata2)

	entEvents := []*ent.Event{
		{ID: uuid.New(), UserID: userID, EventType: "a", Timestamp: time.Now().UTC().Truncate(time.Second), Metadata: string(b1)},
		{ID: uuid.New(), UserID: userID, EventType: "b", Timestamp: time.Now().UTC().Truncate(time.Second), Metadata: string(b2)},
	}

	repo := new(mockEventRepo)
	repo.On("GetEventsByUserID", ctx, userID).Return(entEvents, nil).Once()

	service := NewEventService(repo)
	dtos, err := service.GetEventsByUserID(ctx, userID)
	repo.AssertExpectations(t)
	require.NoError(t, err)
	require.Len(t, dtos, len(entEvents))
	assert.Equal(t, entEvents[0].ID, dtos[0].EventID)
	assert.Equal(t, entEvents[1].ID, dtos[1].EventID)

	var expected1, expected2 map[string]any
	require.NoError(t, json.Unmarshal(b1, &expected1))
	require.NoError(t, json.Unmarshal(b2, &expected2))
	assert.Equal(t, expected1, dtos[0].Metadata)
	assert.Equal(t, expected2, dtos[1].Metadata)
}

func TestEvent_GetEventsByUserID_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	bad := []*ent.Event{{ID: uuid.New(), UserID: userID, EventType: "x", Timestamp: time.Now(), Metadata: "{bad-json"}}

	repo := new(mockEventRepo)
	repo.On("GetEventsByUserID", ctx, userID).Return(bad, nil).Once()

	service := NewEventService(repo)
	res, err := service.GetEventsByUserID(ctx, userID)
	repo.AssertExpectations(t)
	require.Nil(t, res)
	require.Error(t, err)
	assert.ErrorIs(t, err, JsonUnmarshalErr)
}

func TestEvent_GetEventsByUserID_RepoError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := new(mockEventRepo)
	repo.On("GetEventsByUserID", ctx, userID).Return(([]*ent.Event)(nil), errors.New("db error")).Once()

	service := NewEventService(repo)
	res, err := service.GetEventsByUserID(ctx, userID)
	repo.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, res)
}
