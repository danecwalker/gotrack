package db

import "github.com/danecwalker/analytics/pkg/event"

type DBClient interface {
	InsertEvent(event *event.Event) error
	InsertSession(session *event.Session) error
}
