import { isMockEnabled, request } from './http';
import { mockBanUser, mockGetAdminDashboard, mockReviewGame, mockReviewPost } from './mockAdapter';
import { normalizeAdminDashboard } from './normalizers';

export async function getAdminDashboard() {
  return normalizeAdminDashboard(isMockEnabled() ? await mockGetAdminDashboard() : await request('/api/admin/dashboard'));
}

export function reviewGame(id, payload) {
  return isMockEnabled()
    ? mockReviewGame(id, payload)
    : request(`/api/admin/games/${id}/review`, { method: 'POST', body: payload });
}

export function reviewPost(id, payload) {
  return isMockEnabled()
    ? mockReviewPost(id, payload)
    : request(`/api/admin/posts/${id}/review`, { method: 'POST', body: payload });
}

export function banUser(id, payload) {
  return isMockEnabled()
    ? mockBanUser(id, payload)
    : request(`/api/admin/users/${id}/ban`, { method: 'POST', body: payload });
}
