<template>
  <div class="oauth-buttons">
    <div class="divider">
      <span>或使用第三方账号登录</span>
    </div>
    <div class="buttons">
      <button @click="handleOAuthLogin('google')" class="oauth-btn google" :disabled="loading">
        <svg class="icon" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
          <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
          <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
          <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
        </svg>
        <span>使用 Google 登录</span>
      </button>
      <button @click="handleOAuthLogin('github')" class="oauth-btn github" :disabled="loading">
        <svg class="icon" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
          <path fill-rule="evenodd" clip-rule="evenodd" d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.603-3.369-1.34-3.369-1.34-.454-1.156-1.11-1.463-1.11-1.463-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.831.092-.646.35-1.086.636-1.336-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C19.138 20.167 22 16.418 22 12c0-5.523-4.477-10-10-10z" fill="currentColor"/>
        </svg>
        <span>使用 GitHub 登录</span>
      </button>
    </div>
    <div v-if="error" class="error-message">
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useMessage } from 'naive-ui'
import client from '@/api/client'

const route = useRoute()
const message = useMessage()
const loading = ref(false)
const error = ref('')

// OAuth 错误码映射
const errorMessages: Record<string, string> = {
  'auth_cancelled': '您已取消授权',
  'invalid_state': '授权已过期或无效，请重新登录',
  'invalid_code': '授权码无效，请重试',
  'provider_error': 'OAuth 提供商返回错误，请稍后重试',
  'email_conflict': '该邮箱已被其他账号使用',
  'internal_error': '系统内部错误，请稍后重试'
}

// 检查 URL 参数中的错误信息
onMounted(() => {
  const errorCode = route.query.error as string
  if (errorCode) {
    const errorMsg = errorMessages[errorCode] || '登录失败，请重试'
    error.value = errorMsg
    message.error(errorMsg)
  }
})

async function handleOAuthLogin(provider: 'google' | 'github') {
  try {
    loading.value = true
    error.value = ''
    
    // 添加时间戳参数防止浏览器缓存
    const timestamp = Date.now()
    const response = await client.get(`/api/auth/${provider}/login?t=${timestamp}`)
    const authUrl = response.data.authorization_url
    
    if (!authUrl) {
      throw new Error('未能获取授权 URL')
    }
    
    // 使用 location.replace 而不是 location.href，防止浏览器后退按钮问题
    window.location.replace(authUrl)
  } catch (err: any) {
    console.error(`OAuth login error (${provider}):`, err)
    const errorMsg = err.response?.data?.error?.message || `${provider === 'google' ? 'Google' : 'GitHub'} 登录失败，请稍后重试`
    error.value = errorMsg
    message.error(errorMsg)
    loading.value = false
  }
}
</script>

<style scoped>
.oauth-buttons {
  margin-top: 2rem;
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.divider {
  position: relative;
  text-align: center;
  margin-bottom: 1.5rem;
}

.divider::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(to right, transparent, #e5e7eb 20%, #e5e7eb 80%, transparent);
}

.divider span {
  position: relative;
  display: inline-block;
  padding: 0 1rem;
  background: #ffffff;
  color: #6b7280;
  font-size: 0.875rem;
  font-weight: 500;
}

.buttons {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.oauth-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  width: 100%;
  padding: 0.75rem 1.5rem;
  border: 1px solid #d1d5db;
  border-radius: 8px;
  background: #ffffff;
  color: #374151;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s ease;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.oauth-btn:hover:not(:disabled) {
  background: #f9fafb;
  border-color: #9ca3af;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  transform: translateY(-1px);
}

.oauth-btn:active:not(:disabled) {
  transform: translateY(0);
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
}

.oauth-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.oauth-btn .icon {
  width: 20px;
  height: 20px;
  flex-shrink: 0;
}

.oauth-btn.google:hover:not(:disabled) {
  border-color: #4285F4;
}

.oauth-btn.github {
  color: #1f2937;
}

.oauth-btn.github .icon {
  color: #1f2937;
}

.oauth-btn.github:hover:not(:disabled) {
  border-color: #1f2937;
}

.error-message {
  margin-top: 1rem;
  padding: 0.75rem;
  background: #fef2f2;
  border: 1px solid #fecaca;
  border-radius: 8px;
  color: #dc2626;
  font-size: 0.875rem;
  text-align: center;
  animation: slideIn 0.3s ease-out;
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(-10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 响应式设计 */
@media (max-width: 640px) {
  .oauth-btn {
    padding: 0.65rem 1.25rem;
    font-size: 0.9rem;
  }

  .oauth-btn .icon {
    width: 18px;
    height: 18px;
  }

  .divider span {
    font-size: 0.8rem;
  }
}
</style>
