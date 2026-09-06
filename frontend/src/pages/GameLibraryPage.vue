<script setup>
import StateBlock from '../components/StateBlock.vue';

defineProps({
  games: {
    type: Array,
    required: true,
  },
});

const emit = defineEmits(['play-game', 'download-source']);
</script>

<template>
  <section class="view active">
    <div class="container page-hero">
      <div class="crumbs"><a href="#/">首页</a><span>/</span><span>游戏库</span></div>
      <h1 class="page-title">游戏库</h1>
      <p class="page-sub">共 {{ games.length }} 款开源游戏 · 在线试玩 · 源码下载</p>
    </div>
    <div v-if="games.length" class="container library-list">
      <div v-for="game in games" :key="game.title" class="list-row" @click="emit('play-game', game)">
        <div class="list-cover"><span>{{ game.glyph }}</span></div>
        <div class="list-info">
          <div class="list-title">{{ game.title }} <span v-if="game.hasSource" class="src-tag">SRC</span></div>
          <div class="list-desc">{{ game.engine }} 引擎 · {{ game.genre }} · {{ game.license }} 协议 · 由 {{ game.author }} 发布。</div>
          <div class="list-tags"><span class="tag">{{ game.engine }}</span><span class="tag">{{ game.genre }}</span><span class="tag">{{ game.license }}</span></div>
        </div>
        <div class="list-meta">
          <span>▶ {{ game.plays }}</span>
          <button class="game-act-btn play" @click.stop="emit('play-game', game)">试玩</button>
          <button v-if="game.hasSource" class="game-act-btn" @click.stop="emit('download-source', game)">下载</button>
        </div>
      </div>
    </div>
    <div v-else class="container">
      <StateBlock icon="游" title="游戏库暂无内容" text="没有拿到游戏列表时，用户会看到明确提示，而不是空白页面。" />
    </div>
  </section>
</template>
