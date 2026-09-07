ALTER TABLE comments
  ADD COLUMN IF NOT EXISTS parent_id BIGINT REFERENCES comments(id) ON DELETE CASCADE,
  ADD COLUMN IF NOT EXISTS reply_to_comment_id BIGINT REFERENCES comments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_comments_post_parent_created_at
  ON comments (post_id, parent_id, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_comments_parent_created_at
  ON comments (parent_id, created_at ASC, id ASC);

CREATE INDEX IF NOT EXISTS idx_comments_reply_to_comment_id
  ON comments (reply_to_comment_id);
