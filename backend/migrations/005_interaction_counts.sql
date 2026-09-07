CREATE TABLE IF NOT EXISTS comment_likes (
  comment_id BIGINT NOT NULL REFERENCES comments(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (comment_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_comments_post_id
  ON comments (post_id);

CREATE INDEX IF NOT EXISTS idx_forum_posts_status_created_at
  ON forum_posts (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_game_files_game_kind
  ON game_files (game_id, kind);
