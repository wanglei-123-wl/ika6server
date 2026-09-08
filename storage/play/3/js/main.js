import { SCREEN_WIDTH, SCREEN_HEIGHT, canvas } from './render.js';
import GameUI from './ui.js';
import CAVE_BUILDINGS from './cave.js';
import DUNGEONS from './dungeons.js';
import {
  addToStacks,
  itemDefinition,
  LEGACY_RESOURCE_ITEM_IDS,
  normalizeItemStacks,
  removeFromStacks,
  resolveItemId,
  setStackCount,
  stackCount,
  techniqueItemId,
  techniqueNameForItemId,
} from './items.js';

const ctx = canvas.getContext('2d');
const ui = new GameUI(SCREEN_WIDTH, SCREEN_HEIGHT);
const SAVE_KEY = 'wenda_life_save_v1';
const SAVE_VERSION = 4;

const EQUIPMENT_SLOT_BY_CATEGORY = Object.freeze({
  weapon: 'weapon',
  artifact: 'artifact',
  armor: 'armor',
  storage: 'storage',
  technique: 'technique',
  spell: 'spell',
});
const EQUIPMENT_SLOTS = Object.freeze(['weapon', 'artifact', 'armor', 'storage', 'technique', 'spell']);

const REALMS = [
  { id: 'mortal', family: '凡人', stageName: '', stageIndex: 0, threshold: 0, color: '#ffffff', major: false },
  { id: 'qi_1', family: '练气', stageName: '一层', stageIndex: 1, threshold: 30, color: '#ffffff', major: false },
  { id: 'qi_2', family: '练气', stageName: '二层', stageIndex: 2, threshold: 60, color: '#ffffff', major: false },
  { id: 'qi_3', family: '练气', stageName: '三层', stageIndex: 3, threshold: 95, color: '#ffffff', major: false },
  { id: 'qi_4', family: '练气', stageName: '四层', stageIndex: 4, threshold: 135, color: '#ffffff', major: false },
  { id: 'qi_5', family: '练气', stageName: '五层', stageIndex: 5, threshold: 180, color: '#ffffff', major: false },
  { id: 'qi_6', family: '练气', stageName: '六层', stageIndex: 6, threshold: 230, color: '#ffffff', major: false },
  { id: 'qi_7', family: '练气', stageName: '七层', stageIndex: 7, threshold: 285, color: '#ffffff', major: false },
  { id: 'qi_8', family: '练气', stageName: '八层', stageIndex: 8, threshold: 345, color: '#ffffff', major: false },
  { id: 'qi_9', family: '练气', stageName: '九层', stageIndex: 9, threshold: 410, color: '#ffffff', major: false },
  { id: 'foundation_early', family: '筑基', stageName: '初期', stageIndex: 1, threshold: 480, color: '#70d58c', major: true },
  { id: 'foundation_middle', family: '筑基', stageName: '中期', stageIndex: 2, threshold: 560, color: '#70d58c', major: false },
  { id: 'foundation_late', family: '筑基', stageName: '后期', stageIndex: 3, threshold: 650, color: '#70d58c', major: false },
  { id: 'foundation_peak', family: '筑基', stageName: '巅峰', stageIndex: 4, threshold: 750, color: '#70d58c', major: false },
  { id: 'gold_core', family: '金丹', stageName: '', stageIndex: 1, threshold: 900, color: '#f0c85a', major: true },
  { id: 'nascent_soul_early', family: '元婴', stageName: '初期', stageIndex: 1, threshold: 1100, color: '#b68cff', major: true },
  { id: 'nascent_soul_middle', family: '元婴', stageName: '中期', stageIndex: 2, threshold: 1350, color: '#b68cff', major: false },
  { id: 'nascent_soul_late', family: '元婴', stageName: '后期', stageIndex: 3, threshold: 1650, color: '#b68cff', major: false },
  { id: 'nascent_soul_peak', family: '元婴', stageName: '巅峰', stageIndex: 4, threshold: 2000, color: '#b68cff', major: false },
  { id: 'nascent_soul_transform', family: '元婴', stageName: '婴变', stageIndex: 5, threshold: 2400, color: '#b68cff', major: false },
  { id: 'deity_transform', family: '化神', stageName: '', stageIndex: 1, threshold: 3000, color: '#ef6666', major: true },
];

const CULTIVATION_SOURCES = Object.freeze({
  meditation: '静室修炼',
  pill: '炼化丹药',
  spiritStone: '炼化灵石',
  treasure: '天材地宝',
  dungeon: '秘境探索',
});

const BREAKTHROUGH_FAILURE_RATES = Object.freeze({
  foundation_early: 0.2,
  gold_core: 0.3,
  nascent_soul_early: 0.38,
  deity_transform: 0.5,
});

const REALM_FAMILY_RANK = Object.freeze({
  凡人: 0,
  练气: 1,
  筑基: 2,
  金丹: 3,
  元婴: 4,
  化神: 5,
});

const FEATURE_CONFIG = {
  spiritField: { duration: 60 * 1000, activeStatus: 'growing', rewardName: '灵草', rewardKey: 'spiritHerbs', rewardAmount: 1 },
  alchemyFurnace: { duration: 90 * 1000, activeStatus: 'brewing', rewardName: '回气丹', rewardKey: 'qiPills', rewardAmount: 1, materialKey: 'spiritHerbs', materialName: '灵草', materialAmount: 1 },
  meditationRoom: { duration: 120 * 1000, activeStatus: 'cultivating', rewardName: '修为', rewardKey: 'cultivation', rewardAmount: 10 },
  scripturePavilion: { duration: 45 * 1000, activeStatus: 'researching', rewardName: '引气诀', rewardKey: 'techniques', rewardAmount: 1 },
};

function createCaveBuildings() {
  return CAVE_BUILDINGS.reduce((result, building) => {
    result[building.id] = { level: 0, status: 'ruined' };
    return result;
  }, {});
}

function createFeatureStates() {
  return {
    spiritField: { status: 'empty', startedAt: 0, duration: FEATURE_CONFIG.spiritField.duration, completedAt: 0, progress: 0, remainingMs: 0 },
    alchemyFurnace: { status: 'idle', startedAt: 0, duration: FEATURE_CONFIG.alchemyFurnace.duration, completedAt: 0, progress: 0, remainingMs: 0 },
    meditationRoom: { status: 'idle', startedAt: 0, duration: FEATURE_CONFIG.meditationRoom.duration, completedAt: 0, progress: 0, remainingMs: 0 },
    scripturePavilion: { status: 'idle', startedAt: 0, duration: FEATURE_CONFIG.scripturePavilion.duration, completedAt: 0, progress: 0, remainingMs: 0 },
  };
}

function createAlchemyState() {
  return {
    heat: 0.5,
    materialScrollX: 0,
    upgradeOpen: false,
  };
}

function createDungeonStats() {
  return DUNGEONS.reduce((result, dungeon) => {
    result[dungeon.name] = { count: 0, lastAt: 0, bestCultivation: 0, bestHerbs: 0 };
    return result;
  }, {});
}

function normalizeDungeonStats(savedStats) {
  const stats = createDungeonStats();
  if (!savedStats || typeof savedStats !== 'object') return stats;
  DUNGEONS.forEach((dungeon) => {
    const source = savedStats[dungeon.name];
    if (!source || typeof source !== 'object') return;
    stats[dungeon.name] = {
      count: Number.isFinite(source.count) ? Math.max(0, Math.floor(source.count)) : 0,
      lastAt: Number.isFinite(source.lastAt) ? Math.max(0, source.lastAt) : 0,
      bestCultivation: Number.isFinite(source.bestCultivation) ? Math.max(0, Math.floor(source.bestCultivation)) : 0,
      bestHerbs: Number.isFinite(source.bestHerbs) ? Math.max(0, Math.floor(source.bestHerbs)) : 0,
    };
  });
  return stats;
}

function normalizeInventory(savedInventory) {
  const inventory = savedInventory || {};
  const level = Number.isFinite(inventory.level) ? Math.max(1, Math.min(5, Math.floor(inventory.level))) : 1;
  const scrollY = Number.isFinite(inventory.scrollY) ? Math.max(0, inventory.scrollY) : 0;
  const slotCount = 12 + ((level - 1) * 4);
  const rawItems = Array.isArray(inventory.items) ? inventory.items : null;
  const hasFixedSlots = Boolean(rawItems && (inventory.slotted === true || rawItems.some((entry) => entry == null)));
  const compactItems = rawItems && !hasFixedSlots ? normalizeItemStacks(rawItems) : normalizeItemStacks(inventory.itemStacks);
  const items = hasFixedSlots
    ? rawItems.map((entry) => {
      if (!entry || typeof entry.id !== 'string' || !entry.id) return null;
      const count = Math.max(0, Math.floor(Number(entry.count) || 0));
      return count > 0 ? { id: entry.id, count } : null;
    })
    : compactItems;
  while (items.length < slotCount) items.push(null);
  return { level, scrollY, items, open: false, selectedItemId: null, drag: null };
}

