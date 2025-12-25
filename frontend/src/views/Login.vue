<template>
  <div class="login-container">
    <n-card class="login-card" :bordered="false">
      <div class="logo">
        <h1>🎯 Curry2API</h1>
        <p>欢迎使用 Curry2API 服务</p>
      </div>

      <n-tabs v-model:value="activeTab" type="line" animated>
        <n-tab-pane name="login" tab="登录">
          <n-form
            ref="loginFormRef"
            :model="loginForm"
            :rules="loginRules"
            label-placement="left"
            label-width="auto"
            require-mark-placement="right-hanging"
            @submit.prevent="handleLogin"
          >
            <n-form-item label="用户名或邮箱" path="username_or_email">
              <n-input
                v-model:value="loginForm.username_or_email"
                placeholder="请输入用户名或邮箱"
                size="large"
              />
            </n-form-item>
            <n-form-item label="密码" path="password">
              <n-input
                v-model:value="loginForm.password"
                type="password"
                show-password-on="click"
                placeholder="请输入密码"
                size="large"
              />
            </n-form-item>
            <n-button
              type="primary"
              size="large"
              block
              :loading="loginLoading"
              attr-type="submit"
            >
              登录
            </n-button>
          </n-form>
          
          <!-- OAuth 登录按钮 -->
          <OAuthButtons />
        </n-tab-pane>

        <n-tab-pane name="register" tab="注册">
          <n-form
            ref="registerFormRef"
            :model="registerForm"
            :rules="registerRules"
            label-placement="left"
            label-width="auto"
            require-mark-placement="right-hanging"
            @submit.prevent="handleRegister"
          >
            <n-form-item label="用户名" path="username">
              <n-input
                v-model:value="registerForm.username"
                placeholder="3-32个字符"
                size="large"
              />
            </n-form-item>
            <n-form-item label="邮箱" path="email">
              <n-input
                v-model:value="registerForm.email"
                placeholder="your@example.com"
                size="large"
              />
            </n-form-item>
            <n-form-item label="验证码" path="code">
              <n-input-group>
                <n-input
                  v-model:value="registerForm.code"
                  placeholder="请输入6位验证码"
                  size="large"
                  maxlength="6"
                />
                <n-button
                  type="primary"
                  size="large"
                  :disabled="codeCountdown > 0 || !registerForm.turnstileToken"
                  :loading="codeLoading"
                  @click="handleSendCode"
                >
                  {{ codeCountdown > 0 ? `${codeCountdown}秒后重试` : '发送验证码' }}
                </n-button>
              </n-input-group>
            </n-form-item>
            <n-form-item label="密码" path="password">
              <n-input
                v-model:value="registerForm.password"
                type="password"
                show-password-on="click"
                placeholder="至少6个字符"
                size="large"
              />
            </n-form-item>
            <n-form-item label="邀请码" path="referral_code">
              <n-input
                v-model:value="registerForm.referral_code"
                placeholder="选填，输入邀请码可获得额外奖励"
                size="large"
                :disabled="!!referralCodeFromUrl"
              />
            </n-form-item>
            <n-button
              type="primary"
              size="large"
              block
              :loading="registerLoading"
              attr-type="submit"
            >
              注册
            </n-button>
          </n-form>
          
          <!-- OAuth 登录按钮 -->
          <OAuthButtons />
        </n-tab-pane>
      </n-tabs>
    </n-card>

    <!-- Cloudflare Turnstile 人机验证 - 卡片外部 -->
    <div v-if="activeTab === 'register'" class="turnstile-wrapper">
      <div 
        ref="turnstileRef" 
        class="cf-turnstile"
        :data-sitekey="turnstilesiteKey"
        data-theme="light"
        data-callback="onTurnstileSuccess"
        data-error-callback="onTurnstileError"
        data-expired-callback="onTurnstileExpired"
      ></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import type { FormInst, FormRules } from 'naive-ui'
import { authApi } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import type { LoginRequest, RegisterRequest } from '@/types'
import OAuthButtons from '@/components/OAuthButtons.vue'

const router = useRouter()
const route = useRoute()
const message = useMessage()
const authStore = useAuthStore()

