<script setup>
import { computed, ref } from 'vue';
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
});

const emit = defineEmits(['update:activeDocKey', 'upload', 'request-auth', 'notice']);

const activeTab = ref('all');

const tabs = [
  { key: 'all', label: '全部作品' },
  { key: 'published', label: '已发布' },
  { key: 'reviewing', label: '审核中' },
  { key: 'rejected', label: '未通过' },
  { key: 'draft', label: '草稿箱' },
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

const docs = computed(() => ({ ...fallbackDocs, ...props.devDocs }));

const currentDoc = computed(() => docs.value[props.activeDocKey] || {
  title: '内容加载中',
  sub: '正在读取开发者中心内容。',
  steps: [],
  code: '',
});

const works = ref([
  { id: 1, title: '像素猫大冒险', glyph: '▲', cover: 1, status: 'published', genre: '平台跳跃', engine: 'Godot 4.3', version: 'v1.4.2', plays: '8.2k', downloads: '3.1k', likes: '1.2k', date: '2026-08-12 发布' },
  { id: 2, title: '霓虹骑士', glyph: '►', cover: 2, status: 'published', genre: '竞速', engine: 'Godot 4.2', version: 'v2.0.1', plays: '5.4k', downloads: '2.2k', likes: '980', date: '2026-07-30 发布' },
  { id: 3, title: '苔藓纪元', glyph: '✿', cover: 4, status: 'published', genre: '模拟经营', engine: 'Unity 6', version: 'v0.9.7', plays: '2.1k', downloads: '860', likes: '412', date: '2026-07-02 发布' },
  { id: 4, title: '深渊回响 Demo', glyph: '◇', cover: 6, status: 'reviewing', genre: 'Roguelike', engine: 'Godot 4.3', version: 'v0.3.0', progress: 65, eta: '预计 2 小时内完成审核', date: '今天 09:12 提交' },
  { id: 5, title: '末班地铁（重制版）', glyph: '✚', cover: 8, status: 'reviewing', genre: '恐怖解谜', engine: 'Unity 6', version: 'v1.0.0-rc', progress: 30, eta: '预计 6 小时内完成审核', date: '昨天 22:47 提交' },
  { id: 6, title: '星河咖啡馆', glyph: '✿', cover: 5, status: 'rejected', genre: '休闲养成', engine: 'Godot 4.3', version: 'v0.8.0', reason: '含未授权 BGM 素材，请替换为原创或已授权素材后重新提交。', date: '2026-09-06 被驳回' },
  { id: 7, title: '虚空回廊 Prologue', glyph: '◆', cover: 3, status: 'draft', genre: 'RPG', engine: 'Godot 4.3', version: '草稿', date: '昨天 23:41 编辑' },
]);

const statusMeta = {
  published: { label: '已发布', className: 'published' },
  reviewing: { label: '审核中', className: 'reviewing' },
  rejected: { label: '未通过', className: 'rejected' },
  draft: { label: '草稿', className: 'draft' },
};

const profileName = computed(() => props.currentUser?.name || 'Luna');
const profileInitial = computed(() => props.currentUser?.initial || String(profileName.value).charAt(0).toUpperCase() || 'L');
const profileLevel = computed(() => props.currentUser?.level || 'lv7');

const tabCounts = computed(() => works.value.reduce((result, work) => {
  result.all += 1;
  result[work.status] = (result[work.status] || 0) + 1;
  return result;
}, { all: 0, published: 0, reviewing: 0, rejected: 0, draft: 0 }));

const visibleWorks = computed(() => (
  activeTab.value === 'all'
    ? works.value
    : works.value.filter((work) => work.status === activeTab.value)
));

function switchTab(key) {
  activeTab.value = key;
  if (key === 'docs' && !docs.value[props.activeDocKey]) {
    emit('update:activeDocKey', 'quickstart');
  }
}

function selectDoc(key) {
  emit('update:activeDocKey', key);
}

function requestUpload() {
  if (!props.currentUser) {
    emit('request-auth');
    emit('notice', '请先登录后再上传作品');
    return;
  }
  emit('upload');
}

function resubmitWork(work) {
  work.status = 'reviewing';
  work.progress = 5;
  work.eta = '已重新排队，等待审核';
  work.date = '刚刚重新提交';
  emit('notice', `已重新提交「${work.title}」`);
}

function notify(message) {
  emit('notice', message);
}
</script>

<template>
  <section class="view active dev-page">
    <div class="container">
      <div class="profile-banner">
        <div class="pb-bg"></div>
        <div class="pb-inner">
          <div class="pb-ava-wrap">
            <div class="pb-ava">{{ profileInitial }}</div>
            <div class="pb-ava-badge" title="已认证开发者">✓</div>
          </div>
          <div class="pb-info">
            <div class="pb-name-row">
              <span class="pb-name">{{ profileName }}</span>
              <span class="pb-verified">✓ 已认证开发者</span>
              <span class="pb-lv">{{ profileLevel }}</span>
              <button v-if="!currentUser" class="pb-login-hint" type="button" @click="emit('request-auth')">登录后管理你自己的作品</button>
            </div>
            <p class="pb-bio">像素是浪漫的最小单位。独立游戏开发者主页，用来管理作品、审核状态、数据表现和发布文档。</p>
            <div class="pb-addr">IP·上海 · 2025-03 加入 ika6 · 常用引擎 Godot 4 / Unity</div>
            <div class="pb-stats">
              <div class="pb-stat"><b>32</b><span>关注</span></div>
              <div class="pb-stat"><b>18.2k</b><span>粉丝</span></div>
              <div class="pb-stat"><b>124k</b><span>获赞</span></div>
              <div class="pb-stat"><b>{{ tabCounts.all }}</b><span>作品</span></div>
            </div>
          </div>
          <div class="pb-actions">
            <button class="btn btn-primary" type="button" @click="requestUpload"><span>＋</span>上传新作品</button>
            <button class="btn btn-ghost" type="button" @click="notify('资料编辑功能需要后端接口接入后开放')">编辑资料</button>
          </div>
        </div>
      </div>

      <div class="dash-strip">
        <div class="dash-card"><div class="dc-ic p1">▶</div><div><div class="dc-num">12,847</div><div class="dc-lbl">总试玩 <span class="trend up">↑ 12.5%</span></div></div></div>
        <div class="dash-card"><div class="dc-ic p2">↓</div><div><div class="dc-num">8,632</div><div class="dc-lbl">总下载 <span class="trend up">↑ 8.3%</span></div></div></div>
        <div class="dash-card"><div class="dc-ic p3">♥</div><div><div class="dc-num">24.1k</div><div class="dc-lbl">总获赞 <span class="trend up">↑ 21.7%</span></div></div></div>
        <div class="dash-card"><div class="dc-ic p4">¥</div><div><div class="dc-num">¥3,420</div><div class="dc-lbl">赞助收入 <span class="trend up">↑ 5.1%</span></div></div></div>
        <div class="dash-card"><div class="dc-ic p5">＋</div><div><div class="dc-num">+486</div><div class="dc-lbl">本周新增粉丝</div></div></div>
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

      <div v-if="activeTab !== 'docs'" class="works-grid">
        <article v-for="work in visibleWorks" :key="work.id" class="wk-card">
          <div class="wk-cover" :class="`cover-${work.cover}`">
            <span class="glyph">{{ work.glyph }}</span>
            <span class="wk-status" :class="statusMeta[work.status]?.className">{{ statusMeta[work.status]?.label || '未知' }}</span>
          </div>
          <div class="wk-body">
            <h2 class="wk-title">{{ work.title }}</h2>
            <div class="wk-sub">{{ work.genre }} · {{ work.engine }} · {{ work.version }} · {{ work.date }}</div>
            <div v-if="work.status === 'published'" class="wk-stats">
              <span>试玩 <b>{{ work.plays }}</b></span>
              <span>下载 <b>{{ work.downloads }}</b></span>
              <span>点赞 <b>{{ work.likes }}</b></span>
            </div>
            <template v-else-if="work.status === 'reviewing'">
              <div class="wk-progress"><i :style="{ width: `${work.progress}%` }"></i></div>
              <div class="wk-eta">审核进度 {{ work.progress }}% · {{ work.eta }}</div>
            </template>
            <div v-else-if="work.status === 'rejected'" class="wk-reason"><b>驳回原因：</b>{{ work.reason }}</div>
            <div v-else class="wk-eta">{{ work.date }} · 上传未完成</div>
            <div class="wk-actions">
              <button v-if="work.status === 'published'" class="btn btn-ghost" type="button" @click="notify(`「${work.title}」管理面板需要后端接口接入后开放`)">管理</button>
              <button v-if="work.status === 'reviewing'" class="btn btn-ghost" type="button" @click="notify('已提交催审提醒')">催审</button>
              <button v-if="work.status === 'rejected'" class="btn btn-primary" type="button" @click="resubmitWork(work)">重新提交</button>
              <button v-if="work.status === 'draft'" class="btn btn-primary" type="button" @click="requestUpload">继续编辑</button>
            </div>
          </div>
        </article>
        <StateBlock v-if="!visibleWorks.length" icon="作" title="这个分类下暂时没有作品" text="上传或审核完成后，作品会自动出现在对应分类里。" />
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

.pb-ava {
  width: 98px;
  height: 98px;
  border: 4px solid var(--bg);
  border-radius: 50%;
  display: grid;
  place-items: center;
  color: #fff;
  font-size: 40px;
  font-weight: 900;
  background: linear-gradient(135deg, #8b5cf6, #c026d3);
  box-shadow: 0 10px 34px rgba(139, 92, 246, 0.42);
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
