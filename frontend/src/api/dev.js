import { isMockEnabled, request } from './http';
import { mockGetDevDocs } from './mockAdapter';
import { normalizeDevDocs } from './normalizers';

export async function getDevDocs() {
  return normalizeDevDocs(isMockEnabled() ? await mockGetDevDocs() : await request('/api/dev-docs'));
}
