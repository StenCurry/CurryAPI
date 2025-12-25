<template>
  <div class="personal-settings">
    <!-- 页面标题 -->
    <div class="page-header glass-card">
      <h1 class="gradient-text">⚙️ 个人设置</h1>
      <p class="subtitle">管理您的账户信息和偏好设置</p>
    </div>

    <n-space vertical size="large">
      <!-- 用户信息 -->
      <div class="settings-card glass-card">
        <h3 class="card-title">👤 用户信息</h3>
        <n-descriptions :column="2" class="user-info">
          <n-descriptions-item label="用户名">
            <span class="info-value">{{ authStore.user?.username }}</span>
          </n-descriptions-item>
          <n-descriptions-item label="邮箱">
            <span class="info-value">{{ authStore.user?.email }}</span>
          </n-descriptions-item>
          <n-descriptions-item label="角色">
            <n-tag :type="authStore.isAdmin ? 'success' : 'default'">
              {{ authStore.isAdmin ? '管理员' : '普通用户' }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="注册时间">
            <span class="info-value">{{ formatDate(authStore.user?.created_at) }}</span>
          </n-descriptions-item>
        </n-descriptions>
      </div>

      <!-- 账户统计信息 - Requirements: 2.1, 2.2 -->
      <div class="settings-card glass-card">
        <h3 class="card-title">📊 账户统计</h3>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-icon">📡</div>
            <div class="stat-content">
              <div class="stat-value">{{ formatNumber(usageStats.total_requests) }}</div>
              <div class="stat-label">总 API 调用</div>
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-icon">🔤</div>
            <div class="stat-content">
              <div class="stat-value">{{ formatNumber(usageStats.total_tokens) }}</div>
              <div class="stat-label">总 Token 使用</div>
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-icon">📅</div>
            <div class="stat-content">
              <div class="stat-value">{{ accountAge }}</div>
              <div class="stat-label">账户年龄（天）</div>
            </div>
          </div>
          <div class="stat-item">
            <div class="stat-icon">🤖</div>
            <div class="stat-content">
              <div class="stat-value">{{ usageStats.by_model?.length || 0 }}</div>
              <div class="stat-label">使用模型数</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 最近 7 天使用历史图表 - Requirements: 2.3 -->
      <div class="settings-card glass-card">
        <h3 class="card-title">📈 最近 7 天使用趋势</h3>
        <div class="chart-container" v-if="!trendsLoading && usageTrends.length > 0">
          <Line :data="chartData" :options="chartOptions" />
        </div>
        <div class="chart-loading" v-else-if="trendsLoading">
          <n-spin size="medium" />
          <span>加载中...</span>
        </div>
        <div class="chart-empty" v-else>
          <n-empty description="暂无使用数据" />
        </div>
      </div>

      <!-- OAuth 关联账户 - Requirements: 2.4 -->
      <div class="settings-card glass-card" v-if="oauthAccounts.length > 0">
        <h3 class="card-title">🔗 关联账户</h3>
        <div class="oauth-list">
          <div 
            v-for="account in oauthAccounts" 
            :key="account.provider"
            class="oauth-item"
          >
            <div class="oauth-icon">
              <span v-if="account.provider === 'google'">🔵</span>
              <span v-else-if="account.provider === 'github'">⚫</span>
              <span v-else>🔗</span>
            </div>
            <div class="oauth-info">
              <div class="oauth-provider">{{ getProviderName(account.provider) }}</div>
              <div class="oauth-email">{{ account.email || '已关联' }}</div>
            </div>
            <n-tag type="success" size="small">已连接</n-tag>
          </div>
        </div>
      </div>

      <!-- 修改用户名 - Requirements: 2.5 -->
      <div class="settings-card glass-card">
        <h3 class="card-title">✏️ 修改用户名</h3>
        <n-form ref="usernameFormRef" :model="usernameForm" :rules="usernameRules" class="settings-form">
          <n-form-item label="新用户名" path="username">
            <n-input
              v-model:value="usernameForm.username"
              placeholder="请输入新用户名（3-32个字符）"
              :maxlength="32"
              size="large"
            />
          </n-form-item>
          <n-form-item>
            <n-button type="primary" @click="handleUpdateUsername" :loading="usernameLoading" size="large" class="submit-btn">
              更新用户名
            </n-button>
          </n-form-item>
        </n-form>
      </div>

      <!-- 修改密码 -->
      <div class="settings-card glass-card">
        <h3 class="card-title">🔒 修改密码</h3>
        <n-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" class="settings-form">
          <n-form-item label="原密码" path="oldPassword">
            <n-input
              v-model:value="passwordForm.oldPassword"
              type="password"
              placeholder="请输入原密码"
              show-password-on="click"
              size="large"
            />
          </n-form-item>
          <n-form-item label="新密码" path="newPassword">
            <n-input
              v-model:value="passwordForm.newPassword"
              type="password"
              placeholder="请输入新密码（至少6个字符）"
              show-password-on="click"
              size="large"
            />
          </n-form-item>
          <n-form-item label="确认新密码" path="confirmPassword">
            <n-input
              v-model:value="passwordForm.confirmPassword"
              type="password"
              placeholder="请再次输入新密码"
              show-password-on="click"
              size="large"
            />
          </n-form-item>
          <n-form-item>
            <n-button type="primary" @click="handleUpdatePassword" :loading="passwordLoading" size="large" class="submit-btn">
              更新密码
            </n-button>
          </n-form-item>
        </n-form>
      </div>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useMessage, type FormInst, type FormRules } from 'naive-ui'
