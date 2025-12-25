<template>
  <div class="usage-dashboard">
    <!-- Page Header -->
    <div class="page-header glass-card">
      <div class="header-content">
        <h1 class="gradient-text">📊 使用统计</h1>
        <p class="header-subtitle">查看您的 API 使用情况和统计数据</p>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <n-spin size="large" />
      <p>加载使用数据中...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-container glass-card">
      <div class="error-icon">⚠️</div>
      <h3>加载失败</h3>
      <p>{{ error }}</p>
      <n-button type="primary" @click="fetchData">重试</n-button>
    </div>

    <!-- Empty State -->
    <div v-else-if="isEmpty" class="empty-state glass-card">
      <div class="empty-icon">📭</div>
      <h3>暂无使用数据</h3>
      <p>您还没有任何 API 调用记录</p>
      <div class="empty-actions">
        <n-button type="primary" @click="router.push('/tokens')">
          创建 API 令牌
        </n-button>
        <n-button @click="router.push('/docs')">
          查看文档
        </n-button>
      </div>
    </div>

    <!-- Main Content -->
    <div v-else class="dashboard-content">
      <!-- Date Range Selector -->
      <div class="date-range-selector glass-card">
        <div class="selector-label">时间范围：</div>
        <n-button-group>
          <n-button
            v-for="preset in datePresets"
            :key="preset.value"
            :type="selectedPreset === preset.value ? 'primary' : 'default'"
            @click="selectDatePreset(preset.value)"
          >
            {{ preset.label }}
          </n-button>
        </n-button-group>
        <n-date-picker
          v-model:value="customDateRange"
          type="daterange"
          clearable
          @update:value="handleCustomDateChange"
          class="custom-date-picker"
        />
      </div>

      <!-- Statistics Cards -->
      <div class="stats-grid">
        <!-- Total Tokens Card -->
        <div class="stat-card glass-card">
          <div class="stat-icon primary">🎯</div>
          <div class="stat-content">
            <div class="stat-label">总 Token 数</div>
            <div class="stat-value">{{ formatNumber(stats.total_tokens) }}</div>
            <div class="stat-breakdown">
              <span>输入: {{ formatNumber(stats.prompt_tokens) }}</span>
              <span>输出: {{ formatNumber(stats.completion_tokens) }}</span>
            </div>
          </div>
        </div>

        <!-- Total Cost Card -->
        <div class="stat-card glass-card">
          <div class="stat-icon info">💰</div>
          <div class="stat-content">
            <div class="stat-label">总消费</div>
            <div class="stat-value">{{ formatCost(totalCost) }}</div>
            <div class="stat-breakdown">
              <span>$1 = 1,000,000 tokens</span>
            </div>
          </div>
        </div>

        <!-- Total Requests Card -->
        <div class="stat-card glass-card">
          <div class="stat-icon success">📡</div>
          <div class="stat-content">
            <div class="stat-label">请求次数</div>
            <div class="stat-value">{{ formatNumber(stats.total_requests) }}</div>
          </div>
        </div>

        <!-- Average Tokens Card -->
        <div class="stat-card glass-card">
          <div class="stat-icon warning">📈</div>
          <div class="stat-content">
            <div class="stat-label">平均 Token/请求</div>
            <div class="stat-value">{{ formatNumber(averageTokensPerRequest) }}</div>
          </div>
        </div>
      </div>

      <!-- Charts Section -->
      <div class="charts-section">
        <!-- Time Series Chart -->
        <div class="chart-card glass-card">
          <UsageTimeSeriesChart
            :data="trendsData"
            :loading="trendsLoading"
            @range-change="handleTrendsRangeChange"
          />
        </div>

        <!-- Model Breakdown Chart -->
        <div class="chart-card glass-card">
          <ModelBreakdownChart
            :data="stats.by_model"
            :loading="loading"
          />
        </div>
      </div>

      <!-- Recent Calls Table -->
      <div class="recent-calls glass-card">
        <h3 class="section-title">
          <span class="title-icon">📋</span>
          最近调用记录
        </h3>
        <div v-if="recentCalls.length > 0" class="calls-table-container">
          <table class="calls-table">
            <thead>
              <tr>
                <th>时间</th>
                <th>模型</th>
                <th>Token 数</th>
                <th>费用</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="call in recentCalls" :key="call.id" class="call-row">
                <td>{{ formatDateTime(call.timestamp) }}</td>
                <td>
                  <span class="model-badge">{{ call.model }}</span>
                </td>
                <td>
                  <div class="token-breakdown">
                    <span class="token-total">{{ formatNumber(call.total_tokens) }}</span>
                    <span class="token-detail">
                      ({{ call.prompt_tokens }}/{{ call.completion_tokens }})
                    </span>
                  </div>
                </td>
                <td>
                  <span class="cost-value">{{ formatCost(calculateCost(call.total_tokens)) }}</span>
                </td>
                <td>
                  <span :class="['status-badge', getStatusClass(call.status)]">
                    {{ getStatusText(call.status) }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="no-data">暂无调用记录</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NButtonGroup, NDatePicker, NSpin, useMessage } from 'naive-ui'
