import { request } from './http';
import { normalizeSearch } from './normalizers';

export async function searchSite(params) {
  return normalizeSearch(await request('/api/search', { params }));
}
