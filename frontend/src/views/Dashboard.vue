<template>
  <div class="dashboard">
    <!-- 欢迎标题 -->
    <div class="welcome-header glass-card">
      <div class="welcome-content">
        <h1 class="gradient-text">👋 {{ getGreeting() }}，{{ authStore.user?.username }}</h1>
        <p class="welcome-subtitle">欢迎回到 Curry2API 管理平台</p>
      </div>
    </div>

    <!-- 快速统计卡片 -->
    <div class="section-title">
      <span class="title-icon">📊</span>
      <h2>快速统计</h2>
    </div>
    <div class="stats-grid">
      <StatCard
        title="今日 API 调用"
        :value="statsLoading ? '...' : todayApiCalls.toString()"
        icon="📡"
        color="primary"
      />
      <StatCard
        title="账户余额"
        :value="statsLoading ? '...' : `$${balance.toFixed(2)}`"
        icon="💰"
        color="success"
      />
      <StatCard
        title="活跃令牌"
        :value="statsLoading ? '...' : activeTokens.toString()"
        icon="🔑"
        color="warning"
      />
    </div>

    <!-- 最近活动动态 -->
    <div class="section-title">
      <span class="title-icon">⚡</span>
      <h2>最近活动</h2>
    </div>
    <div class="activity-feed glass-card">
      <div v-if="activityLoading" class="activity-loading">
        <n-spin size="small" />
        <span>加载中...</span>
      </div>
      <div v-else-if="recentCalls.length === 0" class="activity-empty">
        <div class="empty-icon">📭</div>
        <p>暂无 API 调用记录</p>
        <p class="empty-hint">开始使用 API 后，您的活动将显示在这里</p>
      </div>
      <div v-else class="activity-list">
        <div
          v-for="call in recentCalls"
          :key="call.id"
          class="activity-item"
          :class="{ 'activity-error': call.status !== 200 }"
        >
          <div class="activity-icon">
            {{ call.status === 200 ? '✅' : '❌' }}
          </div>
          <div class="activity-content">
            <div class="activity-model">{{ call.model }}</div>
            <div class="activity-details">
              <span class="activity-tokens">{{ call.total_tokens }} tokens</span>
              <span class="activity-duration">{{ call.duration_ms }}ms</span>
            </div>
          </div>
          <div class="activity-time">
            {{ formatTime(call.timestamp) }}
          </div>
        </div>
      </div>
    </div>

    <!-- 账户余额卡片 -->
    <div class="section-title">
      <span class="title-icon">💳</span>
      <h2>账户信息</h2>
    </div>
    <BalanceCard />

    <!-- 快捷操作 -->
    <div class="section-title">
      <span class="title-icon">🚀</span>
      <h2>快捷操作</h2>
    </div>
    <div class="action-cards">
      <div class="action-card glass-card" @click="router.push('/tokens')">
        <div class="action-card-inner">
          <div class="action-icon">
            <n-icon size="56">
              <KeyOutline />
            </n-icon>
          </div>
          <h3>创建 API 令牌</h3>
          <p>生成访问令牌用于 API 调用</p>
          <div class="action-arrow">→</div>
        </div>
      </div>
      <div class="action-card glass-card" @click="router.push('/docs')">
        <div class="action-card-inner">
          <div class="action-icon">
            <n-icon size="56">
              <BookOutline />
            </n-icon>
          </div>
          <h3>查看文档</h3>
          <p>了解如何使用 API</p>
          <div class="action-arrow">→</div>
        </div>
      </div>
      <div class="action-card glass-card game-card" @click="router.push('/games')">
        <div class="action-card-inner">
          <div class="action-icon game-icon">
            <n-icon size="56">
              <GameControllerOutline />
            </n-icon>
          </div>
          <h3>游戏中心</h3>
          <p>使用游戏币参与趣味游戏</p>
          <div class="action-arrow">→</div>
        </div>
      </div>
    </div>

    <!-- API 信息卡片 -->
    <div class="section-title">
      <span class="title-icon">🔗</span>
      <h2>API 信息</h2>
    </div>
    
    <!-- API 地址卡片 -->
    <div class="api-cards-grid">
      <!-- 主站API -->
      <div class="modern-api-card glass-card">
        <div class="card-badge primary-badge">推荐使用</div>
        <div class="api-icon-large">🌐</div>
        <h3 class="api-card-title">主站后端直连服务</h3>
        <div class="api-url-container" @click="copyToClipboard('https://www.kesug.icu')" title="点击复制">
          <code class="modern-api-url">https://www.kesug.icu</code>
        </div>
        <p class="api-feature">✓ 稳定可靠，连接速度快</p>
      </div>

      <!-- 备用API -->
      <div class="modern-api-card glass-card">
        <div class="card-badge warning-badge">备用</div>
        <div class="api-icon-large">🔄</div>
        <h3 class="api-card-title">备用 API 域名</h3>
        <div class="api-url-container" @click="copyToClipboard('https://api.kesug.icu')" title="点击复制">
          <code class="modern-api-url">https://api.kesug.icu</code>
        </div>
        <div class="warning-box">
          <span class="warning-icon">⚠️</span>
          <span class="warning-text">该域名可能被 DNS 污染</span>
        </div>
      </div>
    </div>

    <!-- 兼容性信息 -->
    <div class="info-banner glass-card">
      <div class="banner-icon">✨</div>
      <div class="banner-content">
        <h4>API 兼容性</h4>
        <p>支持 
          <a href="https://developers.openai.com/codex/cli" 
             target="_blank" 
             class="compatibility-link openai-link"
             @click.stop>
            <strong>OpenAI</strong>
          </a> 
          和 
          <a href="https://code.claude.com/docs/zh-CN/overview" 
             target="_blank" 
             class="compatibility-link claude-link"
             @click.stop>
            <strong>Claude Code</strong>
          </a> 
          兼容的 API 调用
        </p>
      </div>
      <div class="banner-note">
        💡 点击上方链接查看官方配置文档
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NIcon, NSpin, useMessage } from 'naive-ui'
import { KeyOutline, BookOutline, GameControllerOutline } from '@vicons/ionicons5'
import { useAuthStore } from '@/stores/auth'
import BalanceCard from '@/components/BalanceCard.vue'
import StatCard from '@/components/StatCard.vue'
import { getBalance } from '@/api/balance'
import { getUsageStats, getRecentCalls } from '@/api/usage'
import { listKeys } from '@/api/admin'
import type { RecentCall, AdminKey } from '@/types'

