package schema

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/ClickHouse/clickhouse-go/v2/lib/chcol"
	"github.com/google/uuid"
)

type MetadataEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Event holds the schema definition for the Event entity.
type Event struct {
	ent.Schema
}

type ClickhouseJSON map[string]any

func (m *ClickhouseJSON) Scan(src any) error {
	switch v := src.(type) {
	case *chcol.JSON:
		b, err := v.MarshalJSON()
		if err != nil {
			return err
		}
		return json.Unmarshal(b, m)
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	}
	return fmt.Errorf("unsupported type %T", src)
}

func (m *ClickhouseJSON) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Fields of the Event.
func (Event) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Unique().Default(uuid.New).Immutable().StorageKey("eventID"),
		field.UUID("userID", uuid.UUID{}).Immutable().StorageKey("userID"),
		field.String("eventType").NotEmpty().StorageKey("event_type"),
		field.Time("timestamp").Default(time.Now).Immutable(),
		field.String("metadata").StorageKey("metadata").Optional(),
	}
}

// Edges of the Event.
func (Event) Edges() []ent.Edge {
	return nil
}
