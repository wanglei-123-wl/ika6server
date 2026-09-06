<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import CreatePostModal from './components/CreatePostModal.vue';
import AuthModal from './components/AuthModal.vue';
import UploadGameModal from './components/UploadGameModal.vue';
import { AUTH_EXPIRED_EVENT, createForumPost, downloadRepo, getDevDocs, getForumBars, getForumPosts, getGameList, getGameSource, getHomeData, getRepoList, likeForumPost, searchSite, trackGamePlay, uploadGame } from './api';
import AdminPage from './pages/AdminPage.vue';
import DevPage from './pages/DevPage.vue';
import ForumPage from './pages/ForumPage.vue';
import GameLibraryPage from './pages/GameLibraryPage.vue';
import HomePage from './pages/HomePage.vue';
import MarketPage from './pages/MarketPage.vue';
import NotFoundPage from './pages/NotFoundPage.vue';
import { useAuthStore } from './stores/authStore';
import ika6Logo from './assets/ika6-logo.png';

const routes = [
  { key: 'home', label: '广场', hash: '#/' },
  { key: 'library', label: '游戏库', hash: '#/library' },
  { key: 'forum', label: '论坛', hash: '#/forum' },
  { key: 'market', label: '源码集市', hash: '#/market' },
  { key: 'dev', label: '开发者', hash: '#/dev' },
  { key: 'admin', label: '管理后台', hash: '#/admin', adminOnly: true },
];

const activeView = ref('home');
const currentPath = ref('#/');
const activeForumCat = ref('全部');
const activeBar = ref({ icon: '◆', name: '独立游戏吧', desc: '正在加载社区数据', posts: '0', members: '0' });
const activeDocKey = ref('quickstart');
const uploadOpen = ref(false);
const uploadSubmitting = ref(false);
const postModalOpen = ref(false);
const postSubmitting = ref(false);
const authOpen = ref(false);
const accountMenuOpen = ref(false);
const playerOpen = ref(false);
const currentGame = ref(null);
const toast = ref('');
const homeStats = ref({ projects: 0, plays: 0, contributors: 0, price: 0 });
const marketStats = ref({ projects: 0, weeklyActive: 0, downloads: '0', price: 0 });
const homeGames = ref([]);
const homeHotPosts = ref([]);
const homeFeedPosts = ref([]);
const gameList = ref([]);
const postList = ref([]);
const barList = ref([]);
const repoList = ref([]);
const devDocsMap = ref({});
const pageLoading = ref(false);
const pageError = ref('');
const searchKeyword = ref('');
const searchLoading = ref(false);
const searchOpen = ref(false);
const searchResults = ref({ games: [], posts: [], repos: [] });
let searchTimer = 0;
let searchRequestId = 0;
const theme = ref('dark');
const authStore = useAuthStore();
const currentUser = computed(() => authStore.user.value);
const isAdmin = computed(() => authStore.isAdmin.value);
const navRoutes = computed(() => routes.filter((item) => !item.adminOnly || isAdmin.value));
const searchTotal = computed(() => (
  searchResults.value.games.length
  + searchResults.value.posts.length
  + searchResults.value.repos.length
));

const forumCats = ['全部', '精华', '互助', '作品发布', '招募', '源码分享', '灌水'];
const forumPosts = computed(() => (
  activeForumCat.value === '全部'
    ? postList.value
    : postList.value.filter((post) => post.cat === activeForumCat.value)
));

function route() {
  currentPath.value = window.location.hash || '#/';
  const key = currentPath.value.replace('#/', '') || 'home';
  activeView.value = routes.some((item) => item.key === key) ? key : 'notFound';
}

function openPlayer(game) {
  if (!game) return;
  currentGame.value = game;
  playerOpen.value = true;
  trackGamePlay(game.id).catch(() => {});
}

function showToast(message) {
  toast.value = message;
  window.clearTimeout(showToast.timer);
  showToast.timer = window.setTimeout(() => {
    toast.value = '';
  }, 2400);
}

