//go:build e2e

package e2e

import (
	"context"
	"os"
	"sort"
	"testing"
	"time"

	chdriver "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/h22k/analyzify/config"
	"github.com/h22k/analyzify/internal/db/clickhouse"
	"github.com/h22k/analyzify/internal/dto"
	"github.com/h22k/analyzify/internal/service"
)

var (
	conn   *clickhouse.Conn
	seeded struct {
		user1 uuid.UUID
		user2 uuid.UUID
		from  time.Time
		to    time.Time
	}
)

func TestMain(m *testing.M) {
	// Ensure test env
	_ = os.Setenv("APP_ENV", "test")

	cfg := config.Load()
	conn = clickhouse.NewConn(cfg.DbHost(), cfg.DbUser(), cfg.DbPass(), cfg.DbPort(), cfg.DbName())

	// Migrate fresh
	if err := conn.Migrate(); err != nil {
		_ = conn.Close()
		os.Exit(1)
	}

	// Seed data
	seedData()

	code := m.Run()

	// Teardown: drop table and close connections
	dropTable(cfg)
	_ = conn.Close()

	os.Exit(code)
}

func seedData() {
	svc := service.NewEventService(conn)
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Second)
	seeded.from = now.Add(-2 * time.Hour)
	seeded.to = now.Add(2 * time.Hour)
	seeded.user1 = uuid.New()
	seeded.user2 = uuid.New()

	// user1: 3 click, 1 view
	for i := 0; i < 3; i++ {
		_, err := svc.CreateEvent(ctx, dto.CreateEventDTO{
			EventID:   uuid.New(),
			UserID:    seeded.user1,
			EventType: "click",
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Metadata:  map[string]any{"i": i, "user": "u1"},
		})
		if err != nil {
			panic(err)
		}
	}
	_, err := svc.CreateEvent(ctx, dto.CreateEventDTO{
		EventID:   uuid.New(),
		UserID:    seeded.user1,
		EventType: "view",
		Timestamp: now.Add(10 * time.Minute),
		Metadata:  map[string]any{"p": "/home"},
	})
	if err != nil {
		panic(err)
	}

	// user2: 2 click, 2 signup
	for i := 0; i < 2; i++ {
		_, err := svc.CreateEvent(ctx, dto.CreateEventDTO{
			EventID:   uuid.New(),
			UserID:    seeded.user2,
			EventType: "click",
			Timestamp: now.Add(time.Duration(20+i) * time.Minute),
			Metadata:  map[string]any{"i": i, "user": "u2"},
		})
		if err != nil {
			panic(err)
		}
	}
	for i := 0; i < 2; i++ {
		_, err := svc.CreateEvent(ctx, dto.CreateEventDTO{
			EventID:   uuid.New(),
			UserID:    seeded.user2,
			EventType: "signup",
			Timestamp: now.Add(time.Duration(30+i) * time.Minute),
			Metadata:  map[string]any{"i": i, "user": "u2"},
		})
		if err != nil {
			panic(err)
		}
	}
}

func dropTable(cfg config.Config) {
	db := chdriver.OpenDB(&chdriver.Options{
		Addr: []string{cfg.DbHost() + ":" + cfg.DbPort()},
		Auth: chdriver.Auth{
			Username: cfg.DbUser(),
			Password: cfg.DbPass(),
			Database: cfg.DbName(),
		},
	})
	_, _ = db.Exec("DROP TABLE IF EXISTS analyzify.events")
	_ = db.Close()
}

func Test_EventCountByEventType_All(t *testing.T) {
	svc := service.NewEventService(conn)
	ctx := context.Background()

	counts, err := svc.GetEventCountByEventType(ctx, nil, seeded.from, seeded.to)
	require.NoError(t, err)

	// Convert to map for easy assertion
	got := map[string]int{}
	for _, c := range counts {
		got[c.EventType] = c.Count
	}
	assert.Equal(t, 5, got["click"])  // 3 + 2
	assert.Equal(t, 1, got["view"])   // 1
	assert.Equal(t, 2, got["signup"]) // 2
}

func Test_EventCountByEventType_Filtered(t *testing.T) {
	svc := service.NewEventService(conn)
	ctx := context.Background()
	evt := "click"
	counts, err := svc.GetEventCountByEventType(ctx, &evt, seeded.from, seeded.to)
	require.NoError(t, err)
	require.Len(t, counts, 1)
	assert.Equal(t, "click", counts[0].EventType)
	assert.Equal(t, 5, counts[0].Count)
}

func Test_GetEventsByUserID_OrderAndMetadata(t *testing.T) {
	svc := service.NewEventService(conn)
	ctx := context.Background()
	events, err := svc.GetEventsByUserID(ctx, seeded.user1)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(events), 4)

	// Sort by timestamp asc to have stable expectations
	sort.Slice(events, func(i, j int) bool { return events[i].Timestamp.Before(events[j].Timestamp) })

	// First three should be clicks for user1
	for i := 0; i < 3; i++ {
		assert.Equal(t, "click", events[i].EventType)
		assert.Equal(t, seeded.user1, events[i].UserID)
		// metadata number may be float64 due to JSON, so just check key presence
		_, hasI := events[i].Metadata["i"]
		assert.True(t, hasI)
		assert.Equal(t, "u1", events[i].Metadata["user"])
	}
}