function normalizePlayerProfile(savedPlayer) {
  const player = savedPlayer || {};
  const maxHp = Number.isFinite(player.maxHp) ? Math.max(1, Math.floor(player.maxHp)) : 100;
  const hp = Number.isFinite(player.hp) ? Math.max(0, Math.min(maxHp, Math.floor(player.hp))) : maxHp;
  const sourceAttributes = player.attributes && typeof player.attributes === 'object' ? player.attributes : {};
  const attributes = {
    physique: Number.isFinite(sourceAttributes.physique) ? Math.max(0, Math.floor(sourceAttributes.physique)) : 10,
    spirit: Number.isFinite(sourceAttributes.spirit) ? Math.max(0, Math.floor(sourceAttributes.spirit)) : 10,
    vitality: Number.isFinite(sourceAttributes.vitality)
      ? Math.max(0, Math.floor(sourceAttributes.vitality))
      : (Number.isFinite(sourceAttributes.agility) ? Math.max(0, Math.floor(sourceAttributes.agility)) : 10),
    sense: Number.isFinite(sourceAttributes.sense) ? Math.max(0, Math.floor(sourceAttributes.sense)) : 10,
  };
  const sourceEquipment = player.equipment && typeof player.equipment === 'object' ? player.equipment : {};
  // Keep existing saves compatible while using the six current equipment categories.
  const equipmentSources = {
    weapon: ['weapon'],
    artifact: ['artifact', 'head'],
    armor: ['armor'],
    storage: ['storage', 'wrist'],
    technique: ['technique', 'waist'],
    spell: ['spell', 'feet'],
  };
  const equipment = Object.entries(equipmentSources).reduce((result, [slot, sourceSlots]) => {
    const value = sourceSlots.map((sourceSlot) => sourceEquipment[sourceSlot]).find((candidate) => (
      typeof candidate === 'string' || (candidate && typeof candidate === 'object')
    ));
    result[slot] = typeof value === 'string' || (value && typeof value === 'object') ? value : null;
    return result;
  }, {});
  const breakthrough = player.breakthrough && typeof player.breakthrough === 'object' ? player.breakthrough : {};
  return {
    hp,
    maxHp,
    attributes,
    equipment,
    breakthrough: {
      attempts: Number.isFinite(breakthrough.attempts) ? Math.max(0, Math.floor(breakthrough.attempts)) : 0,
      failures: Number.isFinite(breakthrough.failures) ? Math.max(0, Math.floor(breakthrough.failures)) : 0,
      lastResult: typeof breakthrough.lastResult === 'string' ? breakthrough.lastResult : '',
    },
  };
}

const state = {
  scene: 'world',
  player: {
    row: 4,
    col: 3,
    name: '林',
    realmIndex: 0,
    realmId: 'mortal',
    realmFamily: '凡人',
    realmStage: '',
    realm: '凡人',
    realmThreshold: 0,
    realmMajor: false,
    hp: 100,
    maxHp: 100,
    attributes: { physique: 10, spirit: 10, vitality: 10, sense: 10 },
    breakthrough: { attempts: 0, failures: 0, lastResult: '' },
    equipment: { weapon: null, artifact: null, armor: null, storage: null, technique: null, spell: null },
  },
  lastMove: '站在空地',
  qi: 0,
  qiCap: 100,
  pendingQi: 0,
  resources: {
    spiritHerbs: 0,
    qiPills: 0,
    cultivation: 0,
    techniques: [],
  },
  dungeonStats: createDungeonStats(),
  inventory: { level: 1, scrollY: 0, items: Array(12).fill(null), open: false, selectedItemId: null, drag: null },
  cave: {
    panel: 'map',
    selectedBuildingId: null,
    activeBuildingId: null,
    player: { row: 8, col: 7 },
    buildings: createCaveBuildings(),
    features: createFeatureStates(),
    alchemy: createAlchemyState(),
  },
  dungeon: {
    name: null,
    result: null,
  },
  attributesOpen: false,
  itemDetailOpen: false,
};

let lastUpdateAt = Date.now();
let lastSavedAt = Date.now();
let lastPersistAt = 0;

function isValidCell(cell) {
  return cell && Number.isInteger(cell.row) && Number.isInteger(cell.col)
    && cell.row >= 0 && cell.row < ui.rows && cell.col >= 0 && cell.col < ui.cols;
}

function isCaveBuildingCell(row, col) {
  return CAVE_BUILDINGS.some((building) => building.row === row && building.col === col);
}

function caveBuildingAt(row, col) {
  return CAVE_BUILDINGS.find((building) => building.row === row && building.col === col) || null;
}

function syncCaveBuildingSelection() {
  const building = caveBuildingAt(state.cave.player.row, state.cave.player.col);
  state.cave.selectedBuildingId = building ? building.id : null;
  return building;
}

function copyBuildingState(source, fallback) {
  const level = source && Number.isInteger(source.level) ? Math.max(0, Math.min(6, source.level)) : fallback.level;
  return { level, status: level > 0 ? 'repaired' : 'ruined' };
}

function copyFeatureState(source, fallback, id) {
  const config = FEATURE_CONFIG[id];
  const base = fallback || { ...createFeatureStates()[id] };
  const statusList = id === 'spiritField'
    ? ['empty', config.activeStatus, 'completed']
    : id === 'scripturePavilion'
      ? ['idle', config.activeStatus, 'completed', 'learned', 'researched']
      : ['idle', config.activeStatus, 'completed'];
  let status = source && statusList.includes(source.status) ? source.status : base.status;
  if (id === 'scripturePavilion' && status === 'researched') status = 'learned';
  const startedAt = source && Number.isFinite(source.startedAt) ? Math.max(0, source.startedAt) : 0;
  const completedAt = source && Number.isFinite(source.completedAt) ? Math.max(0, source.completedAt) : 0;
  const duration = source && Number.isFinite(source.duration) && source.duration > 0
    ? source.duration
    : config.duration;
  const progress = source && Number.isFinite(source.progress) ? Math.max(0, Math.min(1, source.progress)) : 0;
  const remainingMs = source && Number.isFinite(source.remainingMs) ? Math.max(0, source.remainingMs) : 0;
  return { status, startedAt, duration, completedAt, progress, remainingMs };
}

function normalizeAlchemyState(savedAlchemy) {
  const source = savedAlchemy && typeof savedAlchemy === 'object' ? savedAlchemy : {};
  const heat = Number.isFinite(source.heat) ? Math.max(0, Math.min(1, source.heat)) : 0.5;
  const materialScrollX = Number.isFinite(source.materialScrollX) ? Math.max(0, source.materialScrollX) : 0;
  return { heat, materialScrollX, upgradeOpen: false };
}

function normalizeResources(savedResources) {
  const resources = savedResources || {};
  const techniques = Array.isArray(resources.techniques)
    ? resources.techniques.filter((name) => typeof name === 'string').slice(0, 50)
    : [];
  return {
    spiritHerbs: Number.isFinite(resources.spiritHerbs) ? Math.max(0, Math.floor(resources.spiritHerbs)) : 0,
    qiPills: Number.isFinite(resources.qiPills) ? Math.max(0, Math.floor(resources.qiPills)) : 0,
    cultivation: Number.isFinite(resources.cultivation) ? Math.max(0, Math.floor(resources.cultivation)) : 0,
    techniques,
  };
}

function syncLegacyResourcesToItems() {
  if (!Array.isArray(state.inventory.items)) state.inventory.items = [];
  setStackCount(state.inventory.items, LEGACY_RESOURCE_ITEM_IDS.spiritHerbs, state.resources.spiritHerbs);
  setStackCount(state.inventory.items, LEGACY_RESOURCE_ITEM_IDS.qiPills, state.resources.qiPills);
  state.resources.techniques.forEach((name) => {
    const itemId = techniqueItemId(name);
    if (itemId) setStackCount(state.inventory.items, itemId, 1);
  });
}

