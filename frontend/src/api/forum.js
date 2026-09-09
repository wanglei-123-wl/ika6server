import { ApiError, request } from './http';
import { normalizeBar, normalizePage, normalizePost, normalizeReply } from './normalizers';

export async function getForumBars() {
  const result = await request('/api/forum/bars');
  return (Array.isArray(result) ? result : result?.items || result?.list || []).map(normalizeBar);
}

export async function getForumPosts(params) {
  const result = await request('/api/forum/posts', { params });
  return normalizePage(result, normalizePost);
}

export async function getForumPostDetail(id) {
  return normalizePost(await request(`/api/forum/posts/${id}`));
}

export async function createForumPost(payload) {
  return normalizePost(await request('/api/forum/posts', { method: 'POST', body: payload }));
}

export async function likeForumPost(id) {
  return normalizePost(await request(`/api/forum/posts/${id}/like`, { method: 'POST' }));
}

export async function likeForumReply(id) {
  try {
    return await request(`/api/forum/comments/${id}/like`, { method: 'POST' });
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      return request(`/api/forum/replies/${id}/like`, { method: 'POST' });
    }
    throw error;
  }
}

export async function getForumReplies(postId, params) {
  let result;
  try {
    result = await request(`/api/forum/posts/${postId}/comments`, { params });
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) {
      result = await request(`/api/forum/posts/${postId}/replies`, { params });
    } else {
      throw error;
    }
  }
  return normalizePage(result, normalizeReply);
}

export async function getForumCommentReplies(commentId, params) {
  const result = await request(`/api/forum/comments/${commentId}/replies`, { params });
  return normalizePage(result, normalizeReply);
}

export async function createForumReply(postId, payload) {
  let result;
  try {
    result = await request(`/api/forum/posts/${postId}/comments`, { method: 'POST', body: payload });
  } catch (error) {
    if (error instanceof ApiError && error.status === 404 && !payload.parentId && !payload.replyToCommentId) {
      result = await request(`/api/forum/posts/${postId}/replies`, { method: 'POST', body: payload });
    } else {
      throw error;
    }
  }
  return normalizeReply(result);
}
