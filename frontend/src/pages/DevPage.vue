<script setup>
import { computed, onMounted, ref, watch } from 'vue';
import { deleteDeveloperGame, getDeveloperCenter, getDeveloperGames, resubmitDeveloperGame, urgeDeveloperGameReview } from '../api';
import StateBlock from '../components/StateBlock.vue';

const props = defineProps({
  devDocs: {
    type: Object,
    required: true,
  },
  activeDocKey: {
    type: String,
    required: true,
  },
  currentUser: {
    type: Object,
    default: null,
  },
  refreshKey: {
    type: Number,
    default: 0,
  },
});

const emit = defineEmits(['update:activeDocKey', 'upload', 'request-auth', 'notice', 'play-game']);

const activeTab = ref('all');
const loading = ref(false);
const actionLoadingId = ref('');
const errorText = ref('');
const profile = ref(null);
const stats = ref({
  following: '0',
  followers: '0',
  totalLikes: '0',
  works: 0,
  totalPlays: '0',
  totalDownloads: '0',
  sponsorIncome: '0',
  weeklyNewFollowers: '0',
});
const works = ref([]);

const tabs = [
  { key: 'all', label: '全部作品' },
  { key: 'published', label: '已发布' },
  { key: 'reviewing', label: '审核中' },
  { key: 'rejected', label: '未通过' },
  { key: 'draft', label: '草稿箱' },
  { key: 'offline', label: '已下架' },
  { key: 'docs', label: '开发者文档' },
];

const docGroups = [
  { title: '入门', keys: ['quickstart', 'upload', 'web'] },
  { title: '进阶', keys: ['sdk', 'monetize', 'tips'] },
];

const fallbackDocs = {
  monetize: {
    title: '免费变现（赞助）',
    sub: '开源项目如何体面地获得支持',
    steps: ['开启吧内赞助', '赞助打榜', '收益归开发者'],
    code: '赞助功能接通后，会在作品管理页展示收入、支持者与提现记录。',
  },
  tips: {
    title: '最佳实践',
    sub: '让玩家更容易理解、试玩和反馈你的作品',
    steps: ['标题说明游戏类型', '封面展示真实画面', '附上玩法说明和源码协议'],
    code: 'README.md\nLICENSE\nscreenshots/\ncontrols.md',
  },
};

const statusMeta = {
  published: { label: '已发布', className: 'published' },
  reviewing: { label: '审核中', className: 'reviewing' },
  rejected: { label: '未通过', className: 'rejected' },
  draft: { label: '草稿', className: 'draft' },
  offline: { label: '已下架', className: 'draft' },
};

const docs = computed(() => ({ ...fallbackDocs, ...props.devDocs }));
const currentDoc = computed(() => docs.value[props.activeDocKey] || {
  title: '内容加载中',
  sub: '正在读取开发者中心内容。',
  steps: [],
  code: '',
});

const displayProfile = computed(() => profile.value || {
  name: props.currentUser?.name || '开发者',
  initial: props.currentUser?.initial || 'U',
  level: props.currentUser?.level || 'lv1',
  verified: false,
  bio: '',
  location: '',
  joinedAt: '',
  engines: [],
});

const tabCounts = computed(() => works.value.reduce((result, work) => {
  result.all += 1;
  result[work.status] = (result[work.status] || 0) + 1;
  return result;
}, { all: 0, published: 0, reviewing: 0, rejected: 0, draft: 0, offline: 0 }));

const visibleWorks = computed(() => (
  activeTab.value === 'all'
    ? works.value
    : works.value.filter((work) => work.status === activeTab.value)
));

function canDeleteWork(work) {
  return ['draft', 'rejected', 'offline'].includes(work?.status);
}

const profileMeta = computed(() => {
  const parts = [];
  if (displayProfile.value.location) parts.push(`IP·${displayProfile.value.location}`);
  if (displayProfile.value.joinedAt) parts.push(`${formatDate(displayProfile.value.joinedAt)} 加入 ika6`);
  if (displayProfile.value.engines?.length) parts.push(`常用引擎 ${displayProfile.value.engines.join(' / ')}`);
  return parts.join(' · ');
});

