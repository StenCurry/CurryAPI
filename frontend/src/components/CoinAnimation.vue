<template>
  <div class="coin-container">
    <div 
      class="coin" 
      :class="{ flipping: flipping, 'show-result': showResult }"
      :style="coinStyle"
    >
      <!-- Heads side -->
      <div class="coin-face coin-heads">
        <div class="coin-inner">
          <span class="coin-symbol">👑</span>
          <span class="coin-text">正面</span>
        </div>
      </div>
      <!-- Tails side -->
      <div class="coin-face coin-tails">
        <div class="coin-inner">
          <span class="coin-symbol">🌙</span>
          <span class="coin-text">反面</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * CoinAnimation Component
 * 3D 硬币翻转动画组件
 * 
 * Requirements: 7.1, 7.4
 */

import { computed, ref, watch } from 'vue'

export type CoinResult = 'heads' | 'tails'

interface Props {
  flipping: boolean
  result?: CoinResult
}

const props = withDefaults(defineProps<Props>(), {
  result: undefined
})

const emit = defineEmits<{
  (e: 'flipEnd'): void
}>()

// Track when to show the final result
const showResult = ref(false)

// Calculate coin rotation style based on result
const coinStyle = computed(() => {
  if (showResult.value && props.result) {
    // Final position: heads = 0deg, tails = 180deg
    const finalRotation = props.result === 'heads' ? 0 : 180
    return {
      transform: `rotateY(${finalRotation}deg)`,
      transition: 'none'
    }
  }
  return {}
})

// Watch for flipping state changes
watch(() => props.flipping, (newVal) => {
  if (newVal) {
    // Start flipping - hide result
    showResult.value = false
  } else if (!newVal && props.result) {
    // Flipping ended - show result after animation
    setTimeout(() => {
      showResult.value = true
      emit('flipEnd')
    }, 100)
  }
})
</script>


<style scoped>
.coin-container {
  display: flex;
  justify-content: center;
  align-items: center;
  perspective: 1000px;
  padding: 40px;
}

.coin {
  width: 180px;
  height: 180px;
  position: relative;
  transform-style: preserve-3d;
  transition: transform 0.1s ease-out;
}

.coin-face {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  backface-visibility: hidden;
  display: flex;
  justify-content: center;
  align-items: center;
  box-shadow: 
    0 0 0 8px rgba(255, 215, 0, 0.3),
    0 10px 30px rgba(0, 0, 0, 0.4),
    inset 0 -5px 20px rgba(0, 0, 0, 0.2),
    inset 0 5px 20px rgba(255, 255, 255, 0.2);
}

.coin-heads {
  background: linear-gradient(145deg, #ffd700, #ffb700, #ffd700);
  transform: rotateY(0deg);
}

.coin-tails {
  background: linear-gradient(145deg, #c0c0c0, #a0a0a0, #c0c0c0);
  transform: rotateY(180deg);
}

.coin-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.coin-symbol {
  font-size: 48px;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
}

.coin-text {
  font-size: 18px;
  font-weight: bold;
  color: #1f2937;
  text-shadow: 0 1px 2px rgba(255, 255, 255, 0.5);
}

/* Flipping animation */
.coin.flipping {
  animation: coinFlip 2s ease-in-out forwards;
}

@keyframes coinFlip {
  0% {
    transform: rotateY(0deg);
  }
  25% {
    transform: rotateY(900deg) scale(1.1);
  }
  50% {
    transform: rotateY(1800deg) scale(1.2);
  }
  75% {
    transform: rotateY(2520deg) scale(1.1);
  }
  100% {
    transform: rotateY(3240deg);
  }
}

/* Result state - override animation */
.coin.show-result {
  animation: none;
}

/* Hover effect when not flipping */
.coin:not(.flipping):not(.show-result):hover {
  transform: rotateY(15deg) scale(1.05);
  transition: transform 0.3s ease;
}

/* Shine effect */
.coin-face::before {
  content: '';
  position: absolute;
  top: 10%;
  left: 10%;
  width: 30%;
  height: 30%;
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.6) 0%,
    rgba(255, 255, 255, 0) 50%
  );
  border-radius: 50%;
}

/* Edge effect */
.coin-face::after {
  content: '';
  position: absolute;
  top: 5%;
  left: 5%;
  right: 5%;
  bottom: 5%;
  border: 3px solid rgba(255, 255, 255, 0.2);
  border-radius: 50%;
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .coin-container {
    padding: 30px;
  }

  .coin {
    width: 150px;
    height: 150px;
  }
  
  .coin-symbol {
    font-size: 40px;
  }
  
  .coin-text {
    font-size: 16px;
  }

  .coin-face {
    box-shadow: 
      0 0 0 6px rgba(255, 215, 0, 0.3),
      0 8px 24px rgba(0, 0, 0, 0.4),
      inset 0 -4px 16px rgba(0, 0, 0, 0.2),
      inset 0 4px 16px rgba(255, 255, 255, 0.2);
  }
}

@media (max-width: 480px) {
  .coin-container {
    padding: 20px;
  }

  .coin {
    width: 130px;
    height: 130px;
  }
  
  .coin-symbol {
    font-size: 36px;
  }
  
  .coin-text {
    font-size: 14px;
  }

  .coin-inner {
    gap: 6px;
  }

  .coin-face {
    box-shadow: 
      0 0 0 5px rgba(255, 215, 0, 0.3),
      0 6px 20px rgba(0, 0, 0, 0.4),
      inset 0 -3px 12px rgba(0, 0, 0, 0.2),
      inset 0 3px 12px rgba(255, 255, 255, 0.2);
  }
}

@media (max-width: 360px) {
  .coin {
    width: 110px;
    height: 110px;
  }

  .coin-symbol {
    font-size: 30px;
  }

  .coin-text {
    font-size: 12px;
  }
}
</style>