// Get referral code from URL query parameter (e.g., /login?ref=ABC12345)
const referralCodeFromUrl = computed(() => {
  return (route.query.ref as string) || (route.query.referral as string) || ''
})

const activeTab = ref('login')
const loginLoading = ref(false)
const registerLoading = ref(false)
const codeLoading = ref(false)
const codeCountdown = ref(0)

const loginFormRef = ref<FormInst | null>(null)
const registerFormRef = ref<FormInst | null>(null)
const turnstileRef = ref<HTMLElement | null>(null)

// Cloudflare Turnstile Site Key (从环境变量获取)
// 默认使用测试密钥（总是通过）
const turnstilesiteKey = import.meta.env.VITE_TURNSTILE_SITE_KEY || '1x00000000000000000000AA'

const loginForm = ref<LoginRequest>({
  username_or_email: '',
  password: ''
})

const registerForm = ref<RegisterRequest & { turnstileToken?: string }>({
  username: '',
  email: '',
  password: '',
  code: '',
  turnstileToken: '',
  referral_code: ''
})

// Turnstile 回调函数
declare global {
  interface Window {
    onTurnstileSuccess: (token: string) => void
    onTurnstileError: () => void
    onTurnstileExpired: () => void
    turnstile?: any
  }
}

window.onTurnstileSuccess = (token: string) => {
  registerForm.value.turnstileToken = token
  message.success('人机验证通过')
}

window.onTurnstileError = () => {
  registerForm.value.turnstileToken = ''
  message.error('人机验证失败，请刷新页面重试')
}

window.onTurnstileExpired = () => {
  registerForm.value.turnstileToken = ''
  message.warning('人机验证已过期，请重新验证')
  // 重置 Turnstile
  if (window.turnstile && turnstileRef.value) {
    window.turnstile.reset(turnstileRef.value)
  }
}

// 加载 Turnstile 脚本
function loadTurnstileScript() {
  return new Promise((resolve, reject) => {
    if (window.turnstile) {
      resolve(window.turnstile)
      return
    }

    const script = document.createElement('script')
    script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js'
    script.async = true
    script.defer = true
    script.onload = () => resolve(window.turnstile)
    script.onerror = reject
    document.head.appendChild(script)
  })
}

onMounted(async () => {
  console.log('🔑 Turnstile Site Key:', turnstilesiteKey)
  console.log('📍 Current Tab:', activeTab.value)
  
  // Check for referral code in URL and switch to register tab if present
  // Requirements: 5.1, 5.5
  if (referralCodeFromUrl.value) {
    console.log('🎁 Referral code detected:', referralCodeFromUrl.value)
    registerForm.value.referral_code = referralCodeFromUrl.value
    activeTab.value = 'register'
    message.info(`邀请码已填入: ${referralCodeFromUrl.value}`)
  }
  
  try {
    await loadTurnstileScript()
    console.log('✅ Turnstile script loaded successfully')
    console.log('🎯 Turnstile object:', window.turnstile)
  } catch (error) {
    console.error('❌ Failed to load Turnstile script:', error)
    message.error('人机验证组件加载失败')
  }
})

// 监控标签页切换
watch(activeTab, async (newTab) => {
  console.log('📑 Tab changed to:', newTab)
  if (newTab === 'register') {
    console.log('🎯 Register tab active - Turnstile should be visible')
    
    // 等待 DOM 更新
    await nextTick()
    console.log('📦 Turnstile ref after nextTick:', turnstileRef.value)
    
    // 如果 Turnstile 已加载但组件未渲染，手动渲染
    if (window.turnstile && turnstileRef.value) {
      console.log('🔄 Manually rendering Turnstile...')
      try {
        window.turnstile.render(turnstileRef.value, {
          sitekey: turnstilesiteKey,
          theme: 'light',
          callback: window.onTurnstileSuccess,
          'error-callback': window.onTurnstileError,
          'expired-callback': window.onTurnstileExpired
        })
        console.log('✅ Turnstile rendered successfully')
      } catch (error) {
        console.error('❌ Failed to render Turnstile:', error)
      }
    }
  }
})

