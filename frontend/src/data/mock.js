export const games = [
  { title: '虚空回廊 · Void Corridor', glyph: '◆', cover: 1, badge: 'new', author: '@quietforge', engine: 'Godot 4', size: '24 MB', plays: '18.4k', likes: '2.1k', hasSource: true, genre: 'Roguelike', license: 'MIT' },
  { title: '霓虹骑士 · Neon Rider', glyph: '►', cover: 2, badge: 'hot', author: '@neonrider', engine: 'Unity 6', size: '62 MB', plays: '42.7k', likes: '5.8k', hasSource: true, genre: '平台跳跃', license: 'Apache 2.0' },
  { title: '像素猫大冒险', glyph: '▲', cover: 3, badge: 'open', author: '@pixelwitch', engine: 'Godot 4', size: '18 MB', plays: '31.2k', likes: '4.2k', hasSource: true, genre: '平台跳跃', license: 'MIT' },
  { title: '苔藓纪元 · Mossy Era', glyph: '❄', cover: 4, badge: 'new', author: '@mossybyte', engine: 'Bevy', size: '36 MB', plays: '8.9k', likes: '1.4k', hasSource: true, genre: '策略', license: 'MIT' },
  { title: '像素女巫 · Pixel Witch', glyph: '✦', cover: 5, badge: 'hot', author: '@pixelwitch', engine: 'LÖVE', size: '12 MB', plays: '27.3k', likes: '3.8k', hasSource: true, genre: 'Roguelike', license: 'GPL v3' },
  { title: '深渊回响 · Abyss Echo', glyph: '◇', cover: 6, badge: 'open', author: '@catloaf', engine: 'Godot 4', size: '58 MB', plays: '5.6k', likes: '820', hasSource: true, genre: '恐怖', license: 'CC0' },
  { title: '星河咖啡馆', glyph: '✿', cover: 7, badge: 'new', author: '@stargazer', engine: "Ren'Py", size: '86 MB', plays: '14.1k', likes: '2.9k', hasSource: false, genre: '养成/恋爱', license: 'MIT' },
  { title: '末班地铁', glyph: '✚', cover: 8, badge: 'open', author: '@laststop', engine: 'Godot 4', size: '44 MB', plays: '21.7k', likes: '3.1k', hasSource: true, genre: '解谜', license: 'Apache 2.0' },
];

export const posts = [
  { ava: 'P', bg: 'linear-gradient(135deg,#F472B6,#8B5CF6)', name: 'pixelwitch', level: 'lv5', time: '23 分钟前', cat: '作品发布', tags: ['自荐'], title: '【自荐】像素猫大冒险 v1.2 发布，全开源！', excerpt: '肝了两个月，加入了新的 BOSS 战和天气系统。在线直接玩，完整源码也已放出。欢迎来试玩给建议。', media: '▲ 像素猫大冒险 · 新 BOSS 实机演示', replies: '86', views: '1.8k', likes: 142, liked: true },
  { ava: 'N', bg: 'linear-gradient(135deg,#FACC15,#F472B6)', name: 'neonrider', level: 'lv3', time: '1 小时前', cat: '互助', tags: ['新手求助'], title: '第一次做 Roguelike，被玩家骂哭了怎么办', excerpt: '做了两周发出去，3 个玩家评论说平衡稀烂、没手感。这里是怎么从批评里活下来的大佬们？', replies: '203', views: '4.1k', likes: 512, liked: false },
  { ava: 'M', bg: 'linear-gradient(135deg,#10B981,#06B6D4)', name: 'mossybyte', level: 'lv5', time: '2 小时前', cat: '源码分享', tags: ['开源', '教程'], title: '【开源】塔防模板 v2.0，含可视化编辑器', excerpt: 'MIT 协议。包含 8 种塔、12 种敌人、5 种地形，纯 Godot 4 写的，注释齐全。', media: '❄ 塔防模板 v2.0 · 编辑器界面', replies: '156', views: '5.6k', likes: 789, liked: true },
  { ava: 'Q', bg: 'linear-gradient(135deg,#06B6D4,#3B82F6)', name: 'quietforge', level: 'lv7', time: '5 小时前', cat: '作品发布', tags: ['开发日志'], title: '【开发日志 #14】虚空回廊终于做完 BOSS 战', excerpt: '肝了整整 18 天，BOSS 一共有三个阶段。在线试玩和完整源码都已发布，求各位大佬轻喷。', replies: '142', views: '3.2k', likes: 328, liked: false },
  { ava: 'L', bg: 'linear-gradient(135deg,#EF4444,#F59E0B)', name: 'laststop', level: 'lv3', time: '8 小时前', cat: '灌水', tags: ['吐槽'], title: '吐槽：为什么我做的游戏上线 3 天没人玩', excerpt: '标题党封面有、原画有、demo 视频有。下一步应该怎么办，求大神指点。', replies: '412', views: '7.2k', likes: 623, liked: false },
  { ava: 'C', bg: 'linear-gradient(135deg,#1E293B,#475569)', name: 'catloaf', level: 'lv3', time: '1 天前', cat: '招募', tags: ['招募'], title: '[招募] 找一位会写 shader 的队友一起做 HD2D 项目', excerpt: '我们组现在 3 个人，缺一个会写屏幕空间光照和后处理的 shader 大佬。', replies: '87', views: '1.8k', likes: 142, liked: false },
];

