export const canvas = wx.createCanvas();
const gameGlobal = typeof GameGlobal === 'undefined' ? {} : GameGlobal;
gameGlobal.canvas = canvas;

const windowInfo = wx.getWindowInfo ? wx.getWindowInfo() : wx.getSystemInfoSync();
const pixelRatio = windowInfo.pixelRatio || 1;
canvas.width = windowInfo.screenWidth * pixelRatio;
canvas.height = windowInfo.screenHeight * pixelRatio;
canvas.getContext('2d').scale(pixelRatio, pixelRatio);

export const SCREEN_WIDTH = windowInfo.screenWidth;
export const SCREEN_HEIGHT = windowInfo.screenHeight;
export const PIXEL_RATIO = pixelRatio;