const router = useRouter()
const authStore = useAuthStore()
const message = useMessage()

// Stats data
const statsLoading = ref(true)
const todayApiCalls = ref(0)
const balance = ref(0)
const activeTokens = ref(0)

// Activity data
const activityLoading = ref(true)
const recentCalls = ref<RecentCall[]>([])

function getGreeting() {
  const hour = new Date().getHours()
  if (hour < 12) return '早上好'
  if (hour < 18) return '下午好'
  return '晚上好'
}

function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text)
  message.success('API 地址已复制到剪贴板')
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  // Less than 1 minute
  if (diff < 60000) {
    return '刚刚'
  }
  // Less than 1 hour
  if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return `${minutes} 分钟前`
  }
  // Less than 24 hours
  if (diff < 86400000) {
    const hours = Math.floor(diff / 3600000)
    return `${hours} 小时前`
  }
  // More than 24 hours
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}

async function loadStats() {
  statsLoading.value = true
  try {
    // Load balance
    const balanceRes = await getBalance()
    balance.value = balanceRes.data?.balance ?? 0

    // Load today's API calls
    const today = new Date()
    const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate())
    const usageRes = await getUsageStats({
      start_date: startOfDay.toISOString().split('T')[0],
      end_date: today.toISOString().split('T')[0]
    })
    todayApiCalls.value = usageRes.total_requests ?? 0

    // Load active tokens count (same logic as TokenManagement.vue)
    const tokensRes = await listKeys()
    if (tokensRes.data && tokensRes.data.keys) {
      // Filter out the default '0000' key and count only active tokens
      const tokens = tokensRes.data.keys.filter((key: AdminKey) => key.key !== '0000')
      activeTokens.value = tokens.filter((token: AdminKey) => token.is_active).length
    } else {
      activeTokens.value = 0
    }
  } catch (error) {
    console.error('Failed to load stats:', error)
  } finally {
    statsLoading.value = false
  }
}

async function loadRecentActivity() {
  activityLoading.value = true
  try {
    const res = await getRecentCalls({ limit: 5 })
    recentCalls.value = res.calls ?? []
  } catch (error) {
    console.error('Failed to load recent activity:', error)
    recentCalls.value = []
  } finally {
    activityLoading.value = false
  }
}

onMounted(() => {
  loadStats()
  loadRecentActivity()
})
</script>


<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
  max-width: 1400px;
  margin: 0 auto;
  padding: var(--spacing-xl);
}

/* 欢迎头部 - 增强视觉冲击力 */
.welcome-header {
  padding: var(--spacing-xl) var(--spacing-2xl);
  border: 1px solid var(--border-color);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.8) 0%, rgba(248, 250, 252, 0.8) 100%);
  position: relative;
  overflow: hidden;
  border-radius: var(--border-radius-lg);
}

.dark-theme .welcome-header {
  background: linear-gradient(135deg, rgba(30, 41, 59, 0.8) 0%, rgba(15, 23, 42, 0.8) 100%);
}

