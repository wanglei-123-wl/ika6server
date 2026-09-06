function pick(source, keys, fallback = undefined) {
  for (const key of keys) {
    if (source?.[key] !== undefined && source?.[key] !== null) return source[key];
  }

  return fallback;
}

function toArray(value) {
  if (Array.isArray(value)) return value;
  if (Array.isArray(value?.items)) return value.items;
  if (Array.isArray(value?.list)) return value.list;
  if (Array.isArray(value?.records)) return value.records;
  if (Array.isArray(value?.rows)) return value.rows;
  if (Array.isArray(value?.data)) return value.data;
  return [];
}

function formatCount(value, fallback = '0') {
  if (value === undefined || value === null || value === '') return fallback;
  if (typeof value === 'string') return value;
  if (value >= 1000) return `${(value / 1000).toFixed(value >= 10000 ? 1 : 0)}k`;
  return String(value);
}

function normalizeTags(value) {
  if (Array.isArray(value)) return value;
  if (typeof value === 'string' && value.trim()) return value.split(',').map((tag) => tag.trim());
  return [];
}

export function normalizePage(raw, itemNormalizer) {
  const items = toArray(raw);

  return {
    page: Number(pick(raw, ['page', 'currentPage', 'pageNo', 'current'], 1)),
    pageSize: Number(pick(raw, ['pageSize', 'size', 'limit', 'perPage'], items.length || 20)),
    total: Number(pick(raw, ['total', 'totalCount', 'count'], items.length)),
    items: items.map(itemNormalizer),
  };
}

export function normalizeUser(raw = {}) {
  const name = pick(raw, ['name', 'nickname', 'username', 'displayName'], '玩家');

  return {
    id: pick(raw, ['id', 'userId', 'user_id'], ''),
    name,
    initial: pick(raw, ['initial'], String(name).charAt(0).toUpperCase() || 'U'),
    avatar: pick(raw, ['avatar', 'avatarUrl', 'avatar_url'], ''),
    level: pick(raw, ['level', 'rank'], 'lv1'),
    role: pick(raw, ['role', 'userRole', 'user_role'], 'user'),
    method: pick(raw, ['method', 'loginMethod', 'login_method'], 'password'),
  };
}

export function normalizeAuthResult(raw = {}) {
  const user = pick(raw, ['user', 'profile', 'account'], raw);

  return {
    token: pick(raw, ['token', 'accessToken', 'access_token', 'jwt'], ''),
    user: normalizeUser(user),
  };
}

export function normalizeGame(raw = {}) {
  const title = pick(raw, ['title', 'name', 'gameName', 'game_name'], '未命名游戏');
  const sourceUrl = pick(raw, ['sourceUrl', 'source_url', 'downloadUrl', 'download_url'], '');

  return {
    id: pick(raw, ['id', 'gameId', 'game_id', 'slug'], title),
    title,
    glyph: pick(raw, ['glyph', 'icon', 'mark'], '◆'),
    cover: Number(pick(raw, ['cover', 'coverIndex', 'cover_index'], 1)),
    coverUrl: pick(raw, ['coverUrl', 'cover_url', 'coverImage'], ''),
    badge: pick(raw, ['badge', 'tag', 'statusBadge'], ''),
    author: pick(raw, ['author', 'authorName', 'author_name', 'publisher'], '@unknown'),
    engine: pick(raw, ['engine', 'engineName', 'engine_name'], '未知引擎'),
    size: pick(raw, ['size', 'fileSize', 'file_size'], '未知大小'),
    plays: formatCount(pick(raw, ['plays', 'playCount', 'play_count', 'views'], 0)),
    likes: pick(raw, ['likes', 'likeCount', 'like_count'], 0),
    liked: Boolean(pick(raw, ['liked', 'isLiked', 'is_liked'], false)),
    hasSource: Boolean(pick(raw, ['hasSource', 'has_source', 'sourceAvailable'], sourceUrl)),
    genre: pick(raw, ['genre', 'category', 'type'], '其他'),
    license: pick(raw, ['license', 'openSourceLicense', 'open_source_license'], '未知协议'),
    status: pick(raw, ['status', 'reviewStatus', 'review_status'], 'published'),
    playUrl: pick(raw, ['playUrl', 'play_url'], ''),
    sourceUrl,
  };
}

export function normalizePost(raw = {}) {
  const name = pick(raw, ['name', 'author', 'authorName', 'author_name'], '匿名玩家');

  return {
    id: pick(raw, ['id', 'postId', 'post_id'], ''),
    ava: pick(raw, ['ava', 'avatarText', 'avatar_text'], String(name).charAt(0).toUpperCase() || 'U'),
    bg: pick(raw, ['bg', 'avatarBg', 'avatar_bg'], 'linear-gradient(135deg,#06B6D4,#3B82F6)'),
    name,
    level: pick(raw, ['level', 'authorLevel', 'author_level'], ''),
    time: pick(raw, ['time', 'createdAt', 'created_at'], ''),
    cat: pick(raw, ['cat', 'category', 'categoryName', 'category_name'], '灌水'),
    tags: normalizeTags(pick(raw, ['tags', 'labels'], [])),
    title: pick(raw, ['title', 'subject'], '无标题帖子'),
    excerpt: pick(raw, ['excerpt', 'summary', 'content'], ''),
    media: pick(raw, ['media', 'mediaText', 'media_text'], ''),
    replies: formatCount(pick(raw, ['replies', 'replyCount', 'reply_count'], 0)),
    views: formatCount(pick(raw, ['views', 'viewCount', 'view_count'], 0)),
    likes: Number(pick(raw, ['likes', 'likeCount', 'like_count'], 0)),
    liked: Boolean(pick(raw, ['liked', 'isLiked', 'is_liked'], false)),
    barId: pick(raw, ['barId', 'bar_id', 'forumId', 'forum_id'], ''),
  };
}

