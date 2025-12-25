<template>
  <div class="stat-card glass-card" :class="[colorClass]">
    <div class="stat-icon" :class="iconColorClass">{{ icon }}</div>
    <div class="stat-content">
      <div class="stat-value">{{ value }}</div>
      <div class="stat-label">{{ title }}</div>
      <div v-if="trend !== undefined" class="stat-trend" :class="trendClass">
        <span class="trend-icon">{{ trend >= 0 ? '↑' : '↓' }}</span>
        {{ Math.abs(trend) }}%
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  title: string
  value: string | number
  icon: string
  trend?: number
  color?: 'primary' | 'success' | 'warning' | 'error'
}

const props = withDefaults(defineProps<Props>(), {
  color: 'primary'
})

const colorClass = computed(() => `color-${props.color}`)
const iconColorClass = computed(() => `icon-${props.color}`)

const trendClass = computed(() => {
  if (props.trend === undefined) return ''
  return props.trend >= 0 ? 'trend-up' : 'trend-down'
})
</script>

<style scoped>
.stat-card {
  padding: 1.5rem;
  display: flex;
  align-items: flex-start;
  gap: 1rem;
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
  background: var(--color-primary-light);
}

/* 图标颜色区分类型 */
.icon-primary {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.icon-success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.icon-warning {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.icon-error {
  background: var(--color-error-light);
  color: var(--color-error);
}

.stat-content {
  flex: 1;
  min-width: 0;
}

.stat-value {
  color: var(--text-primary);
  font-size: 1.75rem;
  font-weight: 700;
  line-height: 1.2;
  word-break: break-word;
}

.stat-label {
  color: var(--text-secondary);
  font-size: 0.875rem;
  margin-top: 0.25rem;
}

.stat-trend {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  margin-top: 0.5rem;
  padding: 0.25rem 0.5rem;
  border-radius: var(--border-radius-sm);
  font-size: 0.8rem;
  font-weight: 600;
}

.trend-icon {
  font-size: 0.75rem;
}

.trend-up {
  background: var(--color-success-light);
  color: var(--color-success);
}

.trend-down {
  background: var(--color-error-light);
  color: var(--color-error);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stat-card {
    padding: 1.25rem;
    min-height: 80px;
  }
  
  .stat-icon {
    font-size: 2rem;
    width: 48px;
    height: 48px;
  }
  
  .stat-value {
    font-size: 1.5rem;
  }
  
  .stat-label {
    font-size: 0.8rem;
  }
}

@media (max-width: 480px) {
  .stat-card {
    padding: 1rem;
    gap: 0.75rem;
  }

  .stat-icon {
    font-size: 1.75rem;
    width: 44px;
    height: 44px;
  }

  .stat-value {
    font-size: 1.35rem;
  }

  .stat-label {
    font-size: 0.75rem;
  }

  .stat-trend {
    font-size: 0.75rem;
    padding: 0.2rem 0.4rem;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .stat-card:active {
    transform: scale(0.98);
  }
}
</style>
