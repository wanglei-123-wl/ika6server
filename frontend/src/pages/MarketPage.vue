<script setup>
import StateBlock from '../components/StateBlock.vue';

defineProps({
  repos: {
    type: Array,
    required: true,
  },
  stats: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(['download-repo']);
</script>

<template>
  <section class="view active">
    <div class="container page-hero">
      <div class="crumbs"><a href="#/">首页</a><span>/</span><span>源码集市</span></div>
      <h1 class="page-title">开源源码集市</h1>
      <p class="page-sub">直接下载项目源码 · 自由改造 · 永久免费</p>
    </div>
    <div class="container market-stats">
      <div class="mstat"><div class="n">{{ stats.projects.toLocaleString() }}</div><div class="l">开源项目</div></div>
      <div class="mstat"><div class="n">{{ stats.weeklyActive.toLocaleString() }}</div><div class="l">本周活跃</div></div>
      <div class="mstat"><div class="n">{{ stats.downloads }}</div><div class="l">累计下载</div></div>
      <div class="mstat"><div class="n">¥{{ stats.price }}</div><div class="l">永远免费</div></div>
    </div>
    <div v-if="repos.length" class="container repo-list">
      <article v-for="repo in repos" :key="repo.name" class="repo-card">
        <div class="repo-top"><div class="repo-icon">{{ repo.icon }}</div><div class="repo-name">{{ repo.name }}</div><span class="repo-license">{{ repo.license }}</span><span class="tag">{{ repo.badge }}</span></div>
        <div class="repo-desc">{{ repo.desc }}</div>
        <div class="repo-meta"><span><i :style="{ background: repo.dots }"></i>{{ repo.lang }}</span><span>★ {{ repo.stars }}</span><span>⑂ {{ repo.forks }}</span><span>▼ {{ repo.downloads }}</span><button class="game-act-btn play" @click="emit('download-repo', repo)">下载源码</button></div>
      </article>
    </div>
    <div v-else class="container">
      <StateBlock icon="源" title="源码集市暂无项目" text="后端没有返回源码项目时，客户端会显示空状态。" />
    </div>
  </section>
</template>
