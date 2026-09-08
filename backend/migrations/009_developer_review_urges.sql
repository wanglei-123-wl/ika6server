CREATE TABLE IF NOT EXISTS developer_review_urges (
  id BIGSERIAL PRIMARY KEY,
  game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  urged_on DATE NOT NULL DEFAULT CURRENT_DATE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (game_id, user_id, urged_on)
);

CREATE INDEX IF NOT EXISTS idx_developer_review_urges_game_created_at
  ON developer_review_urges (game_id, created_at DESC);