onBeforeUnmount(() => {
  // 清理全局回调
  if (window.onTurnstileSuccess) window.onTurnstileSuccess = () => {}
  if (window.onTurnstileError) window.onTurnstileError = () => {}
  if (window.onTurnstileExpired) window.onTurnstileExpired = () => {}
})

const loginRules: FormRules = {
  username_or_email: [
    { required: true, message: '请输入用户名或邮箱', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const registerRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 32, message: '用户名长度为3-32个字符', trigger: 'blur' }
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入有效的邮箱地址', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码为6位数字', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少6个字符', trigger: 'blur' }
  ]
}

async function handleLogin() {
  try {
    await loginFormRef.value?.validate()
    loginLoading.value = true
    const data = await authApi.login(loginForm.value)
    
    if (data.user) {
      authStore.setUser(data.user)
      message.success('登录成功！')
      router.push('/dashboard')
    }
  } catch (error: any) {
    message.error(error.response?.data?.error?.message || '登录失败')
  } finally {
    loginLoading.value = false
  }
}

async function handleSendCode() {
  if (!registerForm.value.email) {
    message.warning('请先输入邮箱地址')
    return
  }

  if (!registerForm.value.turnstileToken) {
    message.warning('请先完成人机验证')
    return
  }

  try {
    codeLoading.value = true
    await authApi.sendCode({ 
      email: registerForm.value.email,
      turnstile_token: registerForm.value.turnstileToken
    })
    message.success('验证码已发送到您的邮箱')
    
    // Start countdown
    codeCountdown.value = 60
    const timer = setInterval(() => {
      codeCountdown.value--
      if (codeCountdown.value <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  } catch (error: any) {
    message.error(error.response?.data?.error?.message || '发送验证码失败')
    // 验证失败后重置 Turnstile
    if (window.turnstile && turnstileRef.value) {
      window.turnstile.reset(turnstileRef.value)
      registerForm.value.turnstileToken = ''
    }
  } finally {
    codeLoading.value = false
  }
}

async function handleRegister() {
  try {
    await registerFormRef.value?.validate()
    
    if (!registerForm.value.turnstileToken) {
      message.warning('请完成人机验证')
      return
    }
    
    registerLoading.value = true
    
    // Build registration request with optional referral code
    // Requirements: 5.1, 5.5
    const registerData: any = {
      username: registerForm.value.username,
      email: registerForm.value.email,
      password: registerForm.value.password,
      code: registerForm.value.code,
      turnstile_token: registerForm.value.turnstileToken
    }
    
    // Include referral code if provided
    if (registerForm.value.referral_code) {
      registerData.referral_code = registerForm.value.referral_code
    }
    
    const response = await authApi.register(registerData)
    
    // Show success message with referral bonus info if applicable
    if (response.referral_bonus_applied) {
      message.success('注册成功！邀请奖励已发放，请登录')
    } else {
      message.success('注册成功！请登录')
    }
    
    activeTab.value = 'login'
    registerForm.value = {
      username: '',
      email: '',
      password: '',
      code: '',
      turnstileToken: '',
      referral_code: ''
    }
    // 重置 Turnstile
    if (window.turnstile && turnstileRef.value) {
      window.turnstile.reset(turnstileRef.value)
    }
  } catch (error: any) {
    message.error(error.response?.data?.error?.message || '注册失败')
    // 注册失败后重置 Turnstile
    if (window.turnstile && turnstileRef.value) {
      window.turnstile.reset(turnstileRef.value)
      registerForm.value.turnstileToken = ''
    }
  } finally {
    registerLoading.value = false
  }
}
</script>

<style scoped>
.login-container {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 20px;
  background: radial-gradient(circle at top left, #e0e7ff 0%, transparent 40%),
              radial-gradient(circle at bottom right, #f3e8ff 0%, transparent 40%),
              #f8fafc;
  overflow: hidden;
}

/* 装饰性背景元素 */
.login-container::before,
.login-container::after {
  content: '';
  position: absolute;
  width: 600px;
  height: 600px;
  border-radius: 50%;
  filter: blur(80px);
  opacity: 0.5;
  z-index: 0;
  animation: float 10s ease-in-out infinite;
}

.login-container::before {
  top: -100px;
  left: -100px;
  background: radial-gradient(circle, var(--color-primary-light) 0%, transparent 70%);
  animation-delay: 0s;
}

.login-container::after {
  bottom: -100px;
  right: -100px;
  background: radial-gradient(circle, rgba(139, 92, 246, 0.15) 0%, transparent 70%);
  animation-delay: 5s;
}

@keyframes float {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(20px, 20px); }
}

.login-card {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 460px;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.6);
  border-radius: 24px;
  box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.15), 0 0 0 1px rgba(255, 255, 255, 0.2) inset;
  padding: 2rem;
  animation: cardSlideIn 0.6s cubic-bezier(0.16, 1, 0.3, 1);
}

@keyframes cardSlideIn {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.logo {
  text-align: center;
  margin-bottom: 2rem;
}

.logo h1 {
  font-size: 2.25rem;
  margin: 0 0 0.5rem 0;
  font-weight: 800;
  background: linear-gradient(135deg, var(--color-primary) 0%, #8b5cf6 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -0.02em;
}

.logo p {
  color: var(--text-secondary);
  margin: 0;
  font-size: 1rem;
  font-weight: 500;
  opacity: 0.8;
}

/* 表单样式优化 */
:deep(.n-tabs .n-tabs-nav) {
  background: rgba(241, 245, 249, 0.5);
  border-radius: 12px;
  padding: 4px;
  margin-bottom: 24px;
  border: 1px solid rgba(0, 0, 0, 0.05);
}

:deep(.n-tabs .n-tabs-tab) {
  border-radius: 8px;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  font-weight: 600;
  color: var(--text-muted);
}

:deep(.n-tabs .n-tabs-tab--active) {
  background: #ffffff !important;
  color: var(--color-primary) !important;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);
}

:deep(.n-form-item-label) {
  color: var(--text-primary) !important;
  font-weight: 600;
  font-size: 0.9rem;
}

:deep(.n-input) {
  background: rgba(255, 255, 255, 0.6) !important;
  border: 1px solid var(--border-color) !important;
  border-radius: 10px !important;
  transition: all 0.2s ease;
}

:deep(.n-input:hover) {
  border-color: var(--border-color-hover) !important;
  background: rgba(255, 255, 255, 0.8) !important;
}

:deep(.n-input:focus-within) {
  border-color: var(--color-primary) !important;
  background: #ffffff !important;
  box-shadow: 0 0 0 3px var(--color-primary-light) !important;
}

:deep(.n-button--primary-type) {
  height: 44px;
  font-size: 1rem;
  font-weight: 600;
  border-radius: 10px !important;
  background: linear-gradient(135deg, var(--color-primary) 0%, #4338ca 100%) !important;
  border: none !important;
  box-shadow: 0 4px 6px -1px var(--color-primary-light), 0 2px 4px -1px var(--color-primary-light);
  transition: all 0.2s ease;
}

:deep(.n-button--primary-type:hover:not(:disabled)) {
  transform: translateY(-1px);
  box-shadow: 0 10px 15px -3px var(--color-primary-light), 0 4px 6px -2px var(--color-primary-light);
  filter: brightness(1.1);
}

:deep(.n-button--primary-type:active:not(:disabled)) {
  transform: translateY(0);
}

/* 验证码输入框组合优化 */
:deep(.n-input-group .n-button) {
  border-radius: 0 10px 10px 0 !important;
  margin-left: -1px;
}

:deep(.n-input-group .n-input) {
  border-radius: 10px 0 0 10px !important;
}

/* Turnstile 容器 */
.turnstile-wrapper {
  width: 100%;
  display: flex;
  justify-content: center;
  margin-top: 24px;
  min-height: 65px;
}

/* 响应式设计 */
@media (max-width: 640px) {
  .login-container {
    padding: 16px;
    justify-content: flex-start;
    padding-top: 10vh;
  }

  .login-card {
    padding: 1.5rem;
    box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
  }

  .logo h1 {
    font-size: 1.8rem;
  }
}
</style>