function syncItemsToLegacyResources() {
  state.resources.spiritHerbs = stackCount(state.inventory.items, LEGACY_RESOURCE_ITEM_IDS.spiritHerbs);
  state.resources.qiPills = stackCount(state.inventory.items, LEGACY_RESOURCE_ITEM_IDS.qiPills);
  const techniqueNames = state.inventory.items
    .filter((stack) => stack && typeof stack.id === 'string')
    .map((stack) => techniqueNameForItemId(stack.id))
    .filter(Boolean);
  state.resources.techniques = Array.from(new Set([...state.resources.techniques, ...techniqueNames]));
}

function addItem(itemId, amount = 1) {
  if (!Array.isArray(state.inventory.items)) state.inventory.items = [];
  addToStacks(state.inventory.items, itemId, amount);
  syncItemsToLegacyResources();
}

function removeItem(itemId, amount = 1) {
  if (!Array.isArray(state.inventory.items)) state.inventory.items = [];
  removeFromStacks(state.inventory.items, itemId, amount);
  syncItemsToLegacyResources();
}

function realmDisplayName(realm) {
  if (!realm) return '凡人';
  return realm.stageName ? `${realm.family}${realm.stageName}` : realm.family;
}

function refreshRealm(syncFromCultivation = false) {
  const cultivation = state.resources && Number.isFinite(state.resources.cultivation)
    ? state.resources.cultivation
    : 0;
  let realmIndex = Number.isInteger(state.player.realmIndex)
    ? Math.max(0, Math.min(REALMS.length - 1, state.player.realmIndex))
    : 0;
  if (syncFromCultivation) {
    realmIndex = 0;
    REALMS.forEach((realm, index) => {
      if (cultivation >= realm.threshold) realmIndex = index;
    });
  }
  state.player.realmIndex = realmIndex;
  const currentRealm = REALMS[realmIndex];
  state.player.realmId = currentRealm.id;
  state.player.realmFamily = currentRealm.family;
  state.player.realmStage = currentRealm.stageName;
  state.player.realm = realmDisplayName(currentRealm);
  state.player.realmColor = currentRealm.color;
  const nextRealm = REALMS[realmIndex + 1];
  state.player.nextRealm = nextRealm ? realmDisplayName(nextRealm) : '已至顶峰';
  state.player.nextRealmThreshold = nextRealm ? nextRealm.threshold : cultivation;
  state.player.nextRealmMajor = Boolean(nextRealm && nextRealm.major);
  state.player.nextBreakthroughFailureRate = breakthroughFailureRate(nextRealm);
  state.player.canBreakthrough = Boolean(nextRealm && cultivation >= nextRealm.threshold);
  state.player.realmThreshold = currentRealm.threshold;
  state.player.realmMajor = Boolean(currentRealm.major);
  state.player.realmProgress = nextRealm
    ? `${cultivation} / ${nextRealm.threshold}`
    : `${cultivation}`;
}

function addCultivation(amount, source = '') {
  const value = Number.isFinite(amount) ? Math.max(0, Math.floor(amount)) : 0;
  if (value <= 0) return false;
  const previousRealmIndex = state.player.realmIndex;
  state.resources.cultivation += value;
  refreshRealm(false);
  if (source && CULTIVATION_SOURCES[source]) {
    state.lastMove = `${CULTIVATION_SOURCES[source]}，修为 +${value}`;
  }
  return state.player.realmIndex > previousRealmIndex;
}

function derivedPlayerStats() {
  const attributes = state.player.attributes || {};
  return {
    maxHp: 100 + (Math.max(0, Number(attributes.physique) || 0) * 10),
    qiCap: 100 + (Math.max(0, Number(attributes.spirit) || 0) * 10),
    power: 10 + (Math.max(0, Number(attributes.vitality) || 0) * 2),
    sense: Math.max(0, Number(attributes.sense) || 0),
  };
}

function applyDerivedPlayerStats(preserveHp = true) {
  const stats = derivedPlayerStats();
  const previousMaxHp = Math.max(1, Number(state.player.maxHp) || stats.maxHp);
  const previousHp = Math.max(0, Number(state.player.hp) || 0);
  state.player.maxHp = stats.maxHp;
  state.player.hp = preserveHp
    ? Math.min(stats.maxHp, previousHp + Math.max(0, stats.maxHp - previousMaxHp))
    : stats.maxHp;
  refreshQiCapacity();
  state.qiCap = Math.max(state.qiCap, stats.qiCap);
  state.qi = Math.min(state.qi, state.qiCap);
}

function isMajorBreakthrough(targetRealm) {
  return Boolean(targetRealm && targetRealm.major);
}

function breakthroughFailureRate(targetRealm) {
  if (!targetRealm) return 0;
  return Number.isFinite(BREAKTHROUGH_FAILURE_RATES[targetRealm.id])
    ? BREAKTHROUGH_FAILURE_RATES[targetRealm.id]
    : 0;
}

function breakthroughAttributeReward(targetRealm) {
  if (!targetRealm) return { physique: 0, spirit: 0, vitality: 0, sense: 0 };
  const major = isMajorBreakthrough(targetRealm);
  const familyRank = REALM_FAMILY_RANK[targetRealm.family] || 0;
  const stageIndex = Math.max(1, Number(targetRealm.stageIndex) || 1);
  const base = major ? (4 + familyRank * 2) : (1 + Math.floor(familyRank / 2));
  if (major) {
    return {
      physique: base + 1,
      spirit: base + 2,
      vitality: base + 1,
      sense: base,
    };
  }
  return {
    physique: base + (stageIndex % 3 === 0 ? 1 : 0),
    spirit: base + (stageIndex % 4 === 0 ? 1 : 0),
    vitality: base + (stageIndex % 2 === 0 ? 1 : 0),
    sense: base + (stageIndex >= 5 ? 1 : 0),
  };
}

function attemptBreakthrough() {
  const currentIndex = Math.max(0, Math.min(REALMS.length - 1, state.player.realmIndex || 0));
  const currentRealm = REALMS[currentIndex];
  const targetRealm = REALMS[currentIndex + 1];
  const cultivation = state.resources.cultivation;
  if (!targetRealm) {
    state.lastMove = '已至最高境界';
    return;
  }
  if (cultivation < targetRealm.threshold) {
    state.lastMove = `修为不足，还需${targetRealm.threshold - cultivation}`;
    return;
  }

  const breakthrough = state.player.breakthrough || { attempts: 0, failures: 0, lastResult: '' };
  breakthrough.attempts += 1;
  const failed = isMajorBreakthrough(targetRealm) && Math.random() < breakthroughFailureRate(targetRealm);
  if (failed) {
    breakthrough.failures += 1;
    breakthrough.lastResult = 'failed';
    const loss = Math.max(1, Math.floor((targetRealm.threshold - currentRealm.threshold) * 0.12));
    state.resources.cultivation = Math.max(currentRealm.threshold, cultivation - loss);
    state.player.breakthrough = breakthrough;
    refreshRealm(false);
    state.lastMove = `突破失败，修为损耗${loss}`;
    saveState();
    return;
  }

  const reward = breakthroughAttributeReward(targetRealm);
  state.player.attributes.physique += reward.physique;
  state.player.attributes.spirit += reward.spirit;
  state.player.attributes.vitality += reward.vitality;
  state.player.attributes.sense += reward.sense;
  breakthrough.lastResult = 'success';
  state.player.breakthrough = breakthrough;
  state.player.realmIndex = currentIndex + 1;
  refreshRealm(false);
  applyDerivedPlayerStats(true);
  state.lastMove = `突破至${state.player.realm}`;
  saveState();
}

