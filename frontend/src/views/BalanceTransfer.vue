<template>
  <div class="balance-transfer">
    <!-- Balance Overview Section -->
    <div class="balance-overview">
      <div class="balance-card glass-card game-balance" :class="{ active: direction === 'toUSD' }">
        <div class="balance-icon">🎮</div>
        <div class="balance-info">
          <span class="balance-label">游戏币余额</span>
          <span class="balance-value">{{ formatNumber(gameBalance) }}</span>
        </div>
      </div>
      
      <div class="transfer-arrow" @click="toggleDirection">
        <span class="arrow-icon" :class="{ reversed: direction === 'toGame' }">⇄</span>
        <span class="arrow-hint">点击切换</span>
      </div>
      
      <div class="balance-card glass-card account-balance" :class="{ active: direction === 'toGame' }">
        <div class="balance-icon">💰</div>
        <div class="balance-info">
          <span class="balance-label">账户余额</span>
          <span class="balance-value">${{ formatNumber(accountBalance) }}</span>
        </div>
      </div>
    </div>

    <!-- Exchange Form Section -->
    <div class="exchange-section glass-card">
      <div class="section-header">
        <h2 class="section-title">💱 余额划转</h2>
        <div class="direction-tabs">
          <n-button 
            :type="direction === 'toUSD' ? 'primary' : 'default'"
            size="small"
            @click="direction = 'toUSD'"
          >
            🎮 → 💰 兑换余额
          </n-button>
          <n-button 
            :type="direction === 'toGame' ? 'primary' : 'default'"
            size="small"
            @click="direction = 'toGame'"
          >
            💰 → 🎮 购买游戏币
          </n-button>
        </div>
      </div>

      <div class="exchange-rate-badge">
        <span class="rate-label">兑换比例</span>
        <span class="rate-value">1 游戏币 = $1 USD</span>
      </div>

      <!-- Daily Limit Info (only for toUSD direction) -->
      <div v-if="direction === 'toUSD'" class="limit-info">
        <div class="limit-item">
          <span class="limit-label">今日已兑换</span>
          <span class="limit-value">{{ formatNumber(todayExchanged) }} / {{ formatNumber(dailyLimit) }}</span>
        </div>
        <div class="limit-progress">
          <div class="progress-bar" :style="{ width: limitPercentage + '%' }"></div>
        </div>
        <div class="limit-remaining">
          剩余可兑换: <strong>{{ formatNumber(remainingLimit) }}</strong> 游戏币
        </div>
      </div>

      <!-- Exchange Form -->
      <div class="exchange-form">
        <div class="form-group">
          <label class="form-label">{{ direction === 'toUSD' ? '兑换数量' : '购买数量' }}</label>
          <n-input-number
            v-model:value="exchangeAmount"
            :min="1"
            :max="maxExchangeAmount"
            :precision="2"
            :placeholder="direction === 'toUSD' ? '请输入兑换数量' : '请输入购买金额'"
            size="large"
            class="amount-input"
            :disabled="loading"
          >
            <template #suffix>
              <span class="input-suffix">{{ direction === 'toUSD' ? '游戏币' : 'USD' }}</span>
            </template>
          </n-input-number>
          <div class="quick-amounts">
            <n-button 
              v-for="amount in quickAmounts" 
              :key="amount"
              size="small"
              :disabled="amount > maxExchangeAmount || loading"
              @click="exchangeAmount = amount"
            >
              {{ direction === 'toUSD' ? amount : '$' + amount }}
            </n-button>
            <n-button 
              size="small"
              :disabled="maxExchangeAmount <= 0 || loading"
              @click="exchangeAmount = maxExchangeAmount"
            >
              全部
            </n-button>
          </div>
        </div>

        <!-- Preview -->
        <div class="exchange-preview" v-if="exchangeAmount && exchangeAmount > 0">
          <template v-if="direction === 'toUSD'">
            <div class="preview-row">
              <span class="preview-label">扣除游戏币</span>
              <span class="preview-value deduct">-{{ formatNumber(exchangeAmount) }}</span>
            </div>
            <div class="preview-row">
              <span class="preview-label">获得账户余额</span>
              <span class="preview-value add">+${{ formatNumber(exchangeAmount) }}</span>
            </div>
          </template>
          <template v-else>
            <div class="preview-row">
              <span class="preview-label">扣除账户余额</span>
              <span class="preview-value deduct">-${{ formatNumber(exchangeAmount) }}</span>
            </div>
            <div class="preview-row">
              <span class="preview-label">获得游戏币</span>
              <span class="preview-value add">+{{ formatNumber(exchangeAmount) }}</span>
            </div>
          </template>
        </div>

        <!-- Exchange Button -->
        <n-button
          type="primary"
          size="large"
          block
          :loading="loading"
          :disabled="!canExchange"
          @click="showConfirmDialog = true"
        >
          {{ loading ? '处理中...' : (direction === 'toUSD' ? '确认兑换' : '确认购买') }}
        </n-button>

        <!-- Validation Messages -->
        <div v-if="validationError" class="validation-error">
          <span class="error-icon">⚠️</span>
          {{ validationError }}
        </div>
      </div>
    </div>

    <!-- Exchange History Section -->
    <div class="history-section glass-card">
      <div class="section-header">
        <h2 class="section-title">📜 兑换记录</h2>
      </div>

      <div v-if="historyLoading" class="history-loading">
        <n-spin size="medium" />
        <span>加载中...</span>
      </div>

      <div v-else-if="exchangeHistory.length === 0" class="history-empty">
        <span class="empty-icon">📭</span>
        <p>暂无兑换记录</p>
      </div>

      <div v-else class="history-list">
        <div 
          v-for="record in exchangeHistory" 
          :key="record.id" 
          class="history-item"
        >
          <div class="history-info">
            <div class="history-amount">
              <span class="game-coins">{{ formatNumber(record.game_coins_amount) }} 游戏币</span>
              <span class="arrow">→</span>
              <span class="usd-amount">${{ formatNumber(record.usd_amount) }}</span>
            </div>
            <div class="history-time">{{ formatTime(record.created_at) }}</div>
          </div>
          <div class="history-status" :class="record.status">
            {{ record.status === 'completed' ? '成功' : '失败' }}
          </div>
        </div>
      </div>

      <!-- Pagination -->
      <div v-if="historyTotal > pageSize" class="pagination-container">
        <n-pagination
          v-model:page="currentPage"
          :page-size="pageSize"
          :item-count="historyTotal"
          @update:page="loadExchangeHistory"
        />
      </div>
    </div>

    <!-- Confirmation Dialog -->
    <n-modal v-model:show="showConfirmDialog" preset="dialog" :title="direction === 'toUSD' ? '确认兑换' : '确认购买'">
      <template #icon>
        <span style="font-size: 1.5rem;">💱</span>
      </template>
      <div class="confirm-content">
        <p>您确定要进行以下{{ direction === 'toUSD' ? '兑换' : '购买' }}吗？</p>
        <div class="confirm-details">
          <template v-if="direction === 'toUSD'">
            <div class="confirm-row">
              <span>扣除游戏币:</span>
              <strong class="deduct">{{ formatNumber(exchangeAmount || 0) }}</strong>
            </div>
            <div class="confirm-row">
              <span>获得账户余额:</span>
              <strong class="add">${{ formatNumber(exchangeAmount || 0) }}</strong>
            </div>
          </template>
          <template v-else>
            <div class="confirm-row">
              <span>扣除账户余额:</span>
              <strong class="deduct">${{ formatNumber(exchangeAmount || 0) }}</strong>
            </div>
            <div class="confirm-row">
              <span>获得游戏币:</span>
              <strong class="add">{{ formatNumber(exchangeAmount || 0) }}</strong>
            </div>
          </template>
        </div>
        <p class="confirm-note">此操作不可撤销</p>
      </div>
      <template #action>
        <n-button @click="showConfirmDialog = false">取消</n-button>
        <n-button type="primary" :loading="loading" @click="handleExchange">
          {{ direction === 'toUSD' ? '确认兑换' : '确认购买' }}
        </n-button>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useMessage } from 'naive-ui'
