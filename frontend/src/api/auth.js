import { isMockEnabled, request, setAuthToken } from './http';
import { mockGetCurrentUser, mockLogin, mockRegister, mockSocialLogin } from './mockAdapter';
import { normalizeAuthResult, normalizeUser } from './normalizers';
import { requestWithFallback } from './runtime';

export async function login(payload) {
  const result = normalizeAuthResult(await requestWithFallback(
    () => request('/api/auth/login', { method: 'POST', body: payload, auth: false }),
    () => mockLogin(payload),
  ));

  setAuthToken(result.token, { persist: payload.remember !== false });
  return result;
}

export async function register(payload) {
  const result = normalizeAuthResult(await requestWithFallback(
    () => request('/api/auth/register', { method: 'POST', body: payload, auth: false }),
    () => mockRegister(payload),
  ));

  setAuthToken(result.token);
  return result;
}

export async function socialLogin(provider) {
  const result = normalizeAuthResult(await requestWithFallback(
    () => request('/api/auth/social-login', { method: 'POST', body: { provider }, auth: false }),
    () => mockSocialLogin(provider),
  ));

  setAuthToken(result.token);
  return result;
}

export async function logout() {
  try {
    await requestWithFallback(
      () => request('/api/auth/logout', { method: 'POST' }),
      () => Promise.resolve({}),
    );
  } finally {
    setAuthToken('');
  }
}

export async function getCurrentUser() {
  return normalizeUser(isMockEnabled()
    ? await mockGetCurrentUser()
    : await request('/api/auth/me'));
}