export const repos = [
  { icon: '◆', name: 'void-corridor', desc: 'Roguelike 动作 RPG 完整源码，含程序化地牢、对话系统、存档系统。', lang: 'GDScript', dots: '#478CBF', license: 'MIT', stars: '8.2k', forks: '1.1k', downloads: '24.3k', size: '24 MB', badge: '完整模板' },
  { icon: '►', name: 'neon-rider', desc: '2D 横版赛车 + 无限疾驰，支持手柄，内置关卡编辑器。', lang: 'C#', dots: '#178600', license: 'Apache 2.0', stars: '12.6k', forks: '2.4k', downloads: '41.2k', size: '62 MB', badge: '本月热门' },
  { icon: '❄', name: 'mossy-era', desc: '纯 Rust Bevy ECS 策略游戏框架，适合学习数据驱动游戏开发。', lang: 'Rust', dots: '#dea584', license: 'MIT', stars: '4.7k', forks: '680', downloads: '9.8k', size: '36 MB', badge: 'Demo 参考' },
  { icon: '▲', name: 'pixel-cat-adventure', desc: '横版平台跳跃完整作品，含动画状态机、可破坏地形、QTE。', lang: 'GDScript', dots: '#478CBF', license: 'GPL v3', stars: '6.4k', forks: '940', downloads: '31.2k', size: '18 MB', badge: '完整模板' },
];

export const bars = [
  { icon: '◆', name: '独立游戏吧', desc: '独立游戏开发交流 · 开源互助 · 作品发布', posts: '247,503', members: '38,412', hot: true },
  { icon: '▲', name: '像素猫大冒险吧', desc: '横版平台跳跃 · 全开源 · 1.2万行 GDScript', posts: '12,847', members: '3,204' },
  { icon: '►', name: '霓虹骑士吧', desc: '2D 横版赛车 + 无限疾驰 · Unity 6 完整工程', posts: '18,221', members: '4,128' },
  { icon: '❄', name: '苔藓纪元吧', desc: '纯 Rust + Bevy ECS · 数据驱动游戏开发', posts: '6,432', members: '1,847' },
];

export const devDocs = {
  quickstart: {
    title: '快速开始',
    sub: '从 0 到 1，把游戏发布到 PIXEL FORGE 只需要 4 步',
    steps: ['注册并登录', '上传游戏文件', '写一篇帖子介绍', '一键发布'],
    code: 'npm install -g @pixelforge/cli\npf login\npf deploy ./build/web --game-name "虚空回廊"',
  },
  upload: {
    title: '上传指南',
    sub: '不同引擎的打包导出要求，照着做就能一次通过',
    steps: ['Godot 4 导出 Web', 'Unity 构建 WebGL', 'HTML5 静态站点托管'],
    code: 'index.html\nbuild/\nREADME.md\nLICENSE\nassets/',
  },
  web: {
    title: 'Web 构建 & 适配',
    sub: '确保你的游戏在不同浏览器配置上都能流畅运行',
    steps: ['分辨率适配', '触屏适配', '性能预算'],
    code: '<canvas id="game"></canvas>\n<script src="build/game.js"></script>',
  },
  sdk: {
    title: 'SDK / API 参考',
    sub: '用几行代码接入排行榜、成就与社区',
    steps: ['初始化 SDK', '提交分数', '解锁成就'],
    code: "const pf = await PixelForge.init({ gameId: 'void-corridor' });\nawait pf.leaderboard.submit('main', score);\npf.achievements.unlock('first-boss');",
  },
};