import { useGameStore } from '@/stores/game'
import { getBalance } from '@/api/balance'
import {
  exchangeGameCoins,
  purchaseGameCoins,
  getExchangeHistory,
  getTodayExchangeAmount,
  type ExchangeRecord
} from '@/api/gameCoin'

const message = useMessage()
const gameStore = useGameStore()

// Balance states
const accountBalance = ref(0)
const gameBalance = computed(() => gameStore.gameCoins)

// Exchange form states
const exchangeAmount = ref<number | null>(null)
const loading = ref(false)
const showConfirmDialog = ref(false)
const direction = ref<'toUSD' | 'toGame'>('toUSD') // 兑换方向

// Daily limit states
const todayExchanged = ref(0)
const dailyLimit = ref(1000)
const remainingLimit = computed(() => Math.max(0, dailyLimit.value - todayExchanged.value))

// Toggle direction
function toggleDirection() {
  direction.value = direction.value === 'toUSD' ? 'toGame' : 'toUSD'
  exchangeAmount.value = null // Reset amount when switching direction
}

// History states
const exchangeHistory = ref<ExchangeRecord[]>([])
const historyLoading = ref(false)
const historyTotal = ref(0)
const currentPage = ref(1)
const pageSize = 10

// Quick amount buttons
const quickAmounts = [10, 50, 100, 500]

