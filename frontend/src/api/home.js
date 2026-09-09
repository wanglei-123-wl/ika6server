import { request } from './http';
import { normalizeHome } from './normalizers';

export async function getHomeData() {
  return normalizeHome(await request('/api/home'));
}
