CREATE TABLE IF NOT EXISTS game_play_deployments (
  game_id BIGINT PRIMARY KEY REFERENCES games(id) ON DELETE CASCADE,
  file_id BIGINT REFERENCES game_files(id) ON DELETE SET NULL,
  root_path TEXT NOT NULL,
  entry_path TEXT NOT NULL,
  public_url TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'reviewing',
  error_message TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_game_play_deployments_status
  ON game_play_deployments (status);