function loadState() {
  let saved;
  try {
    saved = wx.getStorageSync(SAVE_KEY);
  } catch (error) {
    saved = null;
  }
  if (!saved || !Number.isFinite(saved.version) || saved.version > SAVE_VERSION) return;

  if (isValidCell(saved.player)) {
    state.player.row = saved.player.row;
    state.player.col = saved.player.col;
  }
  if (isValidCell(saved.cavePlayer)) {
    state.cave.player.row = saved.cavePlayer.row;
    state.cave.player.col = saved.cavePlayer.col;
  }
  if (Number.isFinite(saved.qi)) state.qi = Math.max(0, saved.qi);
  if (Number.isFinite(saved.pendingQi)) state.pendingQi = Math.max(0, saved.pendingQi);
  state.resources = normalizeResources(saved.resources);
  state.dungeonStats = normalizeDungeonStats(saved.dungeonStats);
  state.inventory = normalizeInventory(saved.inventory);
  if (state.inventory.items.length === 0) syncLegacyResourcesToItems();
  else syncItemsToLegacyResources();
  const profile = normalizePlayerProfile(saved.player);
  state.player.hp = profile.hp;
  state.player.maxHp = profile.maxHp;
  state.player.attributes = profile.attributes;
  state.player.equipment = profile.equipment;
  state.player.breakthrough = profile.breakthrough;
  const savedRealmIndex = saved.player && Number.isInteger(saved.player.realmIndex)
    ? saved.player.realmIndex
    : (Number.isInteger(saved.realmIndex) ? saved.realmIndex : 0);
  state.player.realmIndex = Math.max(0, Math.min(REALMS.length - 1, savedRealmIndex));
  refreshRealm(saved.version < 3);

  CAVE_BUILDINGS.forEach((building) => {
    state.cave.buildings[building.id] = copyBuildingState(saved.buildings && saved.buildings[building.id], state.cave.buildings[building.id]);
  });
  Object.keys(state.cave.features).forEach((id) => {
    const savedFeature = saved.features && saved.features[id];
    state.cave.features[id] = copyFeatureState(savedFeature, state.cave.features[id], id);
  });
  state.cave.alchemy = normalizeAlchemyState(saved.alchemy);

  const savedAt = Number.isFinite(saved.savedAt) ? saved.savedAt : Date.now();
  const now = Date.now();
  applyDerivedPlayerStats(true);
  applyOfflineProduction(savedAt, now);
  advanceFeatureProgress(now);
  lastSavedAt = now;
  lastUpdateAt = now;
  state.scene = 'world';
  state.cave.panel = 'map';
  state.cave.activeBuildingId = null;
  syncCaveBuildingSelection();
  state.lastMove = '已恢复洞府状态';
}

function saveState() {
  updateProduction(Date.now());
  const now = Date.now();
  const payload = {
    version: SAVE_VERSION,
    savedAt: now,
    player: {
      row: state.player.row,
      col: state.player.col,
      realmIndex: state.player.realmIndex,
      realmId: state.player.realmId,
      realmFamily: state.player.realmFamily,
      realmStage: state.player.realmStage,
      hp: state.player.hp,
      maxHp: state.player.maxHp,
      attributes: state.player.attributes,
      breakthrough: state.player.breakthrough,
      equipment: state.player.equipment,
    },
    realmIndex: state.player.realmIndex,
    qi: state.qi,
    pendingQi: state.pendingQi,
    resources: state.resources,
    dungeonStats: state.dungeonStats,
    inventory: { level: state.inventory.level, scrollY: state.inventory.scrollY, slotted: true, items: state.inventory.items },
    cavePlayer: { row: state.cave.player.row, col: state.cave.player.col },
    buildings: state.cave.buildings,
    features: state.cave.features,
    alchemy: {
      heat: state.cave.alchemy.heat,
      materialScrollX: state.cave.alchemy.materialScrollX,
    },
  };
  try {
    wx.setStorageSync(SAVE_KEY, payload);
    lastSavedAt = now;
  } catch (error) {
    // 本地缓存不可用时继续运行，退出前仍会再次尝试保存。
  }
}

function refreshQiCapacity() {
  const mainHall = state.cave.buildings.mainHall;
  const mainHallLevel = mainHall ? mainHall.level : 0;
  const caveCapacity = [100, 100, 180, 300, 500, 800, 1200][mainHallLevel] || 100;
  const spirit = Math.max(0, Number(state.player.attributes && state.player.attributes.spirit) || 0);
  const attributeCapacity = 100 + (spirit * 10);
  state.qiCap = Math.max(caveCapacity, attributeCapacity);
  state.qi = Math.min(state.qi, state.qiCap);
  state.pendingQi = Math.min(state.pendingQi, state.qiCap);
}

function qiPerMinute() {
  const spiritArray = state.cave.buildings.spiritArray;
  const level = spiritArray ? spiritArray.level : 0;
  return [0, 2, 4, 7, 11, 16, 22][level] || 0;
}

function applyOfflineProduction(fromTime, toTime) {
  const elapsedSeconds = Math.min(8 * 60 * 60, Math.max(0, (toTime - fromTime) / 1000));
  state.pendingQi = Math.min(state.qiCap, state.pendingQi + (qiPerMinute() * elapsedSeconds) / 60);
}

function featureBuildingLevel(id) {
  const building = state.cave.buildings[id];
  return building && building.level > 0 ? Math.min(6, building.level) : 1;
}

function featureDuration(id) {
  const config = FEATURE_CONFIG[id];
  const level = featureBuildingLevel(id);
  const speedFactors = [1, 1, 0.92, 0.84, 0.76, 0.68, 0.58];
  return Math.max(1000, Math.round(config.duration * speedFactors[level]));
}

function featureRewardAmount(id) {
  const level = featureBuildingLevel(id);
  if (id === 'spiritField') return 1 + Math.floor((level - 1) / 2);
  if (id === 'alchemyFurnace') return 1 + Math.floor((level - 1) / 3);
  if (id === 'meditationRoom') {
    const baseReward = 10 + ((level - 1) * 3);
    const hasTechnique = state.resources.techniques.includes('引气诀');
    return hasTechnique ? Math.round(baseReward * 1.25) : baseReward;
  }
  return 1;
}

function advanceFeatureProgress(now) {
  Object.keys(FEATURE_CONFIG).forEach((id) => {
    const feature = state.cave.features[id];
    const config = FEATURE_CONFIG[id];
    if (!feature || feature.status !== config.activeStatus) return;
    const startedAt = Number.isFinite(feature.startedAt) && feature.startedAt > 0 ? feature.startedAt : now;
    const duration = Number.isFinite(feature.duration) && feature.duration > 0 ? feature.duration : config.duration;
    const elapsed = Math.max(0, now - startedAt);
    feature.startedAt = startedAt;
    feature.duration = duration;
    feature.progress = Math.min(1, elapsed / duration);
    feature.remainingMs = Math.max(0, duration - elapsed);
    if (elapsed >= duration) {
      feature.status = 'completed';
      feature.completedAt = startedAt + duration;
      feature.progress = 1;
      feature.remainingMs = 0;
    }
  });
}

function resetFeature(feature, status) {
  feature.status = status;
  feature.startedAt = 0;
  feature.completedAt = 0;
  feature.progress = 0;
  feature.remainingMs = 0;
}

function startFeature(id, feature, now) {
  const config = FEATURE_CONFIG[id];
  feature.status = config.activeStatus;
  feature.startedAt = now;
  feature.completedAt = 0;
  feature.duration = featureDuration(id);
  feature.progress = 0;
  feature.remainingMs = feature.duration;
}

function rebaseFeatureAfterUpgrade(id) {
  const feature = state.cave.features[id];
  const config = FEATURE_CONFIG[id];
  if (!feature || !config || feature.status !== config.activeStatus) return;
  const progress = Math.max(0, Math.min(1, Number.isFinite(feature.progress) ? feature.progress : 0));
  const duration = featureDuration(id);
  const now = Date.now();
  feature.duration = duration;
  feature.startedAt = now - Math.round(duration * progress);
  feature.remainingMs = Math.max(0, duration - Math.round(duration * progress));
  advanceFeatureProgress(now);
}

function claimFeatureReward(id, feature) {
  const config = FEATURE_CONFIG[id];
  if (id === 'scripturePavilion') {
    if (!state.resources.techniques.includes(config.rewardName)) state.resources.techniques.push(config.rewardName);
    const techniqueId = techniqueItemId(config.rewardName);
    if (techniqueId) addItem(techniqueId, 1);
    resetFeature(feature, 'learned');
    state.lastMove = `获得功法：${config.rewardName}`;
    return;
  }
  const rewardAmount = featureRewardAmount(id);
  if (id === 'meditationRoom') {
    addCultivation(rewardAmount, 'meditation');
  } else {
    if (config.rewardKey === 'spiritHerbs') addItem(LEGACY_RESOURCE_ITEM_IDS.spiritHerbs, rewardAmount);
    else if (config.rewardKey === 'qiPills') addItem(LEGACY_RESOURCE_ITEM_IDS.qiPills, rewardAmount);
    else state.resources[config.rewardKey] += rewardAmount;
    refreshRealm(false);
  }
  const amountText = id === 'meditationRoom' ? `${rewardAmount}点` : `${rewardAmount}份`;
  resetFeature(feature, id === 'spiritField' ? 'empty' : 'idle');
  state.lastMove = `领取${config.rewardName}${amountText}`;
  if (id === 'meditationRoom' && state.player.canBreakthrough) {
    state.lastMove += '，修为已足，可尝试突破';
  }
}

function locationAt(row, col) {
  return ui.locations[row] && ui.locations[row][col] ? ui.locations[row][col] : '';
}

function dungeonAt(row, col) {
  const name = locationAt(row, col);
  return DUNGEONS.find((dungeon) => dungeon.name === name) || null;
}

