package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/danecwalker/gotrack/pkg/event"
	"github.com/danecwalker/gotrack/pkg/store"
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
		CreatedAt:   time.Now().UTC(),
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
		CreatedAt:  time.Now().UTC(),
	}).
		On("CONFLICT (id) DO NOTHING").
		Exec(context.Background())

	if err != nil {
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
	graph := &store.GraphStats{
		Period: period,
	}

	dest := make([]struct {
		Views  int       `bun:"views"`
		Visits int       `bun:"visits"`
		Time   time.Time `bun:"time"`
	}, 0)

	time_fmt := "%Y-%m-%d"
	if period == "24h" {
		time_fmt = "%Y-%m-%d %H:00:00.000000+00:00"
	}

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
		ColumnExpr("min(created_at) as min_time").
		TableExpr("(?)", q2).
		GroupExpr("?, ?", 1, 2)

	q4 := s.db.NewSelect().
		ColumnExpr("COUNT(session_id) as visits").
		ColumnExpr("strftime('"+time_fmt+"', min_time) as time").
		TableExpr("(?)", q3).
		Where("min_time BETWEEN ? AND ?", from, to).
		Group("time").
		OrderExpr("time DESC")

	q5 := s.db.NewSelect().
		Model((*store.Event)(nil)).
		ColumnExpr("COUNT(event_name) as views").
		ColumnExpr("strftime('"+time_fmt+"', created_at) as time").
		Where("created_at BETWEEN ? AND ?", from, to).
		Where("event_name = 'pageview'").
		Group("time").
		OrderExpr("time DESC")

	err := s.db.NewSelect().
		With("SessionGroups", q1).
		Column("visits").
		Column("views").
		ColumnExpr("t1.time as time").
		TableExpr("(?) AS t1", q5).
		Join("JOIN (?) AS t2 ON t1.time = t2.time", q4).
		Scan(context.Background(), &dest)

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

		fmt.Println(dest[0].Time)

		if dCount < len(dest) && dest[dCount].Time.Compare(t) >= 0 {
			graph.PageViews[i].Y = dest[dCount].Views
			graph.Visitors[i].Y = dest[dCount].Visits
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
