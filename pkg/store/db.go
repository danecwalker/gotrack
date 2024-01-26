package store

import "github.com/danecwalker/analytics/pkg/event"

type DBClient interface {
	InsertEvent(event *event.WEvent) error
	InsertSession(session *event.Session) error
	GetStats() (*Stats, error)

	Close() error
}
