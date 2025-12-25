<template>
  <div class="referral-center">
    <div class="page-header">
      <h1 class="page-title">🎁 邀请中心</h1>
      <p class="page-subtitle">邀请好友注册，双方各得 $50 奖励</p>
    </div>

    <div class="referral-grid">
      <!-- 邀请链接卡片 -->
      <div class="glass-card referral-link-card">
        <div class="card-header">
          <div class="card-icon">🔗</div>
          <h3 class="card-title">我的邀请链接</h3>
        </div>
        
        <div v-if="loadingCode" class="card-loading">
          <n-spin size="medium" />
        </div>
        
        <div v-else-if="codeError" class="card-error">
          <p>{{ codeError }}</p>
        </div>
        
        <div v-else class="link-content">
          <div class="referral-code-display">
            <span class="code-label">邀请码</span>
            <span class="code-value">{{ referralCode?.referral_code }}</span>
          </div>
          
          <div class="referral-link-display">
            <n-input
              :value="referralCode?.referral_link"
              readonly
              class="link-input"
            />
            <n-button 
              type="primary" 
              @click="copyLink"
              class="copy-button"
            >
              {{ copied ? '✓ 已复制' : '复制链接' }}
            </n-button>
          </div>
          
          <p class="link-hint">分享此链接给好友，好友注册后双方各得 $50</p>
        </div>
      </div>

      <!-- 邀请统计卡片 -->
      <div class="glass-card stats-card">
        <div class="card-header">
          <div class="card-icon">📊</div>
          <h3 class="card-title">邀请统计</h3>
        </div>
        
        <div v-if="loadingStats" class="card-loading">
          <n-spin size="medium" />
        </div>
        
        <div v-else-if="statsError" class="card-error">
          <p>{{ statsError }}</p>
        </div>
        
        <div v-else class="stats-content">
          <div class="stat-item">
            <div class="stat-value">{{ stats?.total_referrals || 0 }}</div>
            <div class="stat-label">成功邀请</div>
          </div>
          <div class="stat-divider"></div>
          <div class="stat-item">
            <div class="stat-value">${{ formatAmount(stats?.total_bonus || 0) }}</div>
            <div class="stat-label">累计奖励</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 邀请列表 -->
    <div class="glass-card referral-list-card">
      <div class="card-header">
        <div class="card-icon">👥</div>
        <h3 class="card-title">邀请记录</h3>
      </div>
      
      <div v-if="loadingList" class="card-loading">
        <n-spin size="large" />
      </div>
      
      <div v-else-if="listError" class="card-error">
        <p>{{ listError }}</p>
      </div>
      
      <div v-else-if="referrals.length === 0" class="empty-state">
        <div class="empty-icon">📭</div>
        <p>暂无邀请记录</p>
        <p class="empty-hint">分享您的邀请链接，邀请好友一起使用</p>
      </div>
      
      <div v-else class="list-content">
        <n-data-table
          :columns="columns"
          :data="referrals"
          :pagination="pagination"
          :remote="true"
          @update:page="handlePageChange"
          class="referral-table"
        />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { NSpin, NInput, NButton, NDataTable, useMessage } from 'naive-ui'
import { 
  getReferralCode, 
  getReferralStats, 
  getReferralList,
  type ReferralCode,
  type ReferralStats,
  type ReferredUser
} from '@/api/referral'

const message = useMessage()

// Loading states
const loadingCode = ref(true)
const loadingStats = ref(true)
const loadingList = ref(true)

// Error states
const codeError = ref('')
const statsError = ref('')
const listError = ref('')

// Data
const referralCode = ref<ReferralCode | null>(null)
const stats = ref<ReferralStats | null>(null)
const referrals = ref<ReferredUser[]>([])
const total = ref(0)
const copied = ref(false)

// Pagination
const pagination = ref({
  page: 1,
  pageSize: 10,
  pageCount: 1,
  itemCount: 0
})

// Table columns
const columns = [
  {
    title: '用户名',
    key: 'username',
    width: 150
  },
  {
    title: '邮箱',
    key: 'email',
    width: 200
  },
  {
    title: '注册时间',
    key: 'registered_at',
    width: 180,
    render: (row: ReferredUser) => formatDate(row.registered_at)
  },
  {
    title: '奖励金额',
    key: 'bonus_amount',
    width: 120,
    render: (row: ReferredUser) => h('span', { class: 'bonus-amount' }, `+$${formatAmount(row.bonus_amount)}`)
  }
]

function formatAmount(value: number): string {
  return value.toFixed(2)
}

