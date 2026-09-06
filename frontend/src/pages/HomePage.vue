<script setup>
import { computed } from 'vue';
import GameCard from '../components/GameCard.vue';
import PostItem from '../components/PostItem.vue';
import StateBlock from '../components/StateBlock.vue';

const props = defineProps({
  homeStats: {
    type: Object,
    required: true,
  },
  games: {
    type: Array,
    required: true,
  },
  hotPosts: {
    type: Array,
    required: true,
  },
  feedPosts: {
    type: Array,
    required: true,
  },
});

const emit = defineEmits(['upload', 'play-game', 'source', 'like-post']);

function compact(value) {
  const number = Number(value) || 0;
  if (number >= 1000) return `${(number / 1000).toFixed(number >= 10000 ? 1 : 0)}k`;
  return String(number);
}

const heroStats = computed(() => [
  { value: compact(props.homeStats.projects), label: '开源作品' },
  { value: compact(props.homeStats.plays), label: '累计试玩' },
  { value: compact(props.homeStats.contributors), label: '贡献者' },
  { value: `${props.homeStats.price ?? 0}元`, label: '永久免费' },
]);
</script>

<template>
  <section class="view active">
    <div class="container home-layout">
      <div class="hero">
        <div class="hero-badge"><span class="dot"></span><span class="mono">{{ homeStats.projects.toLocaleString() }} projects online</span></div>
        <h1><span class="line">开源游戏的</span><span class="line grad">试玩与创作社区</span></h1>
        <p class="lead">上传作品、浏览器试玩、下载源码、贴吧式交流。这个 Vue 版本已经把原型改成组件和状态驱动，方便后面接真实接口。</p>
        <div class="hero-actions">
          <button class="btn btn-primary btn-lg" @click="emit('upload')">上传游戏</button>
          <button class="btn btn-ghost btn-lg" @click="emit('play-game', games[0])">立即试玩</button>
        </div>
        <div class="hero-meta">
          <div v-for="item in heroStats" :key="item.label" class="meta-item"><div class="num">{{ item.value }}</div><div class="lbl">{{ item.label }}</div></div>
        </div>
      </div>
      <aside class="side-stack">
        <button class="upload-cta" @click="emit('upload')">
          <span class="ic">＋</span>
          <strong>发布你的开源游戏</strong>
          <span>拖入构建包，写介绍，进入论坛和游戏库。</span>
        </button>
        <div class="panel">
          <div class="panel-head"><div class="h">24h 热帖</div><a href="#/forum">更多</a></div>
          <div class="panel-body">
            <div v-for="(post, index) in hotPosts.slice(0, 5)" :key="post.title" class="rank-item">
              <div class="rank-num">{{ index + 1 }}</div>
              <div class="t">{{ post.title }}</div>
              <div class="v">{{ post.views }}</div>
            </div>
            <StateBlock v-if="!hotPosts.length" icon="帖" title="暂无热帖" text="等后端返回热门内容后，这里会自动展示。" />
          </div>
        </div>
      </aside>
    </div>

    <div class="container section">
      <div class="section-head">
        <div><h2 class="section-title">最新 <span class="accent">开源游戏</span></h2><p class="section-sub">从原型的游戏卡片拆出的 Vue 组件</p></div>
        <a class="btn btn-ghost" href="#/library">全部游戏</a>
      </div>
      <div v-if="games.length" class="games-grid">
        <GameCard v-for="game in games.slice(0, 6)" :key="game.title" :game="game" @play="emit('play-game', $event)" @source="emit('source', $event)" />
      </div>
      <StateBlock v-else icon="游" title="暂无游戏数据" text="客户端已经准备好接口，等待服务端返回游戏列表。" />
    </div>

    <div class="container section">
      <div class="section-head">
        <div><h2 class="section-title">社区 <span class="accent">帖子流</span></h2><p class="section-sub">点赞和展开回复由 Vue 状态控制</p></div>
      </div>
      <div v-if="feedPosts.length" class="post-list">
        <PostItem v-for="post in feedPosts.slice(0, 4)" :key="post.title" :post="post" @like="emit('like-post', $event)" />
      </div>
      <StateBlock v-else icon="社" title="暂无社区动态" text="后端没有返回帖子时，页面会保持这个空状态，不会白屏。" />
    </div>
  </section>
</template>
