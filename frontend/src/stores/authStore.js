import { computed, reactive } from 'vue';
import { getCurrentUser, login, logout, register, socialLogin } from '../api/auth';
import { getAuthToken, getAuthTokenPersistence } from '../api/http';

const USER_CACHE_KEY = 'ika6_current_user';
const LEGACY_USER_CACHE_KEY = 'pf_user';

const state = reactive({
  user: null,
  initialized: false,
  loading: false,
});

function readCachedUser() {
  try {
    return JSON.parse(
      window.localStorage.getItem(USER_CACHE_KEY)
        || window.sessionStorage.getItem(USER_CACHE_KEY)
        || window.localStorage.getItem(LEGACY_USER_CACHE_KEY)
        || 'null',
    );
  } catch {
    return null;
  }
}

function cacheUser(user, { persist = true } = {}) {
  window.localStorage.removeItem(USER_CACHE_KEY);
  window.sessionStorage.removeItem(USER_CACHE_KEY);
  window.localStorage.removeItem(LEGACY_USER_CACHE_KEY);

  if (user) {
    const storage = persist ? window.localStorage : window.sessionStorage;
    storage.setItem(USER_CACHE_KEY, JSON.stringify(user));
  }
}

async function runAuthAction(action, { persist = true } = {}) {
  state.loading = true;

  try {
    const result = await action();
    state.user = result.user || result;
    cacheUser(state.user, { persist });
    return result;
  } finally {
    state.loading = false;
  }
}

function withAdminAccess(user) {
  if (!user) return user;

  return {
    ...user,
    adminAccess: user.role === 'admin',
  };
}

export function useAuthStore() {
  const isAuthenticated = computed(() => Boolean(state.user));
  const isAdmin = computed(() => state.user?.role === 'admin' && state.user?.adminAccess === true);

  async function bootstrap() {
    if (!getAuthToken()) {
      state.user = null;
      cacheUser(null);
      state.initialized = true;
      return;
    }

    const tokenPersistence = getAuthTokenPersistence();
    state.user = readCachedUser();

    try {
      const user = await getCurrentUser();
      state.user = withAdminAccess(user);
      cacheUser(state.user, { persist: tokenPersistence !== 'session' });
    } catch {
      state.user = null;
      cacheUser(null);
    }

    state.initialized = true;
  }

  async function loginWithPassword(payload) {
    const result = await runAuthAction(() => login(payload), { persist: payload.remember !== false });
    state.user = withAdminAccess(result.user || result);
    cacheUser(state.user, { persist: payload.remember !== false });
    return { ...result, user: state.user };
  }

  async function registerAccount(payload) {
    return runAuthAction(() => register(payload));
  }

  async function loginWithProvider(provider) {
    return runAuthAction(() => socialLogin(provider));
  }

  async function logoutAccount() {
    state.loading = true;

    try {
      await logout();
    } finally {
      state.user = null;
      state.loading = false;
      cacheUser(null);
    }
  }

  function expireSession() {
    state.user = null;
    cacheUser(null);
  }

  return {
    state,
    user: computed(() => state.user),
    loading: computed(() => state.loading),
    initialized: computed(() => state.initialized),
    isAuthenticated,
    isAdmin,
    bootstrap,
    loginWithPassword,
    registerAccount,
    loginWithProvider,
    logoutAccount,
    expireSession,
  };
}
