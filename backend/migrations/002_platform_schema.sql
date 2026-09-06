ALTER TABLE users
  ADD COLUMN IF NOT EXISTS banned_until TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS ban_reason TEXT NOT NULL DEFAULT '';

ALTER TABLE comments
  ADD COLUMN IF NOT EXISTS likes BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS games (
  id BIGSERIAL PRIMARY KEY,
  owner_id BIGINT NOT NULL REFERENCES users(id),
  title TEXT NOT NULL,
  summary TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  engine TEXT NOT NULL DEFAULT '',
  genre TEXT NOT NULL DEFAULT '',
  license TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'reviewing',
  cover_url TEXT NOT NULL DEFAULT '',
  play_url TEXT NOT NULL DEFAULT '',
  source_url TEXT NOT NULL DEFAULT '',
  plays BIGINT NOT NULL DEFAULT 0,
  likes BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS game_likes (
  game_id BIGINT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (game_id, user_id)
);

CREATE TABLE IF NOT EXISTS forum_bars (
  id BIGSERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS forum_posts (
  id BIGSERIAL PRIMARY KEY,
  bar_id BIGINT NOT NULL REFERENCES forum_bars(id),
  author_id BIGINT NOT NULL REFERENCES users(id),
  title TEXT NOT NULL,
  category TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL,
  tags JSONB NOT NULL DEFAULT '[]'::jsonb,
  status TEXT NOT NULL DEFAULT 'pending',
  likes BIGINT NOT NULL DEFAULT 0,
  views BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS forum_post_likes (
  post_id BIGINT NOT NULL REFERENCES forum_posts(id) ON DELETE CASCADE,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (post_id, user_id)
);

CREATE TABLE IF NOT EXISTS repositories (
  id BIGSERIAL PRIMARY KEY,
  owner_id BIGINT REFERENCES users(id),
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  language TEXT NOT NULL DEFAULT '',
  license TEXT NOT NULL DEFAULT '',
  download_url TEXT NOT NULL DEFAULT '',
  downloads BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

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
