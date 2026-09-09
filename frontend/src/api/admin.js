import { request, resolveApiAssetUrl } from './http';
import { normalizeAdminDashboard } from './normalizers';

function toArray(value) {
  if (Array.isArray(value)) return value;
  if (Array.isArray(value?.items)) return value.items;
  if (Array.isArray(value?.data)) return value.data;
  return [];
}

function pick(source, keys, fallback = undefined) {
  for (const key of keys) {
    if (source?.[key] !== undefined && source?.[key] !== null) return source[key];
  }
  return fallback;
}

function normalizeAdminGame(raw = {}) {
  return {
    id: pick(raw, ['id', 'gameId', 'game_id'], ''),
    title: pick(raw, ['title', 'name', 'gameName', 'game_name'], '未命名作品'),
    author: pick(raw, ['author', 'authorName', 'author_name', 'username'], 'unknown'),
    glyph: pick(raw, ['glyph', 'icon', 'mark'], '◆'),
    coverUrl: resolveApiAssetUrl(pick(raw, ['coverUrl', 'cover_url', 'coverImage'], '')),
    status: pick(raw, ['status', 'reviewStatus', 'review_status'], 'reviewing'),
    genre: pick(raw, ['genre', 'category', 'type'], '其他'),
    engine: pick(raw, ['engine', 'engineName', 'engine_name'], '未知引擎'),
    version: pick(raw, ['version', 'ver', 'releaseVersion', 'release_version'], 'v1.0.0'),
    plays: pick(raw, ['plays', 'playCount', 'play_count'], 0),
    downloads: pick(raw, ['downloads', 'downloadCount', 'download_count'], 0),
    likes: pick(raw, ['likes', 'likeCount', 'like_count'], 0),
    createdAt: pick(raw, ['createdAt', 'created_at'], ''),
    updatedAt: pick(raw, ['updatedAt', 'updated_at', 'createdAt', 'created_at'], ''),
    reason: pick(raw, ['rejectReason', 'reject_reason', 'reason'], ''),
    playUrl: pick(raw, ['playUrl', 'play_url'], ''),
    sourceUrl: pick(raw, ['sourceUrl', 'source_url'], ''),
  };
}

function normalizePage(raw, itemNormalizer) {
  const items = toArray(raw);
  return {
    page: Number(pick(raw, ['page', 'currentPage', 'pageNo'], 1)),
    pageSize: Number(pick(raw, ['pageSize', 'size', 'limit'], items.length || 20)),
    total: Number(pick(raw, ['total', 'totalCount', 'count'], items.length)),
    items: items.map(itemNormalizer),
  };
}

function normalizeAdminUser(raw = {}) {
  const name = pick(raw, ['name', 'username', 'nickname'], '未命名用户');
  return {
    id: pick(raw, ['id', 'userId', 'user_id'], ''),
    name,
    initial: pick(raw, ['initial'], String(name).charAt(0).toUpperCase() || 'U'),
    email: pick(raw, ['email'], ''),
    role: pick(raw, ['role'], 'user'),
    level: pick(raw, ['level', 'rank'], 'lv1'),
    status: pick(raw, ['status'], 'normal'),
    games: Number(pick(raw, ['games', 'gameCount', 'game_count'], 0)),
    posts: Number(pick(raw, ['posts', 'postCount', 'post_count'], 0)),
    createdAt: pick(raw, ['createdAt', 'created_at'], ''),
    bannedUntil: pick(raw, ['bannedUntil', 'banned_until'], ''),
    banReason: pick(raw, ['banReason', 'ban_reason'], ''),
  };
}

function normalizeAuditLog(raw = {}) {
  return {
    id: pick(raw, ['id'], ''),
    actorId: pick(raw, ['actorId', 'actor_id'], ''),
    action: pick(raw, ['action'], '未知操作'),
    target: pick(raw, ['target'], ''),
    targetId: pick(raw, ['targetId', 'target_id'], ''),
    details: pick(raw, ['details'], null),
    createdAt: pick(raw, ['createdAt', 'created_at'], ''),
  };
}

function normalizeAdminReport(raw = {}) {
  return {
    id: pick(raw, ['id', 'reportId', 'report_id'], ''),
    type: pick(raw, ['type', 'reportType', 'report_type'], '内容举报'),
    reason: pick(raw, ['reason', 'description', 'message'], ''),
    status: pick(raw, ['status'], 'pending'),
    reporter: pick(raw, ['reporter', 'reporterName', 'reporter_name'], '未知用户'),
    target: pick(raw, ['target', 'targetTitle', 'target_title'], '未知内容'),
    targetId: pick(raw, ['targetId', 'target_id'], ''),
    createdAt: pick(raw, ['createdAt', 'created_at'], ''),
  };
}

export async function getAdminDashboard() {
  return normalizeAdminDashboard(await request('/api/admin/dashboard'));
}

export async function getAdminGames(status = 'all') {
  return toArray(await request('/api/admin/games', { params: { status } })).map(normalizeAdminGame);
}

export function reviewGame(id, payload) {
  return request(`/api/admin/games/${id}/review`, { method: 'POST', body: payload });
}

export function reviewPost(id, payload) {
  return request(`/api/admin/posts/${id}/review`, { method: 'POST', body: payload });
}

export function banUser(id, payload) {
  return request(`/api/admin/users/${id}/ban`, { method: 'POST', body: payload });
}

export function unbanUser(id) {
  return request(`/api/admin/users/${id}/unban`, { method: 'POST' });
}

export async function getAdminUsers(page = 1, pageSize = 20) {
  return normalizePage(await request('/api/admin/users', { params: { page, pageSize } }), normalizeAdminUser);
}

export async function getAuditLogs(page = 1, pageSize = 20) {
  return normalizePage(await request('/api/admin/audit-logs', { params: { page, pageSize } }), normalizeAuditLog);
}

export async function getAdminReports(page = 1, pageSize = 20) {
  return normalizePage(await request('/api/admin/reports', { params: { page, pageSize } }), normalizeAdminReport);
}

export function resolveAdminReport(id, payload = {}) {
  return request(`/api/admin/reports/${id}/resolve`, { method: 'POST', body: payload });
}
