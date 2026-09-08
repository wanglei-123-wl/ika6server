const QUALITY_DEFINITIONS = Object.freeze({
  mortal: { id: 'mortal', name: '凡', color: '#edf1f3' },
  yellow: { id: 'yellow', name: '黄', color: '#70d58c' },
  mystic: { id: 'mystic', name: '玄', color: '#b68cff' },
  earth: { id: 'earth', name: '地', color: '#f0c85a' },
  heaven: { id: 'heaven', name: '天', color: '#ef6666' },
  holy: { id: 'holy', name: '圣', color: '#ff9bd8', rainbow: true },
});

const ITEM_CATEGORY_DEFINITIONS = Object.freeze({
  weapon: { id: 'weapon', name: '武器' },
  artifact: { id: 'artifact', name: '法宝' },
  armor: { id: 'armor', name: '衣甲' },
  storage: { id: 'storage', name: '储物' },
  technique: { id: 'technique', name: '功法' },
  formula: { id: 'formula', name: '丹方' },
  spell: { id: 'spell', name: '法术' },
  material: { id: 'material', name: '材料' },
  herb: { id: 'herb', name: '药草' },
  pill: { id: 'pill', name: '丹药' },
  talisman: { id: 'talisman', name: '灵符' },
});

const ITEM_CATALOG = Object.freeze({
  spirit_herb: {
    id: 'spirit_herb',
    name: '灵草',
    symbol: '草',
    category: 'herb',
    kind: 'consumable',
    usable: false,
    quality: 'mortal',
    maxStack: 99,
    description: '炼丹与修行中常用的基础药草。',
  },
  qi_pill: {
    id: 'qi_pill',
    name: '回气丹',
    symbol: '丹',
    category: 'pill',
    kind: 'consumable',
    usable: true,
    quality: 'yellow',
    maxStack: 99,
    description: '服用后恢复灵气，具体效果由丹药规则决定。',
  },
  qi_gathering_manual: {
    id: 'qi_gathering_manual',
    name: '引气诀',
    symbol: '诀',
    category: 'technique',
    kind: 'nonConsumable',
    usable: false,
    quality: 'mortal',
    maxStack: 1,
    description: '基础引气功法，记录后可提升静室修炼收益。',
  },
});

const LEGACY_RESOURCE_ITEM_IDS = Object.freeze({
  spiritHerbs: 'spirit_herb',
  qiPills: 'qi_pill',
});

const TECHNIQUE_ITEM_IDS = Object.freeze({
  引气诀: 'qi_gathering_manual',
});

function itemDefinition(itemId) {
  return ITEM_CATALOG[itemId] || {
    id: itemId,
    name: '未知物品',
    symbol: '?',
    category: 'material',
    kind: 'nonConsumable',
    usable: false,
    quality: 'mortal',
    maxStack: 1,
    description: '该物品尚未录入图鉴。',
  };
}

function categoryDefinition(categoryId) {
  return ITEM_CATEGORY_DEFINITIONS[categoryId] || ITEM_CATEGORY_DEFINITIONS.material;
}

function qualityDefinition(qualityId) {
  return QUALITY_DEFINITIONS[qualityId] || QUALITY_DEFINITIONS.mortal;
}

function itemView(itemId, stack = null) {
  const item = itemDefinition(itemId);
  const category = categoryDefinition(item.category);
  const quality = qualityDefinition(item.quality);
  return {
    ...item,
    categoryName: category.name,
    kindName: item.kind === 'consumable' ? '消耗品' : '非消耗品',
    usable: Boolean(item.usable),
    qualityName: quality.name,
    qualityColor: quality.color,
    qualityRainbow: Boolean(quality.rainbow),
    count: stack && Number.isFinite(stack.count) ? Math.max(0, Math.floor(stack.count)) : 0,
  };
}

function normalizeItemStacks(source) {
  const entries = Array.isArray(source)
    ? source
    : (source && typeof source === 'object'
      ? Object.entries(source).map(([id, count]) => ({ id, count }))
      : []);
  const merged = new Map();
  entries.forEach((entry) => {
    if (!entry || typeof entry !== 'object' || typeof entry.id !== 'string' || !entry.id) return;
    const count = Number.isFinite(entry.count) ? Math.max(0, Math.floor(entry.count)) : 0;
    if (count <= 0) return;
    const item = itemDefinition(entry.id);
    const maxStack = Math.max(1, Number(item.maxStack) || 1);
    const previous = merged.get(entry.id) || 0;
    merged.set(entry.id, previous + count);
  });
  const stacks = [];
  merged.forEach((total, itemId) => {
    const maxStack = Math.max(1, Number(itemDefinition(itemId).maxStack) || 1);
    let remaining = total;
    while (remaining > 0) {
      const count = Math.min(maxStack, remaining);
      stacks.push({ id: itemId, count });
      remaining -= count;
    }
  });
  return stacks;
}

