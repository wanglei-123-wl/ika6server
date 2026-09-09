<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import {
  banUser,
  getAdminDashboard,
  getAdminGames,
  getAdminReports,
  getAdminUsers,
  getAuditLogs,
  resolveAdminReport,
  reviewGame,
  unbanUser,
} from '../api';
import StateBlock from '../components/StateBlock.vue';

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
const actionLoadingId = ref('');
const activePanel = ref('queue');
const dashboard = ref({ pendingGames: 0, pendingPosts: 0, activeUsers: 0, reports: 0 });
const games = ref([]);
const users = ref([]);
const usersPage = ref(1);
const usersPageSize = ref(20);
const usersTotal = ref(0);
const usersLoading = ref(false);
const usersError = ref('');
const auditOpen = ref(false);
const auditLogs = ref([]);
const auditPage = ref(1);
const auditPageSize = ref(20);
const auditTotal = ref(0);
const auditLoading = ref(false);
const auditError = ref('');
const reports = ref([]);
const reportsPage = ref(1);
const reportsPageSize = ref(20);
const reportsTotal = ref(0);
const reportsLoading = ref(false);
const reportsError = ref('');
const rejectOpen = ref(false);
const rejectTarget = ref(null);
const rejectReason = ref('包含未授权素材（图片 / 音乐 / 字体）');
const rejectNote = ref('');
const banOpen = ref(false);
const banTarget = ref(null);
const banReason = ref('违反社区规则');
const banDays = ref('7');
const gameDetailOpen = ref(false);
const gameDetail = ref(null);

const rejectReasons = [
  '包含未授权素材（图片 / 音乐 / 字体）',
  '游戏存在严重 Bug，无法正常运行',
  '源码不完整或缺少构建文件',
  '含违规 / 侵权内容',
  '描述与作品不符，涉嫌虚假宣传',
];

const statusMeta = {
  reviewing: { label: '待审核', className: 'tag-pending' },
  published: { label: '已上线', className: 'tag-live' },
  rejected: { label: '已驳回', className: 'tag-rejected' },
  offline: { label: '已下架', className: 'tag-offline' },
  draft: { label: '草稿', className: 'tag-offline' },
};

const pendingGames = computed(() => games.value.filter((game) => game.status === 'reviewing'));
const totalWorks = computed(() => games.value.length);
const totalUsers = computed(() => usersTotal.value);
const todayNewWorks = computed(() => {
  const today = new Date();
  return games.value.filter((game) => {
    if (!game.createdAt) return false;
    const created = new Date(game.createdAt);
    return !Number.isNaN(created.getTime())
      && created.getFullYear() === today.getFullYear()
      && created.getMonth() === today.getMonth()
      && created.getDate() === today.getDate();
  }).length;
});
const userPageCount = computed(() => Math.max(1, Math.ceil(usersTotal.value / usersPageSize.value)));
const auditPageCount = computed(() => Math.max(1, Math.ceil(auditTotal.value / auditPageSize.value)));
const reportPageCount = computed(() => Math.max(1, Math.ceil(reportsTotal.value / reportsPageSize.value)));

const panelTabs = computed(() => [
  { key: 'queue', label: '审核队列', count: pendingGames.value.length },
  { key: 'works', label: '全部作品', count: games.value.length },
  { key: 'users', label: '用户管理', count: totalUsers.value },
  { key: 'posts', label: '帖子审核', count: dashboard.value.pendingPosts },
  { key: 'reports', label: '举报处理', count: dashboard.value.reports },
]);

const statCards = computed(() => [
  { icon: '✓', value: pendingGames.value.length || dashboard.value.pendingGames, label: '待审核作品', tone: 'p4' },
  { icon: '+', value: todayNewWorks.value, label: '今日新增作品', tone: 'p2' },
  { icon: '▣', value: totalWorks.value, label: '累计作品', tone: 'p1' },
  { icon: '♙', value: dashboard.value.activeUsers, label: '活跃用户', tone: 'p5' },
  { icon: '!', value: dashboard.value.reports, label: '待处理举报', tone: 'p3' },
]);

