import { request, resolveApiAssetUrl } from './http';

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
  if (Array.isArray(value?.data)) return value.data;
  return [];
}

function formatCount(value, fallback = '0') {
  if (value === undefined || value === null || value === '') return fallback;
  if (typeof value === 'string') return value;
  if (value >= 1000) return `${(value / 1000).toFixed(value >= 10000 ? 1 : 0)}k`;
  return String(value);
}

function normalizeProfile(raw = {}, fallbackUser = null) {
  const name = pick(raw, ['name', 'nickname', 'username', 'displayName'], fallbackUser?.name || '开发者');

  return {
    id: pick(raw, ['id', 'userId', 'user_id'], fallbackUser?.id || ''),
    name,
    initial: pick(raw, ['initial', 'avatarText', 'avatar_text'], fallbackUser?.initial || String(name).charAt(0).toUpperCase() || 'U'),
    avatarUrl: resolveApiAssetUrl(pick(raw, ['avatarUrl', 'avatar_url', 'avatar'], fallbackUser?.avatar || '')),
    level: pick(raw, ['level', 'rank'], fallbackUser?.level || 'lv1'),
    verified: Boolean(pick(raw, ['verified', 'isVerified', 'is_verified'], true)),
    bio: pick(raw, ['bio', 'description', 'summary'], ''),
    location: pick(raw, ['location', 'city'], ''),
    joinedAt: pick(raw, ['joinedAt', 'joined_at', 'createdAt', 'created_at'], ''),
    engines: toArray(pick(raw, ['engines', 'engineTags', 'engine_tags'], [])),
  };
}

function normalizeStats(raw = {}) {
  const following = pick(raw, ['following', 'followings', 'followingCount', 'following_count'], null);
  const followers = pick(raw, ['followers', 'followerCount', 'follower_count'], null);
  const sponsorIncome = pick(raw, ['sponsorIncome', 'sponsor_income', 'income', 'revenue'], null);
  const weeklyNewFollowers = pick(raw, ['weeklyNewFollowers', 'weekly_new_followers', 'newFollowers', 'new_followers'], null);
  const trends = [
    pick(raw, ['playTrend', 'play_trend'], ''),
    pick(raw, ['downloadTrend', 'download_trend'], ''),
    pick(raw, ['likeTrend', 'like_trend'], ''),
    pick(raw, ['incomeTrend', 'income_trend'], ''),
  ];
  const advancedStatsAvailable = [following, followers, sponsorIncome, weeklyNewFollowers]
    .some((value) => value !== null && value !== undefined && value !== '' && Number(value) !== 0)
    || trends.some(Boolean);
  const advancedFallback = advancedStatsAvailable ? '0' : '—';

  return {
    following: formatCount(following, advancedFallback),
    followers: formatCount(followers, advancedFallback),
    totalLikes: formatCount(pick(raw, ['totalLikes', 'total_likes', 'likes', 'likeCount', 'like_count'], 0)),
    works: Number(pick(raw, ['works', 'workCount', 'work_count', 'games', 'gameCount', 'game_count'], 0)),
    totalPlays: formatCount(pick(raw, ['totalPlays', 'total_plays', 'plays', 'playCount', 'play_count'], 0)),
    totalDownloads: formatCount(pick(raw, ['totalDownloads', 'total_downloads', 'downloads', 'downloadCount', 'download_count'], 0)),
    sponsorIncome: formatCount(sponsorIncome, advancedFallback),
    weeklyNewFollowers: formatCount(weeklyNewFollowers, advancedFallback),
    playTrend: trends[0],
    downloadTrend: trends[1],
    likeTrend: trends[2],
    incomeTrend: trends[3],
  };
}

function normalizeWork(raw = {}) {
  const title = pick(raw, ['title', 'name', 'gameName', 'game_name'], '未命名作品');
  const status = pick(raw, ['status', 'reviewStatus', 'review_status'], 'draft');

  return {
    id: pick(raw, ['id', 'gameId', 'game_id'], title),
    title,
    glyph: pick(raw, ['glyph', 'icon', 'mark'], '◆'),
    cover: Number(pick(raw, ['cover', 'coverIndex', 'cover_index'], 1)),
    coverUrl: resolveApiAssetUrl(pick(raw, ['coverUrl', 'cover_url', 'coverImage'], '')),
    status,
    genre: pick(raw, ['genre', 'category', 'type'], '其他'),
    engine: pick(raw, ['engine', 'engineName', 'engine_name'], '未知引擎'),
    version: pick(raw, ['version', 'ver', 'releaseVersion', 'release_version'], ''),
    plays: formatCount(pick(raw, ['plays', 'playCount', 'play_count'], 0)),
    downloads: formatCount(pick(raw, ['downloads', 'downloadCount', 'download_count'], 0)),
    likes: formatCount(pick(raw, ['likes', 'likeCount', 'like_count'], 0)),
    progress: Number(pick(raw, ['reviewProgress', 'review_progress', 'progress'], 0)),
    eta: pick(raw, ['reviewMessage', 'review_message', 'eta', 'message'], ''),
    reason: pick(raw, ['rejectReason', 'reject_reason', 'reason'], ''),
    date: pick(raw, ['updatedAt', 'updated_at', 'createdAt', 'created_at', 'date'], ''),
    playUrl: pick(raw, ['playUrl', 'play_url'], ''),
    sourceUrl: pick(raw, ['sourceUrl', 'source_url'], ''),
  };
}

function normalizeDeveloperCenter(raw = {}, fallbackUser = null) {
  const profileSource = pick(raw, ['profile', 'developer', 'user', 'account'], raw);
  const statsSource = pick(raw, ['stats', 'summary'], {});
  const gamesSource = pick(raw, ['games', 'works', 'items', 'list'], []);
  const games = toArray(gamesSource).map(normalizeWork);
  const stats = normalizeStats(statsSource);

  return {
    profile: normalizeProfile(profileSource, fallbackUser),
    stats: {
      ...stats,
      works: stats.works || games.length,
    },
    games,
  };
}

export async function getDeveloperCenter(fallbackUser) {
  const [profile, stats, games] = await Promise.all([
    request('/api/developer/me'),
    request('/api/developer/stats'),
    request('/api/developer/games', { params: { status: 'all' } }),
  ]);

  return normalizeDeveloperCenter({ profile, stats, games }, fallbackUser);
}

export async function getDeveloperGames(status = 'all') {
  return toArray(await request('/api/developer/games', { params: { status } })).map(normalizeWork);
}

export function resubmitDeveloperGame(id) {
  return request(`/api/developer/games/${id}/resubmit`, { method: 'POST' });
}

export function urgeDeveloperGameReview(id) {
  return request(`/api/developer/games/${id}/urge-review`, { method: 'POST' });
}

export function deleteDeveloperGame(id) {
  return request(`/api/developer/games/${id}`, { method: 'DELETE' });
}
