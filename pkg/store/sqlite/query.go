package sqlite

import (
	"context"
	"database/sql"
	"time"
)

const getStats = `-- name: GetStats :one
WITH cte_sessions AS (
  SELECT
    session_id,
    created_at,
    event_name,
    LAG(created_at) OVER (PARTITION BY session_id ORDER BY created_at) AS prev_timestamp,
    strftime('%s', created_at) - LAG(strftime('%s', created_at)) OVER (PARTITION BY session_id ORDER BY created_at) AS difference,
    CASE WHEN LAG(created_at) OVER (PARTITION BY session_id ORDER BY created_at) IS NULL OR strftime('%s', created_at) - LAG(strftime('%s', created_at)) OVER (PARTITION BY session_id ORDER BY created_at) > ? THEN 1 ELSE 0 END AS new_session_flag
  FROM
    events
)
SELECT
	SUM(t.pageview_count) AS pageviews,
	COUNT(distinct t.session_id) AS unique_visitors,
	SUM(CASE WHEN t.event_count = 1 THEN 1 ELSE 0 END) AS bounces,
	SUM(strftime('%s', t.max_time) - strftime('%s', t.min_time)) / COUNT(distinct t.session_group) AS average_session_length
FROM (
	SELECT
    session_id,
    session_group,
    COUNT(*) AS event_count,
    SUM(CASE WHEN event_name = 'pageview' THEN 1 ELSE 0 END) AS pageview_count,
    MIN(created_at) AS min_time,
    MAX(created_at) AS max_time
  FROM (
    SELECT
      session_id,
      event_name,
      created_at,
      SUM(new_session_flag) OVER (ORDER BY session_id, created_at) AS session_group
    FROM
      cte_sessions
  ) AS q GROUP BY 1, 2
) AS t
WHERE t.min_time BETWEEN ? AND ?
`

type GetStatsParams struct {
	SessionTimeout int
	From           time.Time
	To             time.Time
}

type GetStatsResults struct {
	PageViews            sql.NullInt64
	UniqueVisitors       sql.NullInt64
	Bounces              sql.NullInt64
	AverageSessionLength sql.NullInt64
}

func (q *Queries) GetStats(ctx context.Context, arg GetStatsParams) (GetStatsResults, error) {
	row := q.db.QueryRowContext(ctx, getStats,
		arg.SessionTimeout,
		arg.From,
		arg.To,
	)
	var i GetStatsResults
	err := row.Scan(
		&i.PageViews,
		&i.UniqueVisitors,
		&i.Bounces,
		&i.AverageSessionLength,
	)
	return i, err
}

const getViewsAndVisits = `-- name: GetViewsAndVisits :many
WITH "SessionGroups" AS (
	SELECT
		session_id,
		created_at,
		event_name,
		LAG(created_at) OVER (PARTITION BY session_id
	ORDER BY
		created_at) AS prev_timestamp,
		strftime('%s', created_at) - LAG(strftime('%s', created_at)) OVER (PARTITION BY session_id
	ORDER BY
		created_at) as difference,
		CASE
			WHEN LAG(created_at) OVER (PARTITION BY session_id
		ORDER BY
			created_at) IS NULL
			OR strftime('%s', created_at) - LAG(strftime('%s', created_at)) OVER (PARTITION BY session_id
		ORDER BY
			created_at) > ?1 THEN 1
			ELSE 0
		END AS new_session_flag
	FROM
		"events")
	SELECT
		"visits",
		"views",
		t1.time as time
	FROM
		(
		SELECT
			COUNT(event_name) as views,
			strftime(?4, created_at) as time
		FROM
			"events"
		WHERE
			(created_at BETWEEN ?2 AND ?3)
			AND (event_name = 'pageview')
		GROUP BY
			"time"
		ORDER BY
			time DESC) AS t1
	JOIN (
		SELECT
			COUNT(session_id) as visits,
			strftime(?4, min_time) as time
		FROM
			(
			SELECT
				"session_id",
				"session_group",
				min(created_at) as min_time
			FROM
				(
				SELECT
					"session_id",
					"event_name",
					"created_at",
					SUM(new_session_flag) OVER (
					ORDER BY session_id,
					created_at) AS session_group
				FROM
					"SessionGroups")
			GROUP BY
				1,
				2)
		WHERE
			(min_time BETWEEN ?2 AND ?3)
		GROUP BY
			"time"
		ORDER BY
			time DESC) AS t2 ON
		t1.time = t2.time
`

type GetGraphParams struct {
	SessionTimeout int
	From           time.Time
	To             time.Time
	Format         string
}

type GetGraphResults struct {
	Visits sql.NullInt64
	Views  sql.NullInt64
	Time   time.Time
}

func (q *Queries) GetViewsAndVisits(ctx context.Context, args GetGraphParams) ([]GetGraphResults, error) {
	rows, err := q.db.QueryContext(ctx, getViewsAndVisits, args.SessionTimeout, args.From, args.To, args.Format)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetGraphResults
	for rows.Next() {
		var i GetGraphResults
		var t string
		if err := rows.Scan(
			&i.Visits,
			&i.Views,
			&t,
		); err != nil {
			return nil, err
		}
		i.Time, err = time.Parse("2006-01-02 15:04:05.000000+00:00", t)
		if err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
