CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY NOT NULL UNIQUE,
  language TEXT,
  country TEXT,
  browser TEXT,
  os TEXT,
  screen_type TEXT,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
  session_id TEXT NOT NULL,
  event_name TEXT NOT NULL,
  url TEXT NOT NULL,
  referrer TEXT,
  utm_source TEXT,
  utm_medium TEXT,
  utm_campaign TEXT,
  utm_term TEXT,
  utm_content TEXT,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS props (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
  event_id INTEGER NOT NULL,
  key TEXT NOT NULL,
  value TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS revenues (
  id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL UNIQUE,
  event_id INTEGER NOT NULL,
  key TEXT NOT NULL,
  value TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_event_session_id ON events (session_id);
CREATE INDEX IF NOT EXISTS idx_prop_event_id ON props (event_id);
CREATE INDEX IF NOT EXISTS idx_revenue_event_id ON revenues (event_id);