function likePost(post) {
  if (!currentUser.value) {
    openAuth();
    showToast('请先登录后再点赞');
    return;
  }

  post.liked = !post.liked;
  post.likes += post.liked ? 1 : -1;
  likeForumPost(post.id).catch(() => {
    post.liked = !post.liked;
    post.likes += post.liked ? 1 : -1;
    showToast('点赞失败，请稍后重试');
  });
}

function openPostModal() {
  if (!currentUser.value) {
    openAuth();
    showToast('请先登录后再发帖');
    return;
  }

  postModalOpen.value = true;
}

async function submitPost(payload) {
  postSubmitting.value = true;

  try {
    const post = await createForumPost({ ...payload, barId: activeBar.value.id || 1 });
    postList.value = [post, ...postList.value];
    postModalOpen.value = false;
    showToast('帖子已发布');
  } catch {
    showToast('发帖失败，请稍后重试');
  } finally {
    postSubmitting.value = false;
  }
}

async function submitUpload(payload) {
  if (!currentUser.value) {
    uploadOpen.value = false;
    openAuth();
    showToast('请先登录后再发布游戏');
    return;
  }

  uploadSubmitting.value = true;

  try {
    await uploadGame(payload);
    uploadOpen.value = false;
    showToast('游戏已提交，等待审核');
  } catch {
    showToast('提交失败，请稍后重试');
  } finally {
    uploadSubmitting.value = false;
  }
}

function openAuth() {
  accountMenuOpen.value = false;
  authOpen.value = true;
}

function handleAuthenticated(user) {
  authOpen.value = false;
  showToast(user.method === 'register' ? `注册成功，欢迎加入 ${user.name}` : `欢迎回来，${user.name}`);
}

function toggleAccountMenu() {
  if (!currentUser.value) {
    openAuth();
    return;
  }
  accountMenuOpen.value = !accountMenuOpen.value;
}

async function logout() {
  await authStore.logoutAccount();
  accountMenuOpen.value = false;
  showToast('已退出登录');
}

function handleAuthExpired() {
  authStore.expireSession();
  accountMenuOpen.value = false;
  showToast('登录已过期，请重新登录');
}

function accountAction(label) {
  accountMenuOpen.value = false;
  showToast(`${label}功能即将开放`);
}

function closeAccountMenu(event) {
  if (!event.target.closest('.account-wrap')) accountMenuOpen.value = false;
  if (!event.target.closest('.search-wrap')) searchOpen.value = false;
}

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark';
  window.localStorage.setItem('pixel-forge-theme', theme.value);
}

async function loadHomeData() {
  const data = await getHomeData();
  homeStats.value = data.stats || homeStats.value;
  marketStats.value = {
    projects: data.stats?.projects || 0,
    weeklyActive: 3214,
    downloads: '847K',
    price: data.stats?.price || 0,
  };
  homeGames.value = data.latestGames || [];
  homeHotPosts.value = data.hotPosts || [];
  homeFeedPosts.value = data.feedPosts || [];
}

async function loadGames() {
  const data = await getGameList({ page: 1, pageSize: 50 });
  gameList.value = data.items || [];
  if (!currentGame.value && gameList.value.length) currentGame.value = gameList.value[0];
}

async function loadForumData() {
  const [barsResult, postsResult] = await Promise.all([
    getForumBars(),
    getForumPosts({ page: 1, pageSize: 50 }),
  ]);
  barList.value = barsResult || [];
  postList.value = postsResult.items || [];
  if (barList.value.length && !barList.value.some((bar) => bar.name === activeBar.value.name)) {
    activeBar.value = barList.value[0];
  }
}

async function loadRepos() {
  const data = await getRepoList({ page: 1, pageSize: 50 });
  repoList.value = data.items || [];
}

async function loadDeveloperDocs() {
  devDocsMap.value = await getDevDocs();
  if (!devDocsMap.value[activeDocKey.value]) {
    activeDocKey.value = Object.keys(devDocsMap.value)[0] || 'quickstart';
  }
}

