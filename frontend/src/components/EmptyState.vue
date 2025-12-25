<template>
  <div class="empty-state" :class="[`empty-${size}`]">
    <!-- 插图区域 -->
    <div class="empty-illustration">
      <!-- 默认空状态插图 -->
      <template v-if="type === 'default'">
        <svg viewBox="0 0 200 200" class="empty-svg">
          <defs>
            <linearGradient id="emptyGradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#3b82f6;stop-opacity:0.3" />
              <stop offset="100%" style="stop-color:#8b5cf6;stop-opacity:0.3" />
            </linearGradient>
          </defs>
          <circle cx="100" cy="100" r="80" fill="url(#emptyGradient)" />
          <rect x="60" y="70" width="80" height="60" rx="8" fill="rgba(59, 130, 246, 0.2)" stroke="rgba(59, 130, 246, 0.5)" stroke-width="2" />
          <line x1="70" y1="90" x2="130" y2="90" stroke="rgba(59, 130, 246, 0.4)" stroke-width="2" stroke-linecap="round" />
          <line x1="70" y1="105" x2="110" y2="105" stroke="rgba(59, 130, 246, 0.4)" stroke-width="2" stroke-linecap="round" />
          <line x1="70" y1="120" x2="90" y2="120" stroke="rgba(59, 130, 246, 0.4)" stroke-width="2" stroke-linecap="round" />
        </svg>
      </template>

      <!-- 无数据插图 -->
      <template v-else-if="type === 'no-data'">
        <svg viewBox="0 0 200 200" class="empty-svg">
          <defs>
            <linearGradient id="noDataGradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#f59e0b;stop-opacity:0.3" />
              <stop offset="100%" style="stop-color:#ef4444;stop-opacity:0.3" />
            </linearGradient>
          </defs>
          <circle cx="100" cy="100" r="80" fill="url(#noDataGradient)" />
          <circle cx="100" cy="90" r="35" fill="none" stroke="rgba(245, 158, 11, 0.5)" stroke-width="3" />
          <line x1="125" y1="115" x2="145" y2="135" stroke="rgba(245, 158, 11, 0.5)" stroke-width="4" stroke-linecap="round" />
          <text x="100" y="95" text-anchor="middle" fill="rgba(245, 158, 11, 0.6)" font-size="24">?</text>
        </svg>
      </template>

      <!-- 无结果插图 -->
      <template v-else-if="type === 'no-results'">
        <svg viewBox="0 0 200 200" class="empty-svg">
          <defs>
            <linearGradient id="noResultsGradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#10b981;stop-opacity:0.3" />
              <stop offset="100%" style="stop-color:#3b82f6;stop-opacity:0.3" />
            </linearGradient>
          </defs>
          <circle cx="100" cy="100" r="80" fill="url(#noResultsGradient)" />
          <rect x="55" y="60" width="90" height="80" rx="8" fill="rgba(16, 185, 129, 0.2)" stroke="rgba(16, 185, 129, 0.5)" stroke-width="2" />
          <line x1="75" y1="85" x2="125" y2="85" stroke="rgba(16, 185, 129, 0.3)" stroke-width="2" stroke-dasharray="4,4" />
          <line x1="75" y1="100" x2="125" y2="100" stroke="rgba(16, 185, 129, 0.3)" stroke-width="2" stroke-dasharray="4,4" />
          <line x1="75" y1="115" x2="125" y2="115" stroke="rgba(16, 185, 129, 0.3)" stroke-width="2" stroke-dasharray="4,4" />
        </svg>
      </template>
    </div>

      <!-- 错误插图 -->
      <template v-else-if="type === 'error'">
        <svg viewBox="0 0 200 200" class="empty-svg">
          <defs>
            <linearGradient id="errorGradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#ef4444;stop-opacity:0.3" />
              <stop offset="100%" style="stop-color:#dc2626;stop-opacity:0.3" />
            </linearGradient>
          </defs>
          <circle cx="100" cy="100" r="80" fill="url(#errorGradient)" />
          <circle cx="100" cy="100" r="40" fill="none" stroke="rgba(239, 68, 68, 0.5)" stroke-width="3" />
          <line x1="85" y1="85" x2="115" y2="115" stroke="rgba(239, 68, 68, 0.6)" stroke-width="4" stroke-linecap="round" />
          <line x1="115" y1="85" x2="85" y2="115" stroke="rgba(239, 68, 68, 0.6)" stroke-width="4" stroke-linecap="round" />
        </svg>
      </template>

      <!-- 游戏空状态插图 -->
      <template v-else-if="type === 'game'">
        <svg viewBox="0 0 200 200" class="empty-svg">
          <defs>
            <linearGradient id="gameGradient" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" style="stop-color:#8b5cf6;stop-opacity:0.3" />
              <stop offset="100%" style="stop-color:#ec4899;stop-opacity:0.3" />
            </linearGradient>
          </defs>
          <circle cx="100" cy="100" r="80" fill="url(#gameGradient)" />
          <rect x="60" y="75" width="80" height="50" rx="10" fill="rgba(139, 92, 246, 0.2)" stroke="rgba(139, 92, 246, 0.5)" stroke-width="2" />
          <circle cx="80" cy="100" r="8" fill="rgba(139, 92, 246, 0.4)" />
          <circle cx="120" cy="100" r="8" fill="rgba(139, 92, 246, 0.4)" />
          <rect x="95" y="85" width="10" height="30" rx="2" fill="rgba(139, 92, 246, 0.4)" />
        </svg>
      </template>
    </div>

    <!-- 标题 -->
    <h3 class="empty-title">{{ title || defaultTitle }}</h3>

    <!-- 描述 -->
    <p class="empty-description">{{ description || defaultDescription }}</p>

    <!-- 操作按钮插槽 -->
    <div v-if="$slots.action" class="empty-action">
      <slot name="action"></slot>
    </div>

    <!-- 默认操作按钮 -->
    <div v-else-if="actionText" class="empty-action">
      <button class="empty-button" @click="$emit('action')">
        {{ actionText }}
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  type?: 'default' | 'no-data' | 'no-results' | 'error' | 'game'
  title?: string
  description?: string
  actionText?: string
  size?: 'small' | 'medium' | 'large'
}

