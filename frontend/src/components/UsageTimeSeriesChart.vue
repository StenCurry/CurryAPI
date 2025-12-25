<template>
  <div class="time-series-chart">
    <div class="chart-header">
      <h4 class="chart-title">
        <span class="title-icon">📈</span>
        {{ title }}
      </h4>
      <div class="chart-controls">
        <n-button-group size="small">
          <n-button
            v-for="range in timeRanges"
            :key="range.value"
            :type="selectedRange === range.value ? 'primary' : 'default'"
            @click="selectRange(range.value)"
          >
            {{ range.label }}
          </n-button>
        </n-button-group>
      </div>
    </div>
    
    <div v-if="loading" class="chart-loading">
      <n-spin size="medium" />
      <span>加载图表数据...</span>
    </div>
    
    <div v-else-if="!hasData" class="chart-empty">
      <span class="empty-icon">📊</span>
      <span>暂无数据</span>
    </div>
    
    <div v-else class="chart-wrapper">
      <!-- 统计摘要 -->
      <div class="chart-summary">
        <div class="summary-item total">
          <span class="summary-label">总计</span>
          <span class="summary-value">{{ formatLargeNumber(totalTokens) }}</span>
        </div>
        <div class="summary-item input">
          <span class="summary-dot input-dot"></span>
          <span class="summary-label">输入</span>
          <span class="summary-value">{{ formatLargeNumber(totalPromptTokens) }}</span>
        </div>
        <div class="summary-item output">
          <span class="summary-dot output-dot"></span>
          <span class="summary-label">输出</span>
          <span class="summary-value">{{ formatLargeNumber(totalCompletionTokens) }}</span>
        </div>
      </div>
      
      <div class="chart-container">
        <Line :data="chartData" :options="chartOptions" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NButton, NButtonGroup, NSpin } from 'naive-ui'
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

interface DailyUsage {
  date: string
  total_tokens: number
  prompt_tokens: number
  completion_tokens: number
  request_count: number
}

interface Props {
  title?: string
  data?: DailyUsage[]
  loading?: boolean
  showPromptCompletion?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Token 使用趋势',
  data: () => [],
  loading: false,
  showPromptCompletion: true,
})

const emit = defineEmits<{
  (e: 'rangeChange', range: string): void
}>()

const selectedRange = ref('week')

const timeRanges = [
  { label: '日', value: 'day' },
  { label: '周', value: 'week' },
  { label: '月', value: 'month' },
]

const hasData = computed(() => props.data && props.data.length > 0)

// 计算总计
const totalTokens = computed(() => props.data.reduce((sum, d) => sum + d.total_tokens, 0))
const totalPromptTokens = computed(() => props.data.reduce((sum, d) => sum + d.prompt_tokens, 0))
const totalCompletionTokens = computed(() => props.data.reduce((sum, d) => sum + d.completion_tokens, 0))

const chartData = computed(() => {
  if (!hasData.value) {
    return { labels: [], datasets: [] }
  }

  const labels = props.data.map(d => formatDateLabel(d.date))
  
  const datasets: any[] = []

  if (props.showPromptCompletion) {
    // 堆叠面积图：输入和输出 Token
    datasets.push(
      {
        label: '输入 Token',
        data: props.data.map(d => d.prompt_tokens),
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
        order: 2,
      },
      {
        label: '输出 Token',
        data: props.data.map(d => d.completion_tokens),
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
        order: 1,
      }
    )
  } else {
    // 单线图：总 Token
    datasets.push({
      label: '总 Token',
      data: props.data.map(d => d.total_tokens),
      borderColor: 'rgba(139, 92, 246, 1)',
      backgroundColor: createGradient(),
      fill: true,
      tension: 0.4,
      pointRadius: 5,
      pointHoverRadius: 8,
      pointBackgroundColor: 'rgba(139, 92, 246, 1)',
      pointBorderColor: '#fff',
      pointBorderWidth: 2,
      borderWidth: 3,
    })
  }

  return { labels, datasets }
})