// Computed properties
const maxExchangeAmount = computed(() => {
  if (direction.value === 'toUSD') {
    return Math.min(gameBalance.value, remainingLimit.value)
  } else {
    return accountBalance.value
  }
})

const limitPercentage = computed(() => {
  if (dailyLimit.value === 0) return 0
  return Math.min(100, (todayExchanged.value / dailyLimit.value) * 100)
})

const validationError = computed(() => {
  if (!exchangeAmount.value || exchangeAmount.value <= 0) return null
  if (exchangeAmount.value < 1) return direction.value === 'toUSD' ? '最小兑换金额为 1 游戏币' : '最小购买金额为 $1'
  
  if (direction.value === 'toUSD') {
    if (exchangeAmount.value > gameBalance.value) return '游戏币余额不足'
    if (exchangeAmount.value > remainingLimit.value) return '超过今日兑换限额'
  } else {
    if (exchangeAmount.value > accountBalance.value) return '账户余额不足'
  }
  return null
})

const canExchange = computed(() => {
  return exchangeAmount.value && 
         exchangeAmount.value >= 1 && 
         exchangeAmount.value <= maxExchangeAmount.value &&
         !validationError.value
})

// Format helpers
function formatNumber(value: number | null | undefined): string {
  if (value === null || value === undefined) return '0.00'
  return value.toFixed(2)
}

