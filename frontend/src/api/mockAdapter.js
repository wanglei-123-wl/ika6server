import { bars, devDocs, games, posts, repos } from '../data/mock';

const currentUser = {
  id: 10001,
  name: 'quietforge',
  initial: 'Q',
  avatar: '',
  level: 'lv7',
  role: 'admin',
  method: 'mock',
};

let sessionUser = null;
let gameItems = games.map((game, index) => ({
  ...game,
  id: index + 1,
  slug: game.title.toLowerCase().replace(/\s+/g, '-'),
  status: 'published',
  createdAt: `2026-08-${String(index + 10).padStart(2, '0')}`,
}));

let postItems = posts.map((post, index) => ({
  ...post,
  id: index + 1,
  barId: 1,
  authorId: index + 10,
  createdAt: `2026-09-${String(index + 1).padStart(2, '0')}`,
}));

let repoItems = repos.map((repo, index) => ({
  ...repo,
  id: index + 1,
  slug: repo.name,
  author: repo.name.split('-')[0],
}));

const wait = (ms = 220) => new Promise((resolve) => window.setTimeout(resolve, ms));
const clone = (value) => JSON.parse(JSON.stringify(value));
const getActor = () => sessionUser || currentUser;

function toNumber(value) {
  if (typeof value === 'number') return value;
  const text = String(value).toLowerCase().replace(/,/g, '');
  if (text.endsWith('k')) return Number.parseFloat(text) * 1000;
  return Number.parseFloat(text) || 0;
}

function paginate(items, { page = 1, pageSize = 20 } = {}) {
  const currentPage = Number(page) || 1;
  const size = Number(pageSize) || 20;
  const start = (currentPage - 1) * size;

  return {
    page: currentPage,
    pageSize: size,
    total: items.length,
    items: clone(items.slice(start, start + size)),
  };
}

function filterByKeyword(items, keyword, fields) {
  const text = String(keyword || '').trim().toLowerCase();
  if (!text) return items;

  return items.filter((item) => fields.some((field) => String(item[field] || '').toLowerCase().includes(text)));
}

function toggleLike(items, id) {
  const item = items.find((entry) => String(entry.id) === String(id));
  if (!item) throw new Error('Not found');

  item.liked = !item.liked;
  item.likes = Math.max(0, toNumber(item.likes) + (item.liked ? 1 : -1));
  return clone(item);
}

export async function mockLogin(payload) {
  await wait();
  const account = payload.account || '';
  const name = account.includes('@') ? account.split('@')[0] : account.slice(-4) || currentUser.name;
  sessionUser = {
    ...currentUser,
    name,
    initial: name.charAt(0).toUpperCase(),
    role: account.toLowerCase().startsWith('admin') ? 'admin' : 'user',
    method: 'password',
    remember: payload.remember,
  };

  return {
    token: 'mock-token',
    user: clone(sessionUser),
  };
}

export async function mockRegister(payload) {
  await wait();
  sessionUser = {
    ...currentUser,
    name: payload.name,
    initial: payload.name.charAt(0).toUpperCase(),
    level: 'lv1',
    role: 'user',
    method: 'register',
  };

  return {
    token: 'mock-token',
    user: clone(sessionUser),
  };
}

export async function mockSocialLogin(provider) {
  await wait();
  const names = { GitHub: 'quietforge', Google: 'stargazer', 微信: 'pixelwitch', QQ: 'mossybyte' };
  const name = names[provider] || 'player';
  sessionUser = {
    ...currentUser,
    name,
    initial: name.charAt(0).toUpperCase(),
    level: 'lv5',
    role: provider === 'GitHub' ? 'admin' : 'user',
    method: provider,
  };

  return {
    token: 'mock-token',
    user: clone(sessionUser),
  };
}

export async function mockGetCurrentUser() {
  await wait(120);
  return clone(sessionUser || currentUser);
}

export async function mockGetHome() {
  await wait();

  return {
    stats: {
      projects: 12847,
      plays: 847000,
      contributors: 2100,
      price: 0,
    },
    latestGames: clone(gameItems.slice(0, 6)),
    hotPosts: clone(postItems.slice(0, 5)),
    feedPosts: clone(postItems.slice(0, 4)),
  };
}

export async function mockGetGames(params = {}) {
  await wait();
  let items = [...gameItems];

  items = filterByKeyword(items, params.keyword, ['title', 'author', 'engine', 'genre', 'license']);
  if (params.genre) items = items.filter((game) => game.genre === params.genre);
  if (params.engine) items = items.filter((game) => game.engine === params.engine);
  if (params.license) items = items.filter((game) => game.license === params.license);

  return paginate(items, params);
}

export async function mockGetGameDetail(id) {
  await wait();
  const game = gameItems.find((item) => String(item.id) === String(id));
  if (!game) throw new Error('Game not found');

  return clone({
    ...game,
    description: `${game.title} is a playable open-source game project.`,
    screenshots: [],
    sourceUrl: game.hasSource ? `/mock-downloads/${game.slug}.zip` : '',
    playUrl: `/mock-play/${game.slug}/index.html`,
  });
}

export async function mockLikeGame(id) {
  await wait(160);
  return toggleLike(gameItems, id);
}

export async function mockTrackGamePlay(id) {
  await wait(120);
  const game = gameItems.find((item) => String(item.id) === String(id));
  if (!game) throw new Error('Game not found');

  game.plays = toNumber(game.plays) + 1;
  return clone({ playUrl: `/mock-play/${game.slug}/index.html`, plays: game.plays });
}

