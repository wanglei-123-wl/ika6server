import { request } from './http';
import { mockSearch } from './mockAdapter';
import { normalizeSearch } from './normalizers';
import { requestWithFallback } from './runtime';

export async function searchSite(params) {
  return normalizeSearch(await requestWithFallback(
    () => request('/api/search', { params }),
    () => mockSearch(params),
  ));
}
