package store

import (
	"sort"

	"github.com/danecwalker/analytics/pkg/event"
)

type inMemoryStore struct {
	events []event.Event
}

func NewInMemoryStore() Store {
	return &inMemoryStore{
		events: make([]event.Event, 0),
	}
}

func (s *inMemoryStore) Save(event event.Event) error {
	s.events = append(s.events, event)
	return nil
}

func (s *inMemoryStore) AverageRevenue() (float64, error) {
	total := 0.0
	count := 0
	for _, e := range s.events {
		if e.Revenue != nil {
			total += e.Revenue.Amount
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	return total / float64(count), nil
}

func (s *inMemoryStore) BounceRate() (float64, error) {
	views := make(map[string]int)

	for _, e := range s.events {
		if e.Name == "pageview" {
			views[e.Session]++
		}
	}

	bounces := 0
	for _, v := range views {
		if v == 1 {
			bounces++
		}
	}

	return float64(bounces) / float64(len(views)), nil
}

func (s *inMemoryStore) VisitDuration() (float64, error) {
	views := make(map[string][]event.Event)

	for _, e := range s.events {
		views[e.Session] = append(views[e.Session], e)
	}

	total := 0.0
	for _, v := range views {
		if len(v) > 1 {
			sort.Slice(v, func(i, j int) bool {
				return v[i].Timestamp.Before(v[j].Timestamp)
			})
			total += v[len(v)-1].Timestamp.Sub(v[0].Timestamp).Seconds()
		}
	}

	if len(views) == 0 {
		return 0, nil
	}
	return total / float64(len(views)), nil
}