function formatDate(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}`;
}

function formatWorkDate(value) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`;
}

async function loadCenter() {
  if (!props.currentUser) {
    profile.value = null;
    works.value = [];
    errorText.value = '';
    return;
  }

  loading.value = true;
  errorText.value = '';

  try {
    const data = await getDeveloperCenter(props.currentUser);
    profile.value = data.profile;
    stats.value = data.stats;
    works.value = data.games;
  } catch (error) {
    works.value = [];
    errorText.value = error?.status === 404
      ? '后端还没有提供开发者中心接口'
      : (error?.message || '开发者中心数据加载失败');
  } finally {
    loading.value = false;
  }
}

async function switchTab(key) {
  activeTab.value = key;
  if (key === 'docs') {
    if (!docs.value[props.activeDocKey]) emit('update:activeDocKey', 'quickstart');
    return;
  }
  if (!props.currentUser) return;

  loading.value = true;
  errorText.value = '';
  try {
    works.value = await getDeveloperGames(key);
  } catch (error) {
    works.value = [];
    errorText.value = error?.message || '作品列表加载失败';
  } finally {
    loading.value = false;
  }
}

function requestUpload() {
  if (!props.currentUser) {
    emit('request-auth');
    emit('notice', '请先登录后再上传作品');
    return;
  }
  emit('upload');
}

async function resubmitWork(work) {
  actionLoadingId.value = `resubmit-${work.id}`;
  try {
    await resubmitDeveloperGame(work.id);
    emit('notice', `已重新提交「${work.title}」`);
    await switchTab(activeTab.value);
  } catch (error) {
    emit('notice', error?.message || '重新提交失败');
  } finally {
    actionLoadingId.value = '';
  }
}

async function urgeReview(work) {
  actionLoadingId.value = `urge-${work.id}`;
  try {
    const result = await urgeDeveloperGameReview(work.id);
    emit('notice', result?.changed === false ? '今天已经催过，请明天再试' : '已提交催审提醒');
  } catch (error) {
    emit('notice', error?.message || '催审失败');
  } finally {
    actionLoadingId.value = '';
  }
}

async function deleteWork(work) {
  if (!canDeleteWork(work)) {
    emit('notice', '当前状态不支持删除');
    return;
  }

  const confirmed = window.confirm(`确定删除「${work.title}」吗？删除后不能恢复。`);
  if (!confirmed) return;

  actionLoadingId.value = `delete-${work.id}`;
  try {
    await deleteDeveloperGame(work.id);
    emit('notice', `已删除「${work.title}」`);
    await switchTab(activeTab.value);
  } catch (error) {
    emit('notice', error?.message || '删除失败');
  } finally {
    actionLoadingId.value = '';
  }
}

function selectDoc(key) {
  emit('update:activeDocKey', key);
}

function notice(message) {
  emit('notice', message);
}

onMounted(loadCenter);
watch(() => props.currentUser?.id, loadCenter);
watch(() => props.refreshKey, loadCenter);
</script>

