import { request } from './http';
import { normalizeDownload, normalizeGame, normalizePage } from './normalizers';

export async function getGameList(params) {
  const result = await request('/api/games', { params });
  return normalizePage(result, normalizeGame);
}

export async function getGameDetail(id) {
  return normalizeGame(await request(`/api/games/${id}`));
}

export async function likeGame(id) {
  return normalizeGame(await request(`/api/games/${id}/like`, { method: 'POST' }));
}

export function trackGamePlay(id) {
  return request(`/api/games/${id}/play`, { method: 'POST' });
}

export async function getGameSource(id) {
  return normalizeDownload(await request(`/api/games/${id}/download-source`));
}

export async function submitGame(payload) {
  return normalizeGame(await request('/api/games', { method: 'POST', body: payload }));
}
