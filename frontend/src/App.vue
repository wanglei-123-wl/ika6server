<script setup>
import { computed, onMounted, onUnmounted, ref } from 'vue';
import AppModal from './components/AppModal.vue';
import AuthModal from './components/AuthModal.vue';
import GameCard from './components/GameCard.vue';
import PostItem from './components/PostItem.vue';
import { bars, devDocs, games, posts as seedPosts, repos } from './data/mock';
import ika6Logo from './assets/ika6-logo.png';

const routes = [
  { key: 'home', label: '广场', hash: '#/' },
  { key: 'library', label: '游戏库', hash: '#/library' },
  { key: 'forum', label: '论坛', hash: '#/forum' },
  { key: 'market', label: '源码集市', hash: '#/market' },
  { key: 'dev', label: '开发者', hash: '#/dev' },
];

const activeView = ref('home');
const activeForumCat = ref('全部');
const activeBar = ref(bars[0]);
const activeDocKey = ref('quickstart');
const uploadOpen = ref(false);
const authOpen = ref(false);
const accountMenuOpen = ref(false);
const playerOpen = ref(false);
const currentGame = ref(games[0]);
const toast = ref('');
const postList = ref(seedPosts.map((post) => ({ ...post })));
const theme = ref('dark');
const currentUser = ref(null);

const forumCats = ['全部', '精华', '互助', '作品发布', '招募', '源码分享', '灌水'];
const currentDoc = computed(() => devDocs[activeDocKey.value]);
const forumPosts = computed(() => (
  activeForumCat.value === '全部'
    ? postList.value
    : postList.value.filter((post) => post.cat === activeForumCat.value)
));

function route() {
  const key = window.location.hash.replace('#/', '') || 'home';
  activeView.value = routes.some((item) => item.key === key) ? key : 'home';
}

function openPlayer(game) {
  currentGame.value = game;
  playerOpen.value = true;
}

function showToast(message) {
  toast.value = message;
  window.clearTimeout(showToast.timer);
  showToast.timer = window.setTimeout(() => {
    toast.value = '';
  }, 2400);
}

function likePost(post) {
  post.liked = !post.liked;
  post.likes += post.liked ? 1 : -1;
}

function submitUpload() {
  uploadOpen.value = false;
  showToast('游戏已提交，等待自动部署');
}

function openAuth() {
  accountMenuOpen.value = false;
  authOpen.value = true;
}

function handleAuthenticated(user) {
  currentUser.value = user;
  if (user.remember !== false) window.localStorage.setItem('pf_user', JSON.stringify(user));
  authOpen.value = false;
  showToast(user.method === '注册' ? `注册成功，欢迎加入 ${user.name}` : `欢迎回来，${user.name}`);
}

function toggleAccountMenu() {
  if (!currentUser.value) {
    openAuth();
    return;
  }
  accountMenuOpen.value = !accountMenuOpen.value;
}

function logout() {
  currentUser.value = null;
  accountMenuOpen.value = false;
  window.localStorage.removeItem('pf_user');
  showToast('已退出登录');
}

function accountAction(label) {
  accountMenuOpen.value = false;
  showToast(`${label}功能即将开放`);
}

function closeAccountMenu(event) {
  if (!event.target.closest('.account-wrap')) accountMenuOpen.value = false;
}

function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark';
  window.localStorage.setItem('pixel-forge-theme', theme.value);
}

function onKeydown(event) {
  if (event.key === 'Escape') {
    uploadOpen.value = false;
    authOpen.value = false;
    accountMenuOpen.value = false;
    playerOpen.value = false;
  }
}

onMounted(() => {
  theme.value = window.localStorage.getItem('pixel-forge-theme') || 'dark';
  try {
    const saved = JSON.parse(window.localStorage.getItem('pf_user') || 'null');
    if (saved?.name) currentUser.value = saved;
  } catch {
    currentUser.value = null;
  }
  route();
  window.addEventListener('hashchange', route);
  window.addEventListener('keydown', onKeydown);
  window.addEventListener('click', closeAccountMenu);
});