function formatTime(timestamp: string): string {
  const date = new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Load account balance
async function loadAccountBalance() {
  try {
    const response = await getBalance()
    accountBalance.value = response.data?.balance ?? 0
  } catch (err) {
    console.error('Failed to load account balance:', err)
  }
}

// Load today's exchange amount
async function loadTodayExchange() {
  try {
    const response = await getTodayExchangeAmount()
    todayExchanged.value = response.data.amount
    dailyLimit.value = response.data.limit
  } catch (err) {
    console.error('Failed to load today exchange amount:', err)
  }
}

// Load exchange history
async function loadExchangeHistory(page: number = 1) {
  historyLoading.value = true
  try {
    const offset = (page - 1) * pageSize
    const response = await getExchangeHistory(pageSize, offset)
    exchangeHistory.value = response.data.records || []
    historyTotal.value = response.data.total
    currentPage.value = page
  } catch (err) {
    console.error('Failed to load exchange history:', err)
    exchangeHistory.value = []
  } finally {
    historyLoading.value = false
  }
}

// Handle exchange
async function handleExchange() {
  if (!canExchange.value || !exchangeAmount.value) return

  loading.value = true
  try {
    let response
    if (direction.value === 'toUSD') {
      response = await exchangeGameCoins(exchangeAmount.value)
    } else {
      response = await purchaseGameCoins(exchangeAmount.value)
    }
    
    if (response.data.success) {
      message.success(direction.value === 'toUSD' ? '兑换成功！' : '购买成功！')
      
      // Update balances
      accountBalance.value = response.data.new_account_balance
      await gameStore.loadBalance()
      
      // Refresh today's exchange and history
      if (direction.value === 'toUSD') {
        await loadTodayExchange()
      }
      await loadExchangeHistory(1)
      
      // Reset form
      exchangeAmount.value = null
      showConfirmDialog.value = false
    }
  } catch (err: any) {
    const errorMessage = err.response?.data?.error?.message || (direction.value === 'toUSD' ? '兑换失败，请稍后重试' : '购买失败，请稍后重试')
    message.error(errorMessage)
  } finally {
    loading.value = false
    showConfirmDialog.value = false
  }
}

// Watch for game balance changes
watch(() => gameStore.gameCoins, () => {
  // Ensure exchange amount doesn't exceed available balance
  if (exchangeAmount.value && exchangeAmount.value > maxExchangeAmount.value) {
    exchangeAmount.value = maxExchangeAmount.value > 0 ? maxExchangeAmount.value : null
  }
})

// Initialize
onMounted(async () => {
  await Promise.all([
    gameStore.loadBalance(),
    loadAccountBalance(),
    loadTodayExchange(),
    loadExchangeHistory()
  ])
})
</script>


<style scoped>
.balance-transfer {
  padding: 2rem;
  max-width: 900px;
  margin: 0 auto;
}

/* Balance Overview */
.balance-overview {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  margin-bottom: 2rem;
}

.balance-card {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1.5rem 2rem;
  flex: 1;
  max-width: 320px;
}

.balance-icon {
  font-size: 2.5rem;
}

.balance-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.balance-label {
  color: var(--text-secondary);
  font-size: 0.85rem;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.balance-value {
  color: var(--text-primary);
  font-size: 1.75rem;
  font-weight: 700;
}

.game-balance .balance-value {
  color: var(--color-primary);
}

.account-balance .balance-value {
  color: var(--color-success);
}

.transfer-arrow {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 60px;
  height: 60px;
  background: var(--color-primary-light);
  border-radius: 50%;
  flex-shrink: 0;
  cursor: pointer;
  transition: all var(--transition-fast);
}

.transfer-arrow:hover {
  background: var(--bg-hover);
  transform: scale(1.1);
}

.arrow-icon {
  font-size: 1.5rem;
  color: var(--color-primary);
  transition: transform var(--transition-fast);
}

.arrow-icon.reversed {
  transform: rotate(180deg);
}

.arrow-hint {
  font-size: 0.6rem;
  color: var(--text-muted);
  margin-top: 2px;
}

.balance-card.active {
  border-color: var(--color-success);
}

.direction-tabs {
  display: flex;
  gap: 0.5rem;
}

/* Exchange Section */
.exchange-section {
  padding: 2rem;
  margin-bottom: 2rem;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.section-title {
  color: var(--text-primary);
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0;
}

.exchange-rate-badge {
  margin-bottom: 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: var(--color-success-light);
  border: 1px solid var(--color-success);
  border-radius: 20px;
}

.rate-label {
  color: var(--text-secondary);
  font-size: 0.8rem;
}

.rate-value {
  color: var(--color-success);
  font-weight: 600;
  font-size: 0.9rem;
}

/* Limit Info */
.limit-info {
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
  padding: 1rem 1.5rem;
  margin-bottom: 1.5rem;
}

.limit-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
}

.limit-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.limit-value {
  color: var(--text-primary);
  font-weight: 600;
}

.limit-progress {
  height: 6px;
  background: var(--bg-tertiary);
  border-radius: 3px;
  overflow: hidden;
  margin-bottom: 0.75rem;
}

.progress-bar {
  height: 100%;
  background: var(--color-success);
  border-radius: 3px;
  transition: width var(--transition-fast);
}

.limit-remaining {
  color: var(--text-secondary);
  font-size: 0.85rem;
  text-align: right;
}

.limit-remaining strong {
  color: var(--color-success);
}

/* Exchange Form */
.exchange-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.form-label {
  color: var(--text-primary);
  font-size: 0.9rem;
  font-weight: 600;
}

.amount-input {
  width: 100%;
}

.input-suffix {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.quick-amounts {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
}

/* Exchange Preview */
.exchange-preview {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius);
  padding: 1rem 1.5rem;
}

.preview-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
}

.preview-row:not(:last-child) {
  border-bottom: 1px solid var(--border-color);
}

.preview-label {
  color: var(--text-secondary);
}

.preview-value {
  font-weight: 700;
  font-size: 1.1rem;
}

.preview-value.deduct {
  color: var(--color-error);
}

.preview-value.add {
  color: var(--color-success);
}

/* Validation Error */
.validation-error {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  background: var(--color-error-light);
  border: 1px solid var(--color-error);
  border-radius: var(--border-radius);
  color: var(--color-error);
  font-size: 0.9rem;
}

.error-icon {
  font-size: 1rem;
}

/* History Section */
.history-section {
  padding: 2rem;
}

.history-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 3rem;
  color: var(--text-secondary);
}

