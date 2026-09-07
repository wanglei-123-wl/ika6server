import { request } from './http';
import { mockGetDevDocs } from './mockAdapter';
import { normalizeDevDocs } from './normalizers';
import { requestWithFallback } from './runtime';

export async function getDevDocs() {
  return normalizeDevDocs(await requestWithFallback(
    () => request('/api/dev-docs'),
    () => mockGetDevDocs(),
  ));
}
