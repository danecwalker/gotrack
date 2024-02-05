package store

import (
	"time"

	"github.com/danecwalker/gotrack/pkg/event"
	"github.com/uptrace/bun"
)

type Event struct {
	bun.BaseModel `bun:"events,alias:events"`
	ID            int64      `bun:"id,pk,notnull,unique,autoincrement"`
	SessionID     string     `bun:"session_id,notnull"`
	EventName     string     `bun:"event_name,notnull"`
	Url           string     `bun:"url,notnull"`
	Referrer      string     `bun:"referrer"`
	UTMSource     string     `bun:"utm_source"`
	UTMMedium     string     `bun:"utm_medium"`
	UTMCampaign   string     `bun:"utm_campaign"`
	UTMTerm       string     `bun:"utm_term"`
	UTMContent    string     `bun:"utm_content"`
	Props         []*Prop    `bun:"rel:has-many,join:id=event_id"`
	Revenue       []*Revenue `bun:"rel:has-many,join:id=event_id"`
	CreatedAt     time.Time  `bun:"created_at,notnull"`
}

type Session struct {
	bun.BaseModel `bun:"sessions,alias:sessions"`
	ID            string           `bun:"id,pk,notnull,unique"`
	Language      string           `bun:"language"`
	Country       string           `bun:"country"`
	Browser       string           `bun:"browser"`
	Os            string           `bun:"os"`
	ScreenType    event.ScreenType `bun:"screen_type"`
	Events        []*Event         `bun:"rel:has-many,join:id=session_id"`
	CreatedAt     time.Time        `bun:"created_at,notnull"`
}

type Prop struct {
	bun.BaseModel `bun:"props,alias:props"`
	ID            int64     `bun:"id,pk,notnull,unique,autoincrement"`
	EventID       int64     `bun:"event_id,notnull"`
	Key           string    `bun:"key,notnull"`
	Value         string    `bun:"value,notnull"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
}

type Revenue struct {
	bun.BaseModel `bun:"revenue,alias:revenue"`
	ID            int64     `bun:"id,pk,notnull,unique,autoincrement"`
	EventID       int64     `bun:"event_id,notnull"`
	Key           string    `bun:"key,notnull"`
	Value         string    `bun:"value,notnull"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
}
