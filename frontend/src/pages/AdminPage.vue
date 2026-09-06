<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { banUser, getAdminDashboard, reviewGame, reviewPost } from '../api';

const props = defineProps({
  currentUser: {
    type: Object,
    default: null,
  },
  isAdmin: {
    type: Boolean,
    required: true,
  },
});

const emit = defineEmits(['notice', 'request-auth']);

const loading = ref(false);
const dashboard = ref({ pendingGames: 0, pendingPosts: 0, activeUsers: 0, reports: 0 });
const activePanel = ref('games');
const pendingGames = ref([
  { id: 901, title: '星坠矿车', author: 'railfox', engine: 'Godot 4', size: '48 MB', status: '待审核' },
  { id: 902, title: '废土电台', author: 'dustfm', engine: 'Unity 6', size: '128 MB', status: '待审核' },
  { id: 903, title: '云上棋局', author: 'cloudrook', engine: 'HTML5', size: '9 MB', status: '待审核' },
]);
const pendingPosts = ref([
  { id: 801, title: '【招募】找像素画师一起做赛博酒馆', author: 'neonrider', cat: '招募', status: '待审核' },
  { id: 802, title: '【源码】轻量背包系统分享', author: 'quietforge', cat: '源码分享', status: '待审核' },
  { id: 803, title: '这个 Boss 数值是不是太变态了', author: 'pixelwitch', cat: '互助', status: '待审核' },
]);
const users = ref([
  { id: 701, name: 'laststop', level: 'lv3', role: 'user', status: '正常', reports: 2 },
  { id: 702, name: 'catloaf', level: 'lv3', role: 'user', status: '正常', reports: 1 },
  { id: 703, name: 'mossybyte', level: 'lv5', role: 'moderator', status: '正常', reports: 0 },
]);

const panels = [
  { key: 'games', label: '游戏审核' },
  { key: 'posts', label: '帖子审核' },
  { key: 'users', label: '用户管理' },
  { key: 'settings', label: '站点配置' },
];

const summaryCards = computed(() => [
  { label: '待审游戏', value: dashboard.value.pendingGames },
  { label: '待审帖子', value: dashboard.value.pendingPosts },
  { label: '活跃用户', value: dashboard.value.activeUsers },
  { label: '举报待处理', value: dashboard.value.reports },
]);

async function loadDashboard() {
  if (!props.isAdmin) return;
  loading.value = true;

  try {
    dashboard.value = await getAdminDashboard();
  } catch {
    emit('notice', '后台数据加载失败');
  } finally {
    loading.value = false;
  }
}

async function handleGameReview(game, status) {
  game.status = status === 'approved' ? '已通过' : '已驳回';
  await reviewGame(game.id, { status });
  emit('notice', `${game.title} ${game.status}`);
}

async function handlePostReview(post, status) {
  post.status = status === 'approved' ? '已通过' : '已隐藏';
  await reviewPost(post.id, { status });
  emit('notice', `${post.title} ${post.status}`);
}

async function handleBanUser(user) {
  user.status = '已禁言';
  await banUser(user.id, { reason: '管理后台手动处理', until: null });
  emit('notice', `${user.name} 已禁言`);
}

onMounted(loadDashboard);

watch(() => props.isAdmin, (isAdmin) => {
  if (isAdmin) loadDashboard();
});
</script>

