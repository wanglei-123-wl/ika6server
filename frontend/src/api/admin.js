import { request } from './http';
import { mockBanUser, mockGetAdminDashboard, mockReviewGame, mockReviewPost } from './mockAdapter';
import { normalizeAdminDashboard } from './normalizers';
import { requestWithFallback } from './runtime';

export async function getAdminDashboard() {
  return normalizeAdminDashboard(await requestWithFallback(
    () => request('/api/admin/dashboard'),
    () => mockGetAdminDashboard(),
  ));
}

export function reviewGame(id, payload) {
  return requestWithFallback(
    () => request(`/api/admin/games/${id}/review`, { method: 'POST', body: payload }),
    () => mockReviewGame(id, payload),
  );
}

export function reviewPost(id, payload) {
  return requestWithFallback(
    () => request(`/api/admin/posts/${id}/review`, { method: 'POST', body: payload }),
    () => mockReviewPost(id, payload),
  );
}

export function banUser(id, payload) {
  return requestWithFallback(
    () => request(`/api/admin/users/${id}/ban`, { method: 'POST', body: payload }),
    () => mockBanUser(id, payload),
  );
}
