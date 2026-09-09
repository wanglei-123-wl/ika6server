import { request } from './http';
import { normalizeDownload, normalizePage, normalizeRepo } from './normalizers';

export async function getRepoList(params) {
  const result = await request('/api/repos', { params });
  return normalizePage(result, normalizeRepo);
}

export async function getRepoDetail(id) {
  return normalizeRepo(await request(`/api/repos/${id}`));
}

export async function downloadRepo(id) {
  return normalizeDownload(await request(`/api/repos/${id}/download`));
}
