package graph

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/h22k/analyzify/ent"
	"github.com/h22k/analyzify/graph/model"
	"github.com/h22k/analyzify/internal/dto"
	"github.com/h22k/analyzify/internal/service"
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

func TestMutation_CreateEvent_Success(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	input := model.NewEvent{UserID: uid, EventType: "click", Metadata: map[string]any{"x": 1}}

	// Expect CreateEvent to be called with matching fields (ID is generated, Timestamp is now)
	matcher := mock.MatchedBy(func(d dto.CreateEventDTO) bool {
		if d.UserID != input.UserID || d.EventType != input.EventType {
			return false
		}
		return assert.ObjectsAreEqualValues(input.Metadata, d.Metadata) && !d.Timestamp.IsZero()
	})

	metadataJSON, _ := json.Marshal(input.Metadata)
	returnedID := uuid.New()
	returnedEnt := &ent.Event{ID: returnedID, UserID: uid, EventType: input.EventType, Timestamp: time.Now(), Metadata: string(metadataJSON)}

	repo := new(mockEventRepo)
	repo.On("CreateEvent", mock.Anything, matcher).Return(returnedEnt, nil).Once()

	resolver := &Resolver{EventService: service.NewEventService(repo)}
	got, err := resolver.Mutation().CreateEvent(ctx, input)
	repo.AssertExpectations(t)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, returnedID, got.EventID)
	assert.Equal(t, uid, got.UserID)
	assert.Equal(t, input.EventType, got.EventType)
	var expectedMeta map[string]any
	require.NoError(t, json.Unmarshal(metadataJSON, &expectedMeta))
	assert.Equal(t, expectedMeta, got.Metadata)
}

func TestMutation_CreateEvent_Error(t *testing.T) {
	ctx := context.Background()
	input := model.NewEvent{UserID: uuid.New(), EventType: "view", Metadata: map[string]any{"ok": true}}

	repo := new(mockEventRepo)
	repo.On("CreateEvent", mock.Anything, mock.AnythingOfType("dto.CreateEventDTO")).Return((*ent.Event)(nil), errors.New("db fail")).Once()

	resolver := &Resolver{EventService: service.NewEventService(repo)}
	got, err := resolver.Mutation().CreateEvent(ctx, input)
	repo.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, got)
	assert.ErrorIs(t, err, EventCreateErr)
	assert.ErrorIs(t, err, service.EventCreationErr)
}

func TestQuery_EventsByUserID_Success(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	m1 := map[string]any{"a": 1}
	m2 := map[string]any{"b": "x"}
	b1, _ := json.Marshal(m1)
	b2, _ := json.Marshal(m2)
	entEvents := []*ent.Event{
		{ID: uuid.New(), UserID: uid, EventType: "a", Timestamp: time.Now(), Metadata: string(b1)},
		{ID: uuid.New(), UserID: uid, EventType: "b", Timestamp: time.Now(), Metadata: string(b2)},
	}

	repo := new(mockEventRepo)
	repo.On("GetEventsByUserID", mock.Anything, uid).Return(entEvents, nil).Once()

	resolver := &Resolver{EventService: service.NewEventService(repo)}
	got, err := resolver.Query().EventsByUserID(ctx, uid)
	repo.AssertExpectations(t)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, entEvents[0].ID, got[0].EventID)
	assert.Equal(t, entEvents[1].ID, got[1].EventID)
	assert.Equal(t, "a", got[0].EventType)
	assert.Equal(t, "b", got[1].EventType)
	var exp1, exp2 map[string]any
	require.NoError(t, json.Unmarshal(b1, &exp1))
	require.NoError(t, json.Unmarshal(b2, &exp2))
	assert.Equal(t, exp1, got[0].Metadata)
	assert.Equal(t, exp2, got[1].Metadata)
}

func TestQuery_EventsByUserID_Error(t *testing.T) {
	ctx := context.Background()
	uid := uuid.New()
	repo := new(mockEventRepo)
	repo.On("GetEventsByUserID", mock.Anything, uid).Return(([]*ent.Event)(nil), errors.New("boom")).Once()

	resolver := &Resolver{EventService: service.NewEventService(repo)}
	got, err := resolver.Query().EventsByUserID(ctx, uid)
	repo.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, got)
	assert.ErrorIs(t, err, EventFetchErr)
}

func TestQuery_EventCountByEventType_Success(t *testing.T) {
	ctx := context.Background()
	evtType := "x"
	from := time.Now().Add(-time.Hour)
	to := time.Now()
	counts := []dto.EventCountDTO{{EventType: "x", Count: 3}, {EventType: "y", Count: 1}}

	repo := new(mockEventRepo)
	repo.On("GetEventCountByEventType", mock.Anything, &evtType, from, to).Return(counts, nil).Once()

	resolver := &Resolver{EventService: service.NewEventService(repo)}
	got, err := resolver.Query().EventCountByEventType(ctx, &evtType, from, to)
	repo.AssertExpectations(t)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, counts[0].EventType, got[0].EventType)
	assert.Equal(t, int32(counts[0].Count), got[0].Count)
	assert.Equal(t, counts[1].EventType, got[1].EventType)
	assert.Equal(t, int32(counts[1].Count), got[1].Count)
}

func TestQuery_EventCountByEventType_Error(t *testing.T) {
	ctx := context.Background()
	from := time.Now().Add(-time.Hour)
	to := time.Now()
	repo := new(mockEventRepo)
	repo.On("GetEventCountByEventType", mock.Anything, (*string)(nil), from, to).Return(([]dto.EventCountDTO)(nil), errors.New("fail")).Once()

	resolver := &Resolver{EventService: service.NewEventService(repo)}
	got, err := resolver.Query().EventCountByEventType(ctx, nil, from, to)
	repo.AssertExpectations(t)
	require.Error(t, err)
	require.Nil(t, got)
	assert.ErrorIs(t, err, EventFetchErr)
}
