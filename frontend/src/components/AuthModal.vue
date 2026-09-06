<script setup>
import { ref, watch } from 'vue';
import { useAuthStore } from '../stores/authStore';

const props = defineProps({
  show: {
    type: Boolean,
    required: true,
  },
  defaultTab: {
    type: String,
    default: 'login',
  },
});

const emit = defineEmits(['close', 'authenticated', 'notice']);
const authStore = useAuthStore();

const tab = ref('login');
const loginAccount = ref('');
const loginPassword = ref('');
const rememberMe = ref(true);
const registerName = ref('');
const registerAccount = ref('');
const registerPassword = ref('');
const registerPasswordConfirm = ref('');
const agree = ref(false);
const showLoginPassword = ref(false);
const showRegisterPassword = ref(false);
const showRegisterPasswordConfirm = ref(false);
const errors = ref({});
const loading = ref(false);

function reset() {
  tab.value = props.defaultTab;
  loginAccount.value = '';
  loginPassword.value = '';
  rememberMe.value = true;
  registerName.value = '';
  registerAccount.value = '';
  registerPassword.value = '';
  registerPasswordConfirm.value = '';
  agree.value = false;
  showLoginPassword.value = false;
  showRegisterPassword.value = false;
  showRegisterPasswordConfirm.value = false;
  errors.value = {};
  loading.value = false;
}

watch(() => props.show, (visible) => {
  if (visible) reset();
});

watch(tab, () => {
  errors.value = {};
});

function switchTab(nextTab) {
  if (!loading.value) tab.value = nextTab;
}

function isValidAccount(value) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value) || /^1[3-9]\d{9}$/.test(value);
}

async function submitLogin() {
  const nextErrors = {};
  const account = loginAccount.value.trim();

  if (!isValidAccount(account)) nextErrors.loginAccount = '请输入有效的邮箱或手机号';
  if (loginPassword.value.length < 6) nextErrors.loginPassword = '密码至少 6 位';

  errors.value = nextErrors;
  if (Object.keys(nextErrors).length || loading.value) return;

  loading.value = true;
  try {
    const result = await authStore.loginWithPassword({
      account,
      password: loginPassword.value,
      remember: rememberMe.value,
    });
    emit('authenticated', result.user);
  } catch (error) {
    emit('notice', error.message || '登录失败，请稍后重试');
  } finally {
    loading.value = false;
  }
}

async function submitRegister() {
  const nextErrors = {};
  const name = registerName.value.trim();
  const account = registerAccount.value.trim();

  if (name.length < 2) nextErrors.registerName = '昵称至少 2 个字符';
  if (!isValidAccount(account)) nextErrors.registerAccount = '请输入有效的邮箱或手机号';
  if (registerPassword.value.length < 6) nextErrors.registerPassword = '密码至少 6 位';
  if (registerPasswordConfirm.value !== registerPassword.value) nextErrors.registerPasswordConfirm = '两次输入的密码不一致';

  errors.value = nextErrors;
  if (!agree.value) {
    emit('notice', '请先同意《用户协议》与《开源规范》');
    return;
  }
  if (Object.keys(nextErrors).length || loading.value) return;

  loading.value = true;
  try {
    const result = await authStore.registerAccount({
      name,
      account,
      password: registerPassword.value,
    });
    emit('authenticated', result.user);
  } catch (error) {
    emit('notice', error.message || '注册失败，请稍后重试');
  } finally {
    loading.value = false;
  }
}

