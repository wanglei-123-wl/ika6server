<script setup>
import { computed } from 'vue';
import StateBlock from '../components/StateBlock.vue';

const props = defineProps({
  devDocs: {
    type: Object,
    required: true,
  },
  activeDocKey: {
    type: String,
    required: true,
  },
});

const emit = defineEmits(['update:activeDocKey']);
const currentDoc = computed(() => props.devDocs[props.activeDocKey] || {
  title: '内容加载中',
  sub: '正在读取开发者中心内容。',
  steps: [],
  code: '',
});
</script>

<template>
  <section class="view active">
    <div class="container page-hero">
      <div class="crumbs"><a href="#/">首页</a><span>/</span><span>开发者</span></div>
      <h1 class="page-title">开发者中心</h1>
      <p class="page-sub">上传指南 · SDK 文档 · API 参考 · 从 0 到 1 发布你的开源游戏</p>
    </div>
    <div v-if="Object.keys(devDocs).length" class="container dev-layout">
      <aside class="dev-side">
        <button v-for="(doc, key) in devDocs" :key="key" class="dev-link" :class="{ active: activeDocKey === key }" @click="emit('update:activeDocKey', key)">{{ doc.title }}</button>
      </aside>
      <article class="dev-content">
        <h2>{{ currentDoc.title }}</h2>
        <p>{{ currentDoc.sub }}</p>
        <div class="step-flow">
          <div v-for="(step, index) in currentDoc.steps" :key="step" class="step-item"><div class="step-num">{{ index + 1 }}</div><div>{{ step }}</div></div>
        </div>
        <pre class="code-block">{{ currentDoc.code }}</pre>
      </article>
    </div>
    <div v-else class="container">
      <StateBlock icon="文" title="开发者内容加载中" text="开发者中心内容会通过接口提供，当前暂无可展示数据。" />
    </div>
  </section>
</template>
