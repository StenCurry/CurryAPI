<template>
  <div class="model-breakdown-chart">
    <div class="chart-header">
      <h4 class="chart-title">
        <span class="title-icon">🤖</span>
        {{ title }}
      </h4>
      <div class="chart-type-toggle">
        <n-button-group size="small">
          <n-button
            :type="chartType === 'doughnut' ? 'primary' : 'default'"
            @click="chartType = 'doughnut'"
          >
            环形图
          </n-button>
          <n-button
            :type="chartType === 'bar' ? 'primary' : 'default'"
            @click="chartType = 'bar'"
          >
            柱状图
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
      <span>暂无模型使用数据</span>
    </div>
    
    <div v-else class="chart-content">
      <div class="chart-container">
        <Doughnut v-if="chartType === 'doughnut'" :data="doughnutData" :options="doughnutOptions" />
        <Bar v-else :data="barData" :options="barOptions" />
      </div>
      
      <!-- Legend with details -->
      <div class="chart-legend">
        <div
          v-for="(model, index) in modelData"
          :key="model.model"
          class="legend-item"
        >
          <div class="legend-color" :style="{ backgroundColor: getColor(index) }"></div>
          <div class="legend-info">
            <span class="legend-name">{{ model.model }}</span>
            <span class="legend-value">{{ formatNumber(model.total_tokens) }} tokens</span>
            <span class="legend-percent">{{ getPercentage(model.total_tokens) }}%</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { NButton, NButtonGroup, NSpin } from 'naive-ui'
import {
  Chart as ChartJS,
  ArcElement,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend,
} from 'chart.js'
import { Doughnut, Bar } from 'vue-chartjs'
import { useChartConfig } from '@/composables/useChartConfig'

// Register Chart.js components
ChartJS.register(
  ArcElement,
  CategoryScale,
  LinearScale,
  BarElement,
  Title,
  Tooltip,
  Legend
)

interface ModelStats {
  model: string
  request_count: number
  total_tokens: number
  prompt_tokens: number
  completion_tokens: number
}

interface Props {
  title?: string
  data?: ModelStats[]
  loading?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  title: '模型使用分布',
  data: () => [],
  loading: false,
})

const { colorPalette, borderColorPalette, formatNumber: formatNum } = useChartConfig()

const chartType = ref<'doughnut' | 'bar'>('doughnut')

const hasData = computed(() => props.data && props.data.length > 0)

const modelData = computed(() => {
  if (!hasData.value) return []
  // Sort by total_tokens descending
  return [...props.data].sort((a, b) => b.total_tokens - a.total_tokens)
})

const totalTokens = computed(() => {
  return modelData.value.reduce((sum, m) => sum + m.total_tokens, 0)
})

function getColor(index: number): string {
  return colorPalette[index % colorPalette.length] || 'rgba(59, 130, 246, 0.8)'
}

function getBorderColor(index: number): string {
  return borderColorPalette[index % borderColorPalette.length] || 'rgba(59, 130, 246, 1)'
}

function getPercentage(tokens: number): string {
  if (totalTokens.value === 0) return '0'
  return ((tokens / totalTokens.value) * 100).toFixed(1)
}

function formatNumber(num: number): string {
  return num.toLocaleString('zh-CN')
}

const doughnutData = computed(() => ({
  labels: modelData.value.map(m => m.model),
  datasets: [
    {
      data: modelData.value.map(m => m.total_tokens),
      backgroundColor: modelData.value.map((_, i) => getColor(i)),
      borderColor: modelData.value.map((_, i) => getBorderColor(i)),
      borderWidth: 2,
      hoverOffset: 8,
    },
  ],
}))

const doughnutOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  cutout: '60%',
  plugins: {
    legend: {
      display: false, // We use custom legend
    },
    tooltip: {
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      titleColor: 'rgba(255, 255, 255, 0.9)',
      bodyColor: 'rgba(255, 255, 255, 0.8)',
      borderColor: 'rgba(59, 130, 246, 0.5)',
      borderWidth: 1,
      padding: 12,
      cornerRadius: 8,
      callbacks: {
        label: (context: any) => {
          const label = context.label || ''
          const value = formatNumber(context.parsed)
          const percent = getPercentage(context.parsed)
          return `${label}: ${value} tokens (${percent}%)`
        },
      },
    },
  },
}))