<template>
  <section class="view active dev-page">
    <div class="container">
      <div class="profile-banner">
        <div class="pb-bg"></div>
        <div class="pb-inner">
          <div class="pb-ava-wrap">
            <img v-if="displayProfile.avatarUrl" class="pb-ava-img" :src="displayProfile.avatarUrl" :alt="displayProfile.name">
            <div v-else class="pb-ava">{{ displayProfile.initial }}</div>
            <div v-if="displayProfile.verified" class="pb-ava-badge" title="已认证开发者">✓</div>
          </div>
          <div class="pb-info">
            <div class="pb-name-row">
              <span class="pb-name">{{ displayProfile.name }}</span>
              <span v-if="displayProfile.verified" class="pb-verified">✓ 已认证开发者</span>
              <span class="pb-lv">{{ displayProfile.level }}</span>
              <button v-if="!currentUser" class="pb-login-hint" type="button" @click="emit('request-auth')">登录后管理你自己的作品</button>
            </div>
            <p class="pb-bio">{{ displayProfile.bio || '登录后这里会展示你的开发者资料、作品状态和真实运营数据。' }}</p>
            <div v-if="profileMeta" class="pb-addr">{{ profileMeta }}</div>
            <div class="pb-stats">
              <div class="pb-stat"><b>{{ stats.following }}</b><span>关注</span></div>
              <div class="pb-stat"><b>{{ stats.followers }}</b><span>粉丝</span></div>
              <div class="pb-stat"><b>{{ stats.totalLikes }}</b><span>获赞</span></div>
              <div class="pb-stat"><b>{{ stats.works || tabCounts.all }}</b><span>作品</span></div>
            </div>
          </div>
          <div class="pb-actions">
            <button class="btn btn-primary" type="button" @click="requestUpload"><span>＋</span>上传新作品</button>
            <button class="btn btn-ghost" type="button" @click="notice('资料编辑接口接通后即可开放')">编辑资料</button>
          </div>
        </div>
      </div>

      <div class="dash-strip">
        <div class="dash-card"><div class="dc-ic p1">▶</div><div><div class="dc-num">{{ stats.totalPlays }}</div><div class="dc-lbl">总试玩 <span v-if="stats.playTrend" class="trend">{{ stats.playTrend }}</span></div></div></div>
        <div class="dash-card"><div class="dc-ic p2">↓</div><div><div class="dc-num">{{ stats.totalDownloads }}</div><div class="dc-lbl">总下载 <span v-if="stats.downloadTrend" class="trend">{{ stats.downloadTrend }}</span></div></div></div>
        <div class="dash-card"><div class="dc-ic p3">♥</div><div><div class="dc-num">{{ stats.totalLikes }}</div><div class="dc-lbl">总获赞 <span v-if="stats.likeTrend" class="trend">{{ stats.likeTrend }}</span></div></div></div>
        <div class="dash-card"><div class="dc-ic p4">¥</div><div><div class="dc-num">{{ stats.sponsorIncome }}</div><div class="dc-lbl">赞助收入 <span v-if="stats.incomeTrend" class="trend">{{ stats.incomeTrend }}</span></div></div></div>
        <div class="dash-card"><div class="dc-ic p5">＋</div><div><div class="dc-num">{{ stats.weeklyNewFollowers }}</div><div class="dc-lbl">本周新增粉丝</div></div></div>
      </div>

      <div class="ptab-row">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="ptab"
          :class="{ active: activeTab === tab.key }"
          type="button"
          @click="switchTab(tab.key)"
        >
          {{ tab.label }}
          <span v-if="tab.key !== 'docs'" class="ptab-cnt">{{ tabCounts[tab.key] || 0 }}</span>
        </button>
      </div>

      <div v-if="activeTab !== 'docs'">
        <StateBlock v-if="!currentUser" icon="登" title="请先登录" text="登录后可以查看你自己的作品、审核状态和真实数据。" />
        <StateBlock v-else-if="loading" icon="载" title="正在加载开发者数据" text="正在从服务器读取你的作品和统计信息。" />
        <StateBlock v-else-if="errorText" icon="错" title="开发者数据加载失败" :text="errorText" />
        <div v-else-if="visibleWorks.length" class="works-grid">
          <article v-for="work in visibleWorks" :key="work.id" class="wk-card">
            <div class="wk-cover" :class="`cover-${work.cover}`">
              <img v-if="work.coverUrl" class="wk-cover-img" :src="work.coverUrl" :alt="work.title">
              <span v-else class="glyph">{{ work.glyph }}</span>
              <span class="wk-status" :class="statusMeta[work.status]?.className">{{ statusMeta[work.status]?.label || work.status }}</span>
            </div>
            <div class="wk-body">
              <h2 class="wk-title">{{ work.title }}</h2>
              <div class="wk-sub">{{ work.genre }} · {{ work.engine }}<template v-if="work.version"> · {{ work.version }}</template><template v-if="work.date"> · {{ formatWorkDate(work.date) }}</template></div>
              <div v-if="work.status === 'published'" class="wk-stats">
                <span>试玩 <b>{{ work.plays }}</b></span>
                <span>下载 <b>{{ work.downloads }}</b></span>
                <span>点赞 <b>{{ work.likes }}</b></span>
              </div>
              <template v-else-if="work.status === 'reviewing'">
                <div class="wk-progress"><i :style="{ width: `${work.progress || 0}%` }"></i></div>
                <div class="wk-eta">审核进度 {{ work.progress || 0 }}%<template v-if="work.eta"> · {{ work.eta }}</template></div>
              </template>
              <div v-else-if="work.status === 'rejected'" class="wk-reason"><b>驳回原因：</b>{{ work.reason || '后端暂未返回驳回原因' }}</div>
              <div v-else class="wk-eta">{{ work.date ? formatWorkDate(work.date) : '未发布' }} · 上传未完成</div>
              <div class="wk-actions">
                <button v-if="work.status === 'published'" class="btn btn-ghost" type="button" @click="emit('play-game', work)">查看</button>
                <button v-if="work.status === 'published'" class="btn btn-ghost" type="button" @click="notice('作品编辑接口接通后即可开放')">管理</button>
                <button v-if="work.status === 'reviewing'" class="btn btn-ghost" type="button" :disabled="actionLoadingId === `urge-${work.id}`" @click="urgeReview(work)">催审</button>
                <button v-if="work.status === 'rejected'" class="btn btn-primary" type="button" :disabled="actionLoadingId === `resubmit-${work.id}`" @click="resubmitWork(work)">重新提交</button>
                <button v-if="work.status === 'draft'" class="btn btn-primary" type="button" @click="requestUpload">继续编辑</button>
                <button v-if="canDeleteWork(work)" class="btn btn-danger-soft" type="button" :disabled="actionLoadingId === `delete-${work.id}`" @click="deleteWork(work)">删除</button>
              </div>
            </div>
          </article>
        </div>
        <StateBlock v-else icon="作" title="暂无作品" text="服务器还没有返回你的作品。上传新作品后会出现在这里。" />
      </div>

      <div v-else-if="Object.keys(docs).length" class="dev-layout">
        <aside class="dev-side">
          <div v-for="group in docGroups" :key="group.title" class="grp">
            <div class="grp-h">{{ group.title }}</div>
            <button
              v-for="key in group.keys.filter((item) => docs[item])"
              :key="key"
              class="dev-link"
              :class="{ active: activeDocKey === key }"
              type="button"
              @click="selectDoc(key)"
            >
              <span class="ic">▸</span>{{ docs[key].title }}
            </button>
          </div>
        </aside>
        <article class="dev-doc">
          <h2>{{ currentDoc.title }}</h2>
          <p class="doc-sub">{{ currentDoc.sub }}</p>
          <div class="step-flow">
            <div v-for="(step, index) in currentDoc.steps" :key="step" class="step-item">
              <div class="step-num">{{ index + 1 }}</div>
              <div>
                <div class="s-t">{{ step }}</div>
                <div class="s-d">按照该步骤完成配置后，再进入下一步。</div>
              </div>
            </div>
          </div>
          <div v-if="currentDoc.code" class="doc-code">
            <div class="cb-head"><span>参考</span><span>可按项目实际情况调整</span></div>
            <pre>{{ currentDoc.code }}</pre>
          </div>
        </article>
      </div>
      <div v-else>
        <StateBlock icon="文" title="开发者内容加载中" text="开发者中心内容会通过接口提供，当前暂无可展示数据。" />
      </div>
    </div>
  </section>