.welcome-header::after {
  content: '';
  position: absolute;
  top: 0;
  right: 0;
  width: 300px;
  height: 100%;
  background: radial-gradient(circle at center, var(--color-primary-light) 0%, transparent 70%);
  opacity: 0.5;
  pointer-events: none;
}

.welcome-content {
  position: relative;
  z-index: 1;
}

.welcome-content h1 {
  font-size: 2rem;
  font-weight: 800;
  margin: 0 0 var(--spacing-sm) 0;
  letter-spacing: -0.5px;
  display: flex;
  align-items: center;
  gap: 12px;
  color: var(--text-primary);
}

.welcome-subtitle {
  font-size: 1.1rem;
  margin: 0;
  opacity: 0.8;
  max-width: 600px;
  color: var(--text-secondary);
}

/* 标题区域 */
.section-title {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: var(--spacing-md);
  padding: var(--spacing-sm) var(--spacing-lg);
  background: transparent !important;
  box-shadow: none !important;
  border-radius: 0 !important;
  border-left: 4px solid var(--color-primary);
}

.title-icon {
  font-size: 1.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: var(--bg-card);
  border-radius: 12px;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
}

.section-title h2 {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0;
  letter-spacing: -0.02em;
  color: var(--text-primary);
}

/* 统计卡片网格 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: var(--spacing-lg);
}

/* 最近活动列表优化 */
.activity-feed {
  min-height: 200px;
  max-height: 400px;
  overflow-y: auto;
  padding: 0;
  border-radius: var(--border-radius-lg);
}

.activity-list {
  display: flex;
  flex-direction: column;
}

.activity-item {
  display: flex;
  align-items: center;
  padding: var(--spacing-md) var(--spacing-lg);
  border-bottom: 1px solid var(--border-color-light);
  transition: background-color var(--transition-fast);
  background: transparent !important;
}

.activity-item:last-child {
  border-bottom: none;
}

.activity-item:hover {
  background-color: var(--bg-hover) !important;
}

.activity-icon {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  background: var(--bg-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.25rem;
  margin-right: var(--spacing-md);
  flex-shrink: 0;
  border: 1px solid var(--border-color);
}

.activity-error .activity-icon {
  background: var(--color-error-light);
  border-color: rgba(239, 68, 68, 0.2);
}

.activity-content {
  flex: 1;
  min-width: 0;
}

.activity-model {
  font-weight: 600;
  color: var(--text-primary);
  font-size: 0.95rem;
  margin-bottom: 4px;
}

.activity-details {
  display: flex;
  gap: var(--spacing-md);
  font-size: 0.85rem;
  color: var(--text-secondary);
  background: transparent;
}

.activity-tokens {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  background: var(--color-primary-lighter);
  color: var(--color-primary);
  border-radius: 4px;
  font-weight: 500;
  font-size: 0.75rem;
}

.activity-duration {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  background: var(--color-success-light);
  color: var(--color-success);
  border-radius: 4px;
  font-weight: 500;
  font-size: 0.75rem;
}

.activity-time {
  font-size: 0.85rem;
  color: var(--text-muted);
  white-space: nowrap;
  margin-left: var(--spacing-md);
  font-variant-numeric: tabular-nums;
}

/* 空状态优化 */
.activity-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-2xl);
  color: var(--text-secondary);
  text-align: center;
}

.empty-icon {
  font-size: 3rem;
  margin-bottom: var(--spacing-md);
  opacity: 0.5;
  filter: grayscale(1);
}

.empty-hint {
  font-size: 0.9rem;
  color: var(--text-muted);
  margin-top: var(--spacing-xs);
}

/* 操作卡片优化 */
.action-cards {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: var(--spacing-lg);
}

.action-card {
  cursor: pointer;
  overflow: hidden;
  position: relative;
  transition: all var(--transition-normal);
  border: 1px solid var(--border-color);
  background: var(--bg-card);
  border-radius: var(--border-radius-lg);
}

.action-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
  border-color: var(--color-primary);
}

.action-card-inner {
  padding: var(--spacing-xl);
  text-align: center;
  position: relative;
  z-index: 1;
}

.action-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 72px;
  height: 72px;
  border-radius: 20px;
  margin-bottom: var(--spacing-lg);
  background: var(--bg-secondary);
  transition: all var(--transition-normal);
  color: var(--text-secondary);
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.05);
}

.action-card:hover .action-icon {
  background: var(--color-primary);
  color: #ffffff;
  transform: scale(1.1) rotate(-5deg);
  box-shadow: 0 10px 20px rgba(79, 70, 229, 0.3);
}

/* 游戏卡片特殊处理 */
.game-card:hover .game-icon {
  background: var(--color-warning);
  box-shadow: 0 10px 20px rgba(245, 158, 11, 0.3);
}

.action-card-inner h3 {
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0 0 var(--spacing-sm) 0;
  color: var(--text-primary);
}

