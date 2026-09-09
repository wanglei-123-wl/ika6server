import { request } from './http';
import { normalizeDevDocs } from './normalizers';

export async function getDevDocs() {
  return normalizeDevDocs(await request('/api/dev-docs'));
}