function movePlayer(direction) {
  const next = { ...state.player };
  if (direction === 'up') next.row -= 1;
  if (direction === 'down') next.row += 1;
  if (direction === 'left') next.col -= 1;
  if (direction === 'right') next.col += 1;

  if (next.row < 0 || next.row >= ui.rows || next.col < 0 || next.col >= ui.cols) {
    state.lastMove = '前方没有路了';
    return;
  }

  state.player = next;
  const location = locationAt(next.row, next.col);
  if (location === '洞府') {
    enterCave();
    return;
  }
  const dungeon = dungeonAt(next.row, next.col);
  if (dungeon) {
    state.lastMove = `来到${dungeon.name}`;
    saveState();
    return;
  }
  state.lastMove = location ? `来到${location}` : '来到空地';
  saveState();
}

function enterCave() {
  if (locationAt(state.player.row, state.player.col) !== '洞府') {
    state.lastMove = '需要先到达洞府';
    return;
  }
  state.scene = 'cave';
  state.cave.panel = 'map';
  state.cave.activeBuildingId = null;
  state.cave.player = { row: ui.caveEntrance.row, col: ui.caveEntrance.col };
  state.cave.selectedBuildingId = null;
  state.lastMove = '进入洞府';
  saveState();
}

function enterDungeon(name) {
  const dungeon = DUNGEONS.find((item) => item.name === name);
  if (!dungeon) {
    state.lastMove = '这里没有可探索的秘境';
    return;
  }
  state.scene = 'dungeon';
  state.dungeon.name = dungeon.name;
  state.dungeon.result = null;
  state.lastMove = `抵达${dungeon.name}`;
  saveState();
}

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function exploreDungeon() {
  const dungeon = DUNGEONS.find((item) => item.name === state.dungeon.name);
  if (!dungeon) return;
  if ((state.player.realmIndex || 0) < dungeon.minRealm) {
    const requiredRealm = REALMS[dungeon.minRealm] ? realmDisplayName(REALMS[dungeon.minRealm]) : '更高境界';
    state.dungeon.result = { type: 'blocked', text: `境界不足，需要达到${requiredRealm}` };
    state.lastMove = `无法进入${dungeon.name}`;
    return;
  }
  if (state.qi < dungeon.qiCost) {
    state.dungeon.result = { type: 'blocked', text: `灵气不足，需要${dungeon.qiCost}点灵气` };
    state.lastMove = '灵气不足，暂时无法探索';
    return;
  }

  state.qi -= dungeon.qiCost;
  const cultivation = randomInt(dungeon.cultivation[0], dungeon.cultivation[1]);
  const spiritHerbs = randomInt(dungeon.spiritHerbs[0], dungeon.spiritHerbs[1]);
  const foundPill = Math.random() < dungeon.qiPillChance;
  addCultivation(cultivation, 'dungeon');
  addItem(LEGACY_RESOURCE_ITEM_IDS.spiritHerbs, spiritHerbs);
  if (foundPill) addItem(LEGACY_RESOURCE_ITEM_IDS.qiPills, 1);
  const record = state.dungeonStats[dungeon.name] || { count: 0, lastAt: 0, bestCultivation: 0, bestHerbs: 0 };
  record.count += 1;
  record.lastAt = Date.now();
  record.bestCultivation = Math.max(record.bestCultivation, cultivation);
  record.bestHerbs = Math.max(record.bestHerbs, spiritHerbs);
  state.dungeonStats[dungeon.name] = record;
  const rewardParts = [`修为 +${cultivation}`, `灵草 +${spiritHerbs}`];
  if (foundPill) rewardParts.push('回气丹 +1');
  const resultText = `探索完成：${rewardParts.join('，')}`;
  state.dungeon.result = {
    type: 'success',
    text: resultText,
    cultivation,
    spiritHerbs,
    qiPill: foundPill ? 1 : 0,
    breakthrough: '',
  };
  state.lastMove = state.player.canBreakthrough
    ? `${resultText}，修为已足，可尝试突破`
    : resultText;
  saveState();
}

function leaveDungeon() {
  const name = state.dungeon.name;
  state.scene = 'world';
  state.dungeon.name = null;
  state.dungeon.result = null;
  state.lastMove = name ? `离开${name}` : '回到野外';
  saveState();
}

function moveCavePlayer(direction) {
  const next = { ...state.cave.player };
  if (direction === 'up') next.row -= 1;
  if (direction === 'down') next.row += 1;
  if (direction === 'left') next.col -= 1;
  if (direction === 'right') next.col += 1;

  if (next.row < 0 || next.row >= ui.rows || next.col < 0 || next.col >= ui.cols) {
    state.lastMove = '洞府边缘没有路了';
    return;
  }

  const wasAtEntrance = state.cave.player.row === ui.caveEntrance.row && state.cave.player.col === ui.caveEntrance.col;
  const isAtEntrance = next.row === ui.caveEntrance.row && next.col === ui.caveEntrance.col;
  state.cave.player = next;
  if (!wasAtEntrance && isAtEntrance) {
    state.scene = 'world';
    state.cave.panel = 'map';
    state.cave.activeBuildingId = null;
    state.cave.selectedBuildingId = null;
    state.lastMove = '离开洞府，回到野外';
    saveState();
    return;
  }
  const building = syncCaveBuildingSelection();
  state.lastMove = isAtEntrance
    ? '回到洞府入口'
    : building
      ? `进入${building.name}`
      : '来到洞府空地';
  saveState();
}

function repairSelectedBuilding() {
  const building = CAVE_BUILDINGS.find((item) => item.id === state.cave.selectedBuildingId);
  if (!building) return;
  const buildingState = state.cave.buildings[building.id];
  if (!buildingState || buildingState.level > 0) {
    state.lastMove = `${building.name}已经修缮`;
    return;
  }
  buildingState.level = 1;
  buildingState.status = 'repaired';
  state.lastMove = `${building.name}已经修缮完成`;
  saveState();
}

function upgradeSelectedBuilding() {
  const building = CAVE_BUILDINGS.find((item) => item.id === state.cave.selectedBuildingId);
  if (!building) return;
  const buildingState = state.cave.buildings[building.id];
  if (!buildingState || buildingState.level < 1) {
    state.lastMove = `请先修缮${building.name}`;
    return;
  }
  const maxLevel = building.maxLevel || 6;
  if (buildingState.level >= maxLevel) {
    state.lastMove = `${building.name}已达到最高等级`;
    return;
  }
  buildingState.level += 1;
  rebaseFeatureAfterUpgrade(building.id);
  state.lastMove = `${building.name}已升级至Lv.${buildingState.level}`;
  saveState();
}

function updateProduction(now) {
  const deltaSeconds = Math.min(1, Math.max(0, (now - lastUpdateAt) / 1000));
  lastUpdateAt = now;
  refreshQiCapacity();
  advanceFeatureProgress(now);
  const production = qiPerMinute();
  if (production > 0) {
    state.pendingQi = Math.min(state.qiCap, state.pendingQi + (production * deltaSeconds) / 60);
  }
}

function collectQi() {
  const available = Math.floor(state.pendingQi);
  if (available <= 0) {
    state.lastMove = '聚灵阵暂时没有灵气';
    return;
  }
  const room = Math.max(0, state.qiCap - state.qi);
  if (room <= 0) {
    state.lastMove = '灵气储存已满';
    return;
  }
  const collected = Math.min(room, available);
  state.qi += collected;
  state.pendingQi -= collected;
  state.lastMove = `收取${collected}点灵气`;
  saveState();
}

function useQiPill() {
  const available = state.resources.qiPills;
  if (available <= 0) {
    state.lastMove = '没有可用的回气丹';
    return;
  }
  const room = Math.max(0, state.qiCap - state.qi);
  if (room <= 0) {
    state.lastMove = '当前灵气已满';
    return;
  }
  const restored = Math.min(30, room);
  removeItem(LEGACY_RESOURCE_ITEM_IDS.qiPills, 1);
  state.qi += restored;
  state.lastMove = `服用回气丹，恢复${restored}点灵气`;
  saveState();
}

function toggleInventory() {
  state.inventory.open = !state.inventory.open;
  state.attributesOpen = false;
  state.itemDetailOpen = false;
  state.inventory.selectedItemId = null;
  state.inventory.scrollY = 0;
  state.lastMove = state.inventory.open ? '打开储物袋' : '关闭储物袋';
  saveState();
}