import { updateUsername, updatePassword } from '@/api/user'
import { getUsageStats, getUsageTrends, type DailyUsage } from '@/api/usage'
import { calculateAccountAge } from '@/utils/gameUtils'
import type { UsageStats } from '@/types'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const authStore = useAuthStore()
const message = useMessage()

const usernameFormRef = ref<FormInst | null>(null)
const passwordFormRef = ref<FormInst | null>(null)

const usernameLoading = ref(false)
const passwordLoading = ref(false)
const statsLoading = ref(false)
const trendsLoading = ref(false)

// Usage statistics - Requirements: 2.2
const usageStats = ref<UsageStats>({
  total_requests: 0,
  total_tokens: 0,
  prompt_tokens: 0,
  completion_tokens: 0,
  by_model: [],
  recent_calls: []
})

// Usage trends for chart - Requirements: 2.3
const usageTrends = ref<DailyUsage[]>([])

// OAuth accounts - Requirements: 2.4
interface OAuthAccount {
  provider: string
  email?: string
}
const oauthAccounts = ref<OAuthAccount[]>([])

const usernameForm = reactive({
  username: ''
})

const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const usernameRules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 32, message: '用户名长度应为3-32个字符', trigger: 'blur' }
  ]
}

const passwordRules: FormRules = {
  oldPassword: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少6个字符', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认新密码', trigger: 'blur' },
    {
      validator: (_rule, value) => {
        return value === passwordForm.newPassword
      },
      message: '两次输入的密码不一致',
      trigger: 'blur'
    }
  ]
}

// Computed: Account age in days - Requirements: 2.2
const accountAge = computed(() => {
  if (!authStore.user?.created_at) return 0
  return calculateAccountAge(authStore.user.created_at)
})

function formatDate(dateString?: string) {
  if (!dateString) return '-'
  return new Date(dateString).toLocaleString('zh-CN')
}

function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

function getProviderName(provider: string): string {
  const names: Record<string, string> = {
    google: 'Google',
    github: 'GitHub'
  }
  return names[provider] || provider
}

// Fetch usage statistics - Requirements: 2.2
async function fetchUsageStats() {
  try {
    statsLoading.value = true
    const stats = await getUsageStats()
    usageStats.value = stats
  } catch (error) {
    console.error('Failed to fetch usage stats:', error)
  } finally {
    statsLoading.value = false
  }
}

// Fetch usage trends for last 7 days - Requirements: 2.3
async function fetchUsageTrends() {
  try {
    trendsLoading.value = true
    const response = await getUsageTrends({ days: 7 })
    usageTrends.value = response.trends || []
  } catch (error) {
    console.error('Failed to fetch usage trends:', error)
  } finally {
    trendsLoading.value = false
  }
}

