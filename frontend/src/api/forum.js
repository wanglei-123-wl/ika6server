import { request } from './http';
import {
  mockCreatePost,
  mockCreateReply,
  mockGetBars,
  mockGetPostDetail,
  mockGetPosts,
  mockGetReplies,
  mockLikePost,
  mockLikeReply,
} from './mockAdapter';
import { normalizeBar, normalizePage, normalizePost, normalizeReply } from './normalizers';
import { requestWithFallback } from './runtime';

export async function getForumBars() {
  const result = await requestWithFallback(
    () => request('/api/forum/bars'),
    () => mockGetBars(),
  );
  return (Array.isArray(result) ? result : result?.items || result?.list || []).map(normalizeBar);
}

export async function getForumPosts(params) {
  const result = await requestWithFallback(
    () => request('/api/forum/posts', { params }),
    () => mockGetPosts(params),
  );
  return normalizePage(result, normalizePost);
}

export async function getForumPostDetail(id) {
  return normalizePost(await requestWithFallback(
    () => request(`/api/forum/posts/${id}`),
    () => mockGetPostDetail(id),
  ));
}

export async function createForumPost(payload) {
  return normalizePost(await requestWithFallback(
    () => request('/api/forum/posts', { method: 'POST', body: payload }),
    () => mockCreatePost(payload),
  ));
}

export async function likeForumPost(id) {
  return normalizePost(await requestWithFallback(
    () => request(`/api/forum/posts/${id}/like`, { method: 'POST' }),
    () => mockLikePost(id),
  ));
}

export async function likeForumReply(id) {
  return requestWithFallback(
    () => request(`/api/forum/replies/${id}/like`, { method: 'POST' }),
    () => mockLikeReply(id),
  );
}

export async function getForumReplies(postId, params) {
  const result = await requestWithFallback(
    () => request(`/api/forum/posts/${postId}/replies`, { params }),
    () => mockGetReplies(postId, params),
  );
  return normalizePage(result, normalizeReply);
}

export async function createForumReply(postId, payload) {
  return normalizeReply(await requestWithFallback(
    () => request(`/api/forum/posts/${postId}/replies`, { method: 'POST', body: payload }),
    () => mockCreateReply(postId, payload),
  ));
}
