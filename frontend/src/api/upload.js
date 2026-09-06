import { isMockEnabled, request } from './http';
import { mockSubmitGame } from './mockAdapter';
import { normalizeGame } from './normalizers';

const UPLOAD_TIMEOUT = Number(import.meta.env.VITE_UPLOAD_TIMEOUT || 120000);

export async function uploadGame(payload) {
  if (isMockEnabled()) return normalizeGame(await mockSubmitGame(payload));

  const formData = new FormData();
  Object.entries(payload || {}).forEach(([key, value]) => {
    if (Array.isArray(value)) {
      value.forEach((item) => formData.append(key, item));
    } else if (value !== undefined && value !== null) {
      formData.append(key, value);
    }
  });

  return normalizeGame(await request('/api/games', { method: 'POST', body: formData, timeout: UPLOAD_TIMEOUT }));
}
