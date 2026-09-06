import { isMockEnabled, request, setAuthToken } from './http';
import { mockGetCurrentUser, mockLogin, mockRegister, mockSocialLogin } from './mockAdapter';
import { normalizeAuthResult, normalizeUser } from './normalizers';

export async function login(payload) {
  const result = normalizeAuthResult(isMockEnabled()
    ? await mockLogin(payload)
    : await request('/api/auth/login', { method: 'POST', body: payload, auth: false }));

  setAuthToken(result.token, { persist: payload.remember !== false });
  return result;
}

export async function register(payload) {
  const result = normalizeAuthResult(isMockEnabled()
    ? await mockRegister(payload)
    : await request('/api/auth/register', { method: 'POST', body: payload, auth: false }));

  setAuthToken(result.token);
  return result;
}

export async function socialLogin(provider) {
  const result = normalizeAuthResult(isMockEnabled()
    ? await mockSocialLogin(provider)
    : await request('/api/auth/social-login', { method: 'POST', body: { provider }, auth: false }));

  setAuthToken(result.token);
  return result;
}

export async function logout() {
  if (!isMockEnabled()) {
    await request('/api/auth/logout', { method: 'POST' });
  }

  setAuthToken('');
}

export async function getCurrentUser() {
  return normalizeUser(isMockEnabled() ? await mockGetCurrentUser() : await request('/api/auth/me'));
}
