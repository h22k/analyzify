package clickhouse

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	esql "entgo.io/ent/dialect/sql"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/h22k/analyzify/ent"
	query "github.com/h22k/analyzify/internal/db/clickhouse/sql"
	"github.com/h22k/analyzify/internal/dto"
)

const clickHouseDialect = "clickhouse"

var (
	ConnectionNotInitializedErr = errors.New("ClickHouse connection is not initialized")
	MigrationFailedErr          = errors.New("ClickHouse migration failed")
	ConnectionCloseErr          = errors.New("ClickHouse connection close error")
	EntConnectionCloseErr       = errors.New("ent client close error")
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

func (c *Conn) GetAllEvents(ctx context.Context) ([]*ent.Event, error) {
	return c.client.Debug().Event.Query().All(ctx)
}

func (c *Conn) CreateEvent(ctx context.Context, dto dto.CreateEventDTO) (*ent.Event, error) {
	metadataBytes, err := json.Marshal(dto.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
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
