import { isMockEnabled, request } from './http';
import { mockGetHome } from './mockAdapter';
import { normalizeHome } from './normalizers';

export async function getHomeData() {
  return normalizeHome(isMockEnabled() ? await mockGetHome() : await request('/api/home'));
}