function closeInventory() {
  if (!state.inventory.open) return;
  state.inventory.open = false;
  state.attributesOpen = false;
  state.itemDetailOpen = false;
  state.inventory.selectedItemId = null;
  state.inventory.scrollY = 0;
  state.lastMove = '关闭储物袋';
  saveState();
}

function openAttributes() {
  if (!state.inventory.open) return;
  state.attributesOpen = true;
  state.lastMove = '查看详细属性';
}

function openItemDetails(itemId) {
  if (!state.inventory.open || !itemId) return;
  state.inventory.selectedItemId = itemId;
  state.itemDetailOpen = true;
  state.attributesOpen = false;
  state.lastMove = '查看物品详情';
}

function closeItemDetails() {
  state.itemDetailOpen = false;
  state.inventory.selectedItemId = null;
  state.lastMove = '返回储物袋';
}

function selectedItemId() {
  return state.inventory && typeof state.inventory.selectedItemId === 'string'
    ? state.inventory.selectedItemId
    : null;
}

function equippedSlotForItem(itemId) {
  if (!itemId || !state.player || !state.player.equipment) return null;
  return EQUIPMENT_SLOTS.find((slot) => resolveItemId(state.player.equipment[slot]) === itemId) || null;
}

function useSelectedItem() {
  const itemId = selectedItemId();
  const item = itemId ? itemDefinition(itemId) : null;
  if (!item || !item.usable) {
    state.lastMove = '该物品暂不可使用';
    return;
  }
  if (stackCount(state.inventory.items, itemId) <= 0) {
    state.lastMove = '物品数量不足';
    closeItemDetails();
    return;
  }
  if (itemId === LEGACY_RESOURCE_ITEM_IDS.qiPills) {
    const before = state.qi;
    useQiPill();
    if (state.qi > before && stackCount(state.inventory.items, itemId) <= 0) {
      state.itemDetailOpen = false;
      state.inventory.selectedItemId = null;
    }
    return;
  }
  state.lastMove = `${item.name}暂未配置使用效果`;
}

function equipSelectedItem() {
  const itemId = selectedItemId();
  const item = itemId ? itemDefinition(itemId) : null;
  const slot = item && EQUIPMENT_SLOT_BY_CATEGORY[item.category];
  if (!item || !slot) {
    state.lastMove = '该物品不可装备';
    return;
  }
  const currentSlot = equippedSlotForItem(itemId);
  if (currentSlot === slot) {
    unequipSelectedItem();
    return;
  }
  if (stackCount(state.inventory.items, itemId) <= 0) {
    state.lastMove = '背包中没有该物品';
    return;
  }
  const previousItemId = resolveItemId(state.player.equipment[slot]);
  if (previousItemId) addItem(previousItemId, 1);
  removeItem(itemId, 1);
  state.player.equipment[slot] = itemId;
  state.lastMove = `已装备${item.name}`;
  saveState();
}

function unequipSelectedItem() {
  const itemId = selectedItemId();
  const slot = equippedSlotForItem(itemId);
  if (!itemId || !slot) {
    state.lastMove = '该物品当前未装备';
    return;
  }
  const item = itemDefinition(itemId);
  state.player.equipment[slot] = null;
  addItem(itemId, 1);
  state.lastMove = `已卸下${item.name}`;
  saveState();
}

function runItemDetailAction() {
  const itemId = selectedItemId();
  const item = itemId ? itemDefinition(itemId) : null;
  if (!item) return;
  if (item.usable) useSelectedItem();
  else if (EQUIPMENT_SLOT_BY_CATEGORY[item.category]) {
    if (equippedSlotForItem(itemId)) unequipSelectedItem();
    else equipSelectedItem();
  }
}

function closeAttributes() {
  state.attributesOpen = false;
  state.lastMove = '返回储物袋';
}

function upgradeInventory() {
  const maxLevel = 5;
  if (state.inventory.level >= maxLevel) {
    state.lastMove = '储物袋已达到最高等级';
    return;
  }
  state.inventory.level += 1;
  state.inventory.scrollY = 0;
  state.lastMove = `储物袋扩充至Lv.${state.inventory.level}`;
  saveState();
}

function scrollInventory(deltaY) {
  const maxScroll = ui.getInventoryMaxScroll(state.inventory.level);
  state.inventory.scrollY = Math.max(0, Math.min(maxScroll, state.inventory.scrollY + deltaY));
}

function inventoryItemAt(index) {
  return Number.isInteger(index) && Array.isArray(state.inventory.items)
    ? state.inventory.items[index]
    : null;
}

function itemSlotForCategory(itemId) {
  const item = itemId ? itemDefinition(itemId) : null;
  return item ? EQUIPMENT_SLOT_BY_CATEGORY[item.category] || null : null;
}

function canPlaceInEquipment(itemId, slot) {
  return Boolean(itemId && slot && itemSlotForCategory(itemId) === slot);
}

function swapInventorySlots(sourceIndex, targetIndex) {
  if (!Number.isInteger(sourceIndex) || !Number.isInteger(targetIndex) || sourceIndex === targetIndex) return false;
  const source = inventoryItemAt(sourceIndex);
  const target = inventoryItemAt(targetIndex);
  if (!source) return false;
  state.inventory.items[sourceIndex] = target || null;
  state.inventory.items[targetIndex] = source;
  return true;
}

function finishInventoryDrag(source, x, y) {
  if (!source) return false;
  const targetInventoryIndex = ui.getInventorySlotAtPoint(x, y, state);
  const targetEquipmentSlot = ui.getEquipmentSlotAtPoint(x, y, state);
  const sourceItem = source.type === 'inventory'
    ? inventoryItemAt(source.index)
    : state.player.equipment && state.player.equipment[source.slot];
  const sourceItemId = resolveItemId(sourceItem);
  if (!sourceItemId) return false;

  if (source.type === 'inventory' && Number.isInteger(targetInventoryIndex)) {
    const moved = swapInventorySlots(source.index, targetInventoryIndex);
    if (moved) {
      syncItemsToLegacyResources();
      state.lastMove = '已调整物品位置';
    }
    return moved;
  }

  if (source.type === 'inventory' && targetEquipmentSlot) {
    if (!canPlaceInEquipment(sourceItemId, targetEquipmentSlot)) {
      state.lastMove = '该物品不能放入此装备栏';
      return false;
    }
    const previous = resolveItemId(state.player.equipment[targetEquipmentSlot]);
    state.player.equipment[targetEquipmentSlot] = sourceItemId;
    state.inventory.items[source.index] = previous ? { id: previous, count: 1 } : null;
    syncItemsToLegacyResources();
    state.lastMove = previous ? '已交换装备' : '已装备物品';
    return true;
  }

  if (source.type === 'equipment' && Number.isInteger(targetInventoryIndex)) {
    const target = inventoryItemAt(targetInventoryIndex);
    const targetId = resolveItemId(target);
    if (targetId && !canPlaceInEquipment(targetId, source.slot)) {
      state.lastMove = '目标物品不能放入该装备栏';
      return false;
    }
    state.player.equipment[source.slot] = targetId || null;
    state.inventory.items[targetInventoryIndex] = { id: sourceItemId, count: 1 };
    syncItemsToLegacyResources();
    state.lastMove = targetId ? '已交换装备' : '已卸下物品';
    return true;
  }

  if (source.type === 'equipment' && targetEquipmentSlot && targetEquipmentSlot !== source.slot) {
    const targetItemId = resolveItemId(state.player.equipment[targetEquipmentSlot]);
    if (!canPlaceInEquipment(sourceItemId, targetEquipmentSlot)
      || (targetItemId && !canPlaceInEquipment(targetItemId, source.slot))) {
      state.lastMove = '装备栏类型不匹配';
      return false;
    }
    state.player.equipment[source.slot] = targetItemId || null;
    state.player.equipment[targetEquipmentSlot] = sourceItemId;
    state.lastMove = '已交换装备位置';
    return true;
  }
  return false;
}

function openBuildingFeature() {
  const building = caveBuildingAt(state.cave.player.row, state.cave.player.col);
  if (!building) return;
  state.cave.panel = building.id;
  state.cave.activeBuildingId = building.id;
  state.lastMove = `查看${building.name}`;
  saveState();
}

function closeBuildingFeature() {
  state.cave.panel = 'map';
  state.cave.activeBuildingId = null;
  state.cave.alchemy.upgradeOpen = false;
  state.lastMove = '回到洞府';
  saveState();
}

function openAlchemyUpgrade() {
  if (!state.cave.alchemy) state.cave.alchemy = createAlchemyState();
  state.cave.alchemy.upgradeOpen = true;
  state.lastMove = '查看丹炉升级';
}

