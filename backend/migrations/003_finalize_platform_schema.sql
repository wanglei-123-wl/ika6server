ALTER TABLE games
  ADD COLUMN IF NOT EXISTS description TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS cover_url TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS play_url TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS source_url TEXT NOT NULL DEFAULT '';

ALTER TABLE forum_posts
  ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]'::jsonb;

ALTER TABLE comments
  ADD COLUMN IF NOT EXISTS likes BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS game_files (
  id BIGSERIAL PRIMARY KEY,
  game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  kind TEXT NOT NULL,
  original_name TEXT NOT NULL,
  stored_name TEXT NOT NULL,
  size BIGINT NOT NULL,
  sha256 TEXT NOT NULL DEFAULT '',
  download_url TEXT NOT NULL DEFAULT '',
  downloads BIGINT NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'reviewing',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (game_id, kind)
);

CREATE TABLE IF NOT EXISTS developer_docs (
  id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
  title TEXT NOT NULL,
  subtitle TEXT NOT NULL DEFAULT '',
  steps JSONB NOT NULL DEFAULT '[]'::jsonb,
  code TEXT NOT NULL DEFAULT '',
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO forum_bars (id, name, description)
VALUES (1, '独立游戏吧', '独立游戏开发交流 · 开源互助 · 作品发布')
ON CONFLICT (id) DO NOTHING;

INSERT INTO developer_docs (id, title, subtitle, steps, code)
VALUES (
  1,
  '快速开始',
  '从 0 到 1，把游戏发布到平台',
  '["注册并登录", "上传游戏文件", "写介绍", "一键发布"]'::jsonb,
  'npm install...'
)
ON CONFLICT (id) DO NOTHING;
