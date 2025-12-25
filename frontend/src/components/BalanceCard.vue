<template>
  <div class="balance-card glass-card" :class="balanceStatusClass">
    <div class="balance-header">
      <div class="balance-icon">💰</div>
      <h3 class="balance-title">账户余额</h3>
      <div class="status-indicator" :class="statusIndicatorClass">
        <span class="status-dot"></span>
        <span class="status-text">{{ statusText }}</span>
      </div>
    </div>
    
    <div v-if="loading" class="balance-loading">
      <n-spin size="large" />
    </div>
    
    <div v-else-if="error" class="balance-error">
      <div class="error-icon">⚠️</div>
      <p>{{ error }}</p>
    </div>
    
    <div v-else class="balance-content">
      <div class="balance-amount-container">
        <div class="balance-label">当前余额</div>
        <div class="balance-amount">
          <span class="currency">$</span>
          <span class="amount">{{ formatBalance(balance?.balance || 0) }}</span>
        </div>
      </div>
      
      <!-- 低余额警告 -->
      <div v-if="isLowBalance" class="balance-warning low-balance">
        <div class="warning-icon">⚠️</div>
        <div class="warning-content">
          <strong>余额不足</strong>
          <p>您的余额低于 $10，请及时充值以继续使用服务</p>
        </div>
      </div>
      
      <!-- 余额耗尽警告 -->
      <div v-if="isExhausted" class="balance-warning exhausted">
        <div class="warning-icon">🚫</div>
        <div class="warning-content">
          <strong>余额已耗尽</strong>
          <p>您的所有 API 令牌已被禁用，请充值后继续使用</p>
        </div>
      </div>
      
      <!-- 余额统计 -->
      <div class="balance-stats">
        <div class="stat-item">
          <div class="stat-label">累计消费</div>
          <div class="stat-value">${{ formatBalance(balance?.total_consumed || 0) }}</div>
        </div>
        <div class="stat-divider"></div>
        <div class="stat-item">
          <div class="stat-label">累计充值</div>
          <div class="stat-value">${{ formatBalance(balance?.total_recharged || 0) }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NSpin, useMessage } from 'naive-ui'
import { getBalance, type UserBalance } from '@/api/balance'

const message = useMessage()
const loading = ref(true)
const error = ref('')
const balance = ref<UserBalance | null>(null)

const isLowBalance = computed(() => {
  if (!balance.value) return false
  return balance.value.balance > 0 && balance.value.balance < 10
})

const isExhausted = computed(() => {
  if (!balance.value) return false
  return balance.value.balance <= 0 || balance.value.status === 'exhausted'
})

const balanceStatusClass = computed(() => {
  if (isExhausted.value) return 'status-exhausted'
  if (isLowBalance.value) return 'status-low'
  return 'status-normal'
})

const statusIndicatorClass = computed(() => {
  if (isExhausted.value) return 'indicator-error'
  if (isLowBalance.value) return 'indicator-warning'
  return 'indicator-success'
})

const statusText = computed(() => {
  if (isExhausted.value) return '已耗尽'
  if (isLowBalance.value) return '余额不足'
  return '正常'
})

function formatBalance(value: number): string {
  return value.toFixed(2)
}

async function loadBalance() {
  try {
    loading.value = true
    error.value = ''
    const response = await getBalance()
    balance.value = response.data
  } catch (err: any) {
    error.value = err.response?.data?.error?.message || '加载余额失败'
    message.error(error.value)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadBalance()
})

// 暴露刷新方法供父组件调用
defineExpose({
  refresh: loadBalance
})
</script>

<style scoped>
.balance-card {
  padding: 1.5rem;
}

.balance-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.balance-icon {
  font-size: 1.5rem;
}

.balance-title {
  color: var(--text-primary);
  font-size: 1.125rem;
  font-weight: 600;
  margin: 0;
  flex: 1;
}

/* 简化的状态指示器 */
.status-indicator {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.625rem;
  border-radius: var(--border-radius-sm);
  font-size: 0.75rem;
  font-weight: 500;
}

.status-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
}

.indicator-success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.indicator-success .status-dot {
  background: var(--color-success);
}

.indicator-warning {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.indicator-warning .status-dot {
  background: var(--color-warning);
}

.indicator-error {
  background: var(--color-error-light);
  color: var(--color-error);
}

.indicator-error .status-dot {
  background: var(--color-error);
}

.balance-loading,
.balance-error {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  min-height: 180px;
}

.error-icon {
  font-size: 2.5rem;
  margin-bottom: 1rem;
}

.balance-error p {
  color: var(--text-secondary);
  margin: 0;
}

.balance-content {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.balance-amount-container {
  text-align: center;
  padding: 1.25rem;
  background: var(--bg-secondary);
  border-radius: var(--border-radius-md);
  border: 1px solid var(--border-color);
}

.balance-label {
  color: var(--text-muted);
  font-size: 0.875rem;
  margin-bottom: 0.5rem;
}

.balance-amount {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 0.25rem;
}

.currency {
  color: var(--color-primary);
  font-size: 1.5rem;
  font-weight: 600;
}

.amount {
  color: var(--text-primary);
  font-size: 2.5rem;
  font-weight: 700;
}

.balance-warning {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem;
  border-radius: var(--border-radius);
}

.low-balance {
  background: var(--color-warning-light);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.exhausted {
  background: var(--color-error-light);
  border: 1px solid rgba(239, 68, 68, 0.2);
}

.warning-icon {
  font-size: 1.25rem;
  flex-shrink: 0;
}

.warning-content {
  flex: 1;
}

.warning-content strong {
  display: block;
  font-size: 0.875rem;
  margin-bottom: 0.25rem;
}

.low-balance .warning-content strong {
  color: var(--color-warning);
}

.exhausted .warning-content strong {
  color: var(--color-error);
}

.warning-content p {
  color: var(--text-secondary);
  font-size: 0.8125rem;
  margin: 0;
  line-height: 1.5;
}

.balance-stats {
  display: flex;
  align-items: center;
  justify-content: space-around;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
  border: 1px solid var(--border-color);
}

.stat-item {
  text-align: center;
  flex: 1;
}

.stat-label {
  color: var(--text-muted);
  font-size: 0.8125rem;
  margin-bottom: 0.375rem;
}

.stat-value {
  color: var(--text-primary);
  font-size: 1.125rem;
  font-weight: 600;
}

.stat-divider {
  width: 1px;
  height: 36px;
  background: var(--border-color);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .balance-card {
    padding: 1.25rem;
  }
  
  .balance-amount-container {
    padding: 1rem;
  }
  
  .amount {
    font-size: 2rem;
  }
  
  .currency {
    font-size: 1.25rem;
  }
  
  .balance-stats {
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .stat-divider {
    width: 100%;
    height: 1px;
  }
}
</style>
