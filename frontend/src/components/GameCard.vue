<script setup>
defineProps({
  game: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(['play', 'source']);
</script>

<template>
  <article class="game-card" @click="emit('play', game)">
    <div class="game-cover" :class="`cover-${game.cover}`">
      <div class="game-badges">
        <span v-if="game.badge === 'new'" class="badge new">NEW</span>
        <span v-if="game.badge === 'hot'" class="badge hot">HOT</span>
        <span v-if="game.badge === 'open'" class="badge open">开源</span>
      </div>
      <span class="glyph">{{ game.glyph }}</span>
    </div>
    <div class="game-info">
      <div class="game-title">
        {{ game.title }}
        <span v-if="game.hasSource" class="src-tag">SRC</span>
      </div>
      <div class="game-author">
        <span>{{ game.author }}</span>
        <span class="dot-sep">·</span>
        <span class="mono">{{ game.engine }}</span>
        <span class="dot-sep">·</span>
        <span class="mono">{{ game.size }}</span>
      </div>
      <div class="game-stats">
        <span class="stat">▶ {{ game.plays }}</span>
        <span class="stat">♥ {{ game.likes }}</span>
        <div class="game-actions">
          <button class="game-act-btn play" @click.stop="emit('play', game)">试玩</button>
          <button v-if="game.hasSource" class="game-act-btn" @click.stop="emit('source', game)">源码</button>
        </div>
      </div>
    </div>
  </article>
</template>
