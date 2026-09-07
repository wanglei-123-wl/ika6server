import { request } from './http';
import {
  mockGetGameDetail,
  mockGetGameSource,
  mockGetGames,
  mockLikeGame,
  mockSubmitGame,
  mockTrackGamePlay,
} from './mockAdapter';
import { normalizeDownload, normalizeGame, normalizePage } from './normalizers';
import { requestWithFallback } from './runtime';

export async function getGameList(params) {
  const result = await requestWithFallback(
    () => request('/api/games', { params }),
    () => mockGetGames(params),
  );
  return normalizePage(result, normalizeGame);
}

export async function getGameDetail(id) {
  return normalizeGame(await requestWithFallback(
    () => request(`/api/games/${id}`),
    () => mockGetGameDetail(id),
  ));
}

export async function likeGame(id) {
  return normalizeGame(await requestWithFallback(
    () => request(`/api/games/${id}/like`, { method: 'POST' }),
    () => mockLikeGame(id),
  ));
}

export function trackGamePlay(id) {
  return requestWithFallback(
    () => request(`/api/games/${id}/play`, { method: 'POST' }),
    () => mockTrackGamePlay(id),
  );
}

export async function getGameSource(id) {
  return normalizeDownload(await requestWithFallback(
    () => request(`/api/games/${id}/download-source`),
    () => mockGetGameSource(id),
  ));
}

export async function submitGame(payload) {
  return normalizeGame(await requestWithFallback(
    () => request('/api/games', { method: 'POST', body: payload }),
    () => mockSubmitGame(payload),
  ));
}