import { getUsageStats, getRecentCalls, getUsageTrends } from '@/api/usage'
import type { UsageStats as UsageStatsType, RecentCall } from '@/types'
import type { DailyUsage } from '@/api/usage'
import UsageTimeSeriesChart from '@/components/UsageTimeSeriesChart.vue'
import ModelBreakdownChart from '@/components/ModelBreakdownChart.vue'

const router = useRouter()
const message = useMessage()

// State
const loading = ref(true)
const trendsLoading = ref(false)
const error = ref<string | null>(null)
const selectedPreset = ref('all')  // Default to all to show all historical data
const customDateRange = ref<[number, number] | null>(null)

// Data
const stats = ref<UsageStatsType>({
  total_requests: 0,
  total_tokens: 0,
  prompt_tokens: 0,
  completion_tokens: 0,
  by_model: [],
  recent_calls: []
})

const recentCalls = ref<RecentCall[]>([])
const trendsData = ref<DailyUsage[]>([])

// Date presets
const datePresets = [
  { label: '全部', value: 'all' },
  { label: '今天', value: 'today' },
  { label: '本周', value: 'week' },
  { label: '本月', value: 'month' }
]

// Computed
const isEmpty = computed(() => {
  return !loading.value && !error.value && stats.value.total_requests === 0
})

const averageTokensPerRequest = computed(() => {
  if (stats.value.total_requests === 0) return 0
  return Math.round(stats.value.total_tokens / stats.value.total_requests)
})

// Calculate total cost based on total tokens ($1 = 1,000,000 tokens)
const totalCost = computed(() => {
  return calculateCost(stats.value.total_tokens)
})

// Methods
function selectDatePreset(preset: string) {
  selectedPreset.value = preset
  customDateRange.value = null
  fetchData()
}

function handleCustomDateChange(value: [number, number] | null) {
  if (value) {
    selectedPreset.value = ''
    fetchData()
  }
}

function handleTrendsRangeChange(range: string) {
  let days = 7
  if (range === 'day') days = 1
  else if (range === 'week') days = 7
  else if (range === 'month') days = 30
  
  fetchTrends(days)
}

async function fetchTrends(days: number) {
  trendsLoading.value = true
  try {
    const trendsResponse = await getUsageTrends({ days })
    trendsData.value = trendsResponse.trends || []
  } catch (err: any) {
    console.error('Failed to fetch trends:', err)
  } finally {
    trendsLoading.value = false
  }
}

