<script setup>
import { onMounted, ref } from 'vue';
import { createForumReply, getForumReplies, likeForumReply } from '../api';

const props = defineProps({
  post: {
    type: Object,
    required: true,
  },
  currentUser: {
    type: Object,
    default: null,
  },
});

const emit = defineEmits(['notice', 'request-auth', 'replied']);

const replies = ref([]);
const loading = ref(false);
const submitting = ref(false);
const likingReplyId = ref('');
const error = ref('');
const content = ref('');

async function loadReplies() {
  if (!props.post?.id) return;

  loading.value = true;
  error.value = '';

  try {
    const result = await getForumReplies(props.post.id, { page: 1, pageSize: 20 });
    replies.value = result.items || [];
  } catch (loadError) {
    error.value = loadError.message || '回复加载失败';
  } finally {
    loading.value = false;
  }
}

async function submitReply() {
  const text = content.value.trim();

  if (!props.currentUser) {
    emit('request-auth');
    emit('notice', '请先登录后再回复');
    return;
  }

  if (text.length < 2) {
    emit('notice', '回复内容至少 2 个字符');
    return;
  }

  submitting.value = true;

  try {
    const reply = await createForumReply(props.post.id, { content: text });
    replies.value = [...replies.value, reply];
    content.value = '';
    emit('replied', { post: props.post, reply });
    emit('notice', '回复已发布');
  } catch (submitError) {
    emit('notice', submitError.message || '回复失败，请稍后重试');
  } finally {
    submitting.value = false;
  }
}

async function likeReply(reply) {
  if (!props.currentUser) {
    emit('request-auth');
    emit('notice', '请先登录后再点赞');
    return;
  }

  if (reply.liked || likingReplyId.value === reply.id) return;

  const previousLikes = Number(reply.likes) || 0;
  reply.liked = true;
  reply.likes = previousLikes + 1;
  likingReplyId.value = reply.id;

  try {
    const result = await likeForumReply(reply.id);
    if (typeof result.liked === 'boolean') reply.liked = result.liked;
    if (result.likes !== undefined) reply.likes = Number(result.likes) || 0;
  } catch (likeError) {
    reply.liked = false;
    reply.likes = previousLikes;
    emit('notice', likeError.message || '点赞失败，请稍后重试');
  } finally {
    likingReplyId.value = '';
  }
}

onMounted(loadReplies);
</script>

<template>
  <div class="reply-list" @click.stop>
    <div v-if="loading" class="reply-state">
      <span class="mini-spinner"></span>
      <strong>正在加载回复</strong>
    </div>

    <div v-else-if="error" class="reply-state">
      <strong>回复加载失败</strong>
      <button class="game-act-btn" type="button" @click="loadReplies">重试</button>
    </div>

    <template v-else>
      <div v-if="replies.length" class="reply-items">
        <article v-for="reply in replies" :key="reply.id" class="reply-item">
          <div class="r-ava">{{ reply.avatarText }}</div>
          <div class="r-body">
            <div class="r-head">
              <span class="r-name">{{ reply.author }}</span>
              <span v-if="reply.floor" class="r-floor">{{ reply.floor }}楼</span>
              <span class="r-time">{{ reply.createdAt }}</span>
            </div>
            <div class="r-text">{{ reply.content }}</div>
            <button
              class="reply-like-btn"
              :class="{ liked: reply.liked }"
              type="button"
              :disabled="reply.liked || likingReplyId === reply.id"
              @click="likeReply(reply)"
            >
              ♥ {{ reply.likes }}
            </button>
          </div>
        </article>
      </div>

      <div v-else class="reply-empty">
        <strong>暂无更多回复</strong>
        <span>可以抢个前排，给作者一点反馈。</span>
      </div>
    </template>

    <form class="reply-compose" @submit.prevent="submitReply">
      <textarea
        v-model="content"
        :disabled="submitting"
        maxlength="500"
        placeholder="写下你的回复..."
      />
      <div class="reply-compose-foot">
        <span>{{ content.trim().length }}/500</span>
        <button class="game-act-btn play" type="submit" :disabled="submitting">
          {{ submitting ? '发布中...' : '回复' }}
        </button>
      </div>
    </form>
  </div>
</template>
