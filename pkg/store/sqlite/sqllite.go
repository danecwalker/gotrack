package sqlite

import (
	"context"
	"database/sql"
	"os"

	"github.com/danecwalker/analytics/pkg/event"
	"github.com/danecwalker/analytics/pkg/store"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	path string
	conn *sql.DB
}

func NewSqlite(path string) (store.DBClient, error) {
	s := &Sqlite{
		path: path,
	}

	if err := s.init(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Sqlite) init() error {
	// check if db exists and create if not
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		if _, err := os.Create(s.path); err != nil {
			return err
		}
	}

	// open db connection ensure WAL mode is enabled and it is not locked
	sq, err := sql.Open("sqlite3", s.path) //+"?_journal_mode=WAL&_busy_timeout=5000&cache=shared&rwc=3"
	if err != nil {
		return err
	}

	s.conn = sq

	// create tables if they don't exist
	if err := s.createTables(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) createTables() error {
	// create tables
	tx, err := s.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		event_name TEXT NOT NULL,
		url TEXT NOT NULL,
		referrer TEXT,
		utm_source TEXT,
		utm_medium TEXT,
		utm_campaign TEXT,
		utm_term TEXT,
		utm_content TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		foreign key(session_id) references sessions(id)
	);`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL UNIQUE,
		device_type TEXT NOT NULL,
		language TEXT,
		country TEXT,
		browser TEXT,
		os TEXT,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS props (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL,
		prop_name TEXT NOT NULL,
		prop_value TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		foreign key(event_id) references events(id)
	);`); err != nil {
		return err
	}

	if _, err := tx.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS revenue (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		event_id INTEGER NOT NULL,
		revenue_name TEXT NOT NULL,
		revenue_value TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		foreign key(event_id) references events(id)
	);`); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) InsertEvent(event *event.WEvent) error {
	// insert event
	tx, err := s.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var r sql.Result
	if event.UTM != nil {
		r, err = tx.ExecContext(context.Background(), `INSERT INTO events (
		session_id,
		event_name,
		url,
		referrer,
		utm_source,
		utm_medium,
		utm_campaign,
		utm_term,
		utm_content,
		created_at,
		updated_at
	) VALUES (
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		?,
		datetime('now'),
		datetime('now')
	);`, event.SessionID, event.EventName, event.Url, event.Referrer, event.UTM.Source, event.UTM.Medium, event.UTM.Campaign, event.UTM.Term, event.UTM.Content)
		if err != nil {
			return err
		}
	} else {
		r, err = tx.ExecContext(context.Background(), `INSERT INTO events (
		session_id,
		event_name,
		url,
		referrer,
		created_at,
		updated_at
	) VALUES (
		?,
		?,
		?,
		?,
		datetime('now'),
		datetime('now')
	);`, event.SessionID, event.EventName, event.Url, event.Referrer)
		if err != nil {
			return err
		}
	}

	// get event id
	eventID, err := r.LastInsertId()
	if err != nil {
		return err
	}

	// insert props
	for k, v := range event.Props {
		if _, err := tx.ExecContext(context.Background(), `INSERT INTO props (
			event_id,
			prop_name,
			prop_value,
			created_at,
			updated_at
		) VALUES (
			?,
			?,
			?,
			datetime('now'),
			datetime('now')
		);`, eventID, k, v); err != nil {
			return err
		}
	}

	// insert revenue
	for k, v := range event.Revenue {
		if _, err := tx.ExecContext(context.Background(), `INSERT INTO revenue (
			event_id,
			revenue_name,
			revenue_value,
			created_at,
			updated_at
		) VALUES (
			?,
			?,
			?,
			datetime('now'),
			datetime('now')
		);`, eventID, k, v); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) InsertSession(session *event.Session) error {
	// insert session
	tx, err := s.conn.BeginTx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(context.Background(), `INSERT INTO sessions (
		session_id,
		device_type,
		language,
		country,
		browser,
		os,
		created_at,
		updated_at
	) VALUES (
		?,
		?,
		?,
		?,
		?,
		?,
		datetime('now'),
		datetime('now')
	);`, session.SessionID, session.DeviceType, session.Language, session.Country, session.Browser, session.Os); err != nil {
		if err.Error() != "UNIQUE constraint failed: sessions.session_id" {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) GetStats() (*store.Stats, error) {
	stats := &store.Stats{}

	q, err := s.conn.QueryContext(
		context.Background(),
		`SELECT SUM("t"."pageview_count") AS "pageviews", 
		COUNT(DISTINCT "t"."session_id") AS "unique_visitors", 
		SUM(CASE WHEN "t"."event_count" = 1 THEN 1 ELSE 0 END) AS "bounces",
		SUM(CAST ((JULIANDAY("t"."max_time") - JULIANDAY("t"."min_time")) * 24 * 60 * 60 AS INTEGER)) / COUNT(DISTINCT "t"."session_id") AS "average_session_length"
		FROM (
			SELECT "events"."session_id" AS "session_id",
			COUNT(*) AS "event_count",
			SUM(CASE WHEN "events"."event_name" = 'pageview' THEN 1 ELSE 0 END) AS "pageview_count",
			MIN("events"."created_at") AS "min_time",
			MAX("events"."created_at") AS "max_time"
			FROM "events"
			JOIN "sessions" ON "sessions"."session_id" = "events"."session_id"
			GROUP BY 1
		) AS "t";`)
	if err != nil {
		return nil, err
	}

	if q.Next() {
		if err := q.Scan(&stats.PageViews, &stats.Visitors, &stats.Bounces, &stats.AverageSessionLength); err != nil {
			return nil, err
		}
	}

	return stats, nil
}

func (s *Sqlite) Close() error {
	return s.conn.Close()
}
