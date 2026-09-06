<script setup>
defineProps({
  path: {
    type: String,
    default: '',
  },
});
</script>

<template>
  <section class="view active">
    <div class="container not-found">
      <div class="not-found-card">
        <span>404</span>
        <h1>这个页面还没有开放</h1>
        <p>当前地址 {{ path || '#/' }} 没有匹配到客户端页面。等后续新增详情页、个人中心或管理子页面时，可以继续接到这里。</p>
        <a class="btn btn-primary" href="#/">回到广场</a>
      </div>
    </div>
  </section>
</template>