async function fetchData() {
  loading.value = true
  error.value = null

  try {
    // Build query parameters based on selected date range
    // 'all' preset means no date filter - show all data
    const params: { start_date?: string; end_date?: string } = {}
    
    if (selectedPreset.value === 'today') {
      const today = new Date()
      params.start_date = today.toISOString().split('T')[0]
      params.end_date = today.toISOString().split('T')[0]
    } else if (selectedPreset.value === 'week') {
      const today = new Date()
      const weekAgo = new Date(today)
      weekAgo.setDate(today.getDate() - 7)
      params.start_date = weekAgo.toISOString().split('T')[0]
      params.end_date = today.toISOString().split('T')[0]
    } else if (selectedPreset.value === 'month') {
      const today = new Date()
      const monthAgo = new Date(today)
      monthAgo.setMonth(today.getMonth() - 1)
      params.start_date = monthAgo.toISOString().split('T')[0]
      params.end_date = today.toISOString().split('T')[0]
    } else if (customDateRange.value) {
      const [start, end] = customDateRange.value
      params.start_date = new Date(start).toISOString().split('T')[0]
      params.end_date = new Date(end).toISOString().split('T')[0]
    }
    // Note: 'all' preset doesn't set any date params, so all data is returned

    // Fetch usage stats and recent calls in parallel
    const [statsData, recentCallsData] = await Promise.all([
      getUsageStats(params),
      getRecentCalls({ limit: 50 })
    ])

    stats.value = statsData
    recentCalls.value = recentCallsData.calls || []
    
    // Fetch initial trends data (7 days by default)
    await fetchTrends(7)
  } catch (err: any) {
    console.error('Failed to fetch usage data:', err)
    const errorMsg = err.message || '加载数据失败'
    error.value = errorMsg
    message.error(errorMsg)
  } finally {
    loading.value = false
  }
}

function formatNumber(num: number): string {
  return num.toLocaleString('zh-CN')
}

// Calculate cost from tokens ($1 = 1,000,000 tokens)
function calculateCost(tokens: number): number {
  return tokens / 1000000
}

// Format cost as dollar amount with appropriate precision
function formatCost(cost: number): string {
  if (cost < 0.01) {
    return `$${cost.toFixed(6)}`
  } else if (cost < 1) {
    return `$${cost.toFixed(4)}`
  } else {
    return `$${cost.toFixed(2)}`
  }
}

function formatDateTime(dateStr: string): string {
  const date = new Date(dateStr)
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hour = String(date.getHours()).padStart(2, '0')
  const minute = String(date.getMinutes()).padStart(2, '0')
  return `${month}-${day} ${hour}:${minute}`
}

function getStatusClass(statusCode: number): string {
  if (statusCode >= 200 && statusCode < 300) return 'status-success'
  if (statusCode >= 400 && statusCode < 500) return 'status-warning'
  if (statusCode >= 500) return 'status-error'
  return 'status-default'
}

function getStatusText(statusCode: number): string {
  if (statusCode >= 200 && statusCode < 300) return '成功'
  if (statusCode === 401) return '未授权'
  if (statusCode === 403) return '禁止访问'
  if (statusCode === 429) return '请求过多'
  if (statusCode >= 400 && statusCode < 500) return '客户端错误'
  if (statusCode >= 500) return '服务器错误'
  return '未知'
}

// Lifecycle
onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.usage-dashboard {
  padding: var(--spacing-xl);
  animation: fadeIn var(--transition-slow) ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Page Header */
.page-header {
  padding: var(--spacing-xl);
  margin-bottom: var(--spacing-xl);
  text-align: center;
}

.header-content h1 {
  font-size: 2rem;
  font-weight: 600;
  margin: 0 0 var(--spacing-sm) 0;
  color: var(--text-primary);
}

.gradient-text {
  color: var(--text-primary) !important;
  background: none !important;
  -webkit-text-fill-color: var(--text-primary) !important;
}

.header-subtitle {
  color: var(--text-secondary);
  font-size: 1rem;
  margin: 0;
}

/* Loading State */
.loading-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: var(--spacing-2xl);
  gap: var(--spacing-md);
  color: var(--text-secondary);
}

/* Error State */
.error-container {
  padding: var(--spacing-xl);
  text-align: center;
}

