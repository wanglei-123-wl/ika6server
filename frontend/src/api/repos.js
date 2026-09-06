import { isMockEnabled, request } from './http';
import { mockDownloadRepo, mockGetRepoDetail, mockGetRepos } from './mockAdapter';
import { normalizeDownload, normalizePage, normalizeRepo } from './normalizers';

export async function getRepoList(params) {
  const result = isMockEnabled() ? await mockGetRepos(params) : await request('/api/repos', { params });
  return normalizePage(result, normalizeRepo);
}

export async function getRepoDetail(id) {
  return normalizeRepo(isMockEnabled() ? await mockGetRepoDetail(id) : await request(`/api/repos/${id}`));
}

export async function downloadRepo(id) {
  return normalizeDownload(isMockEnabled() ? await mockDownloadRepo(id) : await request(`/api/repos/${id}/download`));
}