// Chart.js data computed property - Requirements: 2.3
const chartData = computed(() => {
  if (usageTrends.value.length === 0) {
    return { labels: [], datasets: [] }
  }

  const labels = usageTrends.value.map(d => formatDateLabel(d.date))
  
  return {
    labels,
    datasets: [
      {
        label: 'API 调用',
        data: usageTrends.value.map(d => d.request_count),
        borderColor: 'rgba(59, 130, 246, 1)',
        backgroundColor: 'rgba(59, 130, 246, 0.3)',
        fill: true,
        tension: 0.4,
        pointRadius: 4,
        pointHoverRadius: 7,
        pointBackgroundColor: 'rgba(59, 130, 246, 1)',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        borderWidth: 3,
      },
      {
        label: 'Token 使用',
        data: usageTrends.value.map(d => d.total_tokens),
        borderColor: 'rgba(16, 185, 129, 1)',
        backgroundColor: 'rgba(16, 185, 129, 0.3)',
        fill: true,
        tension: 0.4,
        pointRadius: 4,
        pointHoverRadius: 7,
        pointBackgroundColor: 'rgba(16, 185, 129, 1)',
        pointBorderColor: '#fff',
        pointBorderWidth: 2,
        borderWidth: 3,
        yAxisID: 'y1',
      }
    ]
  }
})

// Chart.js options
const chartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  plugins: {
    legend: {
      display: true,
      position: 'top' as const,
      align: 'end' as const,
      labels: {
        color: 'rgba(255, 255, 255, 0.9)',
        font: {
          family: 'system-ui, -apple-system, sans-serif',
          size: 12,
          weight: 'bold' as const,
        },
        padding: 20,
        usePointStyle: true,
        pointStyle: 'circle',
      },
    },
    tooltip: {
      enabled: true,
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      titleColor: '#fff',
      bodyColor: 'rgba(255, 255, 255, 0.9)',
      borderColor: 'rgba(59, 130, 246, 0.5)',
      borderWidth: 1,
      padding: 16,
      cornerRadius: 12,
      titleFont: {
        size: 14,
        weight: 'bold' as const,
      },
      bodyFont: {
        size: 13,
      },
      displayColors: true,
      boxPadding: 6,
    },
  },
  scales: {
    x: {
      grid: {
        color: 'rgba(255, 255, 255, 0.08)',
      },
      ticks: {
        color: 'rgba(255, 255, 255, 0.7)',
        font: {
          size: 11,
        },
        padding: 8,
      },
    },
    y: {
      type: 'linear' as const,
      display: true,
      position: 'left' as const,
      title: {
        display: true,
        text: 'API 调用',
        color: 'rgba(255, 255, 255, 0.7)',
      },
      grid: {
        color: 'rgba(255, 255, 255, 0.08)',
      },
      ticks: {
        color: 'rgba(255, 255, 255, 0.7)',
        font: {
          size: 11,
        },
        padding: 12,
      },
      beginAtZero: true,
    },
    y1: {
      type: 'linear' as const,
      display: true,
      position: 'right' as const,
      title: {
        display: true,
        text: 'Token 数',
        color: 'rgba(255, 255, 255, 0.7)',
      },
      grid: {
        drawOnChartArea: false,
      },
      ticks: {
        color: 'rgba(255, 255, 255, 0.7)',
        font: {
          size: 11,
        },
        padding: 12,
        callback: (value: any) => formatLargeNumber(Number(value)),
      },
      beginAtZero: true,
    },
  },
}))

function formatDateLabel(dateStr: string): string {
  const date = new Date(dateStr)
  const month = date.getMonth() + 1
  const day = date.getDate()
  return `${month}/${day}`
}

function formatLargeNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toLocaleString()
}

async function handleUpdateUsername() {
  try {
    await usernameFormRef.value?.validate()
    usernameLoading.value = true

    await updateUsername(usernameForm.username)
    message.success('用户名更新成功')
    
    // 刷新用户信息
    await authStore.fetchUser()
    
    // 清空表单
    usernameForm.username = ''
  } catch (error: any) {
    if (error.message) {
      message.error(error.message)
    }
  } finally {
    usernameLoading.value = false
  }
}

async function handleUpdatePassword() {
  try {
    await passwordFormRef.value?.validate()
    passwordLoading.value = true

    await updatePassword(passwordForm.oldPassword, passwordForm.newPassword)
    message.success('密码更新成功')
    
    // 清空表单
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
  } catch (error: any) {
    if (error.message) {
      message.error(error.message)
    }
  } finally {
    passwordLoading.value = false
  }
}

onMounted(() => {
  fetchUsageStats()
  fetchUsageTrends()
})
</script>


<style scoped>
.personal-settings {
  padding: 2rem;
  max-width: 900px;
  margin: 0 auto;
  animation: fadeIn 0.6s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 简约卡片样式 */
.glass-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-md);
  transition: all var(--transition-normal);
}

.glass-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