function stackCount(stacks, itemId) {
  if (!Array.isArray(stacks)) return 0;
  return stacks.reduce((total, entry) => (
    entry && entry.id === itemId ? total + Math.max(0, Math.floor(entry.count || 0)) : total
  ), 0);
}

function setStackCount(stacks, itemId, count) {
  if (!Array.isArray(stacks) || typeof itemId !== 'string') return stacks;
  const normalizedCount = Math.max(0, Math.floor(Number(count) || 0));
  const previousIndexes = [];
  for (let index = stacks.length - 1; index >= 0; index -= 1) {
    if (stacks[index] && stacks[index].id === itemId) {
      previousIndexes.unshift(index);
      stacks[index] = null;
    }
  }
  if (normalizedCount <= 0) {
    return stacks;
  }
  const item = itemDefinition(itemId);
  const maxStack = Math.max(1, Number(item.maxStack) || 1);
  let remaining = normalizedCount;
  const targetIndexes = Array.from(new Set(previousIndexes.concat(
    stacks.map((entry, index) => (entry == null ? index : -1)).filter((index) => index >= 0),
  )));
  let targetCursor = 0;
  while (remaining > 0) {
    const nextCount = Math.min(maxStack, remaining);
    const targetIndex = targetIndexes[targetCursor];
    if (Number.isInteger(targetIndex)) stacks[targetIndex] = { id: itemId, count: nextCount };
    else stacks.push({ id: itemId, count: nextCount });
    targetCursor += 1;
    remaining -= nextCount;
  }
  return stacks;
}

function addToStacks(stacks, itemId, amount) {
  if (!Array.isArray(stacks) || typeof itemId !== 'string') return stacks;
  const amountToAdd = Math.max(0, Math.floor(Number(amount) || 0));
  if (amountToAdd <= 0) return stacks;
  const item = itemDefinition(itemId);
  const maxStack = Math.max(1, Number(item.maxStack) || 1);
  let remaining = amountToAdd;
  for (let index = 0; index < stacks.length && remaining > 0; index += 1) {
    const stack = stacks[index];
    if (!stack || stack.id !== itemId) continue;
    const room = Math.max(0, maxStack - Math.max(0, Math.floor(stack.count || 0)));
    if (room <= 0) continue;
    const added = Math.min(room, remaining);
    stack.count += added;
    remaining -= added;
  }
  while (remaining > 0) {
    const nextCount = Math.min(maxStack, remaining);
    const emptyIndex = stacks.findIndex((entry) => entry == null);
    if (emptyIndex >= 0) stacks[emptyIndex] = { id: itemId, count: nextCount };
    else stacks.push({ id: itemId, count: nextCount });
    remaining -= nextCount;
  }
  return stacks;
}

function removeFromStacks(stacks, itemId, amount) {
  const current = stackCount(stacks, itemId);
  return setStackCount(stacks, itemId, current - Math.max(0, Math.floor(Number(amount) || 0)));
}

function techniqueItemId(name) {
  return TECHNIQUE_ITEM_IDS[name] || null;
}

function techniqueNameForItemId(itemId) {
  return Object.entries(TECHNIQUE_ITEM_IDS).find(([, id]) => id === itemId)?.[0] || null;
}

function resolveItemId(value) {
  if (typeof value === 'string') return techniqueItemId(value) || value;
  if (value && typeof value === 'object' && typeof value.id === 'string') return value.id;
  return null;
}

export {
  ITEM_CATALOG,
  ITEM_CATEGORY_DEFINITIONS,
  LEGACY_RESOURCE_ITEM_IDS,
  QUALITY_DEFINITIONS,
  TECHNIQUE_ITEM_IDS,
  addToStacks,
  categoryDefinition,
  itemDefinition,
  itemView,
  normalizeItemStacks,
  qualityDefinition,
  removeFromStacks,
  resolveItemId,
  setStackCount,
  stackCount,
  techniqueItemId,
  techniqueNameForItemId,
};
