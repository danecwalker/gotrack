package event

import (
	"net/url"
	"strings"
)

type UTM struct {
	Source   string
	Medium   string
	Campaign string
	Term     string
	Content  string
}

type WEvent struct {
	SessionID string
	EventName string
	Url       string
	Referrer  string
	Props     map[string]interface{}
	Revenue   map[string]interface{}
	UTM       *UTM
}

func NewWEvent(session_id string) *WEvent {
	return &WEvent{
		SessionID: session_id,
		Props:     make(map[string]interface{}),
		Revenue:   make(map[string]interface{}),
	}
}

func (e *WEvent) Parse(ev *Event) error {
	e.EventName = strings.TrimSpace(ev.EventName)
	e.Props = ev.Props
	e.Revenue = ev.Revenue

	location, err := url.Parse(strings.TrimSpace(ev.Url))
	if err != nil {
		return err
	}
	utm_source := location.Query().Get("utm_source")
	utm_medium := location.Query().Get("utm_medium")
	utm_campaign := location.Query().Get("utm_campaign")
	utm_term := location.Query().Get("utm_term")
	utm_content := location.Query().Get("utm_content")

	ref := location.Query().Get("ref")

	// ignore all utm params if utm_source is empty
	if utm_source != "" {
		e.UTM = &UTM{}
		e.UTM.Source = utm_source
		if utm_medium != "" {
			e.UTM.Medium = utm_medium
		}
		if utm_campaign != "" {
			e.UTM.Campaign = utm_campaign
		}
		if utm_term != "" {
			e.UTM.Term = utm_term
		}
		if utm_content != "" {
			e.UTM.Content = utm_content
		}
	}

	if ref != "" {
		e.Referrer = ref
	} else if ev.Referrer != "" {
		refer, err := url.Parse(ev.Referrer)
		if err != nil {
			return err
		}

		e.Referrer = refer.Scheme + "://" + refer.Host + refer.Path
	}

	e.Url = location.Scheme + "://" + location.Host + location.Path

	return nil
}