function closeAlchemyUpgrade() {
  if (!state.cave.alchemy) return;
  state.cave.alchemy.upgradeOpen = false;
  state.lastMove = '返回丹炉';
}

function upgradeActiveBuilding() {
  const building = CAVE_BUILDINGS.find((item) => item.id === state.cave.activeBuildingId);
  if (!building) return;
  const buildingState = state.cave.buildings[building.id];
  if (building.id === 'alchemyFurnace') {
    if (!buildingState || buildingState.level < 1) {
      state.cave.selectedBuildingId = building.id;
      repairSelectedBuilding();
      return;
    }
    if (buildingState.level >= (building.maxLevel || 6)) {
      state.lastMove = `${building.name}已达到最高等级`;
      return;
    }
    openAlchemyUpgrade();
    return;
  }
  if (!buildingState || buildingState.level < 1) {
    state.cave.selectedBuildingId = building.id;
    repairSelectedBuilding();
    return;
  }
  state.cave.selectedBuildingId = building.id;
  upgradeSelectedBuilding();
}

function runFeatureAction() {
  const id = state.cave.activeBuildingId;
  const feature = state.cave.features[id];
  const building = CAVE_BUILDINGS.find((item) => item.id === id);
  const buildingState = building && state.cave.buildings[building.id];
  if (id === 'spiritArray') {
    if (!buildingState || buildingState.level < 1) {
      state.lastMove = `请先修缮${building ? building.name : '聚灵阵'}`;
      return;
    }
    collectQi();
    return;
  }
  if (!feature) {
    state.lastMove = building ? `${building.name}暂无可执行功能` : '暂无可执行功能';
    return;
  }
  if (!buildingState || buildingState.level < 1) {
    state.lastMove = `请先修缮${building.name}`;
    return;
  }
  const now = Date.now();
  advanceFeatureProgress(now);
  if (id === 'spiritField') {
    if (feature.status === 'empty') {
      startFeature(id, feature, now);
      state.lastMove = '灵田开始种植灵草';
      saveState();
    } else if (feature.status === 'growing') {
      state.lastMove = '灵草正在生长';
    } else if (feature.status === 'completed') {
      claimFeatureReward(id, feature);
      saveState();
    }
    return;
  }
  if (id === 'alchemyFurnace') {
    if (feature.status === 'idle') {
      startFeature(id, feature, now);
      state.lastMove = '丹炉开始炼制';
      saveState();
    } else if (feature.status === 'brewing') {
      state.lastMove = '丹炉正在炼制';
    } else if (feature.status === 'completed') {
      claimFeatureReward(id, feature);
      saveState();
    }
    return;
  }
  if (id === 'meditationRoom') {
    if (feature.status === 'idle') {
      startFeature(id, feature, now);
      state.lastMove = '林开始在静室修炼';
      saveState();
    } else if (feature.status === 'cultivating') {
      state.lastMove = '正在静室修炼';
    } else if (feature.status === 'completed') {
      claimFeatureReward(id, feature);
      saveState();
    }
    return;
  }
  if (id === 'scripturePavilion') {
    if (feature.status === 'idle') {
      startFeature(id, feature, now);
      state.lastMove = '藏经阁开始研究引气诀';
      saveState();
    } else if (feature.status === 'researching') {
      state.lastMove = '正在研究引气诀';
    } else if (feature.status === 'completed') {
      claimFeatureReward(id, feature);
      saveState();
    } else if (feature.status === 'learned') {
      state.lastMove = '引气诀已经记录';
    }
  }
}

function setAlchemyHeatFromPoint(x) {
  if (!state.cave.alchemy || !ui.alchemyHeatBar) return;
  const bar = ui.alchemyHeatBar;
  state.cave.alchemy.heat = Math.max(0, Math.min(1, (x - bar.x) / bar.width));
  state.lastMove = '调整火候';
}

function scrollAlchemyMaterials(deltaX) {
  if (!state.cave.alchemy) return;
  const maxScroll = ui.getAlchemyMaterialMaxScroll();
  state.cave.alchemy.materialScrollX = Math.max(0, Math.min(maxScroll, state.cave.alchemy.materialScrollX + deltaX));
}