function formatDate(value) {
  if (!value) return '未知时间';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

function formatDateTime(value) {
  if (!value) return '未知时间';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return `${formatDate(value)} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}`;
}

function formatDetails(details) {
  if (!details) return '';
  if (typeof details === 'string') return details;
  try {
    return JSON.stringify(details);
  } catch {
    return String(details);
  }
}

function isBanned(user) {
  if (user?.status === 'banned') return true;
  if (!user?.bannedUntil) return false;
  const until = new Date(user.bannedUntil);
  return !Number.isNaN(until.getTime()) && until.getTime() > Date.now();
}

function canModerateUser(user) {
  return user?.role !== 'admin' && String(user?.id) !== String(props.currentUser?.id);
}

async function loadAdminData() {
  if (!props.isAdmin) return;
  loading.value = true;
  const [statsResult, gamesResult] = await Promise.allSettled([
    getAdminDashboard(),
    getAdminGames('all'),
  ]);

  if (statsResult.status === 'fulfilled') dashboard.value = statsResult.value;
  if (gamesResult.status === 'fulfilled') games.value = gamesResult.value;
  if (statsResult.status === 'rejected' && gamesResult.status === 'rejected') {
    emit('notice', '后台数据加载失败');
  } else if (statsResult.status === 'rejected') {
    emit('notice', statsResult.reason?.message || '后台看板加载失败');
  } else if (gamesResult.status === 'rejected') {
    emit('notice', gamesResult.reason?.message || '作品列表加载失败');
  }
  loading.value = false;
}

async function loadAdminUsers(page = usersPage.value) {
  if (!props.isAdmin) return;
  usersLoading.value = true;
  usersError.value = '';
  try {
    const result = await getAdminUsers(page, usersPageSize.value);
    users.value = result.items;
    usersPage.value = result.page;
    usersTotal.value = result.total;
  } catch (error) {
    users.value = [];
    usersError.value = error?.message || '用户列表加载失败';
  } finally {
    usersLoading.value = false;
  }
}

async function loadAuditLogs(page = auditPage.value) {
  auditLoading.value = true;
  auditError.value = '';
  try {
    const result = await getAuditLogs(page, auditPageSize.value);
    auditLogs.value = result.items;
    auditPage.value = result.page;
    auditTotal.value = result.total;
  } catch (error) {
    auditLogs.value = [];
    auditError.value = error?.message || '操作日志加载失败';
  } finally {
    auditLoading.value = false;
  }
}

async function openAuditLogs() {
  auditOpen.value = true;
  await loadAuditLogs(1);
}

async function loadReports(page = reportsPage.value) {
  reportsLoading.value = true;
  reportsError.value = '';
  try {
    const result = await getAdminReports(page, reportsPageSize.value);
    reports.value = result.items;
    reportsPage.value = result.page;
    reportsTotal.value = result.total;
  } catch (error) {
    reports.value = [];
    reportsError.value = error?.status === 404
      ? '后端举报处理接口暂未实现'
      : (error?.message || '举报列表加载失败');
  } finally {
    reportsLoading.value = false;
  }
}

async function switchPanel(key) {
  activePanel.value = key;
  if (key === 'users') await loadAdminUsers(1);
  if (key === 'reports') await loadReports(1);
}

async function updateGameStatus(game, status, reason, message) {
  actionLoadingId.value = `${status}-${game.id}`;
  try {
    await reviewGame(game.id, { status, reason });
    emit('notice', message);
    gameDetailOpen.value = false;
    await loadAdminData();
  } catch (error) {
    emit('notice', error?.message || '作品状态更新失败');
  } finally {
    actionLoadingId.value = '';
  }
}

function approveGame(game) {
  return updateGameStatus(game, 'approved', '符合发布要求', `已通过「${game.title}」`);
}

function openReject(game) {
  rejectTarget.value = game;
  rejectReason.value = rejectReasons[0];
  rejectNote.value = '';
  rejectOpen.value = true;
}

function closeReject() {
  if (actionLoadingId.value) return;
  rejectOpen.value = false;
  rejectTarget.value = null;
}

async function confirmReject() {
  if (!rejectTarget.value) return;
  const game = rejectTarget.value;
  const reason = rejectNote.value.trim() ? `${rejectReason.value}（${rejectNote.value.trim()}）` : rejectReason.value;
  actionLoadingId.value = `reject-${game.id}`;
  try {
    await reviewGame(game.id, { status: 'rejected', reason });
    emit('notice', `已驳回「${game.title}」`);
    rejectOpen.value = false;
    rejectTarget.value = null;
    await loadAdminData();
  } catch (error) {
    emit('notice', error?.message || '驳回失败');
  } finally {
    actionLoadingId.value = '';
  }
}

function openGameDetail(game) {
  gameDetail.value = game;
  gameDetailOpen.value = true;
}

function closeGameDetail() {
  if (!actionLoadingId.value) gameDetailOpen.value = false;
}

function openBan(user) {
  if (!canModerateUser(user)) {
    emit('notice', '管理员账号不能被封禁');
    return;
  }
  banTarget.value = user;
  banReason.value = user.banReason || '违反社区规则';
  banDays.value = '7';
  banOpen.value = true;
}

function closeBan() {
  if (!actionLoadingId.value) {
    banOpen.value = false;
    banTarget.value = null;
  }
}

async function confirmBan() {
  if (!banTarget.value) return;
  const user = banTarget.value;
  const days = Number(banDays.value);
  if (!Number.isFinite(days) || days <= 0) {
    emit('notice', '请选择有效的封禁时长');
    return;
  }
  actionLoadingId.value = `ban-${user.id}`;
  try {
    const until = new Date(Date.now() + days * 24 * 60 * 60 * 1000).toISOString();
    await banUser(user.id, { reason: banReason.value.trim() || '违反社区规则', until });
    emit('notice', `已封禁用户「${user.name}」${days}天`);
    banOpen.value = false;
    banTarget.value = null;
    await loadAdminUsers(usersPage.value);
  } catch (error) {
    emit('notice', error?.message || '封禁失败');
  } finally {
    actionLoadingId.value = '';
  }
}

async function unbanAccount(user) {
  if (!canModerateUser(user)) return;
  actionLoadingId.value = `unban-${user.id}`;
  try {
    await unbanUser(user.id);
    emit('notice', `已解除「${user.name}」的封禁`);
    await loadAdminUsers(usersPage.value);
  } catch (error) {
    emit('notice', error?.message || '解除封禁失败');
  } finally {
    actionLoadingId.value = '';
  }
}

async function resolveReport(report) {
  if (!report?.id) return;
  actionLoadingId.value = `report-${report.id}`;
  try {
    await resolveAdminReport(report.id, { status: 'resolved' });
    emit('notice', '举报已处理');
    await loadReports(reportsPage.value);
  } catch (error) {
    emit('notice', error?.message || '举报处理失败');
  } finally {
    actionLoadingId.value = '';
  }
}

function changePage(type, direction) {
  const page = type === 'users' ? usersPage.value : type === 'audit' ? auditPage.value : reportsPage.value;
  const last = type === 'users' ? userPageCount.value : type === 'audit' ? auditPageCount.value : reportPageCount.value;
  const next = Math.min(last, Math.max(1, page + direction));
  if (next === page) return;
  if (type === 'users') loadAdminUsers(next);
  if (type === 'audit') loadAuditLogs(next);
  if (type === 'reports') loadReports(next);
}

onMounted(() => {
  if (props.isAdmin) {
    loadAdminData();
    loadAdminUsers(1);
  }
});

watch(() => props.isAdmin, (isAdmin) => {
  if (isAdmin) {
    loadAdminData();
    loadAdminUsers(1);
  }
});
</script>

<template>
  <section class="view active admin-page">
    <div v-if="!currentUser" class="container admin-guard">
      <div class="admin-guard-card">
        <span>LOCK</span>
        <h2>需要先登录</h2>
        <p>管理后台必须识别用户身份。请先登录管理员账号。</p>
        <button class="btn btn-primary" type="button" @click="emit('request-auth')">去登录</button>
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

    <div v-else class="container admin-console">
      <div class="admin-profile">
        <div class="admin-profile-bg"></div>
        <div class="admin-profile-inner">
          <div class="admin-avatar-wrap">
            <div class="admin-avatar">{{ currentUser.initial || 'A' }}</div>
            <div class="admin-avatar-badge">✓</div>
          </div>
          <div class="admin-profile-info">
            <div class="admin-name-row">
              <span class="admin-name">平台管理员</span>
              <span class="admin-shield">超级管理员</span>
              <span class="admin-sys">SYS</span>
            </div>
            <p>负责平台内容审核、作品管理与社区治理，守护开源游戏生态的公平与秩序。</p>
            <div class="admin-profile-meta">全站审核 · 用户治理 · 数据看板 · 最高权限</div>
            <div class="admin-profile-stats">
              <div><b>{{ totalWorks }}</b><span>管理作品</span></div>
              <div><b>{{ totalUsers }}</b><span>注册用户</span></div>
              <div><b>{{ dashboard.pendingGames }}</b><span>待审作品</span></div>
              <div><b>{{ dashboard.reports }}</b><span>待处理举报</span></div>
            </div>
          </div>
          <div class="admin-profile-actions">
            <button class="btn btn-ghost" type="button" @click="openAuditLogs">操作日志</button>
          </div>
        </div>
      </div>

      <div class="admin-dash-strip">
        <div v-for="card in statCards" :key="card.label" class="dash-card">
          <div class="dc-ic" :class="card.tone">{{ card.icon }}</div>
          <div>
            <div class="dc-num">{{ card.value }}</div>
            <div class="dc-lbl">{{ card.label }}</div>
          </div>
        </div>
      </div>

      <div class="admin-tabs">
        <button v-for="tab in panelTabs" :key="tab.key" class="admin-tab-pill" :class="{ active: activePanel === tab.key }" type="button" @click="switchPanel(tab.key)">
          {{ tab.label }} <span>{{ tab.count }}</span>
        </button>
      </div>

      <div v-if="activePanel === 'queue'" class="admin-panel-list">
        <div class="adm-toolbar">
          <span>共 <b>{{ pendingGames.length }}</b> 件作品等待审核</span>
          <span>建议 24 小时内完成处理</span>
          <button class="game-act-btn" type="button" :disabled="loading" @click="loadAdminData">刷新</button>
        </div>
        <div v-if="pendingGames.length" class="adm-list">
          <article v-for="game in pendingGames" :key="game.id" class="adm-row pending">
            <button class="adm-thumb adm-thumb-button" type="button" @click="openGameDetail(game)">
              <img v-if="game.coverUrl" :src="game.coverUrl" :alt="game.title">
              <span v-else>{{ game.glyph }}</span>
            </button>
            <div class="adm-info">
              <div class="adm-title">{{ game.title }} <span class="adm-tag tag-pending">待审核</span></div>
              <div class="adm-meta">
                <span>作者 <b>@{{ game.author }}</b></span>
                <span>{{ game.genre }} · {{ game.engine }}</span>
                <span>{{ game.version }}</span>
                <span>提交于 {{ formatDate(game.updatedAt) }}</span>
              </div>
            </div>
            <div class="adm-actions">
              <button class="btn btn-ghost btn-sm" type="button" @click="openGameDetail(game)">查看</button>
              <button class="btn btn-success btn-sm" type="button" :disabled="actionLoadingId === `approved-${game.id}`" @click="approveGame(game)">通过</button>
              <button class="btn btn-danger btn-sm" type="button" :disabled="actionLoadingId === `reject-${game.id}`" @click="openReject(game)">驳回</button>
            </div>
          </article>
        </div>
        <StateBlock v-else icon="审" title="审核队列为空" text="当前没有服务器返回的待审核作品。" />
      </div>

      <div v-else-if="activePanel === 'works'" class="admin-panel-list">
        <div class="adm-toolbar">
          <span>共 <b>{{ games.length }}</b> 件作品</span>
          <span><span class="adm-tag tag-live">已上线 {{ games.filter((item) => item.status === 'published').length }}</span></span>
          <span><span class="adm-tag tag-pending">待审核 {{ pendingGames.length }}</span></span>
          <button class="game-act-btn" type="button" :disabled="loading" @click="loadAdminData">刷新</button>
        </div>
        <div v-if="games.length" class="adm-list">
          <article v-for="game in games" :key="game.id" class="adm-row">
            <button class="adm-thumb adm-thumb-button" type="button" @click="openGameDetail(game)">
              <img v-if="game.coverUrl" :src="game.coverUrl" :alt="game.title">
              <span v-else>{{ game.glyph }}</span>
            </button>
            <div class="adm-info">
              <div class="adm-title">{{ game.title }} <span class="adm-tag" :class="statusMeta[game.status]?.className">{{ statusMeta[game.status]?.label || game.status }}</span></div>
              <div class="adm-meta">
                <span>作者 <b>@{{ game.author }}</b></span>
                <span>{{ game.genre }} · {{ game.engine }}</span>
                <span>试玩 {{ game.plays }}</span>
                <span>下载 {{ game.downloads }}</span>
              </div>
            </div>
            <div class="adm-actions">
              <button class="btn btn-ghost btn-sm" type="button" @click="openGameDetail(game)">查看</button>
              <button v-if="game.status === 'reviewing'" class="btn btn-success btn-sm" type="button" :disabled="actionLoadingId === `approved-${game.id}`" @click="approveGame(game)">通过</button>
              <button v-if="game.status === 'reviewing'" class="btn btn-danger btn-sm" type="button" :disabled="actionLoadingId === `reject-${game.id}`" @click="openReject(game)">驳回</button>
              <button v-else-if="game.status === 'published'" class="btn btn-danger-soft btn-sm" type="button" :disabled="actionLoadingId === `offline-${game.id}`" @click="updateGameStatus(game, 'offline', '管理员下架', `已下架「${game.title}」`)">下架</button>
              <button v-else-if="game.status === 'offline'" class="btn btn-success btn-sm" type="button" :disabled="actionLoadingId === `published-${game.id}`" @click="updateGameStatus(game, 'published', '管理员恢复上线', `已恢复「${game.title}」上线`)">恢复上线</button>
            </div>
          </article>
        </div>
        <StateBlock v-else icon="作" title="暂无作品数据" text="服务器还没有返回可管理的作品。" />
      </div>

      <div v-else-if="activePanel === 'users'" class="admin-panel-list">
        <div class="adm-toolbar">
          <span>注册用户 <b>{{ usersTotal }}</b></span>
          <span v-if="userPageCount > 1">第 {{ usersPage }} / {{ userPageCount }} 页</span>
          <button class="game-act-btn" type="button" :disabled="usersLoading" @click="loadAdminUsers(usersPage)">刷新</button>
        </div>
        <StateBlock v-if="usersLoading" icon="载" title="正在加载用户" text="正在从服务器读取用户列表。" />
        <StateBlock v-else-if="usersError" icon="错" title="用户列表加载失败" :text="usersError" />
        <div v-else-if="users.length" class="adm-user-list">
          <article v-for="user in users" :key="user.id" class="adm-user-row">
            <div class="adm-user-avatar">{{ user.initial }}</div>
            <div class="adm-user-info">
              <div class="adm-title">{{ user.name }} <span v-if="user.role === 'admin'" class="adm-tag tag-live">管理员</span><span v-if="isBanned(user)" class="adm-tag tag-rejected">已封禁</span></div>
              <div class="adm-meta">
                <span>{{ user.email }}</span>
                <span>{{ user.level }}</span>
                <span>作品 {{ user.games }}</span>
                <span>帖子 {{ user.posts }}</span>
                <span>注册于 {{ formatDate(user.createdAt) }}</span>
              </div>
              <div v-if="isBanned(user)" class="adm-ban-note">封禁至 {{ formatDateTime(user.bannedUntil) }}<template v-if="user.banReason"> · {{ user.banReason }}</template></div>
            </div>
            <div class="adm-actions">
              <button v-if="isBanned(user)" class="btn btn-success btn-sm" type="button" :disabled="actionLoadingId === `unban-${user.id}`" @click="unbanAccount(user)">解除封禁</button>
              <button v-else-if="canModerateUser(user)" class="btn btn-danger-soft btn-sm" type="button" @click="openBan(user)">封禁</button>
              <span v-else class="adm-muted">不可操作</span>
            </div>
          </article>
        </div>
        <StateBlock v-else icon="人" title="暂无用户数据" text="服务器还没有返回用户列表。" />
        <div v-if="!usersLoading && !usersError && userPageCount > 1" class="adm-pagination">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="usersPage <= 1" @click="changePage('users', -1)">上一页</button>
          <span>{{ usersPage }} / {{ userPageCount }}</span>
          <button class="btn btn-ghost btn-sm" type="button" :disabled="usersPage >= userPageCount" @click="changePage('users', 1)">下一页</button>
        </div>
      </div>

      <div v-else-if="activePanel === 'posts'" class="admin-panel-list">
        <StateBlock icon="帖" title="待审核帖子列表尚未接通" text="后端当前只有帖子审核动作接口，没有提供管理员读取待审核帖子列表的接口。" />
      </div>

      <div v-else class="admin-panel-list">
        <div class="adm-toolbar">
          <span>服务器返回 <b>{{ reportsTotal }}</b> 条举报</span>
          <button class="game-act-btn" type="button" :disabled="reportsLoading" @click="loadReports(reportsPage)">刷新</button>
        </div>
        <StateBlock v-if="reportsLoading" icon="载" title="正在加载举报" text="正在从服务器读取举报列表。" />
        <StateBlock v-else-if="reportsError" icon="错" title="举报列表不可用" :text="reportsError" />
        <div v-else-if="reports.length" class="adm-list">
          <article v-for="report in reports" :key="report.id" class="adm-row">
            <div class="adm-info">
              <div class="adm-title">{{ report.type }} <span class="adm-tag tag-pending">{{ report.status }}</span></div>
              <div class="adm-meta"><span>举报人 {{ report.reporter }}</span><span>目标 {{ report.target }}</span><span>{{ formatDateTime(report.createdAt) }}</span></div>
              <div class="adm-ban-note">{{ report.reason || '服务器未返回举报原因' }}</div>
            </div>
            <div class="adm-actions"><button class="btn btn-success btn-sm" type="button" :disabled="actionLoadingId === `report-${report.id}`" @click="resolveReport(report)">处理</button></div>
          </article>
        </div>
        <StateBlock v-else icon="报" title="暂无举报数据" text="服务器当前返回 0 条举报记录；这不是前端生成的占位数据。" />
        <div v-if="!reportsLoading && !reportsError && reportPageCount > 1" class="adm-pagination">
          <button class="btn btn-ghost btn-sm" type="button" :disabled="reportsPage <= 1" @click="changePage('reports', -1)">上一页</button>
          <span>{{ reportsPage }} / {{ reportPageCount }}</span>
          <button class="btn btn-ghost btn-sm" type="button" :disabled="reportsPage >= reportPageCount" @click="changePage('reports', 1)">下一页</button>
        </div>
      </div>
    </div>

    <div v-if="rejectOpen" class="modal-bg" @click.self="closeReject">
      <div class="modal reject-modal">
        <div class="modal-head">
          <h3>驳回作品</h3>
          <button class="icon-btn" type="button" :disabled="Boolean(actionLoadingId)" @click="closeReject">×</button>
        </div>
        <div class="modal-body">
          <div class="field">
            <label>驳回原因（将反馈给作者）</label>
            <div class="reject-reasons">
              <label v-for="reason in rejectReasons" :key="reason" class="rr-item" :class="{ sel: rejectReason === reason }">
                <input v-model="rejectReason" type="radio" :value="reason">
                {{ reason }}
              </label>
            </div>
          </div>
          <div class="field">
            <label for="reject-note">补充说明（可选）</label>
            <textarea id="reject-note" v-model="rejectNote" placeholder="可填写具体问题位置、修改建议..." />
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn btn-ghost" type="button" :disabled="Boolean(actionLoadingId)" @click="closeReject">取消</button>
          <button class="btn btn-danger" type="button" :disabled="Boolean(actionLoadingId)" @click="confirmReject">确认驳回</button>
        </div>
      </div>
    </div>

    <div v-if="gameDetailOpen && gameDetail" class="modal-bg" @click.self="closeGameDetail">
      <div class="modal admin-detail-modal">
        <div class="modal-head">
          <h3>作品详情</h3>
          <button class="icon-btn" type="button" :disabled="Boolean(actionLoadingId)" @click="closeGameDetail">×</button>
        </div>
        <div class="modal-body admin-detail-body">
          <div class="admin-detail-cover">
            <img v-if="gameDetail.coverUrl" :src="gameDetail.coverUrl" :alt="gameDetail.title">
            <span v-else>{{ gameDetail.glyph }}</span>
          </div>
          <div class="admin-detail-content">
            <div class="admin-detail-title">{{ gameDetail.title }} <span class="adm-tag" :class="statusMeta[gameDetail.status]?.className">{{ statusMeta[gameDetail.status]?.label || gameDetail.status }}</span></div>
            <dl class="admin-detail-grid">
              <div><dt>作者</dt><dd>@{{ gameDetail.author }}</dd></div>
              <div><dt>引擎</dt><dd>{{ gameDetail.engine }}</dd></div>
              <div><dt>类型</dt><dd>{{ gameDetail.genre }}</dd></div>
              <div><dt>版本</dt><dd>{{ gameDetail.version }}</dd></div>
              <div><dt>试玩</dt><dd>{{ gameDetail.plays }}</dd></div>
              <div><dt>下载</dt><dd>{{ gameDetail.downloads }}</dd></div>
              <div><dt>提交时间</dt><dd>{{ formatDateTime(gameDetail.createdAt || gameDetail.updatedAt) }}</dd></div>
              <div><dt>驳回原因</dt><dd>{{ gameDetail.reason || '无' }}</dd></div>
            </dl>
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn btn-ghost" type="button" :disabled="Boolean(actionLoadingId)" @click="closeGameDetail">关闭</button>
          <button v-if="gameDetail.status === 'reviewing'" class="btn btn-danger" type="button" :disabled="Boolean(actionLoadingId)" @click="openReject(gameDetail); gameDetailOpen = false">驳回</button>
          <button v-if="gameDetail.status === 'reviewing'" class="btn btn-success" type="button" :disabled="Boolean(actionLoadingId)" @click="approveGame(gameDetail)">通过</button>
          <button v-if="gameDetail.status === 'published'" class="btn btn-danger-soft" type="button" :disabled="Boolean(actionLoadingId)" @click="updateGameStatus(gameDetail, 'offline', '管理员下架', `已下架「${gameDetail.title}」`)">下架</button>
          <button v-if="gameDetail.status === 'offline'" class="btn btn-success" type="button" :disabled="Boolean(actionLoadingId)" @click="updateGameStatus(gameDetail, 'published', '管理员恢复上线', `已恢复「${gameDetail.title}」上线`)">恢复上线</button>
        </div>
      </div>
    </div>

    <div v-if="banOpen && banTarget" class="modal-bg" @click.self="closeBan">
      <div class="modal admin-small-modal">
        <div class="modal-head">
          <h3>封禁用户</h3>
          <button class="icon-btn" type="button" :disabled="Boolean(actionLoadingId)" @click="closeBan">×</button>
        </div>
        <div class="modal-body">
          <p class="modal-note">将封禁用户「{{ banTarget.name }}」，封禁期间该用户不能继续使用需要登录的功能。</p>
          <div class="field">
            <label for="ban-days">封禁时长</label>
            <select id="ban-days" v-model="banDays">
              <option value="1">1 天</option>
              <option value="3">3 天</option>
              <option value="7">7 天</option>
              <option value="30">30 天</option>
              <option value="90">90 天</option>
            </select>
          </div>
          <div class="field">
            <label for="ban-reason">封禁原因</label>
            <textarea id="ban-reason" v-model="banReason" placeholder="填写封禁原因，便于后续追溯..." />
          </div>
        </div>
        <div class="modal-foot">
          <button class="btn btn-ghost" type="button" :disabled="Boolean(actionLoadingId)" @click="closeBan">取消</button>
          <button class="btn btn-danger" type="button" :disabled="Boolean(actionLoadingId)" @click="confirmBan">确认封禁</button>
        </div>
      </div>
    </div>

    <div v-if="auditOpen" class="modal-bg" @click.self="auditOpen = false">
      <div class="modal audit-modal">
        <div class="modal-head">
          <div><h3>管理员操作日志</h3><p class="modal-head-sub">日志来自服务器，不在前端本地生成。</p></div>
          <button class="icon-btn" type="button" @click="auditOpen = false">×</button>
        </div>
        <div class="modal-body">
          <StateBlock v-if="auditLoading" icon="载" title="正在加载日志" text="正在从服务器读取管理员操作记录。" />
          <StateBlock v-else-if="auditError" icon="错" title="日志加载失败" :text="auditError" />
          <div v-else-if="auditLogs.length" class="audit-list">
            <article v-for="log in auditLogs" :key="log.id" class="audit-row">
              <div class="audit-dot"></div>
              <div class="audit-content"><strong>{{ log.action }}</strong><span>{{ log.target }} #{{ log.targetId }} · 操作者 #{{ log.actorId }}</span><small>{{ formatDateTime(log.createdAt) }}<template v-if="log.details"> · {{ formatDetails(log.details) }}</template></small></div>
            </article>
          </div>
          <StateBlock v-else icon="志" title="暂无操作日志" text="服务器当前没有返回日志记录。" />
          <div v-if="!auditLoading && !auditError && auditPageCount > 1" class="adm-pagination">
            <button class="btn btn-ghost btn-sm" type="button" :disabled="auditPage <= 1" @click="changePage('audit', -1)">上一页</button>
            <span>{{ auditPage }} / {{ auditPageCount }}</span>
            <button class="btn btn-ghost btn-sm" type="button" :disabled="auditPage >= auditPageCount" @click="changePage('audit', 1)">下一页</button>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>
