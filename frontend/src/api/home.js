import { request } from './http';
import { mockGetHome } from './mockAdapter';
import { normalizeHome } from './normalizers';
import { requestWithFallback } from './runtime';

export async function getHomeData() {
  return normalizeHome(await requestWithFallback(
    () => request('/api/home'),
    () => mockGetHome(),
  ));
}
