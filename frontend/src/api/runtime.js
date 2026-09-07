import { isMockEnabled } from './http';

export const API_FALLBACK_EVENT = 'ika6:api-fallback';

let fallbackActive = false;
let fallbackNotified = false;

function isFallbackCandidate(error) {
  return error?.code === 'NETWORK_ERROR'
    || error?.code === 'TIMEOUT'
    || error?.status === 0
    || error?.status >= 500;
}

export function isFallbackActive() {
  return fallbackActive;
}

export async function requestWithFallback(realRequest, mockRequest) {
  if (isMockEnabled() || fallbackActive) {
    return mockRequest();
  }

  try {
    return await realRequest();
  } catch (error) {
    if (!isFallbackCandidate(error)) throw error;

    fallbackActive = true;

    if (!fallbackNotified) {
      fallbackNotified = true;
      window.dispatchEvent(new CustomEvent(API_FALLBACK_EVENT, {
        detail: { reason: error.code || error.status || 'backend-unavailable' },
      }));
    }

    return mockRequest();
  }
}
