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
  summary: '',
  description: '',
  engine: 'Godot 4',
  genre: '平台跳跃',
  license: 'MIT',
  coverFile: null,
  buildFile: null,
  sourceFile: null,
});
const errors = ref({});

function reset() {
  form.title = '';
  form.summary = '';
  form.description = '';
  form.engine = 'Godot 4';
  form.genre = '平台跳跃';
  form.license = 'MIT';
  form.coverFile = null;
  form.buildFile = null;
  form.sourceFile = null;
  errors.value = {};
}

watch(() => props.show, (visible) => {
  if (visible) reset();
});

function readableSize(file) {
  if (!file) return '';
  if (file.size >= 1024 * 1024) return `${(file.size / 1024 / 1024).toFixed(1)} MB`;
  return `${Math.max(1, Math.round(file.size / 1024))} KB`;
}

function setFile(field, event) {
  const [file] = event.target.files || [];
  form[field] = file || null;
  errors.value = { ...errors.value, [field]: '' };
}

function validate() {
  const nextErrors = {};
  const maxBuildSize = 500 * 1024 * 1024;
  const maxSourceSize = 300 * 1024 * 1024;
  const maxCoverSize = 8 * 1024 * 1024;

  if (form.title.trim().length < 2) nextErrors.title = '游戏标题至少 2 个字符';
  if (form.summary.trim().length < 6) nextErrors.summary = '一句话简介至少 6 个字符';
  if (form.description.trim().length < 20) nextErrors.description = '详细介绍至少 20 个字符，方便审核和玩家了解';
  if (!form.coverFile) nextErrors.coverFile = '请上传一张游戏封面';
  if (!form.buildFile) nextErrors.buildFile = '请上传可试玩的游戏构建包';
  if (form.coverFile && !form.coverFile.type.startsWith('image/')) nextErrors.coverFile = '封面必须是图片文件';
  if (form.coverFile && form.coverFile.size > maxCoverSize) nextErrors.coverFile = '封面不能超过 8 MB';
  if (form.buildFile && form.buildFile.size > maxBuildSize) nextErrors.buildFile = '游戏构建包不能超过 500 MB';
  if (form.sourceFile && form.sourceFile.size > maxSourceSize) nextErrors.sourceFile = '源码包不能超过 300 MB';

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
    summary: form.summary.trim(),
    description: form.description.trim(),
    engine: form.engine,
    genre: form.genre,
    license: form.license,
    coverFile: form.coverFile,
    buildFile: form.buildFile,
    sourceFile: form.sourceFile,
  });
}
</script>

<template>
  <div v-if="show" class="modal-bg" @click.self="close">
    <div class="modal upload-modal" role="dialog" aria-modal="true" aria-labelledby="upload-title">
      <div class="modal-head">
        <h3 id="upload-title">发布游戏到 PIXEL FORGE</h3>
        <button class="icon-btn" :disabled="submitting" aria-label="关闭发布窗口" @click="close">×</button>
      </div>

      <div class="modal-body upload-body">
        <div class="upload-status">
          <span>1</span>
          <div>
            <strong>提交后进入审核队列</strong>
            <p>现在先走前端 mock 流程，后端接通后会上传文件并返回审核状态。</p>
          </div>
        </div>

        <div class="field" :class="{ invalid: errors.title }">
          <label for="game-title">游戏标题</label>
          <input id="game-title" v-model="form.title" placeholder="给你的作品起个名字..." />
          <span v-if="errors.title" class="auth-error">{{ errors.title }}</span>
        </div>

        <div class="upload-grid">
          <div class="field" :class="{ invalid: errors.summary }">
            <label for="game-summary">一句话简介</label>
            <input id="game-summary" v-model="form.summary" placeholder="一句话讲清你做了什么..." />
            <span v-if="errors.summary" class="auth-error">{{ errors.summary }}</span>
          </div>
          <div class="field">
            <label for="game-engine">游戏引擎</label>
            <select id="game-engine" v-model="form.engine">
              <option>Godot 4</option>
              <option>Unity 6</option>
              <option>Unreal Engine</option>
              <option>Bevy</option>
              <option>LÖVE</option>
              <option>HTML5</option>
              <option>其他</option>
            </select>
          </div>
        </div>

        <div class="upload-grid">
          <div class="field">
            <label for="game-genre">游戏类型</label>
            <select id="game-genre" v-model="form.genre">
              <option>平台跳跃</option>
              <option>Roguelike</option>
              <option>策略</option>
              <option>解谜</option>
              <option>恐怖</option>
              <option>养成/恋爱</option>
              <option>其他</option>
            </select>
          </div>
          <div class="field">
            <label for="game-license">开源协议</label>
            <select id="game-license" v-model="form.license">
              <option>MIT</option>
              <option>Apache 2.0</option>
              <option>GPL v3</option>
              <option>CC0</option>
              <option>暂不公开源码</option>
            </select>
          </div>
        </div>

        <div class="field" :class="{ invalid: errors.description }">
          <label for="game-description">详细介绍</label>
          <textarea id="game-description" v-model="form.description" placeholder="聊聊你的玩法、开发故事、操作方式、适配说明..." />
          <span v-if="errors.description" class="auth-error">{{ errors.description }}</span>
        </div>

        <div class="upload-files">
          <label class="drop-zone upload-drop" :class="{ invalid: errors.coverFile }">
            <input type="file" accept="image/*" @change="setFile('coverFile', $event)" />
            <div class="ico">封</div>
            <div class="t">{{ form.coverFile ? form.coverFile.name : '上传游戏封面' }}</div>
            <div class="s">{{ form.coverFile ? readableSize(form.coverFile) : 'PNG / JPG / WEBP · 最大 8 MB' }}</div>
            <span v-if="errors.coverFile" class="auth-error">{{ errors.coverFile }}</span>
          </label>

          <label class="drop-zone upload-drop" :class="{ invalid: errors.buildFile }">
            <input type="file" accept=".zip,.rar,.7z,.tar,.gz,.html" @change="setFile('buildFile', $event)" />
            <div class="ico">包</div>
            <div class="t">{{ form.buildFile ? form.buildFile.name : '上传试玩构建包' }}</div>
            <div class="s">{{ form.buildFile ? readableSize(form.buildFile) : 'ZIP / HTML5 构建 · 最大 500 MB' }}</div>
            <span v-if="errors.buildFile" class="auth-error">{{ errors.buildFile }}</span>
          </label>

          <label class="drop-zone upload-drop" :class="{ invalid: errors.sourceFile }">
            <input type="file" accept=".zip,.rar,.7z,.tar,.gz" @change="setFile('sourceFile', $event)" />
            <div class="ico">源</div>
            <div class="t">{{ form.sourceFile ? form.sourceFile.name : '上传源码包，可选' }}</div>
            <div class="s">{{ form.sourceFile ? readableSize(form.sourceFile) : '源码公开时使用 · 最大 300 MB' }}</div>
            <span v-if="errors.sourceFile" class="auth-error">{{ errors.sourceFile }}</span>
          </label>
        </div>
      </div>

      <div class="modal-foot">
        <button class="btn btn-ghost" :disabled="submitting" @click="close">取消</button>
        <button class="btn btn-primary" :disabled="submitting" @click="submit">
          <span v-if="submitting" class="auth-spinner"></span>{{ submitting ? '提交中...' : '提交审核' }}
        </button>
      </div>
    </div>
  </div>
</template>
