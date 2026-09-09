import { ApiError, getAuthToken, request, setAuthToken } from './http';
import { normalizeGame } from './normalizers';

const API_BASE_URL = (import.meta.env.VITE_API_BASE_URL || '').replace(/\/$/, '');
const UPLOAD_TIMEOUT = Number(import.meta.env.VITE_UPLOAD_TIMEOUT || 120000);
const TASK_POLL_INTERVAL = Number(import.meta.env.VITE_UPLOAD_TASK_POLL_INTERVAL || 1000);
const TASK_POLL_TIMEOUT = Number(import.meta.env.VITE_UPLOAD_TASK_TIMEOUT || 300000);

function normalizePayload(payload) {
  if (payload && typeof payload === 'object' && 'code' in payload) {
    if (payload.code !== 0) {
      if (payload.code === 401) setAuthToken('');
      throw new ApiError(payload.message || '请求失败，请稍后重试', {
        code: payload.code,
        errorCode: payload.errorCode,
        data: payload.data,
      });
    }

    return payload.data;
  }

  if (payload && typeof payload === 'object' && 'success' in payload) {
    if (!payload.success) {
      throw new ApiError(payload.message || '请求失败，请稍后重试', {
        code: payload.code,
        errorCode: payload.errorCode,
        data: payload.data,
      });
    }

    return payload.data ?? payload.result ?? payload;
  }

  return payload;
}

function buildUploadUrl() {
  return new URL(`${API_BASE_URL}/api/games`, window.location.origin).toString();
}

function getUploadTaskId(payload) {
  return payload?.uploadTaskId || payload?.upload_task_id || payload?.taskId || payload?.task_id || '';
}

function normalizeTaskProgress(raw = {}) {
  const rawPercent = Number(raw.percent ?? raw.progress ?? raw.progressPercent ?? 0);
  const stage = raw.stageText || raw.stage_text || raw.message || raw.stage || '服务器正在处理...';
  const status = String(raw.status || 'processing').toLowerCase();
  const finished = ['completed', 'reviewing', 'success', 'done'].includes(status);

  return {
    percent: finished && (!Number.isFinite(rawPercent) || rawPercent <= 0)
      ? 100
      : (Number.isFinite(rawPercent) ? Math.max(0, Math.min(100, Math.round(rawPercent))) : 0),
    stage,
    status,
  };
}

function sleep(ms) {
  return new Promise((resolve) => {
    window.setTimeout(resolve, ms);
  });
}

async function waitForUploadTask(taskId, onProgress, initialUploadResult) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < TASK_POLL_TIMEOUT) {
    const task = await request(`/api/uploads/tasks/${encodeURIComponent(taskId)}`);
    const progress = normalizeTaskProgress(task);
    const finished = ['completed', 'reviewing', 'success', 'done'].includes(progress.status);
    onProgress?.({
      ...progress,
      percent: finished ? 100 : Math.min(99, 70 + Math.round(progress.percent * 0.3)),
      loaded: undefined,
      total: undefined,
      phase: 'processing',
    });

    if (finished) {
      return task?.result || task?.game || initialUploadResult || task;
    }

    if (progress.status === 'failed' || progress.status === 'error') {
      throw new ApiError(progress.stage || '服务器处理失败，请稍后重试', {
        status: 422,
        code: 'UPLOAD_TASK_FAILED',
        data: task,
      });
    }

    await sleep(TASK_POLL_INTERVAL);
  }

  throw new ApiError('服务器处理超时，请稍后到开发者中心查看作品状态', {
    code: 'UPLOAD_TASK_TIMEOUT',
  });
}

function createGameFormData(payload) {
  const formData = new FormData();
  Object.entries(payload || {}).forEach(([key, value]) => {
    if (Array.isArray(value)) {
      value.forEach((item) => formData.append(key, item));
    } else if (value !== undefined && value !== null) {
      formData.append(key, value);
    }
  });

  return formData;
}

function uploadGameRequest(payload, onProgress) {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const token = getAuthToken();

    xhr.open('POST', buildUploadUrl(), true);
    xhr.timeout = UPLOAD_TIMEOUT;
    if (token) xhr.setRequestHeader('Authorization', `Bearer ${token}`);

    xhr.upload.onprogress = (event) => {
      if (!event.lengthComputable) {
        onProgress?.({ percent: 8, loaded: 0, total: 0, stage: '正在上传游戏文件...', phase: 'uploading' });
        return;
      }

      const uploadPercent = Math.round((event.loaded / event.total) * 100);
      onProgress?.({
        percent: Math.min(70, Math.round(uploadPercent * 0.7)),
        loaded: event.loaded,
        total: event.total,
        stage: '正在上传游戏文件...',
        phase: 'uploading',
      });
    };

    xhr.onload = () => {
      let payloadData = null;
      try {
        payloadData = xhr.responseText ? JSON.parse(xhr.responseText) : null;
      } catch {
        reject(new ApiError('服务器返回的数据格式不正确', { status: xhr.status }));
        return;
      }

      if (xhr.status < 200 || xhr.status >= 300) {
        if (xhr.status === 401) setAuthToken('');
        reject(new ApiError(payloadData?.message || '请求失败，请稍后重试', {
          status: xhr.status,
          code: payloadData?.code,
          errorCode: payloadData?.errorCode,
          data: payloadData?.data,
        }));
        return;
      }

      const data = normalizePayload(payloadData);
      const taskId = getUploadTaskId(data);
      if (taskId) {
        onProgress?.({ percent: 70, loaded: undefined, total: undefined, stage: '文件上传完成，等待服务器处理...', phase: 'processing' });
        waitForUploadTask(taskId, onProgress, data).then(resolve).catch(reject);
        return;
      }

      const status = String(data?.status || '').toLowerCase();
      onProgress?.({
        percent: 100,
        loaded: undefined,
        total: undefined,
        stage: status === 'reviewing' ? '上传完成，已进入审核队列' : '上传请求已完成',
        phase: 'completed',
      });
      resolve(data);
    };

    xhr.onerror = () => reject(new ApiError('网络连接失败，请检查服务是否可用', { code: 'NETWORK_ERROR' }));
    xhr.ontimeout = () => reject(new ApiError('请求超时，请稍后重试', { code: 'TIMEOUT' }));

    onProgress?.({ percent: 3, loaded: 0, total: 0, stage: '准备上传...', phase: 'uploading' });
    xhr.send(createGameFormData(payload));
  });
}

export async function uploadGame(payload, onProgress) {
  return normalizeGame(await uploadGameRequest(payload, onProgress));
}
