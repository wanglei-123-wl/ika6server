<script setup>
import { computed, onMounted, ref } from 'vue';
import {
  createForumReply,
  getForumCommentReplies,
  getForumReplies,
  likeForumReply,
} from '../api';

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
const replyTarget = ref(null);
const loadingChildrenId = ref('');

const composePlaceholder = computed(() => (
  replyTarget.value ? `回复 @${replyTarget.value.author}` : '写下你的评论...'
));

function getRootReply(reply) {
  if (!reply.parentId) return reply;
  return replies.value.find((item) => String(item.id) === String(reply.parentId)) || reply;
}

function buildReplyPayload(text) {
  if (!replyTarget.value) return { content: text };

  const root = getRootReply(replyTarget.value);
  return {
    content: text,
    parentId: root.id,
    replyToCommentId: replyTarget.value.id,
    replyToAuthor: replyTarget.value.author,
  };
}

function appendNestedReply(createdReply) {
  if (!createdReply.parentId) {
    replies.value = [...replies.value, createdReply];
    return;
  }

  replies.value = replies.value.map((reply) => {
    if (String(reply.id) !== String(createdReply.parentId)) return reply;
    return {
      ...reply,
      replies: [...(reply.replies || []), createdReply],
      replyCount: Number(reply.replyCount || 0) + 1,
    };
  });
}

function updateReply(id, updater) {
  replies.value = replies.value.map((reply) => {
    if (String(reply.id) === String(id)) return updater(reply);

    return {
      ...reply,
      replies: (reply.replies || []).map((child) => (
        String(child.id) === String(id) ? updater(child) : child
      )),
    };
  });
}

async function loadReplies() {
  if (!props.post?.id) return;

  loading.value = true;
  error.value = '';

  try {
    const result = await getForumReplies(props.post.id, { page: 1, pageSize: 20 });
    replies.value = result.items || [];
  } catch (loadError) {
    error.value = loadError.message || '评论加载失败';
  } finally {
    loading.value = false;
  }
}

