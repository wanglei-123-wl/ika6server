<script setup>
import { ref } from 'vue';
import ReplyPanel from './ReplyPanel.vue';

const props = defineProps({
  post: {
    type: Object,
    required: true,
  },
  pinned: {
    type: Boolean,
    default: false,
  },
  currentUser: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(['like', 'notice', 'request-auth', 'reply-created']);
const expanded = ref(false);

function canExpandReplies() {
  return !props.pinned && Boolean(props.post?.id);
}

function toggleReplies() {
  if (canExpandReplies()) expanded.value = !expanded.value;
}
</script>

<template>
  <article :class="pinned ? 'pin-thread' : 'post-item'" @click="toggleReplies">
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
        <span v-if="canExpandReplies()" class="post-jump">{{ expanded ? '收起回复区 ▴' : '点击展开回复区 ▾' }}</span>
      </div>
      <ReplyPanel
        v-if="expanded && canExpandReplies()"
        :post="post"
        :current-user="currentUser"
        @notice="emit('notice', $event)"
        @request-auth="emit('request-auth')"
        @replied="emit('reply-created', $event)"
      />
    </div>
  </article>
</template>
