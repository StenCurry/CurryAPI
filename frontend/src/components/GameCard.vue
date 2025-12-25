<template>
  <div class="game-card glass-card" :style="gradientStyle" @click="navigateToGame">
    <div class="game-icon">{{ icon }}</div>
    <div class="game-content">
      <h3 class="game-name">{{ name }}</h3>
      <p class="game-description">{{ description }}</p>
      <div class="game-meta">
        <span class="min-bet">
          <span class="bet-label">最低下注</span>
          <span class="bet-value">{{ minBet }} 游戏币</span>
        </span>
      </div>
    </div>
    <div class="game-arrow">→</div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

interface Props {
  name: string
  description: string
  icon: string
  minBet: number
  route: string
  gradient: string
}

const props = defineProps<Props>()
const router = useRouter()

const gradientStyle = computed(() => ({
  '--card-gradient': props.gradient
}))

function navigateToGame() {
  router.push(props.route)
}
</script>

<style scoped>
.game-card {
  padding: 1.5rem;
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: relative;
  overflow: hidden;
  border-radius: 20px;
  display: flex;
  align-items: center;
  gap: 1.25rem;
  cursor: pointer;
}

.game-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 3px;
  background: var(--card-gradient, linear-gradient(90deg, #3b82f6, #60a5fa));
}

.game-card::after {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--card-gradient, linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(139, 92, 246, 0.1)));
  opacity: 0;
  transition: opacity 0.4s ease;
  pointer-events: none;
}

.game-card:hover {
  transform: translateY(-6px) scale(1.02);
  box-shadow: 
    0 20px 60px rgba(0, 0, 0, 0.5),
    0 0 40px rgba(59, 130, 246, 0.3);
}

.game-card:hover::after {
  opacity: 1;
}

.game-card:active {
  transform: translateY(-2px) scale(0.99);
}

.game-icon {
  font-size: 3rem;
  filter: drop-shadow(0 4px 16px rgba(59, 130, 246, 0.4));
  flex-shrink: 0;
  z-index: 1;
}

.game-content {
  flex: 1;
  min-width: 0;
  z-index: 1;
}

.game-name {
  color: white;
  font-size: 1.25rem;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
  text-shadow: 0 2px 8px rgba(255, 255, 255, 0.15);
}

.game-description {
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.9rem;
  margin: 0 0 0.75rem 0;
  line-height: 1.4;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.game-meta {
  display: flex;
  align-items: center;
}

.min-bet {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.bet-label {
  color: rgba(255, 255, 255, 0.5);
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.bet-value {
  color: #60a5fa;
  font-size: 0.9rem;
  font-weight: 600;
}

.game-arrow {
  font-size: 1.5rem;
  color: rgba(255, 255, 255, 0.4);
  transition: all 0.3s ease;
  flex-shrink: 0;
  z-index: 1;
}

.game-card:hover .game-arrow {
  color: white;
  transform: translateX(4px);
}

/* Glassmorphism 卡片 */
.glass-card {
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(59, 130, 246, 0.4);
  border-radius: 20px;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.4),
    0 0 20px rgba(59, 130, 246, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .game-card {
    padding: 1.25rem;
    gap: 1rem;
    /* Touch-friendly minimum tap target */
    min-height: 100px;
  }
  
  .game-icon {
    font-size: 2.5rem;
  }
  
  .game-name {
    font-size: 1.1rem;
  }
  
  .game-description {
    font-size: 0.85rem;
  }
  
  .game-arrow {
    display: none;
  }
}

@media (max-width: 480px) {
  .game-card {
    padding: 1rem;
    gap: 0.875rem;
  }

  .game-icon {
    font-size: 2.25rem;
  }

  .game-name {
    font-size: 1rem;
  }

  .game-description {
    font-size: 0.8rem;
    -webkit-line-clamp: 2;
  }

  .bet-label {
    font-size: 0.7rem;
  }

  .bet-value {
    font-size: 0.85rem;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .game-card:hover {
    transform: none;
    box-shadow: 
      0 8px 32px rgba(0, 0, 0, 0.4),
      0 0 20px rgba(59, 130, 246, 0.15),
      inset 0 1px 0 rgba(255, 255, 255, 0.1);
  }

  .game-card:hover::after {
    opacity: 0;
  }

  .game-card:active {
    transform: scale(0.98);
    box-shadow: 
      0 4px 16px rgba(0, 0, 0, 0.4),
      0 0 10px rgba(59, 130, 246, 0.2);
  }

  .game-card:active::after {
    opacity: 1;
  }
}
</style>
