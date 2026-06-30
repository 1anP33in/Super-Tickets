CREATE TABLE IF NOT EXISTS guild_configs (
  guild_id TEXT PRIMARY KEY,
  config_json TEXT NOT NULL,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tickets (
  channel_id TEXT PRIMARY KEY,
  guild_id TEXT NOT NULL,
  owner_id TEXT NOT NULL,
  claimed_by TEXT NOT NULL DEFAULT '',
  department TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL,
  last_activity DATETIME NOT NULL,
  closed INTEGER NOT NULL DEFAULT 0,
  close_reason TEXT NOT NULL DEFAULT ''
);
