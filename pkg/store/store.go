package store

import "github.com/danecwalker/analytics/pkg/event"

type Store interface {
	// Save saves the event.
	Save(event event.Event) error
	// AverageRevenue returns the average revenue for the given domain.
	AverageRevenue() (float64, error)
	// BounceRate returns the percentage of sessions that are bounces for the given domain. A bounce is defined as a session with only one page view.
	BounceRate() (float64, error)
	// TimeOnPage returns the average time on page for the given domain. It is calculated as the difference between when a person lands on a page and when they leave the page.
	VisitDuration() (float64, error)
}
