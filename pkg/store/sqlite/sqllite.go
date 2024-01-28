package sqlite

import (
	"context"
	"database/sql"
	"os"
	"time"

	"github.com/danecwalker/analytics/pkg/event"
	"github.com/danecwalker/analytics/pkg/store"
	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/extra/bundebug"
)

type Sqlite struct {
	path string
	db   *bun.DB
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

	s.db = bun.NewDB(sq, sqlitedialect.New())
	s.db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("bundebug"),
	))

	s.db.NewCreateTable().Model((*store.Session)(nil)).IfNotExists().Exec(context.Background())
	s.db.NewCreateTable().Model((*store.Event)(nil)).IfNotExists().Exec(context.Background())
	s.db.NewCreateTable().Model((*store.Prop)(nil)).IfNotExists().Exec(context.Background())
	s.db.NewCreateTable().Model((*store.Revenue)(nil)).IfNotExists().Exec(context.Background())

	return nil
}

func (s *Sqlite) InsertEvent(ev *event.WEvent) error {
	if ev.UTM == nil {
		ev.UTM = new(event.UTM)
	}

	_, err := s.db.NewInsert().Model(&store.Event{
		SessionID:   ev.SessionID,
		EventName:   ev.EventName,
		Url:         ev.Url,
		Referrer:    ev.Referrer,
		UTMSource:   ev.UTM.Source,
		UTMMedium:   ev.UTM.Medium,
		UTMCampaign: ev.UTM.Campaign,
		UTMTerm:     ev.UTM.Term,
		UTMContent:  ev.UTM.Content,
		CreatedAt:   time.Now(),
	}).Exec(context.Background())

	// get event id
	if err != nil {
		return err
	}

	return nil
}

func (s *Sqlite) InsertSession(session *event.Session) error {
	_, err := s.db.NewInsert().Model(&store.Session{
		ID:         session.SessionID,
		Language:   session.Language,
		Country:    session.Country,
		Browser:    session.Browser,
		Os:         session.Os,
		ScreenType: session.ScreenType,
		CreatedAt:  time.Now(),
	}).Exec(context.Background())

	if err != nil {
		if err.Error() == "UNIQUE constraint failed: sessions.id" {
			return nil
		}
		return err
	}

	return nil
}