.error-icon {
  font-size: 3rem;
  margin-bottom: var(--spacing-md);
}

.error-container h3 {
  color: var(--text-primary);
  font-size: 1.25rem;
  margin: 0 0 var(--spacing-md) 0;
}

.error-container p {
  color: var(--text-secondary);
  margin: 0 0 var(--spacing-xl) 0;
}

/* Empty State */
.empty-state {
  padding: var(--spacing-2xl);
  text-align: center;
}

.empty-icon {
  font-size: 4rem;
  margin-bottom: var(--spacing-lg);
}

.empty-state h3 {
  color: var(--text-primary);
  font-size: 1.5rem;
  margin: 0 0 var(--spacing-md) 0;
}

.empty-state p {
  color: var(--text-secondary);
  font-size: 1rem;
  margin: 0 0 var(--spacing-xl) 0;
}

.empty-actions {
  display: flex;
  gap: var(--spacing-md);
  justify-content: center;
}

/* Date Range Selector */
.date-range-selector {
  padding: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-wrap: wrap;
}

.selector-label {
  color: var(--text-primary);
  font-weight: 500;
  font-size: 0.9rem;
}

.custom-date-picker {
  flex: 1;
  min-width: 250px;
}

/* Statistics Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
}

.stat-card {
  padding: var(--spacing-lg);
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.stat-icon {
  font-size: 2.5rem;
  flex-shrink: 0;
  width: 56px;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--border-radius-md);
  background: var(--bg-secondary);
}

.stat-icon.primary {
  background: var(--color-primary-light);
}

.stat-icon.success {
  background: var(--color-success-light);
}

.stat-icon.warning {
  background: var(--color-warning-light);
}

.stat-icon.info {
  background: var(--color-info-light);
}

.stat-content {
  flex: 1;
}

.stat-label {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin-bottom: var(--spacing-xs);
}

.stat-value {
  color: var(--text-primary);
  font-size: 1.75rem;
  font-weight: 600;
  margin-bottom: var(--spacing-xs);
}

.stat-breakdown {
  display: flex;
  gap: var(--spacing-md);
  color: var(--text-muted);
  font-size: 0.8rem;
}

/* Charts Section */
.charts-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: var(--spacing-lg);
  margin-bottom: var(--spacing-xl);
}

.chart-card {
  padding: var(--spacing-lg);
}

/* Recent Calls Table */
.recent-calls {
  padding: var(--spacing-xl);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  color: var(--text-primary);
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 var(--spacing-lg) 0;
}

.title-icon {
  font-size: 1.25rem;
}

.calls-table-container {
  overflow-x: auto;
}

.calls-table {
  width: 100%;
  border-collapse: collapse;
}

.calls-table thead th {
  color: var(--text-primary);
  font-weight: 500;
  text-align: left;
  padding: var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-secondary);
}

.call-row {
  border-bottom: 1px solid var(--border-color-light);
  transition: background var(--transition-fast);
}

.call-row:hover {
  background: var(--bg-hover);
}

.call-row td {
  padding: var(--spacing-md);
  color: var(--text-secondary);
}

.model-badge {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  background: var(--color-primary-light);
  border-radius: var(--border-radius-sm);
  color: var(--color-primary);
  font-size: 0.8rem;
  font-weight: 500;
}