<template>
  <section class="view active admin-page">
    <div class="container page-hero admin-hero">
      <div class="crumbs"><a href="#/">首页</a><span>/</span><span>管理后台</span></div>
      <h1 class="page-title">站内管理后台</h1>
      <p class="page-sub">游戏审核 · 帖子治理 · 用户管理 · 首页配置，后端接通前先使用 mock 管理流程。</p>
    </div>

    <div v-if="!currentUser" class="container admin-guard">
      <div class="admin-guard-card">
        <span>LOCK</span>
        <h2>需要先登录</h2>
        <p>管理后台必须识别用户身份。现在可以用 mock 登录，后端完成后会改成真实权限校验。</p>
        <button class="btn btn-primary" @click="emit('request-auth')">去登录</button>
      </div>
    </div>

    <div v-else-if="!isAdmin" class="container admin-guard">
      <div class="admin-guard-card">
        <span>403</span>
        <h2>当前账号没有管理员权限</h2>
        <p>普通用户只能使用前台功能，审核、封禁、配置这些操作需要管理员角色。</p>
        <a class="btn btn-ghost" href="#/">回到广场</a>
      </div>
    </div>

    <div v-else class="container admin-layout">
      <div class="admin-summary">
        <div v-for="card in summaryCards" :key="card.label" class="admin-stat">
          <div class="admin-stat-value">{{ card.value }}</div>
          <div class="admin-stat-label">{{ card.label }}</div>
        </div>
      </div>

      <div class="admin-workbench">
        <aside class="admin-side">
          <button v-for="panel in panels" :key="panel.key" class="admin-tab" :class="{ active: activePanel === panel.key }" @click="activePanel = panel.key">
            {{ panel.label }}
          </button>
        </aside>

        <section class="admin-panel">
          <div class="admin-panel-head">
            <div>
              <h2>{{ panels.find((panel) => panel.key === activePanel)?.label }}</h2>
              <p>{{ loading ? '正在同步后台数据...' : '这些操作已经走统一 admin API，后端完成后可直接对接。' }}</p>
            </div>
            <button class="game-act-btn" @click="loadDashboard">刷新</button>
          </div>

          <div v-if="activePanel === 'games'" class="admin-list">
            <article v-for="game in pendingGames" :key="game.id" class="admin-row">
              <div>
                <h3>{{ game.title }} <span>{{ game.status }}</span></h3>
                <p>{{ game.engine }} · {{ game.size }} · 作者 {{ game.author }}</p>
              </div>
              <div class="admin-actions">
                <button class="game-act-btn play" @click="handleGameReview(game, 'approved')">通过</button>
                <button class="game-act-btn" @click="handleGameReview(game, 'rejected')">驳回</button>
              </div>
            </article>
          </div>

          <div v-else-if="activePanel === 'posts'" class="admin-list">
            <article v-for="post in pendingPosts" :key="post.id" class="admin-row">
              <div>
                <h3>{{ post.title }} <span>{{ post.status }}</span></h3>
                <p>{{ post.cat }} · 作者 {{ post.author }}</p>
              </div>
              <div class="admin-actions">
                <button class="game-act-btn play" @click="handlePostReview(post, 'approved')">通过</button>
                <button class="game-act-btn" @click="handlePostReview(post, 'hidden')">隐藏</button>
              </div>
            </article>
          </div>

          <div v-else-if="activePanel === 'users'" class="admin-list">
            <article v-for="user in users" :key="user.id" class="admin-row">
              <div>
                <h3>{{ user.name }} <span>{{ user.status }}</span></h3>
                <p>{{ user.level }} · {{ user.role }} · 举报 {{ user.reports }} 次</p>
              </div>
              <div class="admin-actions">
                <button class="game-act-btn" @click="emit('notice', `${user.name} 资料页即将开放`)">查看</button>
                <button class="game-act-btn play" @click="handleBanUser(user)">禁言</button>
              </div>
            </article>
          </div>

          <div v-else class="admin-settings">
            <div class="admin-config-card">
              <h3>首页推荐位</h3>
              <p>后续可在这里配置首页最新游戏、热门帖子、公告和推荐源码，不再进代码改。</p>
              <button class="game-act-btn play" @click="emit('notice', '配置保存接口待后端接入')">保存配置</button>
            </div>
            <div class="admin-config-card">
              <h3>分类与标签</h3>
              <p>管理游戏分类、论坛分类、帖子标签、源码语言筛选项。</p>
              <button class="game-act-btn" @click="emit('notice', '分类管理即将开放')">管理分类</button>
            </div>
          </div>
        </section>
      </div>
    </div>
  </section>
</template>