async function submitReply() {
  const text = content.value.trim();

  if (!props.currentUser) {
    emit('request-auth');
    emit('notice', '请先登录后再评论');
    return;
  }

  if (text.length < 2) {
    emit('notice', '评论内容至少 2 个字符');
    return;
  }

  submitting.value = true;

  try {
    const reply = await createForumReply(props.post.id, buildReplyPayload(text));
    appendNestedReply(reply);
    content.value = '';
    replyTarget.value = null;
    emit('replied', { post: props.post, reply });
    emit('notice', '评论已发布');
  } catch (submitError) {
    emit('notice', submitError.message || '评论失败，请稍后重试');
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
  updateReply(reply.id, (item) => ({ ...item, liked: true, likes: previousLikes + 1 }));
  likingReplyId.value = reply.id;

  try {
    const result = await likeForumReply(reply.id);
    updateReply(reply.id, (item) => ({
      ...item,
      liked: typeof result.liked === 'boolean' ? result.liked : item.liked,
      likes: result.likes !== undefined ? Number(result.likes) || 0 : item.likes,
    }));
  } catch (likeError) {
    updateReply(reply.id, (item) => ({ ...item, liked: false, likes: previousLikes }));
    emit('notice', likeError.message || '点赞失败，请稍后重试');
  } finally {
    likingReplyId.value = '';
  }
}

async function loadChildReplies(reply) {
  if (loadingChildrenId.value === reply.id) return;

  loadingChildrenId.value = reply.id;
  try {
    const result = await getForumCommentReplies(reply.id, { page: 1, pageSize: 20 });
    updateReply(reply.id, (item) => ({
      ...item,
      replies: result.items || [],
      replyCount: Math.max(Number(item.replyCount || 0), Number(result.total || 0)),
    }));
  } catch (loadError) {
    emit('notice', loadError.message || '二级回复加载失败');
  } finally {
    loadingChildrenId.value = '';
  }
}

function replyTo(reply) {
  if (!props.currentUser) {
    emit('request-auth');
    emit('notice', '请先登录后再回复');
    return;
  }

  replyTarget.value = reply;
}

onMounted(loadReplies);
</script>

<template>
  <div class="reply-list" @click.stop>
    <div v-if="loading" class="reply-state">
      <span class="mini-spinner"></span>
      <strong>正在加载评论</strong>
    </div>

    <div v-else-if="error" class="reply-state">
      <strong>评论加载失败</strong>
      <button class="game-act-btn" type="button" @click="loadReplies">重试</button>
    </div>

    <template v-else>
      <div v-if="replies.length" class="reply-items">
        <article v-for="reply in replies" :key="reply.id" class="reply-item comment-item">
          <div class="r-ava">{{ reply.avatarText }}</div>
          <div class="r-body">
            <div class="r-head">
              <span class="r-name">{{ reply.author }}</span>
              <span v-if="reply.floor" class="r-floor">{{ reply.floor }}楼</span>
              <span class="r-time">{{ reply.createdAt }}</span>
            </div>
            <div class="r-text">{{ reply.content }}</div>
            <div class="comment-actions">
              <button
                class="reply-like-btn"
                :class="{ liked: reply.liked }"
                type="button"
                :disabled="reply.liked || likingReplyId === reply.id"
                @click="likeReply(reply)"
              >
                ♥ {{ reply.likes }}
              </button>
              <button class="comment-reply-btn" type="button" @click="replyTo(reply)">回复</button>
            </div>

            <div v-if="reply.replies?.length" class="comment-children">
              <article v-for="child in reply.replies" :key="child.id" class="reply-item comment-child">
                <div class="r-ava child-ava">{{ child.avatarText }}</div>
                <div class="r-body">
                  <div class="r-head">
                    <span class="r-name">{{ child.author }}</span>
                    <span v-if="child.replyToAuthor" class="reply-to">回复 @{{ child.replyToAuthor }}</span>
                    <span class="r-time">{{ child.createdAt }}</span>
                  </div>
                  <div class="r-text">{{ child.content }}</div>
                  <div class="comment-actions">
                    <button
                      class="reply-like-btn"
                      :class="{ liked: child.liked }"
                      type="button"
                      :disabled="child.liked || likingReplyId === child.id"
                      @click="likeReply(child)"
                    >
                      ♥ {{ child.likes }}
                    </button>
                    <button class="comment-reply-btn" type="button" @click="replyTo(child)">回复</button>
                  </div>
                </div>
              </article>
            </div>

            <button
              v-if="Number(reply.replyCount || 0) > (reply.replies?.length || 0)"
              class="comment-more-btn"
              type="button"
              :disabled="loadingChildrenId === reply.id"
              @click="loadChildReplies(reply)"
            >
              {{ loadingChildrenId === reply.id ? '加载中...' : `展开 ${Number(reply.replyCount || 0) - (reply.replies?.length || 0)} 条回复` }}
            </button>
          </div>
        </article>
      </div>

      <div v-else class="reply-empty">
        <strong>暂无评论</strong>
        <span>可以抢个前排，给作者一点反馈。</span>
      </div>
    </template>

    <form class="reply-compose" @submit.prevent="submitReply">
      <div v-if="replyTarget" class="compose-target">
        <span>正在回复 @{{ replyTarget.author }}</span>
        <button type="button" @click="replyTarget = null">取消</button>
      </div>
      <textarea
        v-model="content"
        :disabled="submitting"
        maxlength="500"
        :placeholder="composePlaceholder"
      />
      <div class="reply-compose-foot">
        <span>{{ content.trim().length }}/500</span>
        <button class="game-act-btn play" type="submit" :disabled="submitting">
          {{ submitting ? '发布中...' : replyTarget ? '回复' : '评论' }}
        </button>
      </div>
    </form>
  </div>
</template>