.action-card-inner p {
  font-size: 0.95rem;
  line-height: 1.5;
  margin: 0;
  color: var(--text-secondary);
  padding: 0 var(--spacing-md);
}

.action-arrow {
  position: absolute;
  top: var(--spacing-lg);
  right: var(--spacing-lg);
  font-size: 1.5rem;
  opacity: 0;
  transition: all var(--transition-fast);
  color: var(--color-primary);
  transform: translateX(-10px);
}

.action-card:hover .action-arrow {
  opacity: 1;
  transform: translateX(0);
}

/* API 信息卡片优化 */
.api-cards-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: var(--spacing-lg);
}

.modern-api-card {
  padding: var(--spacing-xl);
  text-align: center;
  position: relative;
  overflow: hidden;
  transition: all var(--transition-normal);
  border-radius: var(--border-radius-lg);
}

.modern-api-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}

.card-badge {
  position: absolute;
  top: 16px;
  right: 16px;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.primary-badge {
  background-color: var(--color-success-light);
  color: var(--color-success);
  border: 1px solid rgba(16, 185, 129, 0.2);
}

.warning-badge {
  background-color: var(--color-warning-light);
  color: var(--color-warning);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.api-icon-large {
  font-size: 3rem;
  margin-bottom: var(--spacing-md);
  display: inline-block;
  filter: drop-shadow(0 4px 6px rgba(0, 0, 0, 0.1));
}

.api-card-title {
  font-size: 1.25rem;
  font-weight: 700;
  margin: var(--spacing-xs) 0;
  color: var(--text-primary);
}

.api-url-container {
  margin: var(--spacing-lg) 0;
  position: relative;
}

.modern-api-url {
  display: block;
  padding: 12px 16px;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 500;
  font-family: 'JetBrains Mono', monospace;
  background: var(--bg-secondary);
  color: var(--color-primary);
  border: 1px dashed var(--border-color);
  transition: all var(--transition-fast);
  cursor: pointer;
  width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.modern-api-url:hover {
  background: var(--bg-primary);
  border-color: var(--color-primary);
  border-style: solid;
  transform: scale(1.02);
}

.api-feature {
  color: var(--text-secondary);
  font-size: 0.9rem;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

/* 警告框 - 简约设计 */
.warning-box {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-xs);
  border-radius: var(--border-radius-sm);
  padding: var(--spacing-xs) var(--spacing-md);
  margin-top: var(--spacing-md);
  background: var(--color-warning-light);
  border: 1px solid var(--color-warning);
}

.warning-icon {
  font-size: 1rem;
}

.warning-text {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--color-warning);
}

/* 底部信息栏优化 */
.info-banner {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  padding: var(--spacing-lg);
  background: linear-gradient(to right, var(--bg-card), var(--bg-secondary));
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
}

.banner-icon {
  font-size: 2rem;
  flex-shrink: 0;
  background: var(--bg-primary);
  width: 60px;
  height: 60px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color-light);
}

.banner-content h4 {
  font-size: 1.1rem;
  font-weight: 700;
  margin: 0 0 4px 0;
  color: var(--text-primary);
}

.banner-content p {
  margin: 0;
  line-height: 1.6;
  color: var(--text-secondary);
}

.compatibility-link {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 4px;
  background: var(--bg-active);
  color: var(--text-primary);
  font-weight: 600;
  margin: 0 4px;
  transition: all var(--transition-fast);
}

.compatibility-link:hover {
  background: var(--color-primary);
  color: #ffffff;
}

.compatibility-link strong {
  color: inherit !important;
}

.banner-note {
  font-size: 0.85rem;
  font-weight: 500;
  padding: 6px 12px;
  border-radius: 20px;
  background: var(--color-info-light);
  color: var(--color-info-hover);
  border: 1px solid rgba(6, 182, 212, 0.2);
}

/* 响应式调整 */
@media (max-width: 768px) {
  .dashboard {
    gap: var(--spacing-lg);
    padding: var(--spacing-md);
  }

  .welcome-header {
    padding: var(--spacing-lg);
  }

  .welcome-content h1 {
    font-size: 1.5rem;
  }

  .section-title {
    padding-left: var(--spacing-md);
    border-left-width: 3px;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .api-cards-grid,
  .action-cards {
    grid-template-columns: 1fr;
  }

  .info-banner {
    flex-direction: column;
    text-align: center;
    padding: var(--spacing-xl);
  }
  
  .banner-content p {
    font-size: 0.9rem;
  }
  
  .activity-item {
    flex-wrap: wrap;
  }

  .activity-time {
    width: 100%;
    margin-top: var(--spacing-xs);
    padding-left: var(--spacing-xl);
  }
}

@media (min-width: 1400px) {
  .dashboard {
    padding: var(--spacing-xl) var(--spacing-2xl);
  }
}
</style>