func (s *Sqlite) GetStats(from time.Time, to time.Time) (*store.Stats, error) {
	stats := &store.Stats{}

	q1 := s.db.NewSelect().
		Model((*store.Event)(nil)).
		ColumnExpr("session_id").
		ColumnExpr("created_at").
		ColumnExpr("event_name").
		ColumnExpr("LAG(created_at) OVER (PARTITION BY session_id ORDER BY created_at) AS prev_timestamp").
		ColumnExpr("strftime('%s', created_at) - LAG(strftime('%s', created_at)) OVER (PARTITION BY session_id ORDER BY created_at) as difference").
		ColumnExpr("CASE WHEN LAG(created_at) OVER (PARTITION BY session_id ORDER BY created_at) IS NULL OR strftime('%s', created_at) - LAG(strftime('%s', created_at)) OVER (PARTITION BY session_id ORDER BY created_at) > ? THEN 1 ELSE 0 END AS new_session_flag", event.SessionTimeout)

	q2 := s.db.NewSelect().
		Column("session_id").
		Column("event_name").
		Column("created_at").
		ColumnExpr("SUM(new_session_flag) OVER (ORDER BY session_id, created_at) AS session_group").
		Table("SessionGroups")

	q3 := s.db.NewSelect().
		Column("session_id").
		Column("session_group").
		ColumnExpr("count(*) as event_count").
		ColumnExpr("SUM(CASE WHEN event_name = 'pageview' THEN 1 ELSE 0 END) AS pageview_count").
		ColumnExpr("min(created_at) as min_time").
		ColumnExpr("max(created_at) as max_time").
		TableExpr("(?)", q2).
		GroupExpr("?, ?", 1, 2)

	err := s.db.NewSelect().
		With("SessionGroups", q1).
		ColumnExpr("SUM(t.pageview_count) AS pageviews").
		ColumnExpr("COUNT(distinct t.session_id) AS unique_visitors").
		ColumnExpr("SUM(CASE WHEN t.event_count = 1 THEN 1 ELSE 0 END) AS bounces").
		ColumnExpr("SUM(strftime('%s', t.max_time) - strftime('%s', t.min_time)) / COUNT(distinct t.session_group) AS average_session_length").
		TableExpr("(?) AS t", q3).
		Where("t.min_time BETWEEN ? AND ?", from, to).
		Scan(context.Background(), stats)

	if err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *Sqlite) GetViewsAndVisits(period string, from time.Time, to time.Time) (*store.GraphStats, error) {
	graph := &store.GraphStats{}

	dest := make([]struct {
		Views  int       `bun:"views"`
		Visits int       `bun:"visits"`
		Time   time.Time `bun:"time"`
	}, 0)

	q1 := s.db.NewSelect().
		Model((*store.Event)(nil)).
		ColumnExpr("session_id").
		ColumnExpr("created_at").
		ColumnExpr("event_name").
		ColumnExpr("LAG(created_at) OVER (PARTITION BY session_id ORDER BY created_at) AS prev_timestamp").
		ColumnExpr("strftime('%s', created_at) - LAG(strftime('%s', created_at)) OVER (PARTITION BY session_id ORDER BY created_at) as difference").
		ColumnExpr("CASE WHEN LAG(created_at) OVER (PARTITION BY session_id ORDER BY created_at) IS NULL OR strftime('%s', created_at) - LAG(strftime('%s', created_at)) OVER (PARTITION BY session_id ORDER BY created_at) > ? THEN 1 ELSE 0 END AS new_session_flag", event.SessionTimeout)

	q2 := s.db.NewSelect().
		Column("session_id").
		Column("session_group").
		ColumnExpr("count(*) as event_count").
		ColumnExpr("SUM(CASE WHEN event_name = 'pageview' THEN 1 ELSE 0 END) AS pageview_count").
		ColumnExpr("min(created_at) as min_time").
		ColumnExpr("max(created_at) as max_time").
		TableExpr("(?)", q1).
		GroupExpr("?, ?", 1, 2)

	err := s.db.NewSelect().
		ColumnExpr("COUNT(session_id) as visits").
		ColumnExpr("SUM(pageview_count) as views").
		ColumnExpr("strftime('%Y-%m-%d', min_time) as time").
		TableExpr("(?) as t", q2).
		Where("t.min_time BETWEEN ? AND ?", from, to).
		Group("time").
		Order("time", "DESC").
		Scan(context.Background(), &dest)

	if err != nil {
		return nil, err
	}

	size := 30
	switch period {
	case "day":
		size = 24
	case "7d":
		size = 7
	case "30d":
		size = 30
	}

	graph.PageViews = make([]*store.Coord, size)
	graph.Visitors = make([]*store.Coord, size)

	for i := 0; i < size; i++ {
		graph.PageViews[i] = &store.Coord{
			X: time.Now().AddDate(0, 0, -i).Format("2006-01-02"),
			Y: dataOrDefaultViews(dest, i),
		}
		graph.Visitors[i] = &store.Coord{
			X: time.Now().AddDate(0, 0, -i).Format("2006-01-02"),
			Y: dataOrDefaultVisits(dest, i),
		}
	}

	return graph, nil
}

func dataOrDefaultViews(data []struct {
	Views  int       `bun:"views"`
	Visits int       `bun:"visits"`
	Time   time.Time `bun:"time"`
}, i int) int {
	if i < len(data) {
		return data[i].Views
	}
	return 0
}

func dataOrDefaultVisits(data []struct {
	Views  int       `bun:"views"`
	Visits int       `bun:"visits"`
	Time   time.Time `bun:"time"`
}, i int) int {
	if i < len(data) {
		return data[i].Visits
	}
	return 0
}