</template>

<style scoped>
.dev-page {
  padding-top: 32px;
}

.profile-banner {
  position: relative;
  margin-bottom: 20px;
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 18px;
  background: var(--surface);
}

.pb-bg {
  height: 168px;
  background:
    radial-gradient(620px 220px at 12% -10%, rgba(139, 92, 246, 0.58), transparent 60%),
    radial-gradient(520px 220px at 88% 110%, rgba(34, 211, 238, 0.36), transparent 60%),
    radial-gradient(420px 200px at 55% 0%, rgba(236, 72, 153, 0.32), transparent 65%),
    linear-gradient(135deg, #312e81, #1e1b4b);
}

.pb-bg::after {
  content: "";
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  background-size: 34px 34px;
}

.pb-inner {
  position: relative;
  display: flex;
  align-items: flex-end;
  gap: 22px;
  margin-top: -44px;
  padding: 0 28px 24px;
  flex-wrap: wrap;
}

.pb-ava-wrap {
  position: relative;
  flex: 0 0 auto;
}

.pb-ava,
.pb-ava-img {
  width: 98px;
  height: 98px;
  border: 4px solid var(--bg);
  border-radius: 50%;
  box-shadow: 0 10px 34px rgba(139, 92, 246, 0.42);
}

.pb-ava {
  display: grid;
  place-items: center;
  color: #fff;
  font-size: 40px;
  font-weight: 900;
  background: linear-gradient(135deg, #8b5cf6, #c026d3);
}

.pb-ava-img {
  display: block;
  object-fit: cover;
}

.pb-ava-badge {
  position: absolute;
  right: 0;
  bottom: 6px;
  width: 28px;
  height: 28px;
  border: 3px solid var(--bg);
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #fff;
  font-size: 12px;
  font-weight: 900;
  background: linear-gradient(135deg, #10b981, #059669);
}

.pb-info {
  flex: 1;
  min-width: 240px;
}

.pb-name-row,
.pb-actions,
.pb-stats,
.dash-card,
.wk-stats,
.wk-actions {
  display: flex;
  align-items: center;
}

.pb-name-row {
  gap: 10px;
  flex-wrap: wrap;
}

.pb-name {
  color: var(--text);
  font-size: 24px;
  font-weight: 900;
}

.pb-verified,
.pb-lv,
.pb-login-hint {
  border-radius: 99px;
  padding: 3px 10px;
  font-size: 11px;
  font-weight: 700;
}

.pb-verified {
  border: 1px solid rgba(251, 191, 36, 0.35);
  color: #fbbf24;
  background: rgba(251, 191, 36, 0.12);
}

.pb-lv {
  border: 1px solid rgba(139, 92, 246, 0.35);
  color: #a5b4fc;
  background: rgba(139, 92, 246, 0.15);
  font-family: "JetBrains Mono", monospace;
}

.pb-login-hint {
  border: 1px dashed rgba(34, 211, 238, 0.45);
  color: #67e8f9;
  background: rgba(34, 211, 238, 0.1);
  cursor: pointer;
}

.pb-bio {
  margin: 8px 0 4px;
  color: var(--text-2);
  font-size: 13px;
  line-height: 1.6;
}

.pb-addr {
  color: var(--text-3);
  font-size: 12px;
}

.pb-stats {
  gap: 26px;
  margin-top: 12px;
}

.pb-stat b,
.dc-num,
.wk-stats b {
  color: var(--text);
  font-family: "JetBrains Mono", monospace;
  font-weight: 800;
}

.pb-stat b {
  display: block;
  font-size: 18px;
}

.pb-stat span,
.dc-lbl,
.wk-sub,
.wk-eta {
  color: var(--text-3);
  font-size: 11px;
}

.pb-actions {
  gap: 10px;
  padding-bottom: 6px;
}

.dash-strip {
  display: grid;
  grid-template-columns: repeat(5, 1fr);
  gap: 14px;
  margin: 20px 0 24px;
}

.dash-card {
  gap: 14px;
  padding: 16px 18px;
  border: 1px solid var(--border);
  border-radius: 14px;
  background: var(--surface);
}

.dc-ic {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  font-weight: 900;
}

.dc-ic.p1 { color: #a78bfa; background: rgba(139, 92, 246, 0.16); }
.dc-ic.p2 { color: #22d3ee; background: rgba(34, 211, 238, 0.14); }
.dc-ic.p3 { color: #f472b6; background: rgba(236, 72, 153, 0.14); }
.dc-ic.p4 { color: #fbbf24; background: rgba(251, 191, 36, 0.14); }
.dc-ic.p5 { color: #34d399; background: rgba(16, 185, 129, 0.14); }

.dc-num {
  font-size: 20px;
  font-weight: 900;
}

.trend {
  margin-left: 4px;
  color: #f87171;
  font-family: "JetBrains Mono", monospace;
  font-weight: 700;
}

.ptab-row {
  display: flex;
  gap: 4px;
  margin-bottom: 22px;
  border-bottom: 1px solid var(--border);
  flex-wrap: wrap;
}

.ptab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  border: 0;
  border-bottom: 2px solid transparent;
  color: var(--text-2);
  background: transparent;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.ptab.active {
  color: var(--text);
  border-bottom-color: var(--primary);
}

.ptab-cnt {
  padding: 1px 8px;
  border-radius: 99px;
  color: var(--text-3);
  background: var(--surface-2);
  font-family: "JetBrains Mono", monospace;
  font-size: 11px;
  font-weight: 700;
}

.ptab.active .ptab-cnt {
  color: #c4b5fd;
  background: rgba(139, 92, 246, 0.2);
}

.works-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 18px;
}

.wk-card {
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 14px;
  background: var(--surface);
}

.wk-cover {
  position: relative;
  min-height: 150px;
  display: grid;
  place-items: center;
  background:
    radial-gradient(circle at 24% 22%, rgba(255, 255, 255, 0.2), transparent 20%),
    linear-gradient(135deg, rgba(139, 92, 246, 0.52), rgba(34, 211, 238, 0.36));
}

.wk-cover-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.cover-2 { background: linear-gradient(135deg, rgba(244, 63, 94, 0.45), rgba(245, 158, 11, 0.36)); }
.cover-3 { background: linear-gradient(135deg, rgba(34, 211, 238, 0.42), rgba(59, 130, 246, 0.38)); }
.cover-4 { background: linear-gradient(135deg, rgba(16, 185, 129, 0.42), rgba(34, 211, 238, 0.34)); }
.cover-5 { background: linear-gradient(135deg, rgba(236, 72, 153, 0.38), rgba(251, 191, 36, 0.34)); }
.cover-6 { background: linear-gradient(135deg, rgba(15, 23, 42, 0.7), rgba(139, 92, 246, 0.38)); }
.cover-8 { background: linear-gradient(135deg, rgba(30, 41, 59, 0.7), rgba(244, 63, 94, 0.38)); }

.glyph {
  color: rgba(255, 255, 255, 0.92);
  font-size: 44px;
  font-weight: 900;
}

.wk-status {
  position: absolute;
  top: 10px;
  left: 10px;
  z-index: 2;
  padding: 3px 10px;
  border-radius: 99px;
  font-size: 11px;
  font-weight: 700;
}

.wk-status.published { color: #fff; background: rgba(16, 185, 129, 0.92); }
.wk-status.reviewing { color: #1f1300; background: rgba(245, 158, 11, 0.92); }
.wk-status.rejected { color: #fff; background: rgba(244, 63, 94, 0.92); }
.wk-status.draft { color: #fff; background: rgba(100, 116, 139, 0.92); }

.wk-body {
  padding: 14px 16px 16px;
}

.wk-title {
  margin: 0;
  color: var(--text);
  font-size: 15px;
  font-weight: 800;
}

.wk-sub {
  margin: 5px 0 12px;
  line-height: 1.5;
}

.wk-stats {
  gap: 14px;
  color: var(--text-2);
  font-size: 12px;
  flex-wrap: wrap;
}

.wk-progress {
  height: 6px;
  border: 1px solid var(--border);
  border-radius: 99px;
  overflow: hidden;
  background: var(--surface-2);
}

.wk-progress i {
  display: block;
  height: 100%;
  border-radius: 99px;
  background: linear-gradient(90deg, #f59e0b, #fbbf24);
}

.wk-eta {
  margin-top: 6px;
  color: #fbbf24;
}

.wk-reason {
  padding: 9px 11px;
  border: 1px solid rgba(244, 63, 94, 0.3);
  border-radius: 10px;
  color: #fda4af;
  background: rgba(244, 63, 94, 0.07);
  font-size: 12px;
  line-height: 1.6;
}

.wk-reason b {
  color: #fb7185;
}

.wk-actions {
  gap: 8px;
  margin-top: 12px;
}

.wk-actions .btn {
  padding: 6px 14px;
  font-size: 12px;
}

.btn-danger-soft {
  border: 1px solid rgba(244, 63, 94, 0.34);
  color: #fb7185;
  background: rgba(244, 63, 94, 0.08);
}

.btn-danger-soft:hover:not(:disabled) {
  border-color: rgba(244, 63, 94, 0.56);
  color: #fff;
  background: rgba(244, 63, 94, 0.76);
}

.dev-layout {
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 24px;
}

.dev-side,
.dev-doc {
  border: 1px solid var(--border);
  background: var(--surface);
}

.dev-side {
  position: sticky;
  top: 80px;
  align-self: start;
  border-radius: 14px;
  padding: 10px;
}

.grp {
  padding: 4px 0;
}

.grp-h {
  padding: 8px 10px 6px;
  color: var(--text-3);
  font-size: 11px;
  font-weight: 700;
}

.dev-link {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  border: 0;
  border-radius: 10px;
  color: var(--text-2);
  background: transparent;
  font-size: 13px;
  font-weight: 500;
  text-align: left;
  cursor: pointer;
}

.dev-link.active,
.dev-link:hover {
  color: var(--text);
  background: var(--surface-2);
}

.ic {
  width: 20px;
  color: var(--text-3);
}

.dev-doc {
  min-width: 0;
  border-radius: 16px;
  padding: 32px;
}

.dev-doc h2 {
  margin: 0 0 6px;
  color: var(--text);
  font-size: 26px;
  font-weight: 900;
}

.doc-sub {
  margin: 0 0 28px;
  color: var(--text-3);
  font-size: 14px;
}

.step-flow {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 32px;
}

.step-item {
  display: flex;
  gap: 16px;
}

.step-num {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  flex: 0 0 auto;
  color: #fff;
  background: var(--gradient-1);
  font-size: 16px;
  font-weight: 900;
}

.s-t {
  margin-bottom: 4px;
  color: var(--text);
  font-size: 16px;
  font-weight: 700;
}

.s-d {
  color: var(--text-2);
  font-size: 13px;
  line-height: 1.6;
}

.doc-code {
  overflow: hidden;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: var(--bg-1);
}

.cb-head {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  padding: 10px 16px;
  border-bottom: 1px solid var(--border);
  color: var(--text-3);
  font-size: 12px;
}

.doc-code pre {
  margin: 0;
  padding: 18px;
  color: var(--text-2);
  white-space: pre-wrap;
}

@media (max-width: 1080px) {
  .dash-strip,
  .works-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 920px) {
  .dev-layout {
    grid-template-columns: 1fr;
  }

  .dev-side {
    position: static;
  }
}

@media (max-width: 640px) {
  .dev-page {
    padding-top: 18px;
  }

  .pb-inner {
    padding: 0 18px 20px;
  }

  .pb-stats {
    gap: 18px;
    flex-wrap: wrap;
  }

  .pb-actions,
  .dash-strip,
  .works-grid {
    width: 100%;
  }

  .dash-strip,
  .works-grid {
    grid-template-columns: 1fr;
  }

  .dev-doc {
    padding: 22px;
  }
}
</style>