async function loadClientData() {
  pageLoading.value = true;
  pageError.value = '';

  try {
    await Promise.all([
      loadHomeData(),
      loadGames(),
      loadForumData(),
      loadRepos(),
      loadDeveloperDocs(),
    ]);
  } catch (error) {
    pageError.value = error.message || '页面数据加载失败';
    showToast('页面数据加载失败，已保留当前界面');
  } finally {
    pageLoading.value = false;
  }
}

async function downloadGameSource(game) {
  try {
    const result = await getGameSource(game.id);
    showToast(`源码已准备：${result.size || game.size}`);
  } catch {
    showToast('源码暂时无法下载');
  }
}

async function downloadSourceRepo(repo) {
  try {
    await downloadRepo(repo.id);
    showToast(`正在下载 ${repo.name}`);
  } catch {
    showToast('源码项目暂时无法下载');
  }
}

async function runSearch({ silent = false } = {}) {
  const keyword = searchKeyword.value.trim();
  if (!keyword) {
    searchResults.value = { games: [], posts: [], repos: [] };
    searchOpen.value = false;
    return;
  }

  const requestId = ++searchRequestId;
  searchLoading.value = true;
  searchOpen.value = true;

  try {
    const result = await searchSite({ keyword });
    if (requestId === searchRequestId) searchResults.value = result;
  } catch {
    if (requestId === searchRequestId) {
      searchResults.value = { games: [], posts: [], repos: [] };
      if (!silent) showToast('搜索失败，请稍后重试');
    }
  } finally {
    if (requestId === searchRequestId) searchLoading.value = false;
  }
}

function clearSearch() {
  window.clearTimeout(searchTimer);
  searchRequestId += 1;
  searchKeyword.value = '';
  searchResults.value = { games: [], posts: [], repos: [] };
  searchOpen.value = false;
  searchLoading.value = false;
}

function jumpSearch(hash) {
  window.location.hash = hash;
  searchOpen.value = false;
}

function onKeydown(event) {
  if (event.key === 'Escape') {
    uploadOpen.value = false;
    authOpen.value = false;
    postModalOpen.value = false;
    accountMenuOpen.value = false;
    playerOpen.value = false;
    searchOpen.value = false;
  }
}

onMounted(() => {
  theme.value = window.localStorage.getItem('pixel-forge-theme') || 'dark';
  window.addEventListener(AUTH_EXPIRED_EVENT, handleAuthExpired);
  authStore.bootstrap();
  loadClientData();
  route();
  window.addEventListener('hashchange', route);
  window.addEventListener('keydown', onKeydown);
  window.addEventListener('click', closeAccountMenu);
});

onUnmounted(() => {
  window.clearTimeout(searchTimer);
  window.removeEventListener(AUTH_EXPIRED_EVENT, handleAuthExpired);
  window.removeEventListener('hashchange', route);
  window.removeEventListener('keydown', onKeydown);
  window.removeEventListener('click', closeAccountMenu);
});

watch(searchKeyword, (keyword) => {
  window.clearTimeout(searchTimer);

  if (!keyword.trim()) {
    clearSearch();
    return;
  }

  searchOpen.value = true;
  searchTimer = window.setTimeout(() => {
    runSearch({ silent: true });
  }, 360);
});

watch(activeForumCat, async (cat) => {
  try {
    const result = await getForumPosts({ cat, page: 1, pageSize: 50 });
    postList.value = result.items || [];
  } catch {
    showToast('帖子分类加载失败');
  }
});
</script>