.history-empty {
  text-align: center;
  padding: 3rem;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.history-empty p {
  margin: 0;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.history-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  background: var(--bg-secondary);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.history-item:hover {
  background: var(--bg-hover);
  transform: translateX(4px);
}

.history-info {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.history-amount {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-weight: 600;
}

.game-coins {
  color: var(--color-primary);
}

.arrow {
  color: var(--text-muted);
}

.usd-amount {
  color: var(--color-success);
}

.history-time {
  color: var(--text-muted);
  font-size: 0.8rem;
}

.history-status {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
}

.history-status.completed {
  background: var(--color-success-light);
  color: var(--color-success);
}

.history-status.failed {
  background: var(--color-error-light);
  color: var(--color-error);
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 1.5rem;
}

/* Confirmation Dialog */
.confirm-content {
  padding: 1rem 0;
}

.confirm-content p {
  margin: 0 0 1rem 0;
  color: var(--text-secondary);
}

.confirm-details {
  background: var(--bg-secondary);
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.confirm-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
}

.confirm-row span {
  color: var(--text-secondary);
}

.confirm-row strong.deduct {
  color: var(--color-error);
}

.confirm-row strong.add {
  color: var(--color-success);
}

.confirm-note {
  font-size: 0.85rem;
  color: var(--text-muted);
  text-align: center;
}

/* Glass Card */
.glass-card {
  background: var(--bg-card);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  box-shadow: var(--shadow-md);
}

/* Responsive Design */
@media (max-width: 768px) {
  .balance-transfer {
    padding: 1rem;
  }

  .balance-overview {
    flex-direction: column;
    gap: 1rem;
  }

  .balance-card {
    max-width: 100%;
    width: 100%;
  }

  .transfer-arrow {
    transform: rotate(90deg);
  }

  .section-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .exchange-section,
  .history-section {
    padding: 1.5rem;
  }

  .history-item {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .history-status {
    align-self: flex-end;
  }

  .quick-amounts {
    justify-content: flex-start;
  }
}

@media (max-width: 480px) {
  .balance-value {
    font-size: 1.5rem;
  }

  .section-title {
    font-size: 1.25rem;
  }

  .exchange-rate-badge {
    flex-direction: column;
    gap: 0.25rem;
    text-align: center;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .history-item:hover {
    transform: none;
    background: rgba(255, 255, 255, 0.05);
  }

  .history-item:active {
    background: rgba(255, 255, 255, 0.1);
  }
}
</style>
