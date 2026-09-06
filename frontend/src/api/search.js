import { isMockEnabled, request } from './http';
import { mockSearch } from './mockAdapter';
import { normalizeSearch } from './normalizers';

export async function searchSite(params) {
  return normalizeSearch(isMockEnabled() ? await mockSearch(params) : await request('/api/search', { params }));
}
