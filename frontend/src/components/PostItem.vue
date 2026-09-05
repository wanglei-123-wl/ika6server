<script setup>
import { ref } from 'vue';

const props = defineProps({
  post: {
    type: Object,
    required: true,
  },
  pinned: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['like']);
const expanded = ref(false);
</script>

<template>
  <article :class="pinned ? 'pin-thread' : 'post-item'" @click="expanded = !expanded">
    <div class="thread-avatar" :style="{ background: post.bg }">{{ post.ava }}</div>
    <div class="post-body">
      <div class="thread-head">
        <span class="thread-name" :class="{ gold: pinned }">{{ post.name }}</span>
        <span v-if="post.level" class="thread-level">{{ post.level }}</span>
        <span class="thread-time">{{ pinned ? '置顶 · 2026-08-01' : post.time }}</span>
      </div>
      <div class="thread-title">
        <span v-if="pinned" class="tag pin">置顶</span>
        <span v-for="tag in post.tags" v-else :key="tag" class="tag">{{ tag }}</span>
        <span>{{ post.title }}</span>
      </div>
      <p class="post-excerpt">{{ post.excerpt }}</p>
      <div v-if="post.media" class="post-media">
        <div class="img">{{ post.media.slice(0, 1) }}</div>
        <div class="cap">{{ post.media }}</div>
      </div>
      <div class="thread-meta">
        <span class="item">💬 {{ post.replies }} 回复</span>
        <span class="item">👁 {{ post.views }}</span>
        <button class="like-btn" :class="{ liked: post.liked }" @click.stop="emit('like', post)">
          ♥ {{ post.likes }}
        </button>
        <span class="post-jump">{{ expanded ? '收起回复区 ▴' : '点击展开回复区 ▾' }}</span>
      </div>
      <div v-if="expanded" class="reply-list">
        <div class="reply-item">
          <div class="r-ava" :style="{ background: post.bg }">{{ post.ava }}</div>
          <div class="r-body">
            <div class="r-head">
              <span class="r-name">{{ post.name }}</span>
              <span class="r-floor">2楼</span>
              <span class="r-time">刚刚</span>
            </div>
            <div class="r-text">这个楼层是 Vue 状态驱动展开的，后面可以接真实评论接口。</div>
          </div>
        </div>
        <div class="more-floors">查看全部回复 ▾</div>
      </div>
    </div>
  </article>
</template>