.token-breakdown {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.token-total {
  color: var(--text-primary);
  font-weight: 500;
}

.token-detail {
  color: var(--text-muted);
  font-size: 0.75rem;
}

.cost-value {
  color: var(--color-info);
  font-weight: 500;
  font-size: 0.875rem;
}

.status-badge {
  display: inline-block;
  padding: var(--spacing-xs) var(--spacing-sm);
  border-radius: var(--border-radius-sm);
  font-size: 0.8rem;
  font-weight: 500;
}

.status-success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.status-warning {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.status-error {
  background: var(--color-error-light);
  color: var(--color-error);
}

.status-default {
  background: var(--bg-tertiary);
  color: var(--text-muted);
}

.no-data {
  text-align: center;
  padding: var(--spacing-xl);
  color: var(--text-muted);
  font-size: 0.9rem;
}

/* Responsive Design */
@media (max-width: 768px) {
  .usage-dashboard {
    padding: var(--spacing-md);
    -webkit-overflow-scrolling: touch;
  }

  .page-header {
    padding: var(--spacing-lg);
  }

  .header-content h1 {
    font-size: 1.5rem;
  }

  .header-subtitle {
    font-size: 0.9rem;
  }

  .stats-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
  }

  .stat-card {
    padding: var(--spacing-md);
  }

  .stat-icon {
    font-size: 2rem;
    width: 48px;
    height: 48px;
  }

  .stat-value {
    font-size: 1.5rem;
  }

  .charts-section {
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
  }

  .chart-card {
    padding: var(--spacing-md);
  }

  .date-range-selector {
    flex-direction: column;
    align-items: stretch;
    padding: var(--spacing-md);
    gap: var(--spacing-sm);
  }

  /* Touch-friendly date preset buttons */
  .date-range-selector :deep(.n-button) {
    min-height: 44px;
    padding: var(--spacing-sm) var(--spacing-md);
  }

  .custom-date-picker {
    min-width: 100%;
  }

  .recent-calls {
    padding: var(--spacing-lg);
  }

  .section-title {
    font-size: 1.1rem;
  }

  .calls-table-container {
    margin: 0 calc(-1 * var(--spacing-md));
    padding: 0 var(--spacing-md);
  }

  .calls-table {
    font-size: 0.85rem;
  }

  .calls-table thead th,
  .call-row td {
    padding: var(--spacing-sm);
  }
}

@media (max-width: 480px) {
  .usage-dashboard {
    padding: var(--spacing-sm);
  }

  .page-header {
    padding: var(--spacing-md);
  }

  .header-content h1 {
    font-size: 1.25rem;
  }

  .header-subtitle {
    font-size: 0.85rem;
  }

  .stat-card {
    padding: var(--spacing-md);
    gap: var(--spacing-md);
  }

  .stat-icon {
    font-size: 1.75rem;
    width: 44px;
    height: 44px;
  }

  .stat-value {
    font-size: 1.25rem;
  }

  .stat-label {
    font-size: 0.8rem;
  }

  .stat-breakdown {
    flex-direction: column;
    gap: var(--spacing-xs);
    font-size: 0.75rem;
  }

  .chart-card {
    padding: var(--spacing-sm);
  }

  .date-range-selector :deep(.n-button-group) {
    width: 100%;
    display: flex;
  }

  .date-range-selector :deep(.n-button) {
    flex: 1;
  }

  .section-title {
    font-size: 1rem;
  }

  .title-icon {
    font-size: 1rem;
  }

  /* Horizontal scroll for table on small screens */
  .calls-table-container {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .calls-table {
    min-width: 500px;
    font-size: 0.8rem;
  }

  .model-badge {
    font-size: 0.7rem;
    padding: 2px var(--spacing-xs);
  }

  .status-badge {
    font-size: 0.7rem;
    padding: 2px var(--spacing-xs);
  }

  .empty-state {
    padding: var(--spacing-xl) var(--spacing-lg);
  }

  .empty-icon {
    font-size: 3rem;
  }

  .empty-state h3 {
    font-size: 1.25rem;
  }

  .empty-actions {
    flex-direction: column;
  }

  .empty-actions :deep(.n-button) {
    width: 100%;
    min-height: 44px;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .stat-card:hover {
    transform: none;
  }

  .stat-card:active {
    transform: scale(0.98);
  }

  .call-row:hover {
    background: transparent;
  }

  .call-row:active {
    background: var(--bg-hover);
  }

  /* Enable smooth scrolling for table */
  .calls-table-container {
    scroll-behavior: smooth;
  }
}
</style>
