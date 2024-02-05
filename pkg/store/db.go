package store

import (
	"time"

	"github.com/danecwalker/gotrack/pkg/event"
)

type DBClient interface {
	InsertSession(session *event.Session) error
	InsertEvent(event *event.WEvent) error

	GetStats(from time.Time, to time.Time) (*Stats, error)
	GetViewsAndVisits(period string, from time.Time, to time.Time) (*GraphStats, error)

	// Close() error
}