<template>
  <div class="site-shell" :data-theme="theme">
    <header class="topbar">
      <div class="topbar-inner">
        <a class="logo" href="#/">
          <span class="logo-mark"><span>◆</span></span>
          <img class="brand-logo" :src="ika6Logo" alt="ika6.com" />
          <span class="brand-sub">v1.0</span>
        </a>
        <nav class="nav-links">
          <a v-for="routeItem in navRoutes" :key="routeItem.key" :href="routeItem.hash" :class="{ active: activeView === routeItem.key }">
            {{ routeItem.label }}
          </a>
        </nav>
        <div class="search-wrap">
          <div class="search-box">
            <span>⌕</span>
            <input v-model="searchKeyword" placeholder="搜索游戏、源码、帖子..." @focus="searchKeyword && (searchOpen = true)" @keydown.enter="runSearch()" />
            <button v-if="searchKeyword" class="search-clear" aria-label="清空搜索" @click="clearSearch">×</button>
          </div>
          <div v-if="searchOpen" class="search-panel">
            <div class="search-panel-head">
              <strong>{{ searchLoading ? '搜索中...' : `搜索「${searchKeyword.trim()}」` }}</strong>
              <button class="game-act-btn" @click="runSearch">刷新</button>
            </div>
            <div v-if="searchLoading" class="search-loading"><span></span>正在匹配游戏、帖子和源码</div>
            <div v-if="!searchLoading && searchTotal === 0" class="search-empty">没有找到相关游戏、帖子或源码。</div>
            <div v-if="searchResults.games.length" class="search-section">
              <div class="search-title">游戏</div>
              <button v-for="game in searchResults.games" :key="`game-${game.id}`" class="search-item" @click="jumpSearch('#/library')">
                <span>{{ game.glyph }}</span><div><b>{{ game.title }}</b><small>{{ game.engine }} · {{ game.genre }}</small></div>
              </button>
            </div>
            <div v-if="searchResults.posts.length" class="search-section">
              <div class="search-title">帖子</div>
              <button v-for="post in searchResults.posts" :key="`post-${post.id}`" class="search-item" @click="jumpSearch('#/forum')">
                <span>{{ post.ava }}</span><div><b>{{ post.title }}</b><small>{{ post.cat }} · {{ post.name }}</small></div>
              </button>
            </div>
            <div v-if="searchResults.repos.length" class="search-section">
              <div class="search-title">源码</div>
              <button v-for="repo in searchResults.repos" :key="`repo-${repo.id}`" class="search-item" @click="jumpSearch('#/market')">
                <span>{{ repo.icon }}</span><div><b>{{ repo.name }}</b><small>{{ repo.lang }} · {{ repo.license }}</small></div>
              </button>
            </div>
          </div>
        </div>
        <button class="theme-toggle" :aria-label="theme === 'dark' ? '切换到浅色主题' : '切换到深色主题'" @click="toggleTheme">
          <span class="theme-icon">{{ theme === 'dark' ? '☀' : '☾' }}</span>
          <span class="theme-label">{{ theme === 'dark' ? '浅色' : '深色' }}</span>
        </button>
        <button v-if="!currentUser" class="btn btn-ghost login-btn" @click="openAuth">登录 / 注册</button>
        <button class="btn btn-primary" @click="uploadOpen = true">发布游戏</button>
        <div class="account-wrap">
          <button class="avatar account-avatar" :title="currentUser ? currentUser.name : '点击登录'" @click.stop="toggleAccountMenu">{{ currentUser?.initial || 'L' }}</button>
          <div v-if="accountMenuOpen" class="account-menu" @click.stop>
            <div class="account-head">
              <div class="account-large-avatar">{{ currentUser?.initial }}</div>
              <div><div class="account-name">{{ currentUser?.name }} <span>{{ currentUser?.level }}</span></div><div class="account-method">{{ currentUser?.method }}登录 · 在线</div></div>
            </div>
            <div class="account-list">
              <div class="account-section">我的空间</div>
              <button class="account-item" @click="accountAction('我的创作')">我的创作 <span>›</span></button>
              <button class="account-item" @click="accountAction('我的帖子')">我的帖子 <span>›</span></button>
              <button class="account-item" @click="accountAction('我的收藏')">我的收藏 <span>›</span></button>
              <div class="account-section">账户</div>
              <button class="account-item" @click="accountAction('个人设置')">个人设置 <span>›</span></button>
              <button class="account-item account-logout" @click="logout">退出登录</button>
            </div>
          </div>
        </div>
      </div>
    </header>

    <main>
      <div v-if="pageLoading" class="loading-ribbon">
        <span></span>
        <strong>正在加载客户端数据</strong>
      </div>
      <div v-if="pageError" class="error-ribbon">
        <span>!</span>
        <strong>{{ pageError }}</strong>
        <button class="game-act-btn" @click="loadClientData">重新加载</button>
      </div>
      <HomePage
        v-if="activeView === 'home'"
        :home-stats="homeStats"
        :games="homeGames.length ? homeGames : gameList"
        :hot-posts="homeHotPosts.length ? homeHotPosts : postList"
        :feed-posts="homeFeedPosts.length ? homeFeedPosts : postList"
        @upload="uploadOpen = true"
        @play-game="openPlayer"
        @source="downloadGameSource"
        @like-post="likePost"
      />
      <GameLibraryPage
        v-else-if="activeView === 'library'"
        :games="gameList"
        @play-game="openPlayer"
        @download-source="downloadGameSource"
      />
      <ForumPage
        v-else-if="activeView === 'forum'"
        v-model:active-forum-cat="activeForumCat"
        v-model:active-bar="activeBar"
        :bars="barList"
        :forum-cats="forumCats"
        :forum-posts="forumPosts"
        @notice="showToast"
        @create-post="openPostModal"
        @like-post="likePost"
      />
      <MarketPage
        v-else-if="activeView === 'market'"
        :repos="repoList"
        :stats="marketStats"
        @download-repo="downloadSourceRepo"
      />
      <DevPage
        v-else-if="activeView === 'dev'"
        v-model:active-doc-key="activeDocKey"
        :dev-docs="devDocsMap"
      />
      <AdminPage
        v-else-if="activeView === 'admin'"
        :current-user="currentUser"
        :is-admin="isAdmin"
        @notice="showToast"
        @request-auth="openAuth"
      />
      <NotFoundPage v-else :path="currentPath" />
    </main>

    <footer class="footer">
      <div class="container footer-grid">
        <div><div class="logo"><span class="logo-mark"><span>◆</span></span><img class="brand-logo footer-brand-logo" :src="ika6Logo" alt="ika6.com" /></div><p>让每一个想法都能变成可玩的作品。开源、独立、互助。</p></div>
        <a href="#/library">游戏库</a>
        <a href="#/market">源码集市</a>
        <a href="#/dev">开发者文档</a>
      </div>
    </footer>

    <UploadGameModal :show="uploadOpen" :submitting="uploadSubmitting" @close="uploadOpen = false" @submit="submitUpload" />

    <CreatePostModal :show="postModalOpen" :submitting="postSubmitting" @close="postModalOpen = false" @submit="submitPost" />

    <AuthModal :show="authOpen" @close="authOpen = false" @authenticated="handleAuthenticated" @notice="showToast" />

    <div v-if="playerOpen && currentGame" class="player-overlay" @click.self="playerOpen = false">
      <div class="player-window">
        <div class="player-bar"><div class="dots"><span></span><span></span><span></span></div><div class="player-title">{{ currentGame.title }}</div><button class="player-close" @click="playerOpen = false">×</button></div>
        <div class="player-canvas">
          <div class="player-demo">
            <div class="logo">{{ currentGame.glyph }}</div>
            <h2>{{ currentGame.title }}</h2>
            <p>WebGL 2.0 · 平均帧率 60 FPS · 体积 {{ currentGame.size }}</p>
            <div class="keys"><span>WASD 移动</span><span>SPACE 跳跃</span><span>SHIFT 冲刺</span><span>ESC 退出</span></div>
          </div>
        </div>
      </div>
    </div>

    <div v-if="toast" class="toast show"><span class="ic">✓</span><span>{{ toast }}</span></div>
  </div>
</template>
