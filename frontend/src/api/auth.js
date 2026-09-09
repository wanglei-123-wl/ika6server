import { ApiError, request, setAuthToken } from './http';
import { normalizeAuthResult, normalizeUser } from './normalizers';

export async function login(payload) {
  const result = normalizeAuthResult(await request('/api/auth/login', {
    method: 'POST',
    body: payload,
    auth: false,
  }));

  setAuthToken(result.token, { persist: payload.remember !== false });
  return result;
}

export async function register(payload) {
  const result = normalizeAuthResult(await request('/api/auth/register', {
    method: 'POST',
    body: payload,
    auth: false,
  }));

  setAuthToken(result.token);
  return result;
}

export async function socialLogin(provider) {
  const result = normalizeAuthResult(await request('/api/auth/social-login', {
    method: 'POST',
    body: { provider },
    auth: false,
  }));

  setAuthToken(result.token);
  return result;
}

export async function logout() {
  try {
    await request('/api/auth/logout', { method: 'POST' });
  } finally {
    setAuthToken('');
  }
}

export async function getCurrentUser() {
  const result = await request('/api/auth/me');
  if (!result?.user || result.user.id == null) {
    throw new ApiError('服务器返回的用户信息不完整', { code: 'INVALID_RESPONSE' });
  }
  return normalizeUser(result.user);
}
