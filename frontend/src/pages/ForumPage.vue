<script setup>
import PostItem from '../components/PostItem.vue';
import StateBlock from '../components/StateBlock.vue';

defineProps({
  bars: {
    type: Array,
    required: true,
  },
  activeBar: {
    type: Object,
    required: true,
  },
  forumCats: {
    type: Array,
    required: true,
  },
  activeForumCat: {
    type: String,
    required: true,
  },
  forumPosts: {
    type: Array,
    required: true,
  },
});

const pinnedPost = {
  ava: '◆',
  bg: 'var(--gradient-3)',
  name: '吧务 · PIXEL FORGE 官方',
  tags: ['置顶'],
  title: '【吧规】发帖必读：晒作品、求帮助、找队友都到这里来',
  excerpt: '发布游戏请附截图和介绍；求助和招募请标注分类；开源作品请注明协议。',
  replies: '108',
  views: '9.8k',
  likes: 999,
  liked: true,
};

const emit = defineEmits(['update:activeForumCat', 'update:activeBar', 'notice', 'create-post', 'like-post']);

function selectBar(bar) {
  emit('update:activeBar', bar);
  emit('notice', `已切换到「${bar.name}」`);
}
</script>

<template>
  <section class="view active">
    <div class="container page-hero">
      <div class="crumbs"><a href="#/">首页</a><span>/</span><span>论坛</span></div>
      <h1 class="page-title">社区论坛</h1>
      <p class="page-sub">贴吧式交流 · 每个游戏一个吧 · 自由盖楼发帖</p>
    </div>
    <div class="container forum-banner">
      <div class="avatar-big">{{ activeBar.icon }}</div>
      <div class="info">
        <h2>{{ activeBar.name }} <span>贴吧 · 官方区</span></h2>
        <div class="sub">{{ activeBar.desc }}</div>
        <div class="stats"><div><b>{{ activeBar.posts }}</b><span>帖子</span></div><div><b>{{ activeBar.members }}</b><span>吧友</span></div><div><b>3,847</b><span>在线</span></div><div><b>96</b><span>吧务</span></div></div>
      </div>
    </div>
    <div class="container forum-layout">
      <div>
        <div class="forum-cats">
          <button v-for="cat in forumCats" :key="cat" class="cat-btn" :class="{ active: activeForumCat === cat }" @click="emit('update:activeForumCat', cat)">{{ cat }}</button>
        </div>
        <PostItem :post="pinnedPost" pinned />
        <div v-if="forumPosts.length" class="post-list">
          <PostItem v-for="post in forumPosts" :key="post.title" :post="post" @like="emit('like-post', $event)" />
        </div>
        <StateBlock v-else icon="帖" title="这个分类暂无帖子" text="换个分类看看，或者等后端返回新的社区内容。" />
      </div>
      <aside class="side-panel">
        <div class="panel">
          <div class="panel-head"><div class="h">游戏吧</div><a href="#" @click.prevent="emit('notice', '查看更多吧')">全部</a></div>
          <div class="panel-body">
            <button v-for="bar in bars" :key="bar.name" class="tieba-row" :class="{ active: activeBar.name === bar.name }" @click="selectBar(bar)">
              <div class="tieba-ico">{{ bar.icon }}</div>
              <div class="tieba-body"><div class="tieba-name">{{ bar.name }} <span v-if="bar.hot" class="tieba-tag">热</span></div><div class="tieba-desc">{{ bar.desc }}</div><div class="tieba-stats">{{ bar.posts }} 帖 · {{ bar.members }} 吧友</div></div>
            </button>
            <StateBlock v-if="!bars.length" icon="吧" title="暂无游戏吧" text="游戏吧列表会从后端接口读取。" />
          </div>
        </div>
        <button class="upload-cta" @click="emit('create-post')"><span class="ic">＋</span><strong>发一篇帖子</strong><span>晒作品、分享源码、招募队友。</span></button>
      </aside>
    </div>
  </section>
</template>