export async function mockGetGameSource(id) {
  await wait(160);
  const game = gameItems.find((item) => String(item.id) === String(id));
  if (!game || !game.hasSource) throw new Error('Source unavailable');

  return clone({ downloadUrl: `/mock-downloads/${game.slug}.zip`, size: game.size });
}

export async function mockGetBars() {
  await wait();
  return clone(bars.map((bar, index) => ({ ...bar, id: index + 1 })));
}

export async function mockGetPosts(params = {}) {
  await wait();
  let items = [...postItems];

  items = filterByKeyword(items, params.keyword, ['title', 'excerpt', 'name', 'cat']);
  if (params.cat && params.cat !== '全部') items = items.filter((post) => post.cat === params.cat);
  if (params.barId) items = items.filter((post) => String(post.barId) === String(params.barId));

  return paginate(items, params);
}

export async function mockGetPostDetail(id) {
  await wait();
  const post = postItems.find((item) => String(item.id) === String(id));
  if (!post) throw new Error('Post not found');

  return clone(post);
}

export async function mockCreatePost(payload) {
  await wait();
  const actor = getActor();
  const post = {
    id: postItems.length + 1,
    ava: actor.initial,
    bg: 'linear-gradient(135deg,#06B6D4,#3B82F6)',
    name: actor.name,
    level: actor.level,
    time: '刚刚',
    cat: payload.cat || '灌水',
    tags: payload.tags || [],
    title: payload.title,
    excerpt: payload.content,
    replies: '0',
    views: '0',
    likes: 0,
    liked: false,
    barId: payload.barId || 1,
    authorId: actor.id,
    createdAt: new Date().toISOString(),
  };

  postItems = [post, ...postItems];
  return clone(post);
}

export async function mockLikePost(id) {
  await wait(160);
  return toggleLike(postItems, id);
}

export async function mockGetReplies(postId, params = {}) {
  await wait();
  const actor = getActor();
  return paginate([
    {
      id: 1,
      postId,
      author: actor.name,
      avatarText: actor.initial,
      floor: 2,
      content: '这个楼层后面会从真实评论接口读取。',
      createdAt: '刚刚',
      likes: 0,
      liked: false,
    },
  ], params);
}

export async function mockCreateReply(postId, payload) {
  await wait();
  const actor = getActor();
  return clone({
    id: Date.now(),
    postId,
    author: actor.name,
    avatarText: actor.initial,
    floor: 2,
    content: payload.content,
    createdAt: '刚刚',
    likes: 0,
    liked: false,
  });
}

export async function mockGetRepos(params = {}) {
  await wait();
  let items = [...repoItems];

  items = filterByKeyword(items, params.keyword, ['name', 'desc', 'lang', 'license', 'badge']);
  if (params.lang) items = items.filter((repo) => repo.lang === params.lang);
  if (params.license) items = items.filter((repo) => repo.license === params.license);

  return paginate(items, params);
}

export async function mockGetRepoDetail(id) {
  await wait();
  const repo = repoItems.find((item) => String(item.id) === String(id));
  if (!repo) throw new Error('Repo not found');

  return clone(repo);
}

export async function mockDownloadRepo(id) {
  await wait(160);
  const repo = repoItems.find((item) => String(item.id) === String(id));
  if (!repo) throw new Error('Repo not found');

  return clone({ downloadUrl: `/mock-downloads/${repo.slug}.zip`, size: repo.size });
}

export async function mockSearch(params = {}) {
  await wait();
  const keyword = params.keyword || '';

  return {
    games: clone(filterByKeyword(gameItems, keyword, ['title', 'author', 'engine', 'genre']).slice(0, 6)),
    posts: clone(filterByKeyword(postItems, keyword, ['title', 'excerpt', 'name', 'cat']).slice(0, 6)),
    repos: clone(filterByKeyword(repoItems, keyword, ['name', 'desc', 'lang', 'license']).slice(0, 6)),
  };
}

export async function mockSubmitGame(payload) {
  await wait(420);
  const actor = getActor();
  const game = {
    id: gameItems.length + 1,
    title: payload.title,
    glyph: '◆',
    cover: (gameItems.length % 8) + 1,
    badge: 'new',
    author: `@${actor.name}`,
    engine: payload.engine || 'Godot 4',
    size: payload.buildFile?.size ? `${(payload.buildFile.size / 1024 / 1024).toFixed(1)} MB` : '待处理',
    plays: '0',
    likes: 0,
    liked: false,
    hasSource: Boolean(payload.sourceFile),
    genre: payload.genre || '其他',
    license: payload.license || 'MIT',
    slug: `submitted-${Date.now()}`,
    status: 'reviewing',
    createdAt: new Date().toISOString(),
  };

  gameItems = [game, ...gameItems];
  return clone({ ...game, message: 'Game submitted for review' });
}

export async function mockGetAdminDashboard() {
  await wait();

  return {
    pendingGames: 8,
    pendingPosts: 16,
    activeUsers: 3847,
    reports: 3,
  };
}

export async function mockReviewGame(id, payload) {
  await wait();
  return clone({ id, status: payload.status, reason: payload.reason || '' });
}

export async function mockReviewPost(id, payload) {
  await wait();
  return clone({ id, status: payload.status, reason: payload.reason || '' });
}

export async function mockBanUser(id, payload) {
  await wait();
  return clone({ id, banned: true, reason: payload.reason || '', until: payload.until || null });
}

export async function mockGetDevDocs() {
  await wait();
  return clone(devDocs);
}