onUnmounted(() => {
  window.removeEventListener('hashchange', route);
  window.removeEventListener('keydown', onKeydown);
  window.removeEventListener('click', closeAccountMenu);
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
          <a v-for="routeItem in routes" :key="routeItem.key" :href="routeItem.hash" :class="{ active: activeView === routeItem.key }">
            {{ routeItem.label }}
          </a>
        </nav>
        <div class="search-box">
          <span>⌕</span>
          <input placeholder="搜索游戏、源码、帖子..." />
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
      <section v-if="activeView === 'home'" class="view active">
        <div class="container home-layout">
          <div class="hero">
            <div class="hero-badge"><span class="dot"></span><span class="mono">12,847 projects online</span></div>
            <h1><span class="line">开源游戏的</span><span class="line grad">试玩与创作社区</span></h1>
            <p class="lead">上传作品、浏览器试玩、下载源码、贴吧式交流。这个 Vue 版本已经把原型改成组件和状态驱动，方便后面接真实接口。</p>
            <div class="hero-actions">
              <button class="btn btn-primary btn-lg" @click="uploadOpen = true">上传游戏</button>
              <button class="btn btn-ghost btn-lg" @click="openPlayer(games[0])">立即试玩</button>
            </div>
            <div class="hero-meta">
              <div class="meta-item"><div class="num">12.8k</div><div class="lbl">开源作品</div></div>
              <div class="meta-item"><div class="num">847k</div><div class="lbl">累计试玩</div></div>
              <div class="meta-item"><div class="num">2.1k</div><div class="lbl">贡献者</div></div>
              <div class="meta-item"><div class="num">0元</div><div class="lbl">永久免费</div></div>
            </div>
          </div>
          <aside class="side-stack">
            <button class="upload-cta" @click="uploadOpen = true">
              <span class="ic">＋</span>
              <strong>发布你的开源游戏</strong>
              <span>拖入构建包，写介绍，进入论坛和游戏库。</span>
            </button>
            <div class="panel">
              <div class="panel-head"><div class="h">24h 热帖</div><a href="#/forum">更多</a></div>
              <div class="panel-body">
                <div v-for="(post, index) in postList.slice(0, 5)" :key="post.title" class="rank-item">
                  <div class="rank-num">{{ index + 1 }}</div>
                  <div class="t">{{ post.title }}</div>
                  <div class="v">{{ post.views }}</div>
                </div>
              </div>
            </div>
          </aside>
        </div>

        <div class="container section">
          <div class="section-head">
            <div><h2 class="section-title">最新 <span class="accent">开源游戏</span></h2><p class="section-sub">从原型的游戏卡片拆出的 Vue 组件</p></div>
            <a class="btn btn-ghost" href="#/library">全部游戏</a>
          </div>
          <div class="games-grid">
            <GameCard v-for="game in games.slice(0, 6)" :key="game.title" :game="game" @play="openPlayer" @source="showToast(`源码已打包：${$event.engine} ${$event.size}`)" />
          </div>
        </div>

        <div class="container section">
          <div class="section-head">
            <div><h2 class="section-title">社区 <span class="accent">帖子流</span></h2><p class="section-sub">点赞和展开回复由 Vue 状态控制</p></div>
          </div>
          <div class="post-list">
            <PostItem v-for="post in postList.slice(0, 4)" :key="post.title" :post="post" @like="likePost" />
          </div>
        </div>
      </section>

      <section v-else-if="activeView === 'library'" class="view active">
        <div class="container page-hero">
          <div class="crumbs"><a href="#/">首页</a><span>/</span><span>游戏库</span></div>
          <h1 class="page-title">游戏库</h1>
          <p class="page-sub">共 {{ games.length }} 款开源游戏 · 在线试玩 · 源码下载</p>
        </div>
        <div class="container library-list">
          <div v-for="game in games" :key="game.title" class="list-row" @click="openPlayer(game)">
            <div class="list-cover"><span>{{ game.glyph }}</span></div>
            <div class="list-info">
              <div class="list-title">{{ game.title }} <span v-if="game.hasSource" class="src-tag">SRC</span></div>
              <div class="list-desc">{{ game.engine }} 引擎 · {{ game.genre }} · {{ game.license }} 协议 · 由 {{ game.author }} 发布。</div>
              <div class="list-tags"><span class="tag">{{ game.engine }}</span><span class="tag">{{ game.genre }}</span><span class="tag">{{ game.license }}</span></div>
            </div>
            <div class="list-meta">
              <span>▶ {{ game.plays }}</span>
              <button class="game-act-btn play" @click.stop="openPlayer(game)">试玩</button>
              <button v-if="game.hasSource" class="game-act-btn" @click.stop="showToast('源码已下载')">下载</button>
            </div>
          </div>
        </div>
      </section>

      <section v-else-if="activeView === 'forum'" class="view active">
        <div class="container page-hero">
          <div class="crumbs"><a href="#/">首页</a><span>/</span><span>论坛</span></div>
          <h1 class="page-title">社区论坛</h1>
          <p class="page-sub">贴吧式交流 · 每个游戏一个吧 · 自由盖楼发帖</p>
        </div>
        <div class="container forum-banner">
          <div class="avatar-big">{{ activeBar.icon }}</div>
          <div class="info">
            <h2>{{ activeBar.name }} <span>贴吧 · 官方区</span></h2>
            <div class="sub">{{ activeBar.desc }}</div>
            <div class="stats"><div><b>{{ activeBar.posts }}</b><span>帖子</span></div><div><b>{{ activeBar.members }}</b><span>吧友</span></div><div><b>3,847</b><span>在线</span></div><div><b>96</b><span>吧务</span></div></div>
          </div>
        </div>
        <div class="container forum-layout">
          <div>
            <div class="forum-cats">
              <button v-for="cat in forumCats" :key="cat" class="cat-btn" :class="{ active: activeForumCat === cat }" @click="activeForumCat = cat">{{ cat }}</button>
            </div>
            <PostItem :post="{ ava: '◆', bg: 'var(--gradient-3)', name: '吧务 · PIXEL FORGE 官方', tags: ['置顶'], title: '【吧规】发帖必读：晒作品、求帮助、找队友都到这里来', excerpt: '发布游戏请附截图和介绍；求助和招募请标注分类；开源作品请注明协议。', replies: '108', views: '9.8k', likes: 999, liked: true }" pinned />
            <div class="post-list">
              <PostItem v-for="post in forumPosts" :key="post.title" :post="post" @like="likePost" />
            </div>
          </div>
          <aside class="side-panel">
            <div class="panel">
              <div class="panel-head"><div class="h">游戏吧</div><a href="#" @click.prevent="showToast('查看更多吧')">全部</a></div>
              <div class="panel-body">
                <button v-for="bar in bars" :key="bar.name" class="tieba-row" :class="{ active: activeBar.name === bar.name }" @click="activeBar = bar; showToast(`已切换到「${bar.name}」`)">
                  <div class="tieba-ico">{{ bar.icon }}</div>
                  <div class="tieba-body"><div class="tieba-name">{{ bar.name }} <span v-if="bar.hot" class="tieba-tag">热</span></div><div class="tieba-desc">{{ bar.desc }}</div><div class="tieba-stats">{{ bar.posts }} 帖 · {{ bar.members }} 吧友</div></div>
                </button>
              </div>
            </div>
            <button class="upload-cta" @click="uploadOpen = true"><span class="ic">＋</span><strong>发一篇帖子</strong><span>晒作品、分享源码、招募队友。</span></button>
          </aside>
        </div>
      </section>

      <section v-else-if="activeView === 'market'" class="view active">
        <div class="container page-hero">
          <div class="crumbs"><a href="#/">首页</a><span>/</span><span>源码集市</span></div>
          <h1 class="page-title">开源源码集市</h1>
          <p class="page-sub">直接下载项目源码 · 自由改造 · 永久免费</p>
        </div>
        <div class="container market-stats">
          <div class="mstat"><div class="n">12,847</div><div class="l">开源项目</div></div>
          <div class="mstat"><div class="n">3,214</div><div class="l">本周活跃</div></div>
          <div class="mstat"><div class="n">847K</div><div class="l">累计下载</div></div>
          <div class="mstat"><div class="n">¥0</div><div class="l">永远免费</div></div>
        </div>
        <div class="container repo-list">
          <article v-for="repo in repos" :key="repo.name" class="repo-card">
            <div class="repo-top"><div class="repo-icon">{{ repo.icon }}</div><div class="repo-name">{{ repo.name }}</div><span class="repo-license">{{ repo.license }}</span><span class="tag">{{ repo.badge }}</span></div>
            <div class="repo-desc">{{ repo.desc }}</div>
            <div class="repo-meta"><span><i :style="{ background: repo.dots }"></i>{{ repo.lang }}</span><span>★ {{ repo.stars }}</span><span>⑂ {{ repo.forks }}</span><span>▼ {{ repo.downloads }}</span><button class="game-act-btn play" @click="showToast(`正在克隆 ${repo.name} 到本地`)">下载源码</button></div>
          </article>
        </div>
      </section>

      <section v-else-if="activeView === 'dev'" class="view active">
        <div class="container page-hero">
          <div class="crumbs"><a href="#/">首页</a><span>/</span><span>开发者</span></div>
          <h1 class="page-title">开发者中心</h1>
          <p class="page-sub">上传指南 · SDK 文档 · API 参考 · 从 0 到 1 发布你的开源游戏</p>
        </div>
        <div class="container dev-layout">
          <aside class="dev-side">
            <button v-for="(doc, key) in devDocs" :key="key" class="dev-link" :class="{ active: activeDocKey === key }" @click="activeDocKey = key">{{ doc.title }}</button>
          </aside>
          <article class="dev-content">
            <h2>{{ currentDoc.title }}</h2>
            <p>{{ currentDoc.sub }}</p>
            <div class="step-flow">
              <div v-for="(step, index) in currentDoc.steps" :key="step" class="step-item"><div class="step-num">{{ index + 1 }}</div><div>{{ step }}</div></div>
            </div>
            <pre class="code-block">{{ currentDoc.code }}</pre>
          </article>
        </div>
      </section>
    </main>

    <footer class="footer">
      <div class="container footer-grid">
        <div><div class="logo"><span class="logo-mark"><span>◆</span></span><img class="brand-logo footer-brand-logo" :src="ika6Logo" alt="ika6.com" /></div><p>让每一个想法都能变成可玩的作品。开源、独立、互助。</p></div>
        <a href="#/library">游戏库</a>
        <a href="#/market">源码集市</a>
        <a href="#/dev">开发者文档</a>
      </div>
    </footer>

    <AppModal :show="uploadOpen" title="发布游戏到 PIXEL FORGE" @close="uploadOpen = false" @submit="submitUpload">
      <div class="field"><label>游戏标题</label><input placeholder="给你的作品起个名字..." /></div>
      <div class="field"><label>一句话简介</label><input placeholder="一句话讲清你做了什么..." /></div>
      <div class="field"><label>详细介绍</label><textarea placeholder="聊聊你的开发故事、玩法、坑点..." /></div>
      <div class="drop-zone"><div class="ico">⇧</div><div class="t">拖入文件 或 点击选择</div><div class="s">最大 500 MB · 自动部署到 CDN</div></div>
      <div class="field"><label>开源协议</label><select><option>MIT</option><option>Apache 2.0</option><option>GPL v3</option><option>CC0</option></select></div>
    </AppModal>

    <AuthModal :show="authOpen" @close="authOpen = false" @authenticated="handleAuthenticated" @notice="showToast" />

    <div v-if="playerOpen" class="player-overlay" @click.self="playerOpen = false">
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
