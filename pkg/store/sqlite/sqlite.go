package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"os"
	"time"

	"github.com/danecwalker/gotrack/pkg/event"
	"github.com/danecwalker/gotrack/pkg/store"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zerologadapter"
)

//go:embed schema.sql
var ddl string

type Sqlite struct {
	path string
	ctx  context.Context
	q    *Queries
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
	s.ctx = context.Background()
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
	loggerAdapter := zerologadapter.New(zerolog.New(os.Stdout))
	sq = sqldblogger.OpenDriver(s.path, sq.Driver(), loggerAdapter /*, using_default_options*/) // db is STILL *sql.DB

	if _, err := sq.ExecContext(s.ctx, ddl); err != nil {
		return err
	}

	queries := New(sq)

	s.q = queries

	return nil
}

func (s *Sqlite) InsertEvent(ev *event.WEvent) error {
	var err error
	if ev.UTM == nil {
		err = s.q.CreateEvent(s.ctx, CreateEventParams{
			SessionID:   ev.SessionID,
			EventName:   ev.EventName,
			Url:         ev.Url,
			Referrer:    sql.NullString{Valid: false},
			UtmSource:   sql.NullString{Valid: false},
			UtmMedium:   sql.NullString{Valid: false},
			UtmCampaign: sql.NullString{Valid: false},
			UtmTerm:     sql.NullString{Valid: false},
			UtmContent:  sql.NullString{Valid: false},
			CreatedAt:   time.Now().UTC(),
		})
	} else {
		err = s.q.CreateEvent(s.ctx, CreateEventParams{
			SessionID:   ev.SessionID,
			EventName:   ev.EventName,
			Url:         ev.Url,
			Referrer:    sql.NullString{String: ev.Referrer, Valid: true},
			UtmSource:   sql.NullString{String: ev.UTM.Source, Valid: true},
			UtmMedium:   sql.NullString{String: ev.UTM.Medium, Valid: true},
			UtmCampaign: sql.NullString{String: ev.UTM.Campaign, Valid: true},
			UtmTerm:     sql.NullString{String: ev.UTM.Term, Valid: true},
			UtmContent:  sql.NullString{String: ev.UTM.Content, Valid: true},
			CreatedAt:   time.Now().UTC(),
		})
	}

	// get event id
	if err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) InsertSession(session *event.Session) error {

	err := s.q.CreateSession(s.ctx, CreateSessionParams{
		ID:         session.SessionID,
		Language:   sql.NullString{String: session.Language, Valid: true},
		Country:    sql.NullString{String: session.Country, Valid: true},
		Browser:    sql.NullString{String: session.Browser, Valid: true},
		Os:         sql.NullString{String: session.Os, Valid: true},
		ScreenType: sql.NullString{String: string(session.ScreenType), Valid: true},
		CreatedAt:  time.Now().UTC(),
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) GetStats(from time.Time, to time.Time) (*store.Stats, error) {
	stats := &store.Stats{}

	st, err := s.q.GetStats(s.ctx, GetStatsParams{
		SessionTimeout: event.SessionTimeout,
		From:           from,
		To:             to,
	})

	if err != nil {
		return nil, err
	}

	if st.PageViews.Valid {
		stats.PageViews = int(st.PageViews.Int64)
	}

	if st.UniqueVisitors.Valid {
		stats.Visitors = int(st.UniqueVisitors.Int64)
	}

	if st.Bounces.Valid {
		stats.Bounces = int(st.Bounces.Int64)
	}

	if st.AverageSessionLength.Valid {
		stats.AverageSessionLength = int(st.AverageSessionLength.Int64)
	}

	return stats, nil
}

func (s *Sqlite) GetViewsAndVisits(period string, from time.Time, to time.Time) (*store.GraphStats, error) {
	graph := &store.GraphStats{
		Period: period,
	}

	time_fmt := "%Y-%m-%d"
	if period == "24h" {
		time_fmt = "%Y-%m-%d %H:00:00.000000+00:00"
	}

	res, err := s.q.GetViewsAndVisits(s.ctx, GetGraphParams{
		SessionTimeout: event.SessionTimeout,
		From:           from,
		To:             to,
		Format:         time_fmt,
	})

	if err != nil {
		return nil, err
	}

	size := 30
	switch period {
	case "24h":
		size = 24
	case "7d":
		size = 7
	case "30d":
		size = 30
	}

	graph.PageViews = make([]*store.Coord, size)
	graph.Visitors = make([]*store.Coord, size)

	var dCount int = 0
	for i := 0; i < size; i++ {
		graph.PageViews[i] = &store.Coord{}
		graph.Visitors[i] = &store.Coord{}

		var t time.Time
		var format string
		switch period {
		case "24h":
			format = "2006-01-02 15:04 +0000 UTC"
			t = to.Add(-1 * time.Hour)
		default:
			format = "2006-01-02 00:00 +0000 UTC"
			t = to.AddDate(0, 0, -1)
		}

		t = t.Add(-parsePeriod(period) * time.Duration(i))

		graph.PageViews[i].X = t.Format(format)
		graph.Visitors[i].X = t.Format(format)

		if dCount < len(res) && res[dCount].Time.Compare(t) >= 0 {
			if res[dCount].Views.Valid {
				graph.PageViews[i].Y = int(res[dCount].Views.Int64)
			} else {
				graph.PageViews[i].Y = 0
			}
			if res[dCount].Visits.Valid {
				graph.Visitors[i].Y = int(res[dCount].Visits.Int64)
			} else {
				graph.Visitors[i].Y = 0
			}
			dCount++
		} else {
			graph.PageViews[i].Y = 0
			graph.Visitors[i].Y = 0
		}
	}

	return graph, nil
}

func parsePeriod(p string) time.Duration {
	switch p {
	case "24h":
		return time.Hour
	default:
		return 24 * time.Hour
	}
}