const props = withDefaults(defineProps<Props>(), {
  type: 'default',
  size: 'medium'
})

defineEmits<{
  action: []
}>()

const defaultTitle = computed(() => {
  const titles: Record<string, string> = {
    'default': '暂无内容',
    'no-data': '暂无数据',
    'no-results': '未找到结果',
    'error': '出错了',
    'game': '暂无游戏记录'
  }
  return titles[props.type] || '暂无内容'
})

const defaultDescription = computed(() => {
  const descriptions: Record<string, string> = {
    'default': '这里还没有任何内容，稍后再来看看吧',
    'no-data': '数据正在加载中，或者还没有相关数据',
    'no-results': '尝试调整搜索条件或筛选器',
    'error': '发生了一些问题，请稍后重试',
    'game': '开始游戏，创造你的第一条记录吧！'
  }
  return descriptions[props.type] || '这里还没有任何内容'
})
</script>


<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 2rem;
}

/* 尺寸变体 */
.empty-small {
  padding: 1rem;
}

.empty-small .empty-svg {
  width: 80px;
  height: 80px;
}

.empty-small .empty-title {
  font-size: 1rem;
}

.empty-small .empty-description {
  font-size: 0.8rem;
}

.empty-medium .empty-svg {
  width: 120px;
  height: 120px;
}

.empty-large {
  padding: 3rem;
}

.empty-large .empty-svg {
  width: 160px;
  height: 160px;
}

.empty-large .empty-title {
  font-size: 1.5rem;
}

.empty-large .empty-description {
  font-size: 1rem;
}

/* 插图样式 */
.empty-illustration {
  margin-bottom: 1.5rem;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.empty-svg {
  width: 120px;
  height: 120px;
  filter: drop-shadow(0 8px 24px rgba(59, 130, 246, 0.2));
}

/* 标题样式 */
.empty-title {
  color: rgba(255, 255, 255, 0.9);
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 0.5rem 0;
  text-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

/* 描述样式 */
.empty-description {
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.9rem;
  margin: 0 0 1.5rem 0;
  max-width: 300px;
  line-height: 1.5;
}

/* 操作按钮区域 */
.empty-action {
  margin-top: 0.5rem;
}

.empty-button {
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  color: white;
  border: none;
  padding: 0.75rem 1.5rem;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  box-shadow: 
    0 4px 16px rgba(59, 130, 246, 0.3),
    0 0 20px rgba(59, 130, 246, 0.1);
}

.empty-button:hover {
  transform: translateY(-2px);
  box-shadow: 
    0 8px 24px rgba(59, 130, 246, 0.4),
    0 0 30px rgba(59, 130, 246, 0.2);
}

.empty-button:active {
  transform: translateY(0);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .empty-state {
    padding: 1.5rem;
  }
  
  .empty-svg {
    width: 100px;
    height: 100px;
  }
  
  .empty-title {
    font-size: 1.1rem;
  }
  
  .empty-description {
    font-size: 0.85rem;
    max-width: 250px;
  }
  
  .empty-button {
    padding: 0.6rem 1.25rem;
    font-size: 0.85rem;
  }
}
</style>
