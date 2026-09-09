import { itemView, qualityDefinition, resolveItemId } from './items.js';

const COLORS = {
  background: '#050608',
  header: '#090b0f',
  surface: '#0d1015',
  cell: '#11161c',
  cellAlt: '#0c1116',
  current: '#1a232c',
  line: '#28323b',
  lineSoft: '#1a232b',
  text: '#edf1f3',
  muted: '#a8b1b8',
  faint: '#68747e',
  player: '#ffffff',
  accent: '#9eb4c2',
};

function roundedRect(ctx, x, y, width, height, radius) {
  const r = Math.min(radius, width / 2, height / 2);
  ctx.beginPath();
  ctx.moveTo(x + r, y);
  ctx.arcTo(x + width, y, x + width, y + height, r);
  ctx.arcTo(x + width, y + height, x, y + height, r);
  ctx.arcTo(x, y + height, x, y, r);
  ctx.arcTo(x, y, x + width, y, r);
  ctx.closePath();
}

export default class GameUI {
  constructor(width, height) {
    this.locations = [
      ['洞府', '', '', '', '宗门', '', '', ''],
      ['', '青云秘境', '', '', '', '', '玄冰窟', ''],
      ['', '血煞洞', '', '', '', '古战场', '', ''],
      ['', '', '', '雷鸣谷', '', '', '', '万兽岭'],
      ['天火秘境', '', '', '', '', '黄泉地宫', '', ''],
      ['', '', '地脉矿窟', '', '', '', '', ''],
      ['', '迷雾禁地', '', '', '幽冥墓', '', '', ''],
      ['', '', '', '', '', '', '妖王巢穴', ''],
      ['', '', '断魂崖', '', '', '', '', ''],
    ];
    this.rows = 9;
    this.cols = 8;
    this.caveEntrance = { row: this.rows - 1, col: this.cols - 1 };
    this.resize(width, height);
  }

  resize(width, height) {
    this.width = width;
    this.height = height;
    this.headerHeight = Math.max(54, Math.min(70, height * 0.095));
    this.controlHeight = Math.max(172, Math.min(206, height * 0.255));
    this.mapTop = this.headerHeight + 12;
    this.mapBottom = height - this.controlHeight - 12;
    this.mapRect = {
      x: 12,
      y: this.mapTop,
      width: width - 24,
      height: Math.max(1, this.mapBottom - this.mapTop),
    };
    const availableMapHeight = this.mapRect.height;
    const squareSize = Math.min(this.mapRect.width / this.cols, availableMapHeight / this.rows);
    this.cellWidth = squareSize;
    this.cellHeight = squareSize;
    this.mapRect.width = squareSize * this.cols;
    this.mapRect.height = squareSize * this.rows;
    this.mapRect.x = (width - this.mapRect.width) / 2;
    this.mapRect.y = this.mapTop + (availableMapHeight - this.mapRect.height) / 2;
    this.controlRect = {
      x: 12,
      y: height - this.controlHeight,
      width: width - 24,
      height: this.controlHeight,
    };
    this.padCenter = {
      x: width * 0.78,
      y: this.controlRect.y + this.controlRect.height * 0.58,
    };
    this.padSize = Math.min(132, width * 0.36, this.controlRect.height * 0.7);
    const unit = this.padSize / 3;
    this.controlButtonSize = Math.max(34, unit - 8);
    this.controlButtons = {
      up: { x: this.padCenter.x, y: this.padCenter.y - unit },
      left: { x: this.padCenter.x - unit, y: this.padCenter.y },
      right: { x: this.padCenter.x + unit, y: this.padCenter.y },
      down: { x: this.padCenter.x, y: this.padCenter.y + unit },
    };

    this.cavePadSize = this.padSize;
    const caveUnit = this.cavePadSize / 3;
    this.caveControlButtonSize = this.controlButtonSize;
    this.cavePadCenter = {
      x: width * 0.78,
      y: this.controlRect.y + this.controlRect.height * 0.58,
    };
    this.caveControlButtons = {
      up: { x: this.cavePadCenter.x, y: this.cavePadCenter.y - caveUnit },
      left: { x: this.cavePadCenter.x - caveUnit, y: this.cavePadCenter.y },
      right: { x: this.cavePadCenter.x + caveUnit, y: this.cavePadCenter.y },
      down: { x: this.cavePadCenter.x, y: this.cavePadCenter.y + caveUnit },
    };
    this.padCenter = { ...this.cavePadCenter };
    this.controlButtons = { ...this.caveControlButtons };
  }

  drawText(ctx, text, x, y, size, color, weight = 'normal', align = 'left') {
    ctx.font = `${weight} ${size}px sans-serif`;
    ctx.fillStyle = color;
    ctx.textAlign = align;
    ctx.textBaseline = 'middle';
    ctx.fillText(text, Math.round(x), Math.round(y));
  }

  drawFittedText(ctx, text, x, y, maxWidth, size, color, weight = 'normal', align = 'center') {
    let fittedSize = size;
    ctx.font = `${weight} ${fittedSize}px sans-serif`;
    while (fittedSize > 8 && ctx.measureText(text).width > maxWidth) {
      fittedSize -= 0.5;
      ctx.font = `${weight} ${fittedSize}px sans-serif`;
    }
    this.drawText(ctx, text, x, y, fittedSize, color, weight, align);
  }

  buildingLevelColor(level) {
    const colors = {
      0: '#3b464f',
      1: COLORS.text,
      2: '#70d58c',
      3: '#5da9ff',
      4: '#b68cff',
      5: '#f0c85a',
      6: '#ef6666',
    };
    return colors[Math.max(0, Math.min(6, Number(level) || 0))] || colors[0];
  }

  buildingVisualStyle(level) {
    const normalizedLevel = Math.max(0, Math.min(6, Number(level) || 0));
    if (normalizedLevel === 0) {
      return { name: COLORS.muted, frame: COLORS.line, status: COLORS.faint };
    }
    const color = this.buildingLevelColor(normalizedLevel);
    return { name: color, frame: color, status: color };
  }

  renderBackground(ctx) {
    ctx.fillStyle = COLORS.background;
    ctx.fillRect(0, 0, this.width, this.height);
    ctx.fillStyle = COLORS.header;
    ctx.fillRect(0, 0, this.width, this.headerHeight);
    ctx.fillStyle = COLORS.surface;
    ctx.fillRect(0, this.controlRect.y, this.width, this.controlHeight);
  }

