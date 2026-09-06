import { isMockEnabled, request } from './http';
import {
  mockGetGameDetail,
  mockGetGameSource,
  mockGetGames,
  mockLikeGame,
  mockSubmitGame,
  mockTrackGamePlay,
} from './mockAdapter';
import { normalizeDownload, normalizeGame, normalizePage } from './normalizers';

export async function getGameList(params) {
  const result = isMockEnabled() ? await mockGetGames(params) : await request('/api/games', { params });
  return normalizePage(result, normalizeGame);
}

export async function getGameDetail(id) {
  return normalizeGame(isMockEnabled() ? await mockGetGameDetail(id) : await request(`/api/games/${id}`));
}

export async function likeGame(id) {
  return normalizeGame(isMockEnabled() ? await mockLikeGame(id) : await request(`/api/games/${id}/like`, { method: 'POST' }));
}

export function trackGamePlay(id) {
  return isMockEnabled() ? mockTrackGamePlay(id) : request(`/api/games/${id}/play`, { method: 'POST' });
}

export async function getGameSource(id) {
  return normalizeDownload(isMockEnabled() ? await mockGetGameSource(id) : await request(`/api/games/${id}/download-source`));
}

export async function submitGame(payload) {
  return normalizeGame(isMockEnabled() ? await mockSubmitGame(payload) : await request('/api/games', { method: 'POST', body: payload }));
}
