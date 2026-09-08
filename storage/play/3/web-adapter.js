(function installWechatGameAdapter() {
  const touchHandlers = {
    start: [],
    move: [],
    end: [],
  };

  function getCanvasPoint(event) {
    const canvas = document.querySelector('canvas');
    const rect = canvas ? canvas.getBoundingClientRect() : { left: 0, top: 0 };
    return {
      x: event.clientX - rect.left,
      y: event.clientY - rect.top,
    };
  }

  function toTouchEvent(event) {
    const point = getCanvasPoint(event);
    const touch = {
      x: point.x,
      y: point.y,
      clientX: event.clientX,
      clientY: event.clientY,
      pageX: event.pageX,
      pageY: event.pageY,
    };
    return {
      touches: event.type === 'pointerup' ? [] : [touch],
      changedTouches: [touch],
    };
  }

  function bindPointer(type, handlers) {
    window.addEventListener(type, (event) => {
      if (event.pointerType === 'mouse' && event.button !== 0 && type === 'pointerdown') return;
      event.preventDefault();
      const payload = toTouchEvent(event);
      handlers.slice().forEach((handler) => handler(payload));
    }, { passive: false });
  }

  bindPointer('pointerdown', touchHandlers.start);
  bindPointer('pointermove', touchHandlers.move);
  bindPointer('pointerup', touchHandlers.end);
  bindPointer('pointercancel', touchHandlers.end);

  const storage = {
    get(key) {
      const value = window.localStorage.getItem(key);
      return value == null ? null : JSON.parse(value);
    },
    set(key, value) {
      window.localStorage.setItem(key, JSON.stringify(value));
    },
  };

  window.wx = {
    createCanvas() {
      const canvas = document.createElement('canvas');
      canvas.id = 'game-canvas';
      canvas.setAttribute('aria-label', '动物混战游戏画布');
      document.getElementById('game-root').appendChild(canvas);
      return canvas;
    },
    getWindowInfo() {
      return {
        screenWidth: window.innerWidth,
        screenHeight: window.innerHeight,
        pixelRatio: Math.min(window.devicePixelRatio || 1, 2),
      };
    },
    getSystemInfoSync() {
      return this.getWindowInfo();
    },
    getStorageSync(key) {
      try {
        return storage.get(key);
      } catch {
        return null;
      }
    },
    setStorageSync(key, value) {
      storage.set(key, value);
    },
    onTouchStart(handler) {
      touchHandlers.start.push(handler);
    },
    onTouchMove(handler) {
      touchHandlers.move.push(handler);
    },
    onTouchEnd(handler) {
      touchHandlers.end.push(handler);
    },
    onHide(handler) {
      document.addEventListener('visibilitychange', () => {
        if (document.hidden) handler();
      });
    },
    onShow(handler) {
      document.addEventListener('visibilitychange', () => {
        if (!document.hidden) handler();
      });
    },
  };
})();