  renderHeader(ctx, state) {
    // 顶部人物信息已移除，储物袋放在地图下方、操控区上方的深黑留白带左侧。
    const bagY = Math.max(this.headerHeight + 8, this.controlRect.y - 40);
    this.inventoryButton = { x: 16, y: bagY, width: 88, height: 30 };
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, this.inventoryButton.x, this.inventoryButton.y, this.inventoryButton.width, this.inventoryButton.height, 5);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();
    this.drawText(ctx, '储物袋', this.inventoryButton.x + this.inventoryButton.width / 2, bagY + 15, 10, COLORS.text, 'bold', 'center');
    ctx.strokeStyle = COLORS.lineSoft;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(12, this.headerHeight - 1);
    ctx.lineTo(this.width - 12, this.headerHeight - 1);
    ctx.stroke();
  }

  cellAtPoint(x, y) {
    const map = this.mapRect;
    if (x < map.x || x > map.x + map.width || y < map.y || y > map.y + map.height) return null;
    const col = Math.floor((x - map.x) / this.cellWidth);
    const row = Math.floor((y - map.y) / this.cellHeight);
    if (row < 0 || row >= this.rows || col < 0 || col >= this.cols) return null;
    return { row, col };
  }

  isWorldLocationHit(x, y, row, col, name) {
    const cell = this.cellAtPoint(x, y);
    return Boolean(cell && cell.row === row && cell.col === col && this.locations[row][col] === name);
  }

  isWorldDungeonEnterHit(x, y) {
    const button = this.worldDungeonEnterButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isCaveActionHit(x, y) {
    const button = this.caveActionButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isCaveCollectHit(x, y) {
    const button = this.caveCollectButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isCaveFeatureHit(x, y) {
    const button = this.caveFeatureButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isCavePanelBackHit(x, y) {
    const button = this.cavePanelBackButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isFeatureActionHit(x, y) {
    const button = this.featureActionButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isFeatureUpgradeHit(x, y) {
    const button = this.featureUpgradeButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isAlchemyUpgradeCloseHit(x, y) {
    const button = this.alchemyUpgradeCloseButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isAlchemyUpgradeConfirmHit(x, y) {
    const button = this.alchemyUpgradeConfirmButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isAlchemyUpgradeCancelHit(x, y) {
    const button = this.alchemyUpgradeCancelButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isAlchemyMaterialBarHit(x, y) {
    const bar = this.alchemyMaterialBar;
    return Boolean(bar && x >= bar.x && x <= bar.x + bar.width && y >= bar.y && y <= bar.y + bar.height);
  }

  isAlchemyHeatHit(x, y) {
    const bar = this.alchemyHeatBar;
    return Boolean(bar && x >= bar.x && x <= bar.x + bar.width && y >= bar.y - 12 && y <= bar.y + bar.height + 12);
  }

  getAlchemyMaterialMaxScroll() {
    const bar = this.alchemyMaterialBar;
    if (!bar) return 0;
    const slotSize = bar.slotSize;
    const gap = bar.gap;
    const totalWidth = bar.slotCount * slotSize + (bar.slotCount - 1) * gap;
    return Math.max(0, totalWidth - bar.width + 8);
  }

  isDungeonActionHit(x, y) {
    const button = this.dungeonActionButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isDungeonBackHit(x, y) {
    const button = this.dungeonBackButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isInventoryButtonHit(x, y) {
    const button = this.inventoryButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isInventoryCloseHit(x, y) {
    const button = this.inventoryCloseButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isInventoryUpgradeHit(x, y) {
    const button = this.inventoryUpgradeButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isAttributeDetailsHit(x, y) {
    const button = this.attributeDetailsButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isAttributeOverlayCloseHit(x, y) {
    const button = this.attributeOverlayCloseButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isBreakthroughHit(x, y) {
    const button = this.breakthroughButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isItemDetailCloseHit(x, y) {
    const button = this.itemDetailCloseButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isItemDetailActionHit(x, y) {
    const button = this.itemDetailActionButton;
    return Boolean(button && x >= button.x && x <= button.x + button.width && y >= button.y && y <= button.y + button.height);
  }

  isInventoryGridHit(x, y) {
    const panel = this.inventoryPanel;
    if (!panel) return false;
    const layout = this.getInventoryLayoutForPanel(panel);
    return x >= panel.x + 10
      && x <= panel.x + panel.width - 10
      && y >= layout.gridY - 8
      && y <= layout.gridY + layout.gridViewportHeight + 8;
  }

  getInventoryItemAtPoint(x, y, state) {
    const index = this.getInventorySlotAtPoint(x, y, state);
    if (!Number.isInteger(index)) return null;
    const items = this.getInventoryItems(state);
    const item = items[index];
    return item ? item.id : null;
  }

  getInventorySlotAtPoint(x, y, state) {
    const panel = this.inventoryPanel;
    if (!panel || !state || !state.inventory) return null;
    const layout = this.getInventoryLayoutForPanel(panel);
    const scrollY = Math.max(0, Math.min(this.getInventoryMaxScroll(state.inventory.level), state.inventory.scrollY || 0));
    const items = this.getInventoryItems(state);
    if (x < layout.gridX || x > layout.gridX + layout.gridWidth || y < layout.gridY || y > layout.gridY + layout.gridViewportHeight) return null;
    const col = Math.floor((x - layout.gridX) / (layout.cellWidth + layout.gap));
    const row = Math.floor((y - layout.gridY + scrollY) / (layout.cellHeight + layout.gap));
    if (col < 0 || col >= layout.columns || row < 0) return null;
    const cellX = layout.gridX + col * (layout.cellWidth + layout.gap);
    const cellY = layout.gridY + row * (layout.cellHeight + layout.gap) - scrollY;
    if (x > cellX + layout.cellWidth || y > cellY + layout.cellHeight) return null;
    return row * layout.columns + col;
  }

  getEquipmentSlotAtPoint(x, y, state) {
    const panel = this.inventoryPanel;
    if (!panel || !state || !state.player) return null;
    const layout = this.getInventoryLayoutForPanel(panel);
    const equipmentGridX = panel.x + 18;
    const equipmentGridY = panel.y + 188;
    const equipmentColumns = 3;
    const equipmentSlots = ['weapon', 'artifact', 'armor', 'storage', 'technique', 'spell'];
    const col = Math.floor((x - equipmentGridX) / (layout.cellWidth + layout.gap));
    const row = Math.floor((y - equipmentGridY) / (layout.cellHeight + layout.gap));
    if (col < 0 || col >= equipmentColumns || row < 0 || row >= 2) return null;
    const cellX = equipmentGridX + col * (layout.cellWidth + layout.gap);
    const cellY = equipmentGridY + row * (layout.cellHeight + layout.gap);
    if (x > cellX + layout.cellWidth || y > cellY + layout.cellHeight) return null;
    return equipmentSlots[row * equipmentColumns + col];
  }

  getEquipmentItemAtPoint(x, y, state) {
    const panel = this.inventoryPanel;
    if (!panel || !state || !state.player) return null;
    const layout = this.getInventoryLayoutForPanel(panel);
    const equipmentGridX = panel.x + 18;
    const equipmentGridY = panel.y + 188;
    const equipmentColumns = 3;
    const equipmentSlots = ['weapon', 'artifact', 'armor', 'storage', 'technique', 'spell'];
    const col = Math.floor((x - equipmentGridX) / (layout.cellWidth + layout.gap));
    const row = Math.floor((y - equipmentGridY) / (layout.cellHeight + layout.gap));
    if (col < 0 || col >= equipmentColumns || row < 0 || row >= 2) return null;
    const cellX = equipmentGridX + col * (layout.cellWidth + layout.gap);
    const cellY = equipmentGridY + row * (layout.cellHeight + layout.gap);
    if (x > cellX + layout.cellWidth || y > cellY + layout.cellHeight) return null;
    const slot = equipmentSlots[row * equipmentColumns + col];
    const item = state.player.equipment && state.player.equipment[slot];
    return resolveItemId(item);
  }

  getInventorySlotCount(level) {
    const normalizedLevel = Math.max(1, Math.min(5, Number.isFinite(level) ? Math.floor(level) : 1));
    return 12 + ((normalizedLevel - 1) * 4);
  }

  getInventoryLayoutForPanel(panel) {
    const columns = 4;
    const gap = 7;
    const sidePadding = 18;
    const cellWidth = Math.max(42, Math.min(66, (panel.width - sidePadding * 2 - gap * (columns - 1) - 12) / columns));
    const cellHeight = cellWidth;
    const gridWidth = (cellWidth * columns) + (gap * (columns - 1));
    const gridX = panel.x + (panel.width - gridWidth) / 2;
    // Reserve room for identity, vitals, six equipment slots, and attributes.
    // The minimum follows the actual cell size so shorter screens retain a
    // usable backpack viewport without shrinking any cells.
    const equipmentBottom = 188 + (cellHeight * 2) + gap + 8;
    const attributesBottom = 342;
    const profileMinimum = Math.max(equipmentBottom, attributesBottom);
    const profileSectionHeight = Math.min(365, Math.max(profileMinimum, panel.height * 0.56));
    const sectionY = panel.y + profileSectionHeight;
    const backpackTitleY = sectionY + 25;
    const gridY = sectionY + 52;
    const gridViewportHeight = Math.min(360, Math.max(1, panel.height - profileSectionHeight - 124));
    return {
      columns,
      gap,
      cellWidth,
      cellHeight,
      gridWidth,
      gridX,
      gridY,
      gridViewportHeight,
      profileSectionHeight,
      sectionY,
      backpackTitleY,
    };
  }

  getInventoryMaxScroll(level) {
    const panel = this.inventoryPanel || {
      x: 12,
      y: this.headerHeight + 10,
      width: this.width - 24,
      height: this.height - this.headerHeight - 20,
    };
    const layout = this.getInventoryLayoutForPanel(panel);
    const rows = Math.ceil(this.getInventorySlotCount(level) / layout.columns);
    const contentHeight = rows * layout.cellHeight + Math.max(0, rows - 1) * layout.gap;
    return Math.max(0, contentHeight - layout.gridViewportHeight);
  }

  getCaveControlDirection(x, y) {
    const half = this.caveControlButtonSize / 2;
    const hitHalf = half + Math.min(9, this.caveControlButtonSize * 0.18);
    const buttons = this.caveControlButtons;
    return ['up', 'down', 'left', 'right'].find((direction) => {
      const button = buttons[direction];
      return Math.abs(x - button.x) <= hitHalf && Math.abs(y - button.y) <= hitHalf;
    }) || null;
  }

  getCaveBuildingAt(x, y, buildings) {
    const cell = this.cellAtPoint(x, y);
    if (!cell) return null;
    return buildings.find((building) => building.row === cell.row && building.col === cell.col) || null;
  }

  renderMap(ctx, state) {
    const map = this.mapRect;
    const realmColor = state.player.realmColor || COLORS.player;
    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, map.x, map.y, map.width, map.height, 6);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();

    for (let row = 0; row < this.rows; row += 1) {
      for (let col = 0; col < this.cols; col += 1) {
        const x = map.x + col * this.cellWidth;
        const y = map.y + row * this.cellHeight;
        const current = row === state.player.row && col === state.player.col;
        // 主角所在格不做特殊高亮，保持与周围地图格一致，只显示主角文字。
        ctx.fillStyle = (row + col) % 2 ? COLORS.cellAlt : COLORS.cell;
        ctx.fillRect(x + 1, y + 1, this.cellWidth - 2, this.cellHeight - 2);
        ctx.strokeStyle = COLORS.lineSoft;
        ctx.lineWidth = 1;
        ctx.strokeRect(x + 0.5, y + 0.5, this.cellWidth, this.cellHeight);

        const location = this.locations[row][col];
        if (location) {
          const locationSize = Math.min(11, this.cellWidth * 0.27);
          const locationY = current ? y + this.cellHeight * 0.24 : y + this.cellHeight / 2;
          this.drawFittedText(
            ctx,
            location,
            x + this.cellWidth / 2,
            locationY,
            this.cellWidth - 6,
            locationSize,
            current ? COLORS.text : COLORS.muted,
            current ? 'bold' : 'normal',
            'center',
          );
        }
        if (current) {
          // 野外地图与洞府地图使用同一套主角字形规格，避免切换场景时体型突变。
          const playerSize = Math.min(18, this.cellHeight * 0.34);
          this.drawText(ctx, state.player.name || '林', x + this.cellWidth / 2, y + this.cellHeight * 0.56, playerSize, realmColor, 'bold', 'center');
        }
      }
    }

  }

  renderInventoryOverlay(ctx, state) {
    const panel = {
      x: 12,
      y: this.headerHeight + 10,
      width: this.width - 24,
      height: this.height - this.headerHeight - 20,
    };
    const level = state.inventory && state.inventory.level ? state.inventory.level : 1;
    this.inventoryPanel = panel;
    const layout = this.getInventoryLayoutForPanel(panel);
    const maxScroll = this.getInventoryMaxScroll(level);
    const scrollY = Math.max(0, Math.min(maxScroll, state.inventory ? state.inventory.scrollY || 0 : 0));
    const items = this.getInventoryItems(state);
    const slotCount = this.getInventorySlotCount(level);

    this.inventoryCloseButton = { x: panel.x + panel.width - 44, y: panel.y + 12, width: 30, height: 30 };
    this.inventoryUpgradeButton = { x: panel.x + panel.width - 118, y: panel.y + panel.height - 48, width: 98, height: 32 };

    ctx.save();
    ctx.fillStyle = 'rgba(0, 0, 0, 0.68)';
    ctx.fillRect(0, 0, this.width, this.height);
    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, panel.x, panel.y, panel.width, panel.height, 8);
    ctx.fill();
    ctx.strokeStyle = COLORS.accent;
    ctx.lineWidth = 1;
    ctx.stroke();

    const realmColor = state.player.realmColor || COLORS.player;
    // 标题保持原有纵向位置，改为面板水平居中；人物信息整体上移，给装备区留出间距。
    this.drawText(ctx, '储物袋', panel.x + panel.width / 2, panel.y + 28, 16, COLORS.text, 'bold', 'center');
    this.drawText(ctx, '人物信息', panel.x + 18, panel.y + 48, 10, COLORS.muted, 'bold');
    this.drawText(ctx, '×', this.inventoryCloseButton.x + 15, this.inventoryCloseButton.y + 15, 20, COLORS.muted, 'normal', 'center');

    const cultivation = state.resources ? state.resources.cultivation || 0 : 0;
    const profileTop = panel.y + 62;
    const profileTextX = panel.x + 98;
    const profileTextWidth = Math.max(90, panel.width - 116);
    const avatarX = panel.x + 49;
    const avatarY = profileTop + 31;
    this.drawText(ctx, state.player.name || '林', avatarX, avatarY, 36, realmColor, 'bold', 'center');
    this.drawText(ctx, state.player.realm || '凡人', profileTextX, profileTop + 2, 13, realmColor, 'bold', 'left');
    this.drawFittedText(ctx, `修为 ${state.player.realmProgress || cultivation}`, profileTextX, profileTop + 24, profileTextWidth, 9, COLORS.muted, 'normal', 'left');
    this.drawFittedText(ctx, `下境界：${state.player.nextRealm || '已至顶峰'}`, profileTextX, profileTop + 43, profileTextWidth, 8, COLORS.faint, 'normal', 'left');

    const maxHp = Math.max(1, Number(state.player.maxHp) || 100);
    const rawHp = Number(state.player.hp);
    const hp = Math.max(0, Math.min(maxHp, Number.isFinite(rawHp) ? rawHp : maxHp));
    const qi = Math.max(0, Number(state.qi) || 0);
    const qiCap = Math.max(1, Number(state.qiCap) || 100);
    const barX = profileTextX;
    const barWidth = Math.max(92, profileTextWidth - 4);
    const drawStatusBar = (label, valueText, ratio, y, color) => {
      this.drawText(ctx, label, barX, y, 8, COLORS.muted, 'bold', 'left');
      const trackX = barX + 28;
      const trackWidth = Math.max(58, barWidth - 28);
      ctx.fillStyle = COLORS.cell;
      roundedRect(ctx, trackX, y - 4, trackWidth, 8, 4);
      ctx.fill();
      ctx.fillStyle = color;
      roundedRect(ctx, trackX, y - 4, Math.max(4, trackWidth * Math.max(0, Math.min(1, ratio))), 8, 4);
      ctx.fill();
      roundedRect(ctx, trackX, y - 4, trackWidth, 8, 4);
      ctx.strokeStyle = COLORS.accent;
      ctx.lineWidth = 1;
      ctx.stroke();
      // 数值保持水平居中，字号略小并沿条中心线绘制，避免上下触碰边框。
      this.drawFittedText(ctx, valueText, trackX + trackWidth / 2, y + 0.5, trackWidth - 8, 6, COLORS.text, 'normal', 'center');
    };
    drawStatusBar('生命', `${Math.floor(hp)} / ${Math.floor(maxHp)}`, hp / maxHp, profileTop + 62, '#d96868');
    drawStatusBar('灵气', `${Math.floor(qi)} / ${Math.floor(qiCap)}`, qi / qiCap, profileTop + 81, '#6e9fe8');
    const nextThreshold = Number(state.player.nextRealmThreshold);
    const currentThreshold = Number(state.player.realmThreshold) || 0;
    const progressRatio = Number.isFinite(nextThreshold) && nextThreshold > currentThreshold
      ? Math.max(0, Math.min(1, (cultivation - currentThreshold) / (nextThreshold - currentThreshold)))
      : 1;
    drawStatusBar(
      '修为',
      Number.isFinite(nextThreshold) && nextThreshold > currentThreshold
        ? `${Math.floor(cultivation)} / ${Math.floor(nextThreshold)}`
        : `${Math.floor(cultivation)}`,
      progressRatio,
      profileTop + 100,
      realmColor,
    );

    const equipmentTitleY = panel.y + 171;
    const equipmentGridX = panel.x + 18;
    const equipmentGridY = panel.y + 188;
    const equipmentColumns = 3;
    const equipmentGap = layout.gap;
    const equipmentSlots = [
      ['weapon', '武器'], ['artifact', '法宝'], ['armor', '衣甲'],
      ['storage', '储物'], ['technique', '功法'], ['spell', '法术'],
    ];
    this.drawText(ctx, '装备', equipmentGridX, equipmentTitleY, 10, COLORS.text, 'bold', 'left');
    equipmentSlots.forEach(([slot, label], index) => {
      const row = Math.floor(index / equipmentColumns);
      const col = index % equipmentColumns;
      const x = equipmentGridX + col * (layout.cellWidth + equipmentGap);
      const y = equipmentGridY + row * (layout.cellHeight + equipmentGap);
      ctx.fillStyle = (row + col) % 2 ? COLORS.cellAlt : COLORS.cell;
      roundedRect(ctx, x, y, layout.cellWidth, layout.cellHeight, 4);
      ctx.fill();
      ctx.strokeStyle = COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
      const item = state.player.equipment && state.player.equipment[slot];
      if (item) {
        const itemId = resolveItemId(item);
        const itemViewData = itemId ? itemView(itemId, { id: itemId, count: 1 }) : null;
        const itemName = itemViewData ? itemViewData.name : (typeof item === 'string' ? item : (item.name || label));
        const itemSymbol = itemViewData ? itemViewData.symbol : (typeof item === 'object' && item.symbol ? item.symbol : itemName.slice(0, 1));
        const itemColor = itemViewData && itemViewData.qualityColor ? itemViewData.qualityColor : realmColor;
        this.drawText(ctx, itemSymbol, x + layout.cellWidth / 2, y + layout.cellHeight * 0.37, Math.min(20, layout.cellWidth * 0.34), itemColor, 'bold', 'center');
        this.drawFittedText(ctx, itemName, x + layout.cellWidth / 2, y + layout.cellHeight * 0.72, layout.cellWidth - 8, 8, itemColor, 'normal', 'center');
      } else {
        this.drawFittedText(ctx, label, x + layout.cellWidth / 2, y + layout.cellHeight / 2, layout.cellWidth - 8, 8, COLORS.faint, 'normal', 'center');
      }
    });

    const attributeX = equipmentGridX + equipmentColumns * layout.cellWidth + (equipmentColumns - 1) * equipmentGap + 18;
    const attributeWidth = Math.max(50, panel.x + panel.width - attributeX - 16);
    const attributes = state.player.attributes || {};
    const attributeRows = [
      ['体魄', attributes.physique, '#d96868'],
      ['灵力', attributes.spirit, '#6e9fe8'],
      ['气血', attributes.vitality, '#70d58c'],
      ['神识', attributes.sense, '#b68cff'],
    ];
    ctx.strokeStyle = COLORS.lineSoft;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(attributeX - 10, equipmentTitleY - 10);
    ctx.lineTo(attributeX - 10, panel.y + 309);
    ctx.stroke();
    this.drawText(ctx, '属性', attributeX, equipmentTitleY, 10, COLORS.text, 'bold', 'left');
    attributeRows.forEach(([label, value, color], index) => {
      const y = panel.y + 201 + index * 27;
      ctx.fillStyle = COLORS.cell;
      roundedRect(ctx, attributeX, y - 11, attributeWidth, 22, 3);
      ctx.fill();
      ctx.strokeStyle = COLORS.lineSoft;
      ctx.stroke();
      ctx.fillStyle = color;
      ctx.fillRect(attributeX + 6, y - 5, 3, 10);
      this.drawText(ctx, label, attributeX + 14, y, 8, COLORS.muted, 'normal', 'left');
      this.drawText(ctx, String(Number.isFinite(value) ? value : 0), attributeX + attributeWidth - 7, y, 9, COLORS.text, 'bold', 'right');
    });

    this.attributeDetailsButton = {
      x: attributeX,
      y: panel.y + 312,
      width: attributeWidth,
      height: 26,
    };
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, this.attributeDetailsButton.x, this.attributeDetailsButton.y, this.attributeDetailsButton.width, this.attributeDetailsButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.stroke();
    this.drawText(ctx, '查看属性', this.attributeDetailsButton.x + this.attributeDetailsButton.width / 2, this.attributeDetailsButton.y + 13, 8, COLORS.text, 'bold', 'center');

    ctx.strokeStyle = COLORS.lineSoft;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(panel.x + 12, layout.sectionY);
    ctx.lineTo(panel.x + panel.width - 12, layout.sectionY);
    ctx.stroke();
    this.drawText(ctx, '背包', panel.x + 18, layout.backpackTitleY, 14, COLORS.text, 'bold');
    this.drawText(ctx, `Lv.${level} · ${slotCount}格`, panel.x + panel.width - 18, layout.backpackTitleY, 9, COLORS.muted, 'normal', 'right');

    ctx.save();
    ctx.beginPath();
    ctx.rect(panel.x + 10, layout.gridY - 8, panel.width - 20, layout.gridViewportHeight + 16);
    ctx.clip();
    for (let index = 0; index < slotCount; index += 1) {
      const row = Math.floor(index / layout.columns);
      const col = index % layout.columns;
      const x = layout.gridX + col * (layout.cellWidth + layout.gap);
      const y = layout.gridY + row * (layout.cellHeight + layout.gap) - scrollY;
      ctx.fillStyle = (row + col) % 2 ? COLORS.cellAlt : COLORS.cell;
      roundedRect(ctx, x, y, layout.cellWidth, layout.cellHeight, 4);
      ctx.fill();
      ctx.strokeStyle = COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
      const item = items[index];
      const isDraggingSource = state.inventory && state.inventory.drag
        && state.inventory.drag.source
        && state.inventory.drag.source.type === 'inventory'
        && state.inventory.drag.source.index === index;
      if (isDraggingSource) {
        ctx.fillStyle = COLORS.current;
        roundedRect(ctx, x + 2, y + 2, layout.cellWidth - 4, layout.cellHeight - 4, 4);
        ctx.fill();
        ctx.strokeStyle = COLORS.accent;
        ctx.setLineDash && ctx.setLineDash([3, 3]);
        ctx.stroke();
        ctx.setLineDash && ctx.setLineDash([]);
      } else if (item) {
        const itemColor = item.qualityColor || COLORS.text;
        this.drawText(ctx, item.symbol, x + layout.cellWidth / 2, y + layout.cellHeight * 0.38, Math.min(21, layout.cellWidth * 0.38), itemColor, 'bold', 'center');
        this.drawFittedText(ctx, item.name, x + layout.cellWidth / 2, y + layout.cellHeight * 0.69, layout.cellWidth - 8, 8, itemColor, 'normal', 'center');
        this.drawText(ctx, String(item.count), x + layout.cellWidth - 7, y + layout.cellHeight - 9, 8, COLORS.accent, 'bold', 'right');
      }
    }
    ctx.restore();

    const drag = state.inventory && state.inventory.drag;
    if (drag && drag.source) {
      const sourceItem = drag.source.type === 'inventory'
        ? state.inventory.items[drag.source.index]
        : state.player.equipment && state.player.equipment[drag.source.slot];
      if (sourceItem && drag.x != null && drag.y != null) {
        const dragView = itemView(resolveItemId(sourceItem), { id: resolveItemId(sourceItem), count: 1 });
        const previewSize = Math.min(layout.cellWidth, layout.cellHeight) * 0.86;
        ctx.save();
        ctx.globalAlpha = 0.82;
        ctx.fillStyle = COLORS.current;
        roundedRect(ctx, drag.x - previewSize / 2, drag.y - previewSize / 2, previewSize, previewSize, 5);
        ctx.fill();
        ctx.strokeStyle = dragView.qualityColor || COLORS.accent;
        ctx.lineWidth = 1.5;
        ctx.stroke();
        this.drawText(ctx, dragView.symbol, drag.x, drag.y - 4, Math.min(21, previewSize * 0.38), dragView.qualityColor || COLORS.text, 'bold', 'center');
        this.drawFittedText(ctx, dragView.name, drag.x, drag.y + previewSize * 0.22, previewSize - 8, 8, dragView.qualityColor || COLORS.text, 'normal', 'center');
        ctx.restore();
      }
      const targetInventoryIndex = this.getInventorySlotAtPoint(drag.x, drag.y, state);
      const targetEquipmentSlot = this.getEquipmentSlotAtPoint(drag.x, drag.y, state);
      if (Number.isInteger(targetInventoryIndex) || targetEquipmentSlot) {
        let targetX;
        let targetY;
        if (Number.isInteger(targetInventoryIndex)) {
          const targetRow = Math.floor(targetInventoryIndex / layout.columns);
          const targetCol = targetInventoryIndex % layout.columns;
          targetX = layout.gridX + targetCol * (layout.cellWidth + layout.gap);
          targetY = layout.gridY + targetRow * (layout.cellHeight + layout.gap) - scrollY;
        } else {
          const equipmentSlots = ['weapon', 'artifact', 'armor', 'storage', 'technique', 'spell'];
          const targetIndex = equipmentSlots.indexOf(targetEquipmentSlot);
          const targetRow = Math.floor(targetIndex / 3);
          const targetCol = targetIndex % 3;
          targetX = panel.x + 18 + targetCol * (layout.cellWidth + layout.gap);
          targetY = panel.y + 188 + targetRow * (layout.cellHeight + layout.gap);
        }
        ctx.save();
        ctx.strokeStyle = COLORS.accent;
        ctx.lineWidth = 2;
        ctx.strokeRect(targetX + 1, targetY + 1, layout.cellWidth - 2, layout.cellHeight - 2);
        ctx.restore();
      }
    }

    if (maxScroll > 0) {
      const trackX = panel.x + panel.width - 12;
      const trackY = layout.gridY;
      const trackHeight = layout.gridViewportHeight;
      const thumbHeight = Math.max(26, trackHeight * (layout.gridViewportHeight / (layout.gridViewportHeight + maxScroll)));
      const thumbY = trackY + (trackHeight - thumbHeight) * (scrollY / maxScroll);
      ctx.fillStyle = COLORS.cellAlt;
      roundedRect(ctx, trackX - 3, trackY, 4, trackHeight, 2);
      ctx.fill();
      ctx.fillStyle = COLORS.accent;
      roundedRect(ctx, trackX - 3, thumbY, 4, thumbHeight, 2);
      ctx.fill();
    }

    ctx.fillStyle = COLORS.header;
    ctx.fillRect(panel.x + 1, panel.y + panel.height - 72, panel.width - 2, 71);
    ctx.strokeStyle = COLORS.lineSoft;
    ctx.beginPath();
    ctx.moveTo(panel.x + 12, panel.y + panel.height - 72);
    ctx.lineTo(panel.x + panel.width - 12, panel.y + panel.height - 72);
    ctx.stroke();
    ctx.fillStyle = COLORS.current;
    roundedRect(ctx, this.inventoryUpgradeButton.x, this.inventoryUpgradeButton.y, this.inventoryUpgradeButton.width, this.inventoryUpgradeButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.stroke();
    this.drawText(ctx, level >= 5 ? '已满级' : '扩充 +4格', this.inventoryUpgradeButton.x + this.inventoryUpgradeButton.width / 2, this.inventoryUpgradeButton.y + 16, 9, level >= 5 ? COLORS.faint : COLORS.text, 'bold', 'center');
    ctx.restore();
  }

  renderAttributeOverlay(ctx, state) {
    const panel = {
      x: 20,
      y: this.headerHeight + 8,
      width: this.width - 40,
      height: Math.min(440, this.height - this.headerHeight - 16),
    };
    const compact = panel.height < 360;
    const attributes = state.player.attributes || {};
    const cultivation = state.resources ? Math.max(0, Number(state.resources.cultivation) || 0) : 0;
    const currentThreshold = Math.max(0, Number(state.player.realmThreshold) || 0);
    const nextThreshold = Math.max(currentThreshold, Number(state.player.nextRealmThreshold) || currentThreshold);
    const hasNextRealm = state.player.nextRealm && state.player.nextRealm !== '已至顶峰';
    const progressRatio = hasNextRealm && nextThreshold > currentThreshold
      ? Math.max(0, Math.min(1, (cultivation - currentThreshold) / (nextThreshold - currentThreshold)))
      : 1;
    const derivedMaxHp = 100 + (Math.max(0, Number(attributes.physique) || 0) * 10);
    const derivedQiCap = Math.max(Number(state.qiCap) || 0, 100 + (Math.max(0, Number(attributes.spirit) || 0) * 10));
    const power = 10 + (Math.max(0, Number(attributes.vitality) || 0) * 2);
    const sense = Math.max(0, Number(attributes.sense) || 0);

    this.attributeOverlayCloseButton = { x: panel.x + panel.width - 42, y: panel.y + 10, width: 28, height: 28 };
    this.breakthroughButton = {
      x: panel.x + 18,
      y: panel.y + panel.height - 54,
      width: panel.width - 36,
      height: 34,
    };

    ctx.save();
    ctx.fillStyle = 'rgba(0, 0, 0, 0.72)';
    ctx.fillRect(0, 0, this.width, this.height);
    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, panel.x, panel.y, panel.width, panel.height, 7);
    ctx.fill();
    ctx.strokeStyle = COLORS.accent;
    ctx.lineWidth = 1;
    ctx.stroke();

    this.drawText(ctx, '人物属性', panel.x + 18, panel.y + (compact ? 22 : 27), compact ? 13 : 15, COLORS.text, 'bold');
    this.drawText(ctx, '×', this.attributeOverlayCloseButton.x + 14, this.attributeOverlayCloseButton.y + 14, 19, COLORS.muted, 'normal', 'center');
    const realmColor = state.player.realmColor || COLORS.text;
    const realmY = panel.y + (compact ? 49 : 58);
    this.drawText(ctx, state.player.realm || '凡人', panel.x + 18, realmY, compact ? 11 : 13, realmColor, 'bold');
    this.drawFittedText(ctx, hasNextRealm ? `下一境界 ${state.player.nextRealm}` : '已至当前版本最高境界', panel.x + panel.width - 18, realmY, panel.width * 0.55, compact ? 8 : 9, COLORS.muted, 'normal', 'right');

    const barX = panel.x + 18;
    const barY = panel.y + (compact ? 69 : 86);
    const barWidth = panel.width - 36;
    ctx.fillStyle = COLORS.cell;
    const progressHeight = compact ? 8 : 10;
    roundedRect(ctx, barX, barY, barWidth, progressHeight, 4);
    ctx.fill();
    if (progressRatio > 0) {
      ctx.fillStyle = realmColor;
      roundedRect(ctx, barX, barY, Math.max(6, barWidth * progressRatio), progressHeight, 4);
      ctx.fill();
    }
    ctx.strokeStyle = COLORS.line;
    ctx.stroke();
    this.drawText(ctx, hasNextRealm ? `修为 ${cultivation} / ${nextThreshold}` : `修为 ${cultivation}`, panel.x + panel.width / 2, barY + (compact ? 16 : 22), compact ? 8 : 9, COLORS.text, 'bold', 'center');

    const rows = [
      ['体魄', attributes.physique, `生命上限 ${derivedMaxHp}`, '#d96868'],
      ['灵力', attributes.spirit, `灵气上限 ${Math.floor(derivedQiCap)}`, '#6e9fe8'],
      ['气血', attributes.vitality, `力量 ${power}`, '#70d58c'],
      ['神识', attributes.sense, `神识强度 ${sense}`, '#b68cff'],
    ];
    const rowTop = panel.y + (compact ? 101 : 126);
    const rowGap = compact ? 38 : 48;
    const rowHeight = compact ? 30 : 38;
    const columnGap = 8;
    const rowWidth = (panel.width - 36 - columnGap) / 2;
    rows.forEach(([label, value, effect, color], index) => {
      const col = index % 2;
      const row = Math.floor(index / 2);
      const x = panel.x + 18 + col * (rowWidth + columnGap);
      const y = rowTop + row * rowGap;
      ctx.fillStyle = COLORS.cell;
      roundedRect(ctx, x, y, rowWidth, rowHeight, 4);
      ctx.fill();
      ctx.strokeStyle = COLORS.lineSoft;
      ctx.stroke();
      ctx.fillStyle = color;
      ctx.fillRect(x + 7, y + (compact ? 6 : 9), 3, compact ? 18 : 20);
      this.drawText(ctx, label, x + 18, y + (compact ? 10 : 13), compact ? 8 : 9, COLORS.muted, 'bold');
      this.drawText(ctx, String(Math.max(0, Number(value) || 0)), x + rowWidth - 9, y + (compact ? 10 : 13), compact ? 9 : 11, COLORS.text, 'bold', 'right');
      this.drawFittedText(ctx, effect, x + 18, y + (compact ? 23 : 28), rowWidth - 28, compact ? 7 : 8, COLORS.faint, 'normal', 'left');
    });

    const targetFailureRate = Math.max(0, Number(state.player.nextBreakthroughFailureRate) || 0);
    const statusY = panel.y + panel.height - 76;
    const ready = Boolean(state.player.canBreakthrough && hasNextRealm);
    const statusText = !hasNextRealm
      ? '当前版本已至顶峰'
      : ready
        ? (targetFailureRate > 0 ? `大境界突破 · 失败率 ${Math.round(targetFailureRate * 100)}%` : '修为已满，可突破')
        : `还需 ${Math.max(0, nextThreshold - cultivation)} 修为`;
    this.drawText(ctx, statusText, panel.x + panel.width / 2, statusY, 8, ready ? COLORS.accent : COLORS.faint, 'normal', 'center');

    ctx.fillStyle = ready ? COLORS.current : COLORS.cellAlt;
    roundedRect(ctx, this.breakthroughButton.x, this.breakthroughButton.y, this.breakthroughButton.width, this.breakthroughButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = ready ? COLORS.accent : COLORS.line;
    ctx.stroke();
    this.drawText(ctx, hasNextRealm ? '尝试突破' : '已至顶峰', this.breakthroughButton.x + this.breakthroughButton.width / 2, this.breakthroughButton.y + 17, 10, ready ? COLORS.text : COLORS.faint, 'bold', 'center');
    ctx.restore();
  }

  getInventoryItems(state) {
    const stacks = state && state.inventory && Array.isArray(state.inventory.items)
      ? state.inventory.items
      : [];
    return stacks.map((stack) => (stack && stack.count > 0 ? itemView(stack.id, stack) : null));
  }

  renderItemDetailOverlay(ctx, state) {
    const itemId = state.inventory && state.inventory.selectedItemId;
    if (!itemId) return;
    const equipmentSlots = ['weapon', 'artifact', 'armor', 'storage', 'technique', 'spell'];
    const equipmentSlot = equipmentSlots.find((slot) => resolveItemId(state.player && state.player.equipment && state.player.equipment[slot]) === itemId) || null;
    const stack = (state.inventory.items || []).find((entry) => entry && entry.id === itemId);
    const item = itemView(itemId, stack || (equipmentSlot ? { id: itemId, count: 1 } : null));
    const panel = {
      x: 24,
      y: this.headerHeight + 26,
      width: this.width - 48,
      height: Math.min(330, this.height - this.headerHeight - 52),
    };
    const quality = qualityDefinition(item.quality);
    const qualityColor = quality.color || COLORS.accent;
    const centerX = panel.x + panel.width / 2;
    this.itemDetailCloseButton = { x: panel.x + panel.width - 42, y: panel.y + 10, width: 28, height: 28 };
    const equipmentCategorySlots = {
      weapon: 'weapon',
      artifact: 'artifact',
      armor: 'armor',
      storage: 'storage',
      technique: 'technique',
      spell: 'spell',
    };
    const canEquip = Boolean(equipmentCategorySlots[item.category]);
    const actionLabel = item.usable ? '使用' : canEquip ? (equipmentSlot ? '卸下' : '装备') : '';
    this.itemDetailActionButton = actionLabel
      ? { x: panel.x + 24, y: panel.y + panel.height - 48, width: panel.width - 48, height: 32 }
      : null;
    ctx.save();
    ctx.fillStyle = 'rgba(0, 0, 0, 0.78)';
    ctx.fillRect(0, 0, this.width, this.height);
    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, panel.x, panel.y, panel.width, panel.height, 7);
    ctx.fill();
    ctx.strokeStyle = qualityColor;
    ctx.lineWidth = 1;
    ctx.stroke();
    this.drawText(ctx, '物品详情', panel.x + 18, panel.y + 27, 15, COLORS.text, 'bold');
    this.drawText(ctx, '×', this.itemDetailCloseButton.x + 14, this.itemDetailCloseButton.y + 14, 19, COLORS.muted, 'normal', 'center');
    this.drawText(ctx, item.symbol, centerX, panel.y + 91, 44, qualityColor, 'bold', 'center');
    this.drawFittedText(ctx, item.name, centerX, panel.y + 135, panel.width - 36, 17, qualityColor, 'bold', 'center');
    this.drawFittedText(ctx, `${item.qualityName}品 · ${item.categoryName}`, centerX, panel.y + 164, panel.width - 36, 10, COLORS.muted, 'normal', 'center');
    this.drawFittedText(ctx, `${item.kindName} · 数量 ${item.count}`, centerX, panel.y + 188, panel.width - 36, 9, COLORS.faint, 'normal', 'center');
    this.drawFittedText(ctx, item.description, centerX, panel.y + 236, panel.width - 48, 10, COLORS.text, 'normal', 'center');
    if (this.itemDetailActionButton) {
      const button = this.itemDetailActionButton;
      ctx.fillStyle = COLORS.current;
      roundedRect(ctx, button.x, button.y, button.width, button.height, 4);
      ctx.fill();
      ctx.strokeStyle = qualityColor;
      ctx.lineWidth = 1;
      ctx.stroke();
      this.drawText(ctx, actionLabel, button.x + button.width / 2, button.y + button.height / 2, 10, COLORS.text, 'bold', 'center');
    }
    ctx.restore();
  }

  renderCaveHeader(ctx, state) {
    this.drawText(ctx, '洞府', 16, 20, 15, COLORS.text, 'bold');
    this.drawFittedText(ctx, `个人修行居所 · ${state.player.realm || '凡人'}`, 16, 43, this.width * 0.5, 10, COLORS.muted, 'normal', 'left');
    this.drawText(ctx, '右下角 · 出口', this.width - 16, 28, 10, COLORS.accent, 'bold', 'right');
    // 洞府页同样保留储物袋快捷入口，位置与野外地图一致。
    const bagY = Math.max(this.headerHeight + 8, this.controlRect.y - 40);
    this.inventoryButton = { x: 16, y: bagY, width: 88, height: 30 };
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, this.inventoryButton.x, this.inventoryButton.y, this.inventoryButton.width, this.inventoryButton.height, 5);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();
    this.drawText(ctx, '储物袋', this.inventoryButton.x + this.inventoryButton.width / 2, bagY + 15, 10, COLORS.text, 'bold', 'center');
    ctx.strokeStyle = COLORS.lineSoft;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(12, this.headerHeight - 1);
    ctx.lineTo(this.width - 12, this.headerHeight - 1);
    ctx.stroke();
  }

  renderDungeonHeader(ctx, state) {
    this.drawText(ctx, '秘境探索', 16, 20, 15, COLORS.text, 'bold');
    this.drawFittedText(ctx, `${state.player.realm || '凡人'} · 灵气 ${Math.floor(state.qi)} / ${state.qiCap}`, 16, 43, this.width * 0.62, 10, COLORS.muted, 'normal', 'left');
    this.drawText(ctx, '单次探索', this.width - 16, 28, 10, COLORS.accent, 'bold', 'right');
    ctx.strokeStyle = COLORS.lineSoft;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(12, this.headerHeight - 1);
    ctx.lineTo(this.width - 12, this.headerHeight - 1);
    ctx.stroke();
  }

  renderDungeonPanel(ctx, state, dungeon) {
    const map = this.mapRect;
    const centerX = map.x + map.width / 2;
    const compact = map.height < 340;
    const veryCompact = map.height < 250;
    const panelPadding = Math.max(14, Math.min(24, map.width * 0.06));
    const titleY = map.y + (veryCompact ? 24 : (compact ? 30 : 44));
    const infoStartY = titleY + (veryCompact ? 28 : (compact ? 34 : 42));
    const lineGap = veryCompact ? 18 : (compact ? 22 : 27);
    const actionHeight = veryCompact ? 28 : (compact ? 30 : 34);
    const actionY = map.y + map.height - (veryCompact ? 42 : (compact ? 70 : 88));

    this.dungeonActionButton = null;
    this.dungeonBackButton = null;

    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, map.x, map.y, map.width, map.height, 6);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();

    if (!dungeon) {
      this.drawText(ctx, '未找到秘境资料', centerX, map.y + map.height / 2, 13, COLORS.text, 'bold', 'center');
      return;
    }

    this.drawFittedText(ctx, dungeon.name, centerX, titleY, map.width - panelPadding * 2, compact ? 18 : 21, COLORS.text, 'bold', 'center');
    this.drawText(ctx, `危险等级 ${'◆'.repeat(dungeon.danger)}`, centerX, infoStartY, 10, COLORS.accent, 'normal', 'center');
    if (veryCompact) {
      this.drawText(ctx, `消耗 ${dungeon.qiCost}灵气 · ${dungeon.minRealm <= (state.player.realmIndex || 0) ? '境界已满足' : '境界不足'}`, centerX, infoStartY + lineGap, 9, COLORS.muted, 'normal', 'center');
      this.drawFittedText(ctx, `收益：修为 ${dungeon.cultivation[0]}-${dungeon.cultivation[1]} · 灵草 ${dungeon.spiritHerbs[0]}-${dungeon.spiritHerbs[1]}`, centerX, infoStartY + lineGap * 2, map.width - panelPadding * 2, 8, COLORS.faint, 'normal', 'center');
    } else {
      this.drawText(ctx, `最低境界：${dungeon.minRealm <= (state.player.realmIndex || 0) ? '已满足' : '不足'}`, centerX, infoStartY + lineGap, 10, COLORS.muted, 'normal', 'center');
      this.drawText(ctx, `探索消耗：${dungeon.qiCost}点灵气`, centerX, infoStartY + lineGap * 2, 10, COLORS.muted, 'normal', 'center');
      this.drawFittedText(ctx, dungeon.description, centerX, infoStartY + lineGap * 3, map.width - panelPadding * 2, 10, COLORS.faint, 'normal', 'center');
      this.drawFittedText(ctx, `预计收益：修为 ${dungeon.cultivation[0]}-${dungeon.cultivation[1]} · 灵草 ${dungeon.spiritHerbs[0]}-${dungeon.spiritHerbs[1]}`, centerX, infoStartY + lineGap * 4, map.width - panelPadding * 2, 9, COLORS.faint, 'normal', 'center');
    }

    const result = state.dungeon && state.dungeon.result;
    if (result) {
      const resultColor = result.type === 'success' ? COLORS.text : COLORS.accent;
      this.drawFittedText(ctx, result.text, centerX, actionY - (compact ? 26 : 32), map.width - panelPadding * 2, 10, resultColor, 'bold', 'center');
      if (result.breakthrough) {
        this.drawText(ctx, `境界突破：${result.breakthrough}`, centerX, actionY - (compact ? 8 : 11), 9, COLORS.accent, 'normal', 'center');
      }
    }

    this.dungeonActionButton = {
      x: centerX - (veryCompact ? 58 : (compact ? 64 : 72)),
      y: actionY,
      width: veryCompact ? 116 : (compact ? 128 : 144),
      height: actionHeight,
    };
    ctx.fillStyle = result && result.type === 'success' ? COLORS.cell : COLORS.current;
    roundedRect(ctx, this.dungeonActionButton.x, this.dungeonActionButton.y, this.dungeonActionButton.width, this.dungeonActionButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = result && result.type === 'success' ? COLORS.line : COLORS.accent;
    ctx.lineWidth = 1;
    ctx.stroke();
    this.drawText(ctx, result && result.type === 'success' ? '再次探索' : '开始探索', centerX, this.dungeonActionButton.y + this.dungeonActionButton.height / 2, veryCompact ? 9 : 10, COLORS.text, 'bold', 'center');

    this.dungeonBackButton = {
      x: map.x + 14,
      y: map.y + 14,
      width: 86,
      height: 28,
    };
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, this.dungeonBackButton.x, this.dungeonBackButton.y, this.dungeonBackButton.width, this.dungeonBackButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();
    this.drawText(ctx, '返回野外', this.dungeonBackButton.x + this.dungeonBackButton.width / 2, this.dungeonBackButton.y + 14, 9, COLORS.text, 'bold', 'center');

    const footerY = this.controlRect.y;
    ctx.strokeStyle = COLORS.lineSoft;
    ctx.beginPath();
    ctx.moveTo(12, footerY);
    ctx.lineTo(this.width - 12, footerY);
    ctx.stroke();
    const resources = state.resources || {};
    this.drawText(ctx, `灵草 ${resources.spiritHerbs || 0} · 回气丹 ${resources.qiPills || 0}`, 16, footerY + 23, 10, COLORS.muted);
    this.drawText(ctx, `修为 ${resources.cultivation || 0}`, 16, footerY + 45, 10, COLORS.muted);
    const record = state.dungeonStats && dungeon ? state.dungeonStats[dungeon.name] : null;
    if (record) {
      this.drawFittedText(ctx, `累计探索 ${record.count || 0} 次 · 单次最高修为 +${record.bestCultivation || 0}`, this.width - 16, footerY + 34, this.width * 0.56, 9, COLORS.faint, 'normal', 'right');
    }
    this.drawText(ctx, state.lastMove, 16, this.height - 18, 9, COLORS.faint);
  }

  renderCaveMap(ctx, state, buildings) {
    const map = this.mapRect;
    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, map.x, map.y, map.width, map.height, 6);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();

    for (let row = 0; row < this.rows; row += 1) {
      for (let col = 0; col < this.cols; col += 1) {
        const x = map.x + col * this.cellWidth;
        const y = map.y + row * this.cellHeight;
        ctx.fillStyle = (row + col) % 2 ? COLORS.cellAlt : COLORS.cell;
        ctx.fillRect(x + 1, y + 1, this.cellWidth - 2, this.cellHeight - 2);
        ctx.strokeStyle = COLORS.lineSoft;
        ctx.lineWidth = 1;
        ctx.strokeRect(x + 0.5, y + 0.5, this.cellWidth, this.cellHeight);
      }
    }

    buildings.forEach((building) => {
      const x = map.x + building.col * this.cellWidth;
      const y = map.y + building.row * this.cellHeight;
      const selected = state.cave.player.row === building.row && state.cave.player.col === building.col;
      const buildingState = state.cave.buildings[building.id];
      const level = buildingState && Number.isInteger(buildingState.level) ? buildingState.level : 0;
      const style = this.buildingVisualStyle(level);
      ctx.fillStyle = selected ? COLORS.current : COLORS.cellAlt;
      ctx.fillRect(x + 2, y + 2, this.cellWidth - 4, this.cellHeight - 4);
      ctx.strokeStyle = style.frame;
      ctx.lineWidth = selected ? 1.5 : 1;
      ctx.strokeRect(x + 1.5, y + 1.5, this.cellWidth - 3, this.cellHeight - 3);
      const damaged = level <= 0;
      const nameY = selected ? y + this.cellHeight * (damaged ? 0.24 : 0.33) : y + this.cellHeight / 2;
      this.drawFittedText(
        ctx,
        building.shortName,
        x + this.cellWidth / 2,
        nameY,
        this.cellWidth - 8,
        Math.min(11, this.cellWidth * 0.28),
        style.name,
        selected ? 'bold' : 'normal',
        'center',
      );
      if (damaged) {
        this.drawText(ctx, '损坏', x + this.cellWidth / 2, y + this.cellHeight * (selected ? 0.43 : 0.72), 8, style.status, 'normal', 'center');
      }
    });

    const entranceX = map.x + this.caveEntrance.col * this.cellWidth;
    const entranceY = map.y + this.caveEntrance.row * this.cellHeight;
    const playerX = map.x + state.cave.player.col * this.cellWidth;
    const playerY = map.y + state.cave.player.row * this.cellHeight;
    const atEntrance = state.cave.player.row === this.caveEntrance.row && state.cave.player.col === this.caveEntrance.col;
    ctx.strokeStyle = COLORS.accent;
    ctx.lineWidth = atEntrance ? 1.5 : 1;
    ctx.strokeRect(entranceX + 1.5, entranceY + 1.5, this.cellWidth - 3, this.cellHeight - 3);
    this.drawText(ctx, '出口', entranceX + this.cellWidth / 2, entranceY + this.cellHeight * 0.22, 8, COLORS.accent, 'bold', 'center');
    const playerSize = Math.min(18, this.cellHeight * 0.34);
    const currentBuilding = buildings.some((building) => building.row === state.cave.player.row && building.col === state.cave.player.col);
    if (currentBuilding) {
      this.drawText(ctx, state.player.name, playerX + this.cellWidth / 2, playerY + this.cellHeight * 0.82, playerSize, state.player.realmColor || COLORS.player, 'bold', 'center');
    } else {
      this.drawText(ctx, state.player.name, playerX + this.cellWidth / 2, playerY + this.cellHeight * 0.56, playerSize, state.player.realmColor || COLORS.player, 'bold', 'center');
    }
  }

  renderCaveFooter(ctx, state, buildings) {
    ctx.strokeStyle = COLORS.lineSoft;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.moveTo(12, this.controlRect.y);
    ctx.lineTo(this.width - 12, this.controlRect.y);
    ctx.stroke();
    this.drawText(ctx, '洞府等级 Lv.1', 16, this.controlRect.y + 18, 11, COLORS.text, 'bold');
    this.drawText(ctx, `灵气 ${Math.floor(state.qi)} / ${state.qiCap}`, 16, this.controlRect.y + 36, 10, COLORS.muted);

    const selected = buildings.find((building) => building.row === state.cave.player.row && building.col === state.cave.player.col);
    this.caveActionButton = null;
    this.caveCollectButton = null;
    this.caveFeatureButton = null;
    const detailsWidth = this.width * 0.52;
    this.drawFittedText(ctx, `境界 ${state.player.realm || '凡人'} · 修为 ${state.player.realmProgress || (state.resources && state.resources.cultivation) || 0}`, 16, this.controlRect.y + 48, detailsWidth - 28, 8, COLORS.faint, 'normal', 'left');
    this.drawFittedText(ctx, `下境界：${state.player.nextRealm || '已至顶峰'}`, 16, this.controlRect.y + 62, detailsWidth - 28, 8, COLORS.faint, 'normal', 'left');
    if (!selected) {
      // Building details only appear while the player occupies a building cell.
    } else {
      const selectedState = state.cave.buildings[selected.id];
      const repaired = selectedState && selectedState.level > 0;
      const currentLevel = selectedState ? selectedState.level : 0;
      this.drawText(ctx, selected.name, 16, this.controlRect.y + 74, 11, COLORS.text, 'bold');
      this.drawText(ctx, `状态：${repaired ? `Lv.${selectedState.level}` : '损坏'}`, 16, this.controlRect.y + 91, 9, COLORS.muted);
      const effectText = repaired
        ? (selected.levelEffects && selected.levelEffects[currentLevel - 1]) || '已恢复基本功能'
        : selected.effect;
      this.drawFittedText(ctx, effectText, 16, this.controlRect.y + 106, detailsWidth - 28, 9, COLORS.muted, 'normal', 'left');
      const featureButton = {
        x: 16,
        y: this.controlRect.y + 118,
        width: 112,
        height: 28,
      };
      this.caveFeatureButton = featureButton;
      const featureLabel = selected.action && selected.action !== '收取灵气'
        ? selected.action
        : `进入${selected.shortName}`;
      ctx.fillStyle = repaired ? COLORS.current : COLORS.cellAlt;
      roundedRect(ctx, featureButton.x, featureButton.y, featureButton.width, featureButton.height, 4);
      ctx.fill();
      ctx.strokeStyle = repaired ? COLORS.accent : COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
      this.drawText(ctx, featureLabel, featureButton.x + featureButton.width / 2, featureButton.y + featureButton.height / 2, 9, repaired ? COLORS.text : COLORS.muted, 'bold', 'center');
    }
    const spiritArrayState = state.cave.buildings.spiritArray;
    if (spiritArrayState && spiritArrayState.level > 0 && selected && selected.id === 'spiritArray') {
      const collectButton = {
        x: 136,
        y: this.controlRect.y + 118,
        width: 80,
        height: 28,
      };
      this.caveCollectButton = collectButton;
      ctx.fillStyle = COLORS.cell;
      roundedRect(ctx, collectButton.x, collectButton.y, collectButton.width, collectButton.height, 4);
      ctx.fill();
      ctx.strokeStyle = COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
      this.drawText(ctx, `收取 ${Math.floor(state.pendingQi)}`, collectButton.x + collectButton.width / 2, collectButton.y + collectButton.height / 2, 9, COLORS.text, 'bold', 'center');
    }
    this.drawText(ctx, state.lastMove, 16, this.controlRect.y + this.controlRect.height - 12, 9, COLORS.muted);
    this.renderCaveControlPad(ctx, state);
  }


  renderCaveControlPad(ctx, state) {
    const center = this.cavePadCenter;
    const button = this.caveControlButtonSize;
    const positions = this.caveControlButtons;
    const resources = state && state.resources
      ? {
        spiritHerbs: state.resources.spiritHerbs || 0,
        qiPills: state.resources.qiPills || 0,
        cultivation: state.resources.cultivation || 0,
        techniques: Array.isArray(state.resources.techniques) ? state.resources.techniques.length : 0,
      }
      : { spiritHerbs: 0, qiPills: 0, cultivation: 0, techniques: 0 };
    this.drawFittedText(ctx, `灵草 ${resources.spiritHerbs} · 丹药 ${resources.qiPills}`, this.width * 0.57, this.controlRect.y + 37, this.width * 0.4, 8, COLORS.faint, 'normal', 'left');
    this.drawFittedText(ctx, `修为 ${resources.cultivation} · 功法 ${resources.techniques}`, this.width * 0.57, this.controlRect.y + 52, this.width * 0.4, 8, COLORS.faint, 'normal', 'left');
    ctx.save();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.arc(center.x, center.y, this.cavePadSize * 0.44, 0, Math.PI * 2);
    ctx.stroke();
    ctx.restore();
    Object.keys(positions).forEach((direction) => {
      const point = positions[direction];
      ctx.fillStyle = COLORS.cell;
      roundedRect(ctx, point.x - button / 2, point.y - button / 2, button, button, 5);
      ctx.fill();
      ctx.strokeStyle = COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
      const arrow = direction === 'up' ? '↑' : direction === 'down' ? '↓' : direction === 'left' ? '←' : '→';
      this.drawText(ctx, arrow, point.x, point.y + 1, Math.max(16, button * 0.45), COLORS.text, 'normal', 'center');
    });
    ctx.fillStyle = COLORS.cellAlt;
    ctx.beginPath();
    ctx.arc(center.x, center.y, Math.max(13, button * 0.3), 0, Math.PI * 2);
    ctx.fill();
    this.drawText(ctx, '林', center.x, center.y, 10, state.player.realmColor || COLORS.player, 'bold', 'center');
  }

  renderFeaturePanel(ctx, state, building) {
    if (building && building.id === 'alchemyFurnace') {
      this.renderAlchemyPanel(ctx, state, building);
      return;
    }
    this.featureActionButton = null;
    this.featureUpgradeButton = null;
    const feature = state.cave.features[building.id];
    const map = this.mapRect;
    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, map.x, map.y, map.width, map.height, 6);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();

    const centerX = map.x + map.width / 2;
    const compact = map.height < 300;
    const titleY = map.y + (compact ? 30 : 48);
    const levelY = titleY + (compact ? 23 : 28);
    const meditationOffset = building.id === 'meditationRoom' ? (compact ? 18 : 20) : 0;
    const descriptionY = levelY + (building.id === 'meditationRoom' ? (compact ? 38 : 58) : (compact ? 32 : 44));
    const statusY = descriptionY + (compact ? 24 : 30);
    const progressY = statusY + (compact ? 18 : 20);
    const progressTextY = progressY + 15;
    const actionY = Math.min(progressTextY + 18, map.y + map.height - 66);
    const rewardY = Math.min(actionY + 43, map.y + map.height - 28);
    const inventoryY = Math.min(rewardY + 15, map.y + map.height - 12);

    this.drawText(ctx, building.name, centerX, titleY, compact ? 14 : 16, COLORS.text, 'bold', 'center');
    const panelLevel = state.cave.buildings[building.id].level;
    const panelLevelColor = this.buildingLevelColor(panelLevel);
    this.drawText(ctx, panelLevel > 0 ? `等级 Lv.${panelLevel}` : '状态：损坏', centerX, levelY, 10, panelLevelColor, 'normal', 'center');
    if (building.id === 'meditationRoom') {
      this.drawText(ctx, `${state.player.realm || '凡人'} · 修为${state.resources ? state.resources.cultivation : 0}`, centerX, levelY + meditationOffset, 9, COLORS.muted, 'normal', 'center');
    }
    this.drawText(ctx, this.featureDescription(building.id), centerX, descriptionY, compact ? 10 : 11, COLORS.muted, 'normal', 'center');
    if (building.id === 'alchemyFurnace') {
      const herbs = state.resources ? state.resources.spiritHerbs || 0 : 0;
      const formulaColor = herbs > 0 ? COLORS.faint : COLORS.text;
      this.drawText(ctx, `配方：灵草 1 → 回气丹 1（库存 ${herbs}）`, centerX, descriptionY + (compact ? 16 : 18), compact ? 8 : 9, formulaColor, 'normal', 'center');
    }
    if (feature) {
      this.drawText(ctx, `当前状态：${this.featureStatusText(building.id, feature)}`, centerX, statusY, 10, COLORS.text, 'normal', 'center');
    } else {
      this.drawText(ctx, '建筑功能面板', centerX, statusY, 10, COLORS.text, 'normal', 'center');
    }

    const progress = feature && Number.isFinite(feature.progress) ? Math.max(0, Math.min(1, feature.progress)) : 0;
    const progressWidth = Math.min(map.width - 44, 220);
    const progressX = centerX - progressWidth / 2;
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, progressX, progressY, progressWidth, 7, 3);
    ctx.fill();
    if (progress > 0) {
      ctx.fillStyle = COLORS.accent;
      roundedRect(ctx, progressX, progressY, Math.max(7, progressWidth * progress), 7, 3);
      ctx.fill();
    }
    if (feature) {
      this.drawText(ctx, this.featureProgressText(feature), centerX, progressTextY, 9, COLORS.faint, 'normal', 'center');
    } else if (building.id === 'spiritArray') {
      this.drawText(ctx, `待收取灵气：${Math.floor(state.pendingQi)}`, centerX, progressTextY, 9, COLORS.faint, 'normal', 'center');
    } else {
      this.drawText(ctx, '可在此查看和升级建筑', centerX, progressTextY, 9, COLORS.faint, 'normal', 'center');
    }

    this.cavePanelBackButton = { x: map.x + 14, y: map.y + 14, width: 86, height: 28 };
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, this.cavePanelBackButton.x, this.cavePanelBackButton.y, this.cavePanelBackButton.width, this.cavePanelBackButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.stroke();
    this.drawText(ctx, '返回洞府', this.cavePanelBackButton.x + this.cavePanelBackButton.width / 2, this.cavePanelBackButton.y + 14, 9, COLORS.text, 'bold', 'center');

    const buildingState = state.cave.buildings[building.id] || { level: 0 };
    const repaired = buildingState.level > 0;
    this.featureActionButton = { x: centerX - 112, y: actionY, width: 104, height: 30 };
    this.featureUpgradeButton = { x: centerX + 8, y: actionY, width: 104, height: 30 };
    ctx.fillStyle = repaired ? COLORS.current : COLORS.cellAlt;
    roundedRect(ctx, this.featureActionButton.x, this.featureActionButton.y, this.featureActionButton.width, this.featureActionButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = repaired ? COLORS.accent : COLORS.line;
    ctx.stroke();
    const actionLabel = !repaired
      ? '需先修缮'
      : feature
        ? this.featureActionLabel(building.id, feature)
        : building.id === 'spiritArray'
          ? '收取灵气'
          : '暂无功能';
    this.drawText(ctx, actionLabel, this.featureActionButton.x + this.featureActionButton.width / 2, this.featureActionButton.y + 15, 9, repaired && (feature || building.id === 'spiritArray') ? COLORS.text : COLORS.faint, 'bold', 'center');

    const maxLevel = building.maxLevel || 5;
    const upgradeLabel = repaired ? (buildingState.level >= maxLevel ? '已满级' : '升级') : '修缮';
    ctx.fillStyle = buildingState.level >= maxLevel ? COLORS.cellAlt : COLORS.cell;
    roundedRect(ctx, this.featureUpgradeButton.x, this.featureUpgradeButton.y, this.featureUpgradeButton.width, this.featureUpgradeButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = buildingState.level >= maxLevel ? COLORS.lineSoft : COLORS.line;
    ctx.stroke();
    this.drawText(ctx, upgradeLabel, this.featureUpgradeButton.x + this.featureUpgradeButton.width / 2, this.featureUpgradeButton.y + 15, 9, buildingState.level >= maxLevel ? COLORS.faint : COLORS.text, 'bold', 'center');
    if (feature) {
      this.drawText(ctx, `本次完成：${this.featureRewardPreview(building.id, state)}`, centerX, rewardY, 9, COLORS.faint, 'normal', 'center');
      this.drawText(ctx, this.featureRewardText(building.id, state), centerX, inventoryY, 9, COLORS.muted, 'normal', 'center');
    } else if (building.id === 'spiritArray') {
      this.drawText(ctx, `灵气产出：${building.levelEffects[buildingState.level - 1] || '基础产出'}`, centerX, rewardY, 9, COLORS.faint, 'normal', 'center');
      this.drawText(ctx, `当前灵气：${Math.floor(state.qi)} / ${state.qiCap}`, centerX, inventoryY, 9, COLORS.muted, 'normal', 'center');
    }
    this.drawText(ctx, state.lastMove, map.x + 12, map.y + map.height - 12, 9, COLORS.muted);
  }

  renderAlchemyPanel(ctx, state, building) {
    this.featureActionButton = null;
    this.featureUpgradeButton = null;
    this.cavePanelBackButton = null;
    this.alchemyMaterialBar = null;
    this.alchemyHeatBar = null;
    this.alchemyHeatSlot = null;
    this.alchemyUpgradeCloseButton = null;
    this.alchemyUpgradeConfirmButton = null;
    this.alchemyUpgradeCancelButton = null;
    // 丹炉页与洞府建筑页共用同一块内容面板尺寸。
    const map = this.mapRect;
    const centerX = map.x + map.width / 2;
    const buildingState = state.cave.buildings[building.id] || { level: 0 };
    const feature = state.cave.features.alchemyFurnace;
    const alchemy = state.cave.alchemy || { heat: 0.5, materialScrollX: 0 };
    const repaired = buildingState.level > 0;
    const slotCount = repaired ? Math.max(1, Math.min(6, Math.floor(Number(buildingState.level) || 1))) : 0;
    const backpackPanel = {
      x: 12,
      y: this.headerHeight + 10,
      width: this.width - 24,
      height: this.height - this.headerHeight - 20,
    };
    const backpackLayout = this.getInventoryLayoutForPanel(backpackPanel);
    const slotGap = backpackLayout.gap;
    const compact = map.height < 300;
    const titleY = map.y + (compact ? 28 : 42);
    const slotTop = map.y + (compact ? 54 : 72);
    const buttonHeight = compact ? 26 : 30;
    const buttonY = map.y + map.height - buttonHeight - (compact ? 8 : 10);
    // 丹炉内所有物品格统一采用储物袋的标准格子尺寸。
    const desiredSlotSize = backpackLayout.cellWidth;
    const materialSlotSize = desiredSlotSize;
    const materialY = buttonY - materialSlotSize - (compact ? 8 : 14);
    const metricBarHeight = compact ? 7 : 9;
    const metricRowHeight = compact ? 22 : 26;
    const metricRowGap = compact ? 4 : 8;
    const metricBottom = materialY - (compact ? 8 : 12);
    // 火候左侧预留槽位按洞府地图格子标准绘制。
    const heatSlotSize = this.cellWidth;
    const heatRowHeight = Math.max(metricRowHeight, heatSlotSize);
    const metricAreaHeight = metricRowHeight + metricRowGap + heatRowHeight;
    const metricTop = metricBottom - metricAreaHeight;
    // 列优先：第一列放 1、2、3，第四格开始换到第二列顶部。
    const slotColumns = slotCount > 3 ? 2 : (slotCount > 0 ? 1 : 0);
    const slotRows = slotCount > 0 ? Math.min(3, slotCount) : 0;
    const slotGapCount = Math.max(0, slotRows - 1);
    const maxSlotSize = slotCount > 0
      ? (metricTop - 14 - slotTop - slotGap * slotGapCount) / slotRows
      : desiredSlotSize;
    const slotSize = Math.min(desiredSlotSize, Math.max(16, maxSlotSize));
    const slotX = map.x + 22;
    const resultX = map.x + map.width - 22 - slotSize;

    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, map.x, map.y, map.width, map.height, 6);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();

    this.cavePanelBackButton = { x: map.x + 14, y: map.y + 14, width: 86, height: 28 };
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, this.cavePanelBackButton.x, this.cavePanelBackButton.y, this.cavePanelBackButton.width, this.cavePanelBackButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.stroke();
    this.drawText(ctx, '返回洞府', this.cavePanelBackButton.x + this.cavePanelBackButton.width / 2, this.cavePanelBackButton.y + 14, 9, COLORS.text, 'bold', 'center');

    this.drawText(ctx, building.name, centerX, titleY, 16, COLORS.text, 'bold', 'center');
    if (!repaired) {
      this.drawText(ctx, '损坏', centerX, titleY + (compact ? 17 : 22), compact ? 8 : 9, this.buildingLevelColor(buildingState.level), 'normal', 'center');
    }

    for (let index = 0; index < slotCount; index += 1) {
      const col = Math.floor(index / 3);
      const row = index % 3;
      const x = slotX + col * (slotSize + slotGap);
      const y = slotTop + row * (slotSize + slotGap);
      ctx.fillStyle = index % 2 ? COLORS.cellAlt : COLORS.cell;
      roundedRect(ctx, x, y, slotSize, slotSize, 4);
      ctx.fill();
      ctx.strokeStyle = COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
    }
    const slotAreaHeight = slotCount > 0 ? (slotSize * slotRows) + (slotGap * slotGapCount) : slotSize;
    const resultY = slotCount > 0
      ? slotTop + Math.max(0, (slotAreaHeight - slotSize) / 2)
      : slotTop;
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, resultX, resultY, slotSize, slotSize, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.accent;
    ctx.lineWidth = 1;
    ctx.stroke();
    if (feature && feature.status === 'completed') {
      this.drawText(ctx, '丹', resultX + slotSize / 2, resultY + slotSize * 0.42, Math.min(22, slotSize * 0.38), '#f0c85a', 'bold', 'center');
    }
    this.drawText(ctx, '→', centerX, resultY + slotSize / 2, Math.max(14, slotSize * 0.42), COLORS.accent, 'normal', 'center');

    const metricWidth = Math.min(220, map.width - 44);
    const metricX = Math.max(map.x + 22, Math.min(
      resultX + slotSize / 2 - metricWidth / 2,
      map.x + map.width - 22 - metricWidth,
    ));
    const metricLabelWidth = compact ? 28 : 34;
    const metricBarX = metricX + metricLabelWidth;
    const metricBarWidth = Math.max(70, metricWidth - metricLabelWidth);
    const drawBar = (label, value, labelY, barY, color) => {
      this.drawText(ctx, label, metricX, labelY, compact ? 8 : 9, COLORS.muted, 'bold', 'left');
      if (label === '进度' || label === '火候') {
        this.drawText(ctx, `${Math.round(value * 100)}%`, metricX + metricWidth, labelY, compact ? 7 : 8, COLORS.faint, 'normal', 'right');
      }
      ctx.fillStyle = COLORS.cell;
      roundedRect(ctx, metricBarX, barY, metricBarWidth, metricBarHeight, 4);
      ctx.fill();
      if (value > 0) {
        ctx.fillStyle = color;
        roundedRect(ctx, metricBarX, barY, Math.max(4, metricBarWidth * value), metricBarHeight, 4);
        ctx.fill();
      }
      // 状态条外框与物品格使用同级暗灰线条，避免黑底上过亮。
      ctx.strokeStyle = COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
    };
    const progress = feature && Number.isFinite(feature.progress) ? Math.max(0, Math.min(1, feature.progress)) : 0;
    const progressLabelY = metricTop + (compact ? 5 : 6);
    const progressBarY = metricTop + (compact ? 12 : 14);
    const heatRowTop = metricTop + metricRowHeight + metricRowGap;
    const heatControlOffset = Math.max(0, (heatRowHeight - metricRowHeight) / 2);
    const heatLabelY = heatRowTop + heatControlOffset + (compact ? 5 : 6);
    const heatBarY = heatRowTop + heatControlOffset + (compact ? 12 : 14);
    drawBar('进度', progress, progressLabelY, progressBarY, COLORS.accent);

    // 火候左侧预留一个地图标准大小的空槽位，后续再接入具体功能。
    const heatSlotGap = compact ? 6 : 10;
    const heatSlotX = map.x + 22;
    const heatSlotY = heatRowTop + (heatRowHeight - heatSlotSize) / 2;
    this.alchemyHeatSlot = { x: heatSlotX, y: heatSlotY, width: heatSlotSize, height: heatSlotSize };
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, heatSlotX, heatSlotY, heatSlotSize, heatSlotSize, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();

    const heatMetricX = heatSlotX + heatSlotSize + heatSlotGap;
    const heatMetricWidth = Math.max(70, map.x + map.width - 22 - heatMetricX);
    const heatLabelWidth = compact ? 28 : 34;
    const heatMetricBarX = heatMetricX + heatLabelWidth;
    const heatMetricBarWidth = Math.max(48, heatMetricWidth - heatLabelWidth);
    const heatValue = Math.max(0, Math.min(1, alchemy.heat));
    this.drawText(ctx, '火候', heatMetricX, heatLabelY, compact ? 8 : 9, COLORS.muted, 'bold', 'left');
    this.drawText(ctx, `${Math.round(heatValue * 100)}%`, heatMetricX + heatMetricWidth, heatLabelY, compact ? 7 : 8, COLORS.faint, 'normal', 'right');
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, heatMetricBarX, heatBarY, heatMetricBarWidth, metricBarHeight, 4);
    ctx.fill();
    if (heatValue > 0) {
      ctx.fillStyle = '#d27b4d';
      roundedRect(ctx, heatMetricBarX, heatBarY, Math.max(4, heatMetricBarWidth * heatValue), metricBarHeight, 4);
      ctx.fill();
    }
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.stroke();
    this.alchemyHeatBar = { x: heatMetricBarX, y: heatBarY, width: heatMetricBarWidth, height: metricBarHeight };

    const materialBarX = map.x + 16;
    const materialBarWidth = map.width - 32;
    const materialSlotGap = 7;
    const materialSlotCount = 8;
    this.alchemyMaterialBar = {
      x: materialBarX,
      y: materialY - 7,
      width: materialBarWidth,
      height: materialSlotSize + 14,
      slotSize: materialSlotSize,
      gap: materialSlotGap,
      slotCount: materialSlotCount,
    };
    ctx.save();
    ctx.beginPath();
    ctx.rect(materialBarX, materialY - 7, materialBarWidth, materialSlotSize + 14);
    ctx.clip();
    const maxMaterialScroll = this.getAlchemyMaterialMaxScroll();
    const materialScroll = Math.max(0, Math.min(maxMaterialScroll, Number(alchemy.materialScrollX) || 0));
    for (let index = 0; index < materialSlotCount; index += 1) {
      const x = materialBarX + 4 + index * (materialSlotSize + materialSlotGap) - materialScroll;
      ctx.fillStyle = index % 2 ? COLORS.cellAlt : COLORS.cell;
      roundedRect(ctx, x, materialY, materialSlotSize, materialSlotSize, 4);
      ctx.fill();
      ctx.strokeStyle = COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
    }
    ctx.restore();
    if (maxMaterialScroll > 0) {
      this.drawText(ctx, '‹', materialBarX + 5, materialY + materialSlotSize / 2, 15, COLORS.faint, 'normal', 'center');
      this.drawText(ctx, '›', materialBarX + materialBarWidth - 5, materialY + materialSlotSize / 2, 15, COLORS.faint, 'normal', 'center');
    }

    const buttonGap = 10;
    const buttonWidth = Math.max(68, Math.min(106, (map.width - 32 - buttonGap) / 2));
    this.featureUpgradeButton = { x: map.x + 16, y: buttonY, width: buttonWidth, height: buttonHeight };
    this.featureActionButton = { x: map.x + map.width - 16 - buttonWidth, y: buttonY, width: buttonWidth, height: buttonHeight };
    const drawButton = (button, label, active) => {
      ctx.fillStyle = active ? COLORS.current : COLORS.cellAlt;
      roundedRect(ctx, button.x, button.y, button.width, button.height, 4);
      ctx.fill();
      ctx.strokeStyle = active ? COLORS.accent : COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
      this.drawText(ctx, label, button.x + button.width / 2, button.y + button.height / 2, compact ? 8 : 9, active ? COLORS.text : COLORS.faint, 'bold', 'center');
    };
    const upgradeLabel = repaired ? (buildingState.level >= (building.maxLevel || 6) ? '已满级' : '升级') : '修缮';
    drawButton(this.featureUpgradeButton, upgradeLabel, !repaired || buildingState.level < (building.maxLevel || 6));
    const actionLabel = !repaired
      ? '需先修缮'
      : feature && feature.status === 'brewing'
        ? '炼制中'
        : feature && feature.status === 'completed'
          ? '领取丹药'
          : '开始炼丹';
    drawButton(this.featureActionButton, actionLabel, repaired);
  }

  renderAlchemyUpgradeOverlay(ctx, state, building) {
    // 升级弹窗覆盖丹炉页原有面板，不改变页面尺寸。
    const map = this.mapRect;
    const centerX = map.x + map.width / 2;
    const panel = { x: map.x, y: map.y, width: map.width, height: map.height };
    ctx.save();
    ctx.fillStyle = 'rgba(0, 0, 0, 0.72)';
    ctx.fillRect(0, 0, this.width, this.height);
    ctx.fillStyle = COLORS.surface;
    roundedRect(ctx, panel.x, panel.y, panel.width, panel.height, 6);
    ctx.fill();
    ctx.strokeStyle = COLORS.accent;
    ctx.lineWidth = 1;
    ctx.stroke();
    this.alchemyUpgradeCloseButton = { x: panel.x + panel.width - 46, y: panel.y + 14, width: 32, height: 28 };
    const upgradeButtonGap = 10;
    const upgradeButtonWidth = Math.max(68, Math.min(106, (panel.width - 32 - upgradeButtonGap) / 2));
    this.alchemyUpgradeConfirmButton = { x: panel.x + panel.width - 16 - upgradeButtonWidth, y: panel.y + panel.height - 48, width: upgradeButtonWidth, height: 30 };
    const cancelButton = { x: panel.x + 16, y: panel.y + panel.height - 48, width: upgradeButtonWidth, height: 30 };
    this.alchemyUpgradeCancelButton = cancelButton;
    this.drawText(ctx, '丹炉升级', centerX, panel.y + 42, 16, COLORS.text, 'bold', 'center');
    this.drawText(ctx, '×', this.alchemyUpgradeCloseButton.x + 16, this.alchemyUpgradeCloseButton.y + 14, 20, COLORS.muted, 'normal', 'center');
    ctx.strokeStyle = COLORS.lineSoft;
    ctx.lineWidth = 1;
    ctx.strokeRect(panel.x + 22, panel.y + 74, panel.width - 44, panel.height - 136);
    ctx.fillStyle = COLORS.cell;
    roundedRect(ctx, cancelButton.x, cancelButton.y, cancelButton.width, cancelButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.line;
    ctx.stroke();
    this.drawText(ctx, '返回', cancelButton.x + cancelButton.width / 2, cancelButton.y + 15, 9, COLORS.text, 'bold', 'center');
    ctx.fillStyle = COLORS.current;
    roundedRect(ctx, this.alchemyUpgradeConfirmButton.x, this.alchemyUpgradeConfirmButton.y, this.alchemyUpgradeConfirmButton.width, this.alchemyUpgradeConfirmButton.height, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.accent;
    ctx.stroke();
    this.drawText(ctx, '升级', this.alchemyUpgradeConfirmButton.x + this.alchemyUpgradeConfirmButton.width / 2, this.alchemyUpgradeConfirmButton.y + 15, 9, COLORS.text, 'bold', 'center');
    this.featureUpgradeButton = null;
    this.featureActionButton = null;
    ctx.restore();
  }

  featureDescription(id) {
    if (id === 'spiritField') return '管理灵田，培育灵草';
    if (id === 'alchemyFurnace') return '控制丹炉，消耗灵草炼制丹药';
    if (id === 'meditationRoom') return '进入静室，开始修炼';
    if (id === 'scripturePavilion') return '翻阅典籍，研究功法';
    return '洞府功能';
  }

  featureStatusText(id, feature) {
    if (!feature) return '未启用';
    if (feature.status === 'completed') return '已完成，等待领取';
    if (id === 'spiritField') return feature.status === 'growing' ? '灵草生长中' : '等待种植';
    if (id === 'alchemyFurnace') return feature.status === 'brewing' ? '炼制进行中' : '丹炉空闲';
    if (id === 'meditationRoom') return feature.status === 'cultivating' ? '修炼进行中' : '静室空闲';
    if (id === 'scripturePavilion') {
      if (feature.status === 'researching') return '研究引气诀中';
      if (feature.status === 'learned') return '已记录引气诀';
      return '尚未研究';
    }
    return '未启用';
  }

  featureActionLabel(id, feature) {
    if (!feature) return '开始';
    if (feature.status === 'completed') {
      if (id === 'spiritField') return '收获灵草';
      if (id === 'alchemyFurnace') return '领取丹药';
      if (id === 'meditationRoom') return '领取修为';
      if (id === 'scripturePavilion') return '领取功法';
    }
    if (id === 'spiritField') return feature.status === 'growing' ? '查看生长' : '开始种植';
    if (id === 'alchemyFurnace') return feature.status === 'brewing' ? '查看炼制' : '开始炼丹';
    if (id === 'meditationRoom') return feature.status === 'cultivating' ? '查看修炼' : '开始修炼';
    if (id === 'scripturePavilion') {
      if (feature.status === 'researching') return '查看研究';
      if (feature.status === 'learned') return '查看功法';
      return '开始研究';
    }
    return '开始';
  }

  featureProgressText(feature) {
    if (!feature || feature.status === 'idle' || feature.status === 'empty') return '尚未开始';
    if (feature.status === 'learned') return '已记录';
    if (feature.status === 'completed') return '进度 100%';
    const percent = Math.round((feature.progress || 0) * 100);
    const remaining = Math.max(0, Math.ceil((feature.remainingMs || 0) / 1000));
    return `进度 ${percent}% · 剩余 ${remaining}秒`;
  }

  featureRewardText(id, state) {
    const resources = state.resources || {};
    if (id === 'spiritField') return `灵草库存：${resources.spiritHerbs || 0}`;
    if (id === 'alchemyFurnace') return `灵草 ${resources.spiritHerbs || 0} · 回气丹 ${resources.qiPills || 0}`;
    if (id === 'meditationRoom') return `累计修为：${resources.cultivation || 0}`;
    if (id === 'scripturePavilion') return `已记录功法：${(resources.techniques || []).length}`;
    return '';
  }

  featureRewardPreview(id, state) {
    const selected = state.cave && state.cave.buildings ? state.cave.buildings[id] : null;
    const level = selected && selected.level > 0 ? Math.min(6, selected.level) : 1;
    if (id === 'spiritField') return `${1 + Math.floor((level - 1) / 2)}份灵草`;
    if (id === 'alchemyFurnace') return `${1 + Math.floor((level - 1) / 3)}份回气丹`;
    if (id === 'meditationRoom') {
      const baseReward = 10 + ((level - 1) * 3);
      const hasTechnique = state.resources && Array.isArray(state.resources.techniques)
        && state.resources.techniques.includes('引气诀');
      const reward = hasTechnique ? Math.round(baseReward * 1.25) : baseReward;
      return `${reward}点修为`;
    }
    if (id === 'scripturePavilion') return '功法引气诀';
    return '基础奖励';
  }

  renderWorldDungeonPrompt(ctx, state, dungeon) {
    this.worldDungeonEnterButton = null;
    if (!dungeon) return;

    const left = 16;
    const top = this.controlRect.y + 18;
    const width = Math.max(150, this.width * 0.49);
    const right = left + width;
    this.drawFittedText(ctx, dungeon.name, left, top, width, 13, COLORS.text, 'bold', 'left');
    this.drawFittedText(ctx, `危险 ${'◆'.repeat(dungeon.danger)} · 消耗 ${dungeon.qiCost}灵气`, left, top + 23, width, 8, COLORS.accent, 'normal', 'left');
    this.drawFittedText(ctx, dungeon.description, left, top + 43, width, 8, COLORS.faint, 'normal', 'left');
    this.drawFittedText(ctx, `收益 修为 ${dungeon.cultivation[0]}-${dungeon.cultivation[1]} · 灵草 ${dungeon.spiritHerbs[0]}-${dungeon.spiritHerbs[1]}`, left, top + 61, width, 7, COLORS.muted, 'normal', 'left');

    const button = {
      x: left,
      y: this.controlRect.y + this.controlHeight - 47,
      width: Math.min(112, width),
      height: 30,
    };
    this.worldDungeonEnterButton = button;
    ctx.fillStyle = COLORS.current;
    roundedRect(ctx, button.x, button.y, button.width, button.height, 4);
    ctx.fill();
    ctx.strokeStyle = COLORS.accent;
    ctx.lineWidth = 1;
    ctx.stroke();
    this.drawText(ctx, '进入秘境', button.x + button.width / 2, button.y + button.height / 2, 9, COLORS.text, 'bold', 'center');
  }

  renderControlPad(ctx, state, dungeon = null) {
    this.renderWorldDungeonPrompt(ctx, state, dungeon);
    const center = this.padCenter;
    const button = this.controlButtonSize;
    const positions = this.controlButtons;

    ctx.save();
    ctx.strokeStyle = COLORS.line;
    ctx.lineWidth = 1;
    ctx.beginPath();
    ctx.arc(center.x, center.y, this.padSize * 0.44, 0, Math.PI * 2);
    ctx.stroke();
    ctx.restore();

    Object.keys(positions).forEach((direction) => {
      const point = positions[direction];
      ctx.fillStyle = COLORS.cell;
      roundedRect(ctx, point.x - button / 2, point.y - button / 2, button, button, 5);
      ctx.fill();
      ctx.strokeStyle = COLORS.line;
      ctx.lineWidth = 1;
      ctx.stroke();
      const arrow = direction === 'up' ? '↑' : direction === 'down' ? '↓' : direction === 'left' ? '←' : '→';
      this.drawText(ctx, arrow, point.x, point.y + 1, Math.max(18, button * 0.45), COLORS.text, 'normal', 'center');
    });

    ctx.fillStyle = COLORS.cellAlt;
    ctx.beginPath();
    ctx.arc(center.x, center.y, Math.max(14, button * 0.3), 0, Math.PI * 2);
    ctx.fill();
    this.drawText(ctx, '林', center.x, center.y, 11, state.player.realmColor || COLORS.player, 'bold', 'center');
  }

  getControlDirection(x, y) {
    const half = this.controlButtonSize / 2;
    const touchPadding = Math.min(10, this.controlButtonSize * 0.18);
    const hitHalf = half + touchPadding;
    const buttons = this.controlButtons;
    const directions = ['up', 'down', 'left', 'right'];
    return directions.find((direction) => {
      const button = buttons[direction];
      return Math.abs(x - button.x) <= hitHalf && Math.abs(y - button.y) <= hitHalf;
    }) || null;
  }
}

export { COLORS };