function createGradient() {
  return 'rgba(139, 92, 246, 0.2)'
}

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
      callbacks: {
        title: (items: any[]) => {
          if (items.length > 0) {
            return `📅 ${items[0].label}`
          }
          return ''
        },
        label: (context: any) => {
          const label = context.dataset.label || ''
          const value = formatLargeNumber(context.parsed.y)
          return ` ${label}: ${value}`
        },
        afterBody: (items: any[]) => {
          if (items.length > 0 && props.showPromptCompletion && props.data) {
            const index = items[0].dataIndex
            const dataItem = props.data[index]
            if (dataItem) {
              const total = dataItem.total_tokens
              return [`\n📊 总计: ${formatLargeNumber(total)}`]
            }
          }
          return []
        },
      },
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
      grid: {
        color: 'rgba(255, 255, 255, 0.08)',
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

function selectRange(range: string) {
  selectedRange.value = range
  emit('rangeChange', range)
}

// Emit initial range on mount
onMounted(() => {
  emit('rangeChange', selectedRange.value)
})
</script>

<style scoped>
.time-series-chart {
  width: 100%;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.chart-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: white;
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0;
}

.title-icon {
  font-size: 1.3rem;
}

.chart-controls {
  display: flex;
  gap: 0.5rem;
}

.chart-wrapper {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* 统计摘要样式 */
.chart-summary {
  display: flex;
  gap: 2rem;
  padding: 1rem 1.5rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.summary-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.input-dot {
  background: rgba(59, 130, 246, 1);
  box-shadow: 0 0 8px rgba(59, 130, 246, 0.5);
}

.output-dot {
  background: rgba(16, 185, 129, 1);
  box-shadow: 0 0 8px rgba(16, 185, 129, 0.5);
}

.summary-label {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.85rem;
}

.summary-value {
  color: white;
  font-size: 1rem;
  font-weight: 600;
}

.summary-item.total .summary-value {
  color: rgba(139, 92, 246, 1);
  font-size: 1.1rem;
}

.chart-container {
  height: 320px;
  position: relative;
  padding: 0.5rem;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 12px;
}

.chart-loading,
.chart-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 250px;
  gap: 1rem;
  color: rgba(255, 255, 255, 0.6);
  background: rgba(255, 255, 255, 0.02);
  border-radius: 12px;
}

.empty-icon {
  font-size: 3rem;
  opacity: 0.5;
}

@media (max-width: 768px) {
  .chart-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .chart-summary {
    flex-wrap: wrap;
    gap: 1rem;
    padding: 0.875rem 1.25rem;
  }
  
  .chart-container {
    height: 280px;
    padding: 0.375rem;
  }

  .chart-title {
    font-size: 1.1rem;
  }

  .title-icon {
    font-size: 1.2rem;
  }

  /* Touch-friendly button sizes */
  .chart-controls :deep(.n-button) {
    min-height: 36px;
    padding: 0.5rem 0.75rem;
  }

  .summary-label {
    font-size: 0.8rem;
  }

  .summary-value {
    font-size: 0.9rem;
  }

  .summary-item.total .summary-value {
    font-size: 1rem;
  }
}

@media (max-width: 480px) {
  .chart-summary {
    flex-direction: column;
    gap: 0.75rem;
    padding: 0.75rem 1rem;
  }

  .chart-container {
    height: 240px;
  }

  .chart-title {
    font-size: 1rem;
  }

  .chart-controls :deep(.n-button-group) {
    width: 100%;
  }

  .chart-controls :deep(.n-button) {
    flex: 1;
    min-height: 40px;
  }

  .chart-loading,
  .chart-empty {
    height: 200px;
  }

  .empty-icon {
    font-size: 2.5rem;
  }
}

/* Touch device optimizations for charts */
@media (hover: none) and (pointer: coarse) {
  .chart-container {
    /* Enable touch scrolling within chart area */
    touch-action: pan-x pan-y;
  }
}
</style>
