package clickhouse

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	esql "entgo.io/ent/dialect/sql"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/google/uuid"
	"github.com/h22k/analyzify/ent"
	"github.com/h22k/analyzify/ent/event"
	query "github.com/h22k/analyzify/internal/db/clickhouse/sql"
	"github.com/h22k/analyzify/internal/dto"
)

const clickHouseDialect = "clickhouse"

var (
	ConnectionNotInitializedErr = errors.New("ClickHouse connection is not initialized")
	MigrationFailedErr          = errors.New("ClickHouse migration failed")
	ConnectionCloseErr          = errors.New("ClickHouse connection close error")
	EntConnectionCloseErr       = errors.New("ent client close error")
	EventCountByEventTypeErr    = errors.New("failed to get event count by event type")
	JsonMarshalErr              = errors.New("failed to marshal metadata to JSON")
)

func NewConn(dbHost, dbUser, dbPass, dbPort, dbName string) *Conn {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{dbHost + ":" + dbPort},
		Auth: clickhouse.Auth{
			Username: dbUser,
			Password: dbPass,
			Database: dbName,
		},
	})

	if err := conn.Ping(); err != nil {
		panic("failed to connect to ClickHouse: " + err.Error())
	}

	return &Conn{
		conn:   conn,
		client: ent.NewClient(ent.Driver(esql.NewDriver(clickHouseDialect, esql.Conn{ExecQuerier: conn}))), // what the f is this
	}
}

type Conn struct {
	conn *sql.DB

	client *ent.Client
}

func (c *Conn) Migrate() error {
	if c.conn == nil {
		return ConnectionNotInitializedErr
	}

	if _, err := c.conn.Exec(query.CreateEventsTableQuery); err != nil {
		return errors.Join(err, MigrationFailedErr)
	}

	return nil
}

func (c *Conn) GetEventCountByEventType(ctx context.Context, eventType *string, from, to time.Time) ([]dto.EventCountDTO, error) {
	var eventCounts []dto.EventCountDTO

	err := c.client.
		Event.
		Query().
		Select(event.FieldEventType).
		Where(
			event.TimestampGTE(from),
			event.TimestampLTE(to),
		).
		Modify(func(s *esql.Selector) {
			s.GroupBy(event.FieldEventType)
			if eventType != nil {
				s.Having(esql.EQ(event.FieldEventType, *eventType))
			}
		}).
		Aggregate(ent.Count()).
		Scan(ctx, &eventCounts)

	if err != nil {
		return nil, errors.Join(err, EventCountByEventTypeErr)
	}

	return eventCounts, nil
}

func (c *Conn) CreateEvent(ctx context.Context, dto dto.CreateEventDTO) (*ent.Event, error) {
	metadataBytes, err := json.Marshal(dto.Metadata)
	if err != nil {
		return nil, errors.Join(err, JsonMarshalErr)
	}

	return c.client.
		Event.
		Create().
		SetID(dto.EventID).
		SetUserID(dto.UserID).
		SetEventType(dto.EventType).
		SetTimestamp(dto.Timestamp).
		SetMetadata(string(metadataBytes)).
		Save(ctx)
}

func (c *Conn) GetEventsByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.Event, error) {
	return c.client.
		Event.
		Query().
		Modify(func(s *esql.Selector) {
			s.AppendSelect(
				// convert metadata to string for ClickHouse and EntGo compatibility
				esql.As("toString(metadata)", "metadata"),
			)
		}).
		Where(event.UserIDEQ(userID)).
		All(ctx)
}

func (c *Conn) Close() error {
	if c.conn == nil {
		return nil
	}
	if err := c.conn.Close(); err != nil {
		return errors.Join(err, ConnectionCloseErr)
	}
	c.conn = nil

	if c.client != nil {
		if err := c.client.Close(); err != nil {
			return errors.Join(err, EntConnectionCloseErr)
		}
		c.client = nil
	}

	return nil
}
