<template>
  <div class="skeleton-wrapper" :class="[`skeleton-${type}`, { 'skeleton-animated': animated }]">
    <!-- 卡片骨架屏 -->
    <template v-if="type === 'card'">
      <div class="skeleton-card glass-card">
        <div class="skeleton-icon skeleton-pulse"></div>
        <div class="skeleton-content">
          <div class="skeleton-line skeleton-line-lg skeleton-pulse"></div>
          <div class="skeleton-line skeleton-line-sm skeleton-pulse"></div>
        </div>
      </div>
    </template>

    <!-- 列表骨架屏 -->
    <template v-else-if="type === 'list'">
      <div class="skeleton-list glass-card">
        <div v-for="i in rows" :key="i" class="skeleton-list-item">
          <div class="skeleton-avatar skeleton-pulse"></div>
          <div class="skeleton-list-content">
            <div class="skeleton-line skeleton-line-md skeleton-pulse"></div>
            <div class="skeleton-line skeleton-line-sm skeleton-pulse"></div>
          </div>
        </div>
      </div>
    </template>

    <!-- 图表骨架屏 -->
    <template v-else-if="type === 'chart'">
      <div class="skeleton-chart glass-card">
        <div class="skeleton-chart-header">
          <div class="skeleton-line skeleton-line-md skeleton-pulse"></div>
        </div>
        <div class="skeleton-chart-body">
          <div class="skeleton-bars">
            <div v-for="i in 7" :key="i" class="skeleton-bar skeleton-pulse" :style="{ height: getRandomHeight() }"></div>
          </div>
          <div class="skeleton-chart-axis skeleton-pulse"></div>
        </div>
      </div>
    </template>

    <!-- 文本骨架屏 -->
    <template v-else-if="type === 'text'">
      <div class="skeleton-text">
        <div v-for="i in rows" :key="i" class="skeleton-line skeleton-pulse" :class="getTextLineClass(i)"></div>
      </div>
    </template>

    <!-- 统计卡片骨架屏 -->
    <template v-else-if="type === 'stat'">
      <div class="skeleton-stat glass-card">
        <div class="skeleton-stat-icon skeleton-pulse"></div>
        <div class="skeleton-stat-content">
          <div class="skeleton-line skeleton-line-lg skeleton-pulse"></div>
          <div class="skeleton-line skeleton-line-xs skeleton-pulse"></div>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
interface Props {
  type?: 'card' | 'list' | 'chart' | 'text' | 'stat'
  rows?: number
  animated?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  type: 'card',
  rows: 3,
  animated: true
})

// 为图表柱状图生成随机高度
function getRandomHeight(): string {
  const heights = ['40%', '60%', '80%', '50%', '70%', '90%', '55%']
  return heights[Math.floor(Math.random() * heights.length)] ?? '50%'
}

// 为文本行生成不同宽度的类
function getTextLineClass(index: number): string {
  if (index === props.rows) return 'skeleton-line-sm'
  if (index % 2 === 0) return 'skeleton-line-lg'
  return 'skeleton-line-md'
}
</script>


<style scoped>
/* 基础骨架屏样式 */
.skeleton-wrapper {
  width: 100%;
}

/* 脉冲动画 */
@keyframes skeleton-pulse {
  0%, 100% {
    opacity: 0.4;
  }
  50% {
    opacity: 0.8;
  }
}

.skeleton-pulse {
  background: linear-gradient(
    90deg,
    rgba(59, 130, 246, 0.1) 0%,
    rgba(59, 130, 246, 0.2) 50%,
    rgba(59, 130, 246, 0.1) 100%
  );
  background-size: 200% 100%;
}

.skeleton-animated .skeleton-pulse {
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

/* 骨架线条 */
.skeleton-line {
  height: 12px;
  border-radius: 6px;
  margin-bottom: 8px;
}

.skeleton-line:last-child {
  margin-bottom: 0;
}

.skeleton-line-xs {
  width: 30%;
}

.skeleton-line-sm {
  width: 50%;
}

.skeleton-line-md {
  width: 75%;
}

.skeleton-line-lg {
  width: 100%;
}

/* Glassmorphism 卡片 */
.glass-card {
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 16px;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    0 0 16px rgba(59, 130, 246, 0.1);
}

/* 卡片骨架屏 */
.skeleton-card {
  padding: 1.5rem;
  display: flex;
  align-items: flex-start;
  gap: 1rem;
}

.skeleton-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  flex-shrink: 0;
}

.skeleton-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 列表骨架屏 */
.skeleton-list {
  padding: 1rem;
}

.skeleton-list-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 0;
  border-bottom: 1px solid rgba(59, 130, 246, 0.1);
}

.skeleton-list-item:last-child {
  border-bottom: none;
}

.skeleton-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  flex-shrink: 0;
}

.skeleton-list-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

/* 图表骨架屏 */
.skeleton-chart {
  padding: 1.5rem;
}

.skeleton-chart-header {
  margin-bottom: 1.5rem;
}

.skeleton-chart-header .skeleton-line {
  width: 40%;
  height: 16px;
}

.skeleton-chart-body {
  height: 200px;
  display: flex;
  flex-direction: column;
}

.skeleton-bars {
  flex: 1;
  display: flex;
  align-items: flex-end;
  justify-content: space-around;
  gap: 12px;
  padding-bottom: 1rem;
}

.skeleton-bar {
  flex: 1;
  max-width: 40px;
  border-radius: 4px 4px 0 0;
  min-height: 20px;
}

.skeleton-chart-axis {
  height: 2px;
  width: 100%;
  border-radius: 1px;
}

/* 文本骨架屏 */
.skeleton-text {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* 统计卡片骨架屏 */
.skeleton-stat {
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
}

.skeleton-stat-icon {
  width: 56px;
  height: 56px;
  border-radius: 14px;
  flex-shrink: 0;
}

.skeleton-stat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.skeleton-stat-content .skeleton-line-lg {
  height: 24px;
  width: 60%;
}

.skeleton-stat-content .skeleton-line-xs {
  height: 10px;
  width: 40%;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .skeleton-card,
  .skeleton-stat {
    padding: 1rem;
  }
  
  .skeleton-icon,
  .skeleton-stat-icon {
    width: 40px;
    height: 40px;
  }
  
  .skeleton-chart-body {
    height: 150px;
  }
  
  .skeleton-bar {
    max-width: 30px;
  }
}
</style>