function formatDate(dateStr: string): string {
  const date = new Date(dateStr)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

async function copyLink() {
  if (!referralCode.value?.referral_link) return
  
  try {
    await navigator.clipboard.writeText(referralCode.value.referral_link)
    copied.value = true
    message.success('邀请链接已复制到剪贴板')
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    message.error('复制失败，请手动复制')
  }
}

async function loadReferralCode() {
  try {
    loadingCode.value = true
    codeError.value = ''
    const response = await getReferralCode()
    referralCode.value = response.data
  } catch (err: any) {
    codeError.value = err.message || '加载邀请码失败'
  } finally {
    loadingCode.value = false
  }
}

async function loadStats() {
  try {
    loadingStats.value = true
    statsError.value = ''
    const response = await getReferralStats()
    stats.value = response.data
  } catch (err: any) {
    statsError.value = err.message || '加载统计数据失败'
  } finally {
    loadingStats.value = false
  }
}

async function loadReferralList(page = 1) {
  try {
    loadingList.value = true
    listError.value = ''
    const offset = (page - 1) * pagination.value.pageSize
    const response = await getReferralList(pagination.value.pageSize, offset)
    referrals.value = response.data.referrals
    total.value = response.data.total
    pagination.value.itemCount = response.data.total
    pagination.value.pageCount = Math.ceil(response.data.total / pagination.value.pageSize)
    pagination.value.page = page
  } catch (err: any) {
    listError.value = err.message || '加载邀请列表失败'
  } finally {
    loadingList.value = false
  }
}

function handlePageChange(page: number) {
  loadReferralList(page)
}

onMounted(() => {
  loadReferralCode()
  loadStats()
  loadReferralList()
})
</script>


<style scoped>
.referral-center {
  padding: 2rem;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  text-align: center;
  margin-bottom: 2rem;
}

.page-title {
  color: var(--text-primary);
  font-size: 2rem;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
}

.page-subtitle {
  color: var(--text-secondary);
  font-size: 1.1rem;
  margin: 0;
}

.referral-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 1.5rem;
  margin-bottom: 1.5rem;
}

.glass-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-md);
  padding: 1.5rem;
  transition: all var(--transition-normal);
}

.glass-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.card-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.card-icon {
  font-size: 1.5rem;
}

.card-title {
  color: var(--text-primary);
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0;
}

.card-loading,
.card-error {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100px;
}

.card-error p {
  color: var(--color-error);
  margin: 0;
}

/* Referral Link Card */
.link-content {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.referral-code-display {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
  border: 1px solid var(--border-color);
}

.code-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.code-value {
  color: var(--color-primary);
  font-size: 1.5rem;
  font-weight: 700;
  font-family: monospace;
  letter-spacing: 2px;
}

.referral-link-display {
  display: flex;
  gap: 0.75rem;
}

.link-input {
  flex: 1;
}

:deep(.link-input .n-input__input-el) {
  font-family: monospace;
  font-size: 0.9rem;
}

.copy-button {
  flex-shrink: 0;
}

.link-hint {
  color: var(--text-muted);
  font-size: 0.85rem;
  margin: 0;
  text-align: center;
}

/* Stats Card */
.stats-content {
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding: 1rem;
}

.stat-item {
  text-align: center;
  flex: 1;
}

.stat-value {
  color: var(--text-primary);
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
}

.stat-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.stat-divider {
  width: 1px;
  height: 60px;
  background: var(--border-color);
}

/* Referral List Card */
.referral-list-card {
  min-height: 300px;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem;
  text-align: center;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.empty-state p {
  color: var(--text-secondary);
  margin: 0;
  font-size: 1.1rem;
}

.empty-hint {
  color: var(--text-muted) !important;
  font-size: 0.9rem !important;
  margin-top: 0.5rem !important;
}

.list-content {
  overflow-x: auto;
}

:deep(.referral-table) {
  background: transparent !important;
}

:deep(.referral-table .n-data-table-th) {
  background: var(--bg-secondary) !important;
  color: var(--text-primary) !important;
  border-bottom: 1px solid var(--border-color) !important;
}

:deep(.referral-table .n-data-table-td) {
  background: transparent !important;
  color: var(--text-secondary) !important;
  border-bottom: 1px solid var(--border-color-light) !important;
}

:deep(.referral-table .n-data-table-tr:hover .n-data-table-td) {
  background: var(--bg-hover) !important;
}

:deep(.bonus-amount) {
  color: var(--color-success);
  font-weight: 600;
}

/* Responsive */
@media (max-width: 768px) {
  .referral-center {
    padding: 1rem;
  }
  
  .page-title {
    font-size: 1.5rem;
  }
  
  .referral-grid {
    grid-template-columns: 1fr;
  }
  
  .referral-link-display {
    flex-direction: column;
  }
  
  .stats-content {
    flex-direction: column;
    gap: 1rem;
  }
  
  .stat-divider {
    width: 100%;
    height: 1px;
  }
}
</style>
