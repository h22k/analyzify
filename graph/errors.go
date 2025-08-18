package graph

import "errors"

var (
	EventCreateErr = errors.New("failed to create event")
	EventFetchErr  = errors.New("failed to fetch events")
)
