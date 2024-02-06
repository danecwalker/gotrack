-- name: GetSession :one
SELECT * FROM sessions
WHERE id = ? LIMIT 1;

-- name: CreateSession :exec
INSERT INTO sessions (id, language, country, browser, os, screen_type, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(id) DO NOTHING;

-- name: CreateEvent :exec
INSERT INTO events (session_id, event_name, url, referrer, utm_source, utm_medium, utm_campaign, utm_term, utm_content, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
