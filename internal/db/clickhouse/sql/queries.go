package sql

import _ "embed"

var (
	//go:embed create_events_table.sql
	CreateEventsTableQuery string
)
