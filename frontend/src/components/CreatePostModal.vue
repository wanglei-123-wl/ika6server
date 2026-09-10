<script setup>
import { reactive, ref, watch } from 'vue';

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
  submitting: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(['close', 'submit']);

const form = reactive({
  title: '',
  cat: '作品发布',
  tags: '',
  content: '',
  files: [],
});
const errors = ref({});

function reset() {
  form.title = '';
  form.cat = '作品发布';
  form.tags = '';
  form.content = '';
  form.files = [];
  errors.value = {};
}

function readableSize(file) {
  if (!file) return '';
  if (file.size >= 1024 * 1024) return `${(file.size / 1024 / 1024).toFixed(1)} MB`;
  return `${Math.max(1, Math.round(file.size / 1024))} KB`;
}

function setFiles(event) {
  form.files = Array.from(event.target.files || []);
  errors.value = { ...errors.value, files: '' };
}

watch(() => props.show, (visible) => {
  if (visible) reset();
});

function validate() {
  const nextErrors = {};

  if (form.title.trim().length < 4) nextErrors.title = '标题至少 4 个字符';
  if (form.content.trim().length < 12) nextErrors.content = '正文至少 12 个字符';
  if (form.files.length > 5) nextErrors.files = '附件最多上传 5 个';
  if (form.files.some((file) => file.size > 50 * 1024 * 1024)) nextErrors.files = '单个附件不能超过 50 MB';

  errors.value = nextErrors;
  return Object.keys(nextErrors).length === 0;
}

function close() {
  if (!props.submitting) emit('close');
}

function submit() {
  if (props.submitting || !validate()) return;

  emit('submit', {
    title: form.title.trim(),
    cat: form.cat,
    tags: form.tags.split(/[,，\s]+/).map((tag) => tag.trim()).filter(Boolean),
    content: form.content.trim(),
    files: form.files,
  });
}
</script>

<template>
  <div v-if="show" class="modal-bg" @click.self="close">
    <div class="modal post-modal" role="dialog" aria-modal="true" aria-labelledby="post-title">
      <div class="modal-head">
        <h3 id="post-title">发布社区帖子</h3>
        <button class="icon-btn" :disabled="submitting" aria-label="关闭发帖窗口" @click="close">×</button>
      </div>

      <div class="modal-body">
        <div class="upload-status post-status">
          <span>帖</span>
          <div>
            <strong>发到社区论坛</strong>
            <p>帖子会发布到当前游戏吧；如选择附件，正文发布成功后会继续上传附件。</p>
          </div>
        </div>

        <div class="field" :class="{ invalid: errors.title }">
          <label for="forum-post-title">帖子标题</label>
          <input id="forum-post-title" v-model="form.title" placeholder="例如：【自荐】我的开源游戏终于能玩了" />
          <span v-if="errors.title" class="auth-error">{{ errors.title }}</span>
        </div>

        <div class="upload-grid">
          <div class="field">
            <label for="forum-post-cat">帖子分类</label>
            <select id="forum-post-cat" v-model="form.cat">
              <option>作品发布</option>
              <option>互助</option>
              <option>源码分享</option>
              <option>招募</option>
              <option>灌水</option>
            </select>
          </div>
          <div class="field">
            <label for="forum-post-tags">标签</label>
            <input id="forum-post-tags" v-model="form.tags" placeholder="自荐, 开源, 教程" />
          </div>
        </div>

        <div class="field" :class="{ invalid: errors.content }">
          <label for="forum-post-content">正文内容</label>
          <textarea id="forum-post-content" v-model="form.content" placeholder="写清楚玩法、截图说明、求助问题或招募要求..." />
          <span v-if="errors.content" class="auth-error">{{ errors.content }}</span>
        </div>

        <div class="field" :class="{ invalid: errors.files }">
          <label for="forum-post-files">帖子附件</label>
          <label class="drop-zone post-file-zone" for="forum-post-files">
            <input id="forum-post-files" type="file" multiple @change="setFiles" />
            <div class="ico">附</div>
            <div>
              <div class="t">{{ form.files.length ? `已选择 ${form.files.length} 个附件` : '上传附件，可选' }}</div>
              <div class="s">支持图片、压缩包或文档；最多 5 个，单个 50 MB</div>
            </div>
          </label>
          <div v-if="form.files.length" class="file-chip-row">
            <span v-for="file in form.files" :key="`${file.name}-${file.size}`" class="file-chip">{{ file.name }} · {{ readableSize(file) }}</span>
          </div>
          <span v-if="errors.files" class="auth-error">{{ errors.files }}</span>
        </div>
      </div>

      <div class="modal-foot">
        <button class="btn btn-ghost" :disabled="submitting" @click="close">取消</button>
        <button class="btn btn-primary" :disabled="submitting" @click="submit">
          <span v-if="submitting" class="auth-spinner"></span>{{ submitting ? '发布中...' : '发布帖子' }}
        </button>
      </div>
    </div>
  </div>
</template>
