<script setup>
defineProps({
  icon: {
    type: String,
    default: '空',
  },
  title: {
    type: String,
    required: true,
  },
  text: {
    type: String,
    default: '',
  },
});
</script>

<template>
  <div class="state-block">
    <div class="state-icon">{{ icon }}</div>
    <h3>{{ title }}</h3>
    <p v-if="text">{{ text }}</p>
    <slot />
  </div>
</template>
