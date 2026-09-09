import assert from 'node:assert/strict';
import { after, before, beforeEach, test } from 'node:test';
import { fileURLToPath } from 'node:url';
import { createServer } from 'vite';

const originalFetch = globalThis.fetch;
const originalWindow = globalThis.window;
const admin = { id: 8, name: 'Test Admin', role: 'admin', adminAccess: true };
let server;
let store;
let http;
let auth;

function storage() {
  const values = new Map();
  return {
    getItem: (key) => values.get(key) ?? null,
    setItem: (key, value) => values.set(key, String(value)),
    removeItem: (key) => values.delete(key),
  };
}

function reply(data, { status = 200, requestId = '', errorCode = '', message = 'success' } = {}) {
  return new Response(JSON.stringify({
    code: status === 200 ? 0 : status, data, message, errorCode, requestId,
  }), {
    status,
    headers: { 'Content-Type': 'application/json', 'X-Request-ID': requestId },
  });
}

before(async () => {
  server = await createServer({
    root: fileURLToPath(new URL('..', import.meta.url)),
    configFile: false,
    envFile: false,
    optimizeDeps: { noDiscovery: true, include: [] },
    server: { middlewareMode: true, hmr: false, watch: null },
    appType: 'custom',
  });
  http = await server.ssrLoadModule('/src/api/http.js');
  auth = await server.ssrLoadModule('/src/api/auth.js');
  const { useAuthStore } = await server.ssrLoadModule('/src/stores/authStore.js');
  store = useAuthStore();
});

beforeEach(() => {
  globalThis.window = {
    localStorage: storage(),
    sessionStorage: storage(),
    location: { origin: 'http://frontend.test' },
    setTimeout,
    clearTimeout,
    dispatchEvent() {},
  };
  store.expireSession();
  globalThis.fetch = async () => { throw new Error('Unexpected fetch'); };
});

after(async () => {
  globalThis.fetch = originalFetch;
  if (originalWindow === undefined) delete globalThis.window;
  else globalThis.window = originalWindow;
  await server?.close();
});

for (const remember of [true, false]) {
  test(`login and refresh preserve admin identity; remember=${remember}`, async () => {
    const payload = { account: 'admin@example.test', password: ' TEST-password$with-spaces ', remember };
    globalThis.fetch = async (url, options) => {
      if (new URL(url).pathname === '/api/auth/login') {
        assert.deepEqual(JSON.parse(options.body), payload);
        assert.equal(options.headers.Authorization, undefined);
        return reply({ user: admin, token: 'test-token' });
      }
      assert.equal(new URL(url).pathname, '/api/auth/me');
      assert.equal(options.headers.Authorization, 'Bearer test-token');
      return reply({ user: admin });
    };
    await store.loginWithPassword(payload);
    assert.equal(store.isAdmin.value, true);
    await store.bootstrap();
    assert.equal(store.user.value.id, admin.id);
    assert.equal(store.user.value.name, admin.name);
    assert.equal(store.isAdmin.value, true);
    assert.equal(http.getAuthTokenPersistence(), remember ? 'local' : 'session');
    const selected = remember ? window.localStorage : window.sessionStorage;
    const other = remember ? window.sessionStorage : window.localStorage;
    assert.equal(JSON.parse(selected.getItem('ika6_current_user')).id, admin.id);
    assert.equal(other.getItem('ika6_current_user'), null);
  });
}

test('role alone cannot override a denied adminAccess flag', async () => {
  globalThis.fetch = async () => reply({ user: { ...admin, adminAccess: false }, token: 'test-token' });
  await store.loginWithPassword({ account: 'admin@example.test', password: 'test-password' });
  assert.equal(store.isAdmin.value, false);
  await store.bootstrap();
  assert.equal(store.isAdmin.value, false);
});

test('missing current user is rejected instead of fabricating a default user', async () => {
  http.setAuthToken('test-token');
  globalThis.fetch = async () => reply({});
  await assert.rejects(auth.getCurrentUser(), { code: 'INVALID_RESPONSE' });
  await store.bootstrap();
  assert.equal(store.isAuthenticated.value, false);
  assert.equal(window.localStorage.getItem('ika6_current_user'), null);
});

for (const [status, errorCode] of [[401, 'INVALID_CREDENTIALS'], [503, 'AUTH_SERVICE_UNAVAILABLE']]) {
  test(`login preserves ${status} error category and request ID without expiring another session`, async () => {
    http.setAuthToken('existing-token');
    globalThis.fetch = async () => reply(null, { status, errorCode, requestId: 'test-request-id' });
    await assert.rejects(auth.login({ account: 'admin@example.test', password: 'test-password' }), {
      status, errorCode, requestId: 'test-request-id',
    });
    assert.equal(http.getAuthToken(), 'existing-token');
  });
}

test('request ID is preserved for legacy envelopes and malformed response bodies', async () => {
  for (const body of [
    JSON.stringify({ code: 503, errorCode: 'AUTH_SERVICE_UNAVAILABLE', message: 'unavailable' }),
    JSON.stringify({ success: false, message: 'unavailable' }),
    'not-json',
  ]) {
    globalThis.fetch = async () => new Response(body, { headers: { 'X-Request-ID': 'header-id' } });
    await assert.rejects(http.request('/api/auth/login', { auth: false }), { requestId: 'header-id' });
  }
});

test('logout clears both identity and token', async () => {
  globalThis.fetch = async () => reply({ user: admin, token: 'test-token' });
  await store.loginWithPassword({ account: 'admin@example.test', password: 'test-password' });
  globalThis.fetch = async () => reply({ loggedOut: true });
  await store.logoutAccount();
  assert.equal(store.isAuthenticated.value, false);
  assert.equal(http.getAuthToken(), null);
});