async function socialLogin(provider) {
  if (loading.value) return;
  loading.value = true;

  try {
    const result = await authStore.loginWithProvider(provider);
    emit('authenticated', result.user);
  } catch (error) {
    emit('notice', error.message || `${provider} 登录失败`);
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div v-if="show" class="modal-bg" @click.self="emit('close')">
    <div class="modal auth-modal" role="dialog" aria-modal="true" aria-labelledby="auth-title">
      <div class="modal-head">
        <h3 id="auth-title">欢迎来到 <span class="grad">PIXEL FORGE</span></h3>
        <button class="icon-btn" aria-label="关闭登录窗口" @click="emit('close')">×</button>
      </div>
      <div class="modal-body auth-body">
        <div class="auth-tabs" role="tablist">
          <button class="auth-tab" :class="{ active: tab === 'login' }" @click="switchTab('login')">登录</button>
          <button class="auth-tab" :class="{ active: tab === 'register' }" @click="switchTab('register')">注册</button>
        </div>

        <form v-if="tab === 'login'" class="auth-pane" @submit.prevent="submitLogin">
          <div class="auth-field" :class="{ invalid: errors.loginAccount }">
            <label for="login-account">邮箱 / 手机号</label>
            <input id="login-account" v-model="loginAccount" class="auth-input" type="text" autocomplete="username" placeholder="you@pixelforge.dev" />
            <span v-if="errors.loginAccount" class="auth-error">{{ errors.loginAccount }}</span>
          </div>
          <div class="auth-field" :class="{ invalid: errors.loginPassword }">
            <label for="login-password">密码</label>
            <div class="auth-input-wrap">
              <input id="login-password" v-model="loginPassword" class="auth-input" :type="showLoginPassword ? 'text' : 'password'" autocomplete="current-password" placeholder="至少 6 位" />
              <button type="button" class="auth-eye" :aria-label="showLoginPassword ? '隐藏密码' : '显示密码'" @click="showLoginPassword = !showLoginPassword">{{ showLoginPassword ? '隐藏' : '显示' }}</button>
            </div>
            <span v-if="errors.loginPassword" class="auth-error">{{ errors.loginPassword }}</span>
          </div>
          <div class="auth-row">
            <label class="auth-check"><input v-model="rememberMe" type="checkbox" /> 记住我</label>
            <button type="button" class="auth-link" @click="emit('notice', '重置密码邮件已发送')">忘记密码？</button>
          </div>
          <button class="btn btn-primary auth-submit" type="submit" :disabled="loading">
            <span v-if="loading" class="auth-spinner"></span>{{ loading ? '登录中...' : '登录' }}
          </button>
          <div class="auth-divider">或使用以下方式登录</div>
          <div class="social-row">
            <button v-for="provider in ['GitHub', 'Google', '微信', 'QQ']" :key="provider" class="social-btn" type="button" @click="socialLogin(provider)">
              <span class="social-icon">{{ provider === 'GitHub' ? '◉' : provider === 'Google' ? 'G' : provider === '微信' ? '微' : 'Q' }}</span>{{ provider }}
            </button>
          </div>
          <div class="auth-hint">还没有账号？ <button type="button" class="auth-link" @click="switchTab('register')">立即注册</button></div>
        </form>

        <form v-else class="auth-pane" @submit.prevent="submitRegister">
          <div class="auth-field" :class="{ invalid: errors.registerName }">
            <label for="register-name">昵称</label>
            <input id="register-name" v-model="registerName" class="auth-input" type="text" autocomplete="nickname" placeholder="给自己起个名字" />
            <span v-if="errors.registerName" class="auth-error">{{ errors.registerName }}</span>
          </div>
          <div class="auth-field" :class="{ invalid: errors.registerAccount }">
            <label for="register-account">邮箱 / 手机号</label>
            <input id="register-account" v-model="registerAccount" class="auth-input" type="text" autocomplete="email" placeholder="you@pixelforge.dev" />
            <span v-if="errors.registerAccount" class="auth-error">{{ errors.registerAccount }}</span>
          </div>
          <div class="auth-field" :class="{ invalid: errors.registerPassword }">
            <label for="register-password">密码</label>
            <div class="auth-input-wrap">
              <input id="register-password" v-model="registerPassword" class="auth-input" :type="showRegisterPassword ? 'text' : 'password'" autocomplete="new-password" placeholder="至少 6 位" />
              <button type="button" class="auth-eye" :aria-label="showRegisterPassword ? '隐藏密码' : '显示密码'" @click="showRegisterPassword = !showRegisterPassword">{{ showRegisterPassword ? '隐藏' : '显示' }}</button>
            </div>
            <span v-if="errors.registerPassword" class="auth-error">{{ errors.registerPassword }}</span>
          </div>
          <div class="auth-field" :class="{ invalid: errors.registerPasswordConfirm }">
            <label for="register-password-confirm">确认密码</label>
            <div class="auth-input-wrap">
              <input id="register-password-confirm" v-model="registerPasswordConfirm" class="auth-input" :type="showRegisterPasswordConfirm ? 'text' : 'password'" autocomplete="new-password" placeholder="再次输入密码" />
              <button type="button" class="auth-eye" :aria-label="showRegisterPasswordConfirm ? '隐藏密码' : '显示密码'" @click="showRegisterPasswordConfirm = !showRegisterPasswordConfirm">{{ showRegisterPasswordConfirm ? '隐藏' : '显示' }}</button>
            </div>
            <span v-if="errors.registerPasswordConfirm" class="auth-error">{{ errors.registerPasswordConfirm }}</span>
          </div>
          <label class="auth-check auth-agree"><input v-model="agree" type="checkbox" /> 我已阅读并同意 <button type="button" class="auth-link" @click="emit('notice', '《PIXEL FORGE 用户协议》')">《用户协议》</button> 与 <button type="button" class="auth-link" @click="emit('notice', '《开源规范》')">《开源规范》</button></label>
          <button class="btn btn-primary auth-submit" type="submit" :disabled="loading">
            <span v-if="loading" class="auth-spinner"></span>{{ loading ? '创建中...' : '创建账号' }}
          </button>
          <div class="auth-divider">注册即代表同意开源共享</div>
          <div class="auth-hint">已有账号？ <button type="button" class="auth-link" @click="switchTab('login')">直接登录</button></div>
        </form>
        <div class="auth-tip">这是交互原型，登录后状态会保存在当前浏览器。</div>
      </div>
    </div>
  </div>
</template>