const barData = computed(() => ({
  labels: modelData.value.map(m => m.model),
  datasets: [
    {
      label: 'Token 使用量',
      data: modelData.value.map(m => m.total_tokens),
      backgroundColor: modelData.value.map((_, i) => getColor(i)),
      borderColor: modelData.value.map((_, i) => getBorderColor(i)),
      borderWidth: 1,
      borderRadius: 6,
    },
  ],
}))

const barOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  indexAxis: 'y' as const,
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      titleColor: 'rgba(255, 255, 255, 0.9)',
      bodyColor: 'rgba(255, 255, 255, 0.8)',
      borderColor: 'rgba(59, 130, 246, 0.5)',
      borderWidth: 1,
      padding: 12,
      cornerRadius: 8,
      callbacks: {
        label: (context: any) => {
          const value = formatNumber(context.parsed.x)
          const percent = getPercentage(context.parsed.x)
          return `${value} tokens (${percent}%)`
        },
      },
    },
  },
  scales: {
    x: {
      grid: {
        color: 'rgba(255, 255, 255, 0.1)',
      },
      ticks: {
        color: 'rgba(255, 255, 255, 0.7)',
        callback: (value: any) => formatNum(Number(value)),
      },
      beginAtZero: true,
    },
    y: {
      grid: {
        display: false,
      },
      ticks: {
        color: 'rgba(255, 255, 255, 0.7)',
      },
    },
  },
}))
</script>

<style scoped>
.model-breakdown-chart {
  width: 100%;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.chart-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: white;
  font-size: 1.1rem;
  font-weight: 600;
  margin: 0;
}

.title-icon {
  font-size: 1.2rem;
}

.chart-content {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  align-items: center;
}

.chart-container {
  height: 250px;
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
}

.chart-legend {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.legend-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.5rem 0.75rem;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  transition: background 0.2s ease;
}

.legend-item:hover {
  background: rgba(255, 255, 255, 0.1);
}

.legend-color {
  width: 12px;
  height: 12px;
  border-radius: 3px;
  flex-shrink: 0;
}

.legend-info {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  flex: 1;
  min-width: 0;
}

.legend-name {
  color: white;
  font-weight: 500;
  font-size: 0.9rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.legend-value {
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.8rem;
}

.legend-percent {
  color: #60a5fa;
  font-weight: 600;
  font-size: 0.85rem;
}

.chart-loading,
.chart-empty {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  gap: 0.75rem;
  color: rgba(255, 255, 255, 0.6);
}

.empty-icon {
  font-size: 2.5rem;
  opacity: 0.5;
}

@media (max-width: 768px) {
  .chart-content {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }
  
  .chart-container {
    height: 220px;
  }
  
  .chart-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .chart-title {
    font-size: 1rem;
  }

  /* Touch-friendly button sizes */
  .chart-type-toggle :deep(.n-button) {
    min-height: 36px;
    padding: 0.5rem 0.75rem;
  }

  .legend-item {
    padding: 0.625rem 0.875rem;
  }

  .legend-name {
    font-size: 0.85rem;
  }

  .legend-value {
    font-size: 0.75rem;
  }

  .legend-percent {
    font-size: 0.8rem;
  }
}

@media (max-width: 480px) {
  .chart-content {
    gap: 1rem;
  }

  .chart-container {
    height: 180px;
  }

  .chart-type-toggle :deep(.n-button-group) {
    width: 100%;
  }

  .chart-type-toggle :deep(.n-button) {
    flex: 1;
    min-height: 40px;
  }

  .chart-legend {
    gap: 0.5rem;
  }

  .legend-item {
    padding: 0.5rem 0.75rem;
  }

  .legend-color {
    width: 10px;
    height: 10px;
  }

  .chart-loading,
  .chart-empty {
    height: 160px;
  }

  .empty-icon {
    font-size: 2rem;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .legend-item:hover {
    background: rgba(255, 255, 255, 0.05);
  }

  .legend-item:active {
    background: rgba(255, 255, 255, 0.1);
  }

  .chart-container {
    /* Enable touch interactions for charts */
    touch-action: manipulation;
  }
}
</style>
