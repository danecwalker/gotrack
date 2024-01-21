package store

import "github.com/danecwalker/analytics/pkg/event"

type inMemoryStore struct {
}

func NewInMemoryStore() Store {
	return &inMemoryStore{}
}

func (s *inMemoryStore) Save(event event.Event) error {
	return nil
}
