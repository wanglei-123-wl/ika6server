import { request } from './http';
import { mockDownloadRepo, mockGetRepoDetail, mockGetRepos } from './mockAdapter';
import { normalizeDownload, normalizePage, normalizeRepo } from './normalizers';
import { requestWithFallback } from './runtime';

export async function getRepoList(params) {
  const result = await requestWithFallback(
    () => request('/api/repos', { params }),
    () => mockGetRepos(params),
  );
  return normalizePage(result, normalizeRepo);
}

export async function getRepoDetail(id) {
  return normalizeRepo(await requestWithFallback(
    () => request(`/api/repos/${id}`),
    () => mockGetRepoDetail(id),
  ));
}

export async function downloadRepo(id) {
  return normalizeDownload(await requestWithFallback(
    () => request(`/api/repos/${id}/download`),
    () => mockDownloadRepo(id),
  ));
}
