const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, '');
const REQUEST_TIMEOUT = Number(import.meta.env.VITE_API_TIMEOUT || 15000);
const TOKEN_KEY = 'ika6_auth_token';
export const AUTH_EXPIRED_EVENT = 'ika6:auth-expired';

export class ApiError extends Error {
  constructor(message, { status = 0, code = -1, errorCode = '', data = null, requestId = '' } = {}) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
    this.errorCode = errorCode;
    this.data = data;
    this.requestId = requestId;
  }
}

export function getAuthToken() {
  return window.localStorage.getItem(TOKEN_KEY) || window.sessionStorage.getItem(TOKEN_KEY);
}

export function getAuthTokenPersistence() {
  if (window.localStorage.getItem(TOKEN_KEY)) return 'local';
  if (window.sessionStorage.getItem(TOKEN_KEY)) return 'session';
  return 'none';
}

export function setAuthToken(token, { persist = true } = {}) {
  window.localStorage.removeItem(TOKEN_KEY);
  window.sessionStorage.removeItem(TOKEN_KEY);

  if (token) {
    const storage = persist ? window.localStorage : window.sessionStorage;
    storage.setItem(TOKEN_KEY, token);
  }
}

function notifyAuthExpired() {
  window.dispatchEvent(new CustomEvent(AUTH_EXPIRED_EVENT));
}

function buildUrl(path, params) {
  const url = new URL(`${API_BASE_URL}${path}`, window.location.origin);

  Object.entries(params || {}).forEach(([key, value]) => {
    if (value !== undefined && value !== null && value !== '') {
      url.searchParams.set(key, value);
    }
  });

  return url.toString();
}

export function resolveApiAssetUrl(path) {
  if (!path) return '';
  if (/^(https?:|data:|blob:)/i.test(path)) return path;
  return new URL(path, API_BASE_URL || window.location.origin).toString();
}

function normalizePayload(payload, { notifyExpired = true, status = 0, requestId = '' } = {}) {
  if (payload && typeof payload === 'object' && 'code' in payload) {
    if (payload.code !== 0) {
      if (payload.code === 401 && notifyExpired) {
        setAuthToken('');
        notifyAuthExpired();
      }

      throw new ApiError(payload.message || '请求失败，请稍后重试', {
        status,
        code: payload.code,
        errorCode: payload.errorCode,
        data: payload.data,
        requestId: payload.requestId || requestId,
      });
    }

    return payload.data;
  }

  if (payload && typeof payload === 'object' && 'success' in payload) {
    if (!payload.success) {
      throw new ApiError(payload.message || '请求失败，请稍后重试', {
        status,
        code: payload.code,
        errorCode: payload.errorCode,
        data: payload.data,
        requestId: payload.requestId || requestId,
      });
    }

    return payload.data ?? payload.result ?? payload;
  }

  return payload;
}

export async function request(path, options = {}) {
  const {
    method = 'GET',
    params,
    body,
    headers = {},
    auth = true,
    timeout = REQUEST_TIMEOUT,
  } = options;
  const token = auth ? getAuthToken() : '';
  const isFormData = body instanceof FormData;
  const controller = new AbortController();
  const timer = window.setTimeout(() => controller.abort(), timeout);

  let response;

  try {
    response = await fetch(buildUrl(path, params), {
      method,
      headers: {
        ...(isFormData ? {} : { 'Content-Type': 'application/json' }),
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...headers,
      },
      body: body ? (isFormData ? body : JSON.stringify(body)) : undefined,
      signal: controller.signal,
    });
  } catch (error) {
    if (error.name === 'AbortError') {
      throw new ApiError('请求超时，请稍后重试', { code: 'TIMEOUT' });
    }

    throw new ApiError('网络连接失败，请检查服务是否可用', { code: 'NETWORK_ERROR' });
  } finally {
    window.clearTimeout(timer);
  }

  const requestId = response.headers.get('X-Request-ID') || '';
  const text = await response.text();
  let payload = null;

  try {
    payload = text ? JSON.parse(text) : null;
  } catch {
    throw new ApiError('服务器返回的数据格式不正确', { status: response.status, requestId });
  }

  if (!response.ok) {
    if (response.status === 401 && auth) {
      setAuthToken('');
      notifyAuthExpired();
    }
    throw new ApiError(payload?.message || '请求失败，请稍后重试', {
      status: response.status,
      code: payload?.code,
      errorCode: payload?.errorCode,
      data: payload?.data,
      requestId: payload?.requestId || requestId,
    });
  }

  return normalizePayload(payload, { notifyExpired: auth, status: response.status, requestId });
}