/* 页面标题 */
.page-header {
  padding: 2.5rem;
  margin-bottom: 2rem;
  text-align: center;
  animation: slideDown 0.8s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page-header h1 {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
  color: var(--text-primary);
}

.gradient-text {
  color: var(--text-primary);
}

.subtitle {
  color: var(--text-secondary);
  font-size: 1.1rem;
  margin: 0;
}

/* 设置卡片 */
.settings-card {
  padding: 2rem;
  margin-bottom: 1.5rem;
  animation: fadeInUp 0.6s ease-out;
  animation-fill-mode: both;
}

.settings-card:nth-child(2) {
  animation-delay: 0.1s;
}

.settings-card:nth-child(3) {
  animation-delay: 0.2s;
}

.settings-card:nth-child(4) {
  animation-delay: 0.3s;
}

.settings-card:nth-child(5) {
  animation-delay: 0.4s;
}

.settings-card:nth-child(6) {
  animation-delay: 0.5s;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.card-title {
  color: var(--text-primary);
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0 0 1.5rem 0;
}

/* 用户信息 */
.user-info {
  background: var(--bg-secondary);
  padding: 1.5rem;
  border-radius: var(--border-radius);
}

:deep(.user-info .n-descriptions-item-label) {
  color: var(--text-secondary) !important;
  font-weight: 600;
}

:deep(.user-info .n-descriptions-item-content) {
  color: var(--text-primary) !important;
}

.info-value {
  color: var(--text-primary);
  font-weight: 500;
}

/* 账户统计网格 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1.5rem;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
  border: 1px solid var(--border-color);
  transition: all var(--transition-fast);
}

.stat-item:hover {
  background: var(--bg-hover);
  transform: translateY(-2px);
}

.stat-icon {
  font-size: 2rem;
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-primary-light);
  border-radius: var(--border-radius);
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.2;
}

.stat-label {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-top: 0.25rem;
}

/* 图表容器 */
.chart-container {
  width: 100%;
  height: 300px;
}

.usage-chart {
  width: 100%;
  height: 100%;
}

.chart-loading,
.chart-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  gap: 1rem;
  color: var(--text-secondary);
}

/* OAuth 关联账户 */
.oauth-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.oauth-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.5rem;
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
  border: 1px solid var(--border-color);
  transition: all var(--transition-fast);
}

.oauth-item:hover {
  background: var(--bg-hover);
}

.oauth-icon {
  font-size: 1.5rem;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-tertiary);
  border-radius: var(--border-radius-sm);
}

.oauth-info {
  flex: 1;
}

.oauth-provider {
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.oauth-email {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-top: 0.25rem;
}

/* 表单样式 */
.settings-form {
  margin-top: 1rem;
}

:deep(.settings-form .n-form-item-label) {
  color: var(--text-primary) !important;
  font-weight: 600;
}

:deep(.settings-form .n-input) {
  background: var(--bg-secondary) !important;
  border: 1px solid var(--border-color) !important;
  border-radius: var(--border-radius) !important;
  color: var(--text-primary) !important;
  transition: all var(--transition-fast);
}

:deep(.settings-form .n-input:hover) {
  background: var(--bg-hover) !important;
  border-color: var(--border-color) !important;
}

:deep(.settings-form .n-input:focus-within) {
  background: var(--bg-card) !important;
  border-color: var(--color-primary) !important;
  box-shadow: 0 0 0 2px var(--color-primary-light);
}

:deep(.settings-form .n-input input) {
  color: var(--text-primary) !important;
}

:deep(.settings-form .n-input input::placeholder) {
  color: var(--text-muted) !important;
}

.submit-btn {
  background: var(--color-primary) !important;
  border: 1px solid var(--color-primary) !important;
  color: var(--text-inverse) !important;
  font-weight: 600;
  border-radius: var(--border-radius) !important;
  padding: 0.75rem 2rem !important;
  transition: all var(--transition-fast);
}

.submit-btn:hover:not(:disabled) {
  background: var(--color-primary-hover) !important;
  border-color: var(--color-primary-hover) !important;
  transform: translateY(-2px);
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (max-width: 768px) {
  .personal-settings {
    padding: 1rem;
  }

  .page-header {
    padding: 1.5rem;
  }

  .page-header h1 {
    font-size: 1.8rem;
  }

  .settings-card {
    padding: 1.5rem;
  }

  .card-title {
    font-size: 1.2rem;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .stat-item {
    padding: 1rem;
  }

  .stat-icon {
    font-size: 1.5rem;
    width: 40px;
    height: 40px;
  }

  .stat-value {
    font-size: 1.25rem;
  }

  .chart-container {
    height: 250px;
  }
}
</style>
