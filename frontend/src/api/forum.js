import { isMockEnabled, request } from './http';
import {
  mockCreatePost,
  mockCreateReply,
  mockGetBars,
  mockGetPostDetail,
  mockGetPosts,
  mockGetReplies,
  mockLikePost,
} from './mockAdapter';
import { normalizeBar, normalizePage, normalizePost, normalizeReply } from './normalizers';

export async function getForumBars() {
  const result = isMockEnabled() ? await mockGetBars() : await request('/api/forum/bars');
  return (Array.isArray(result) ? result : result?.items || result?.list || []).map(normalizeBar);
}

export async function getForumPosts(params) {
  const result = isMockEnabled() ? await mockGetPosts(params) : await request('/api/forum/posts', { params });
  return normalizePage(result, normalizePost);
}

export async function getForumPostDetail(id) {
  return normalizePost(isMockEnabled() ? await mockGetPostDetail(id) : await request(`/api/forum/posts/${id}`));
}

export async function createForumPost(payload) {
  return normalizePost(isMockEnabled() ? await mockCreatePost(payload) : await request('/api/forum/posts', { method: 'POST', body: payload }));
}

export async function likeForumPost(id) {
  return normalizePost(isMockEnabled() ? await mockLikePost(id) : await request(`/api/forum/posts/${id}/like`, { method: 'POST' }));
}

export async function getForumReplies(postId, params) {
  const result = isMockEnabled() ? await mockGetReplies(postId, params) : await request(`/api/forum/posts/${postId}/replies`, { params });
  return normalizePage(result, normalizeReply);
}

export async function createForumReply(postId, payload) {
  return normalizeReply(isMockEnabled()
    ? await mockCreateReply(postId, payload)
    : await request(`/api/forum/posts/${postId}/replies`, { method: 'POST', body: payload }));
}