export function normalizeBar(raw = {}) {
  return {
    id: pick(raw, ['id', 'barId', 'bar_id'], ''),
    icon: pick(raw, ['icon', 'glyph'], '◆'),
    name: pick(raw, ['name', 'barName', 'bar_name'], '未命名游戏吧'),
    desc: pick(raw, ['desc', 'description', 'summary'], ''),
    posts: formatCount(pick(raw, ['posts', 'postCount', 'post_count'], 0)),
    members: formatCount(pick(raw, ['members', 'memberCount', 'member_count'], 0)),
    hot: Boolean(pick(raw, ['hot', 'isHot', 'is_hot'], false)),
  };
}

export function normalizeReply(raw = {}) {
  const author = pick(raw, ['author', 'name', 'authorName', 'author_name'], '匿名玩家');

  return {
    id: pick(raw, ['id', 'replyId', 'reply_id'], ''),
    postId: pick(raw, ['postId', 'post_id'], ''),
    author,
    avatarText: pick(raw, ['avatarText', 'avatar_text'], String(author).charAt(0).toUpperCase() || 'U'),
    floor: pick(raw, ['floor', 'floorNo', 'floor_no'], ''),
    content: pick(raw, ['content', 'text'], ''),
    createdAt: pick(raw, ['createdAt', 'created_at', 'time'], ''),
    likes: Number(pick(raw, ['likes', 'likeCount', 'like_count'], 0)),
    liked: Boolean(pick(raw, ['liked', 'isLiked', 'is_liked'], false)),
  };
}

export function normalizeRepo(raw = {}) {
  const name = pick(raw, ['name', 'repoName', 'repo_name', 'projectName'], 'untitled-repo');

  return {
    id: pick(raw, ['id', 'repoId', 'repo_id', 'slug'], name),
    icon: pick(raw, ['icon', 'glyph'], '◆'),
    name,
    desc: pick(raw, ['desc', 'description', 'summary'], ''),
    lang: pick(raw, ['lang', 'language'], 'Unknown'),
    dots: pick(raw, ['dots', 'languageColor', 'language_color'], '#06B6D4'),
    license: pick(raw, ['license'], '未知协议'),
    stars: formatCount(pick(raw, ['stars', 'starCount', 'star_count'], 0)),
    forks: formatCount(pick(raw, ['forks', 'forkCount', 'fork_count'], 0)),
    downloads: formatCount(pick(raw, ['downloads', 'downloadCount', 'download_count'], 0)),
    size: pick(raw, ['size', 'fileSize', 'file_size'], '未知大小'),
    badge: pick(raw, ['badge', 'tag'], ''),
  };
}

export function normalizeHome(raw = {}) {
  const stats = pick(raw, ['stats', 'summary'], {});

  return {
    stats: {
      projects: Number(pick(stats, ['projects', 'projectCount', 'project_count'], 0)),
      plays: Number(pick(stats, ['plays', 'playCount', 'play_count'], 0)),
      contributors: Number(pick(stats, ['contributors', 'contributorCount', 'contributor_count'], 0)),
      price: Number(pick(stats, ['price'], 0)),
    },
    latestGames: toArray(pick(raw, ['latestGames', 'latest_games', 'games'], [])).map(normalizeGame),
    hotPosts: toArray(pick(raw, ['hotPosts', 'hot_posts'], [])).map(normalizePost),
    feedPosts: toArray(pick(raw, ['feedPosts', 'feed_posts', 'posts'], [])).map(normalizePost),
  };
}

export function normalizeDevDocs(raw = {}) {
  return Object.fromEntries(Object.entries(raw).map(([key, doc]) => [
    key,
    {
      title: pick(doc, ['title', 'name'], key),
      sub: pick(doc, ['sub', 'summary', 'description'], ''),
      steps: toArray(pick(doc, ['steps', 'items'], [])),
      code: pick(doc, ['code', 'example'], ''),
    },
  ]));
}

export function normalizeSearch(raw = {}) {
  return {
    games: toArray(pick(raw, ['games', 'gameResults', 'game_results'], [])).map(normalizeGame),
    posts: toArray(pick(raw, ['posts', 'postResults', 'post_results'], [])).map(normalizePost),
    repos: toArray(pick(raw, ['repos', 'repoResults', 'repo_results'], [])).map(normalizeRepo),
  };
}

export function normalizeDownload(raw = {}) {
  return {
    downloadUrl: pick(raw, ['downloadUrl', 'download_url', 'url'], ''),
    size: pick(raw, ['size', 'fileSize', 'file_size'], ''),
  };
}

export function normalizeAdminDashboard(raw = {}) {
  return {
    pendingGames: Number(pick(raw, ['pendingGames', 'pending_games'], 0)),
    pendingPosts: Number(pick(raw, ['pendingPosts', 'pending_posts'], 0)),
    activeUsers: Number(pick(raw, ['activeUsers', 'active_users'], 0)),
    reports: Number(pick(raw, ['reports', 'reportCount', 'report_count'], 0)),
  };
}
