import { isMockEnabled } from './http';

export const API_FALLBACK_EVENT = 'ika6:api-fallback';

export async function requestWithFallback(realRequest, mockRequest) {
  if (isMockEnabled()) {
    return mockRequest();
  }

  return realRequest();
}