function bindEvents() {
  let inventoryTouchStartY = null;
  let inventoryTouchStartX = null;
  let inventoryDidScroll = false;
  let inventoryTouchInGrid = false;
  let inventoryTouchItemId = null;
  let inventoryTouchSource = null;
  let inventoryTouchCurrentX = null;
  let inventoryTouchCurrentY = null;
  let inventoryLongPressTimer = null;
  let alchemyTouchStartX = null;
  let alchemyTouchInMaterialBar = false;

  const clearInventoryLongPress = () => {
    if (inventoryLongPressTimer != null) clearTimeout(inventoryLongPressTimer);
    inventoryLongPressTimer = null;
  };

  const resetInventoryTouch = () => {
    clearInventoryLongPress();
    inventoryTouchStartY = null;
    inventoryTouchStartX = null;
    inventoryTouchCurrentX = null;
    inventoryTouchCurrentY = null;
    inventoryDidScroll = false;
    inventoryTouchInGrid = false;
    inventoryTouchItemId = null;
    inventoryTouchSource = null;
  };

  wx.onTouchStart((event) => {
    const touch = event.touches && event.touches[0];
    if (!touch) return;
    const touchX = touch.x == null ? (touch.clientX == null ? touch.pageX : touch.clientX) : touch.x;
    const touchY = touch.y == null ? (touch.clientY == null ? touch.pageY : touch.clientY) : touch.y;

    if (state.inventory.open) {
      inventoryTouchStartY = touchY;
      inventoryTouchStartX = touchX;
      inventoryTouchCurrentX = touchX;
      inventoryTouchCurrentY = touchY;
      inventoryDidScroll = false;
      inventoryTouchInGrid = ui.isInventoryGridHit(touchX, touchY);
      inventoryTouchItemId = null;
      inventoryTouchSource = null;
      if (state.itemDetailOpen) {
        if (ui.isItemDetailCloseHit(touchX, touchY)) closeItemDetails();
        else if (ui.isItemDetailActionHit(touchX, touchY)) runItemDetailAction();
        inventoryTouchStartY = null;
        return;
      }
      if (state.attributesOpen) {
        if (ui.isAttributeOverlayCloseHit(touchX, touchY)) {
          closeAttributes();
          inventoryTouchStartY = null;
          return;
        }
        if (ui.isBreakthroughHit(touchX, touchY)) {
          attemptBreakthrough();
          inventoryTouchStartY = null;
        }
        return;
      }
      if (ui.isInventoryCloseHit(touchX, touchY)) {
        closeInventory();
        inventoryTouchStartY = null;
        return;
      }
      if (ui.isInventoryUpgradeHit(touchX, touchY)) {
        upgradeInventory();
        inventoryTouchStartY = null;
        return;
      }
      if (ui.isAttributeDetailsHit(touchX, touchY)) {
        openAttributes();
        inventoryTouchStartY = null;
        return;
      }
      const equipmentSlot = ui.getEquipmentSlotAtPoint(touchX, touchY, state);
      const equipmentItemId = equipmentSlot
        ? resolveItemId(state.player.equipment && state.player.equipment[equipmentSlot])
        : null;
      const inventorySlot = ui.getInventorySlotAtPoint(touchX, touchY, state);
      const inventoryItem = Number.isInteger(inventorySlot) ? inventoryItemAt(inventorySlot) : null;
      if (equipmentItemId) {
        inventoryTouchSource = { type: 'equipment', slot: equipmentSlot, itemId: equipmentItemId };
        inventoryTouchItemId = equipmentItemId;
      } else if (inventoryItem) {
        inventoryTouchSource = { type: 'inventory', index: inventorySlot, itemId: inventoryItem.id };
        inventoryTouchItemId = inventoryItem.id;
      }
      if (inventoryTouchSource) {
        inventoryLongPressTimer = setTimeout(() => {
          if (!state.inventory.open || inventoryDidScroll || !inventoryTouchSource) return;
          state.inventory.drag = {
            source: { ...inventoryTouchSource },
            x: inventoryTouchCurrentX == null ? inventoryTouchStartX : inventoryTouchCurrentX,
            y: inventoryTouchCurrentY == null ? inventoryTouchStartY : inventoryTouchCurrentY,
          };
          inventoryTouchItemId = null;
          inventoryDidScroll = true;
          state.lastMove = '拖动调整物品位置';
        }, 380);
      }
      return;
    }

    if (state.scene === 'world') {
      if (ui.isInventoryButtonHit(touchX, touchY)) {
        toggleInventory();
        return;
      }
      const direction = ui.getControlDirection(touchX, touchY);
      if (direction) {
        movePlayer(direction);
        return;
      }
      if (ui.isWorldLocationHit(touchX, touchY, state.player.row, state.player.col, '洞府')) {
        enterCave();
        return;
      }
      const currentDungeon = dungeonAt(state.player.row, state.player.col);
      if (currentDungeon && ui.isWorldDungeonEnterHit(touchX, touchY)) {
        enterDungeon(currentDungeon.name);
      }
      return;
    }

    if (state.scene === 'dungeon') {
      if (ui.isDungeonBackHit(touchX, touchY)) {
        leaveDungeon();
        return;
      }
      if (ui.isDungeonActionHit(touchX, touchY)) {
        exploreDungeon();
        return;
      }
      return;
    }

    if (ui.isInventoryButtonHit(touchX, touchY)) {
      toggleInventory();
      return;
    }

    if (state.cave.panel !== 'map') {
      if (state.cave.activeBuildingId === 'alchemyFurnace') {
        if (state.cave.alchemy && state.cave.alchemy.upgradeOpen) {
          if (ui.isAlchemyUpgradeCloseHit(touchX, touchY) || ui.isAlchemyUpgradeCancelHit(touchX, touchY)) closeAlchemyUpgrade();
          else if (ui.isAlchemyUpgradeConfirmHit(touchX, touchY)) {
            state.cave.alchemy.upgradeOpen = false;
            state.cave.selectedBuildingId = 'alchemyFurnace';
            upgradeSelectedBuilding();
          }
          return;
        }
        alchemyTouchStartX = touchX;
        alchemyTouchInMaterialBar = ui.isAlchemyMaterialBarHit(touchX, touchY);
        if (ui.isCavePanelBackHit(touchX, touchY)) {
          closeBuildingFeature();
          return;
        }
        if (ui.isFeatureUpgradeHit(touchX, touchY)) {
          upgradeActiveBuilding();
          return;
        }
        if (ui.isFeatureActionHit(touchX, touchY)) {
          runFeatureAction();
          return;
        }
        if (ui.isAlchemyHeatHit(touchX, touchY)) {
          setAlchemyHeatFromPoint(touchX);
          saveState();
          return;
        }
        return;
      }
      if (ui.isCavePanelBackHit(touchX, touchY)) {
        closeBuildingFeature();
        return;
      }
      if (ui.isFeatureActionHit(touchX, touchY)) {
        runFeatureAction();
        return;
      }
      if (ui.isFeatureUpgradeHit(touchX, touchY)) {
        upgradeActiveBuilding();
      }
      return;
    }

    const caveDirection = ui.getCaveControlDirection(touchX, touchY);
    if (caveDirection) {
      moveCavePlayer(caveDirection);
      return;
    }

    if (ui.isCaveCollectHit(touchX, touchY)) {
      collectQi();
      return;
    }

    const building = caveBuildingAt(state.cave.player.row, state.cave.player.col);
    if (building && ui.isCaveFeatureHit(touchX, touchY)) {
      openBuildingFeature();
    }
  });
  if (wx.onTouchMove) {
    wx.onTouchMove((event) => {
      if (state.scene === 'cave' && state.cave.panel !== 'map' && state.cave.activeBuildingId === 'alchemyFurnace') {
        if (state.cave.alchemy && state.cave.alchemy.upgradeOpen) return;
        const touch = event.touches && event.touches[0];
        if (!touch) return;
        const touchX = touch.x == null ? (touch.clientX == null ? touch.pageX : touch.clientX) : touch.x;
        if (alchemyTouchStartX == null) return;
        if (alchemyTouchInMaterialBar) {
          const deltaX = alchemyTouchStartX - touchX;
          if (Math.abs(deltaX) >= 1) {
            scrollAlchemyMaterials(deltaX);
            alchemyTouchStartX = touchX;
          }
          return;
        }
        if (ui.isAlchemyHeatHit(touchX, touch.y == null ? (touch.clientY == null ? touch.pageY : touch.clientY) : touch.y)) {
          setAlchemyHeatFromPoint(touchX);
        }
        return;
      }
      if (!state.inventory.open || inventoryTouchStartY == null) return;
      if (state.attributesOpen) return;
      const touch = event.touches && event.touches[0];
      if (!touch) return;
      const touchX = touch.x == null ? (touch.clientX == null ? touch.pageX : touch.clientX) : touch.x;
      const touchY = touch.y == null ? (touch.clientY == null ? touch.pageY : touch.clientY) : touch.y;
      inventoryTouchCurrentX = touchX;
      inventoryTouchCurrentY = touchY;
      if (state.inventory.drag) {
        state.inventory.drag.x = touchX;
        state.inventory.drag.y = touchY;
        return;
      }
      const movedX = Math.abs(touchX - inventoryTouchStartX);
      const movedY = Math.abs(touchY - inventoryTouchStartY);
      if (movedX >= 3 || movedY >= 3) clearInventoryLongPress();
      if (!inventoryTouchInGrid) return;
      const deltaY = inventoryTouchStartY - touchY;
      if (Math.abs(deltaY) < 2) return;
      clearInventoryLongPress();
      inventoryTouchItemId = null;
      scrollInventory(deltaY);
      inventoryTouchStartY = touchY;
      inventoryDidScroll = true;
    });
  }
  if (wx.onTouchEnd) {
    wx.onTouchEnd((event) => {
      const touch = event && event.changedTouches && event.changedTouches[0];
      if (touch) {
        inventoryTouchCurrentX = touch.x == null ? (touch.clientX == null ? touch.pageX : touch.clientX) : touch.x;
        inventoryTouchCurrentY = touch.y == null ? (touch.clientY == null ? touch.pageY : touch.clientY) : touch.y;
      }
      if (state.inventory.open && state.inventory.drag) {
        const drag = state.inventory.drag;
        const moved = finishInventoryDrag(drag.source, inventoryTouchCurrentX, inventoryTouchCurrentY);
        state.inventory.drag = null;
        if (moved) saveState();
        resetInventoryTouch();
        return;
      }
      if (state.inventory.open && !state.itemDetailOpen && !state.attributesOpen && !inventoryDidScroll && inventoryTouchItemId) {
        openItemDetails(inventoryTouchItemId);
      }
      resetInventoryTouch();
      alchemyTouchStartX = null;
      alchemyTouchInMaterialBar = false;
    });
  }
}

function render() {
  const now = Date.now();
  updateProduction(now);
  if (now - lastPersistAt >= 5000) {
    saveState();
    lastPersistAt = now;
  }
  ui.resize(SCREEN_WIDTH, SCREEN_HEIGHT);
  ui.renderBackground(ctx);
  if (state.scene === 'world') {
    ui.renderHeader(ctx, state);
    ui.renderMap(ctx, state);
    ui.renderControlPad(ctx, state, dungeonAt(state.player.row, state.player.col));
  } else if (state.scene === 'dungeon') {
    const dungeon = DUNGEONS.find((item) => item.name === state.dungeon.name);
    ui.renderDungeonHeader(ctx, state);
    ui.renderDungeonPanel(ctx, state, dungeon);
  } else {
    ui.renderCaveHeader(ctx, state);
    if (state.cave.panel === 'map') {
      ui.renderCaveMap(ctx, state, CAVE_BUILDINGS);
      ui.renderCaveFooter(ctx, state, CAVE_BUILDINGS);
    } else {
      const activeBuilding = CAVE_BUILDINGS.find((building) => building.id === state.cave.activeBuildingId);
      ui.renderFeaturePanel(ctx, state, activeBuilding);
      if (activeBuilding && activeBuilding.id === 'alchemyFurnace' && state.cave.alchemy && state.cave.alchemy.upgradeOpen) {
        ui.renderAlchemyUpgradeOverlay(ctx, state, activeBuilding);
      }
    }
  }
  if (state.inventory.open) {
    ui.renderInventoryOverlay(ctx, state);
    if (state.attributesOpen) ui.renderAttributeOverlay(ctx, state);
    if (state.itemDetailOpen) ui.renderItemDetailOverlay(ctx, state);
  }
}

function loop() {
  render();
  requestAnimationFrame(loop);
}

bindEvents();
refreshRealm();
loadState();
applyDerivedPlayerStats(true);
if (wx.onHide) wx.onHide(saveState);
if (wx.onShow) wx.onShow(() => {
  const now = Date.now();
  applyOfflineProduction(lastSavedAt, now);
  advanceFeatureProgress(now);
  lastSavedAt = now;
  lastUpdateAt = now;
  state.lastMove = '已结算离线灵气';
  saveState();
});
requestAnimationFrame(loop);
