package store

import "github.com/danecwalker/analytics/pkg/event"

type Store interface {
	// Save saves the event.
	Save(event event.Event) error
}
