<template>
  <div class="wheel-container">
    <div class="wheel-wrapper">
      <!-- Pointer indicator -->
      <div class="wheel-pointer">
        <div class="pointer-arrow"></div>
      </div>
      
      <!-- Spinning wheel -->
      <div 
        class="wheel" 
        :style="wheelStyle"
        :class="{ spinning: spinning }"
      >
        <svg viewBox="0 0 200 200" class="wheel-svg">
          <!-- Wheel segments -->
          <g v-for="(segment, index) in segments" :key="index">
            <path
              :d="getSegmentPath(index)"
              :fill="segment.color"
              stroke="#1f2937"
              stroke-width="1"
            />
            <text
              :transform="getTextTransform(index)"
              text-anchor="middle"
              dominant-baseline="middle"
              fill="white"
              font-size="10"
              font-weight="bold"
              class="segment-label"
            >
              {{ segment.label }}
            </text>
          </g>
          
          <!-- Center circle -->
          <circle cx="100" cy="100" r="15" fill="#1f2937" />
          <circle cx="100" cy="100" r="10" fill="#374151" />
        </svg>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import type { WheelSegment } from '@/utils/gameUtils'


/**
 * WheelSpinner Component
 * 转盘动画组件
 * 
 * Requirements: 5.1, 5.5
 */

interface Props {
  segments: WheelSegment[]
  spinning: boolean
  targetIndex?: number
}

const props = withDefaults(defineProps<Props>(), {
  targetIndex: 0
})

const emit = defineEmits<{
  (e: 'spinEnd'): void
}>()

// Current rotation angle (normalized to 0-360 for display)
const currentRotation = ref(0)
// Total accumulated rotation for animation
const totalRotation = ref(0)
// Animation duration in seconds
const spinDuration = 4

// Calculate the angle for each segment
const segmentAngle = computed(() => 360 / props.segments.length)

// Calculate wheel rotation style
const wheelStyle = computed(() => {
  if (props.spinning && props.targetIndex !== undefined) {
    // Calculate target angle from current position
    // - Add multiple full rotations (5 spins) for visual effect
    // - Plus angle to land on target segment
    // 
    // Segments are drawn starting from top (12 o'clock, -90° in SVG coordinates)
    // To land pointer on segment N, we need segment N's center at top
    // 
    // Target final angle (where segment N center is at top):
    const segmentCenterAngle = props.targetIndex * segmentAngle.value + segmentAngle.value / 2
    const targetAngle = 360 - segmentCenterAngle
    
    // Normalize current rotation to 0-360
    const currentNormalized = currentRotation.value % 360
    
    // Calculate how much more we need to rotate
    // Add 5 full rotations for visual effect
    const fullRotations = 5 * 360
    let additionalRotation = targetAngle - currentNormalized
    if (additionalRotation < 0) {
      additionalRotation += 360
    }
    
    const newTotalRotation = currentRotation.value + fullRotations + additionalRotation
    totalRotation.value = newTotalRotation
    
    return {
      transform: `rotate(${newTotalRotation}deg)`,
      transition: `transform ${spinDuration}s cubic-bezier(0.17, 0.67, 0.12, 0.99)`
    }
  }
  
  return {
    transform: `rotate(${currentRotation.value}deg)`,
    transition: 'none'
  }
})

// Watch for spinning state changes
watch(() => props.spinning, (newVal, oldVal) => {
  if (oldVal && !newVal) {
    // Spinning ended - update current rotation to final position
    currentRotation.value = totalRotation.value
    emit('spinEnd')
  }
})

// Calculate SVG path for a segment
function getSegmentPath(index: number): string {
  const centerX = 100
  const centerY = 100
  const radius = 95
  const startAngle = (index * segmentAngle.value - 90) * (Math.PI / 180)
  const endAngle = ((index + 1) * segmentAngle.value - 90) * (Math.PI / 180)
  
  const x1 = centerX + radius * Math.cos(startAngle)
  const y1 = centerY + radius * Math.sin(startAngle)
  const x2 = centerX + radius * Math.cos(endAngle)
  const y2 = centerY + radius * Math.sin(endAngle)
  
  const largeArcFlag = segmentAngle.value > 180 ? 1 : 0
  
  return `M ${centerX} ${centerY} L ${x1} ${y1} A ${radius} ${radius} 0 ${largeArcFlag} 1 ${x2} ${y2} Z`
}

// Calculate text transform for segment label
function getTextTransform(index: number): string {
  const centerX = 100
  const centerY = 100
  const textRadius = 65
  const angle = (index * segmentAngle.value + segmentAngle.value / 2 - 90) * (Math.PI / 180)
  
  const x = centerX + textRadius * Math.cos(angle)
  const y = centerY + textRadius * Math.sin(angle)
  const rotation = index * segmentAngle.value + segmentAngle.value / 2
  
  return `translate(${x}, ${y}) rotate(${rotation})`
}
</script>


<style scoped>
.wheel-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
}

.wheel-wrapper {
  position: relative;
  width: 300px;
  height: 300px;
}

.wheel-pointer {
  position: absolute;
  top: -10px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 10;
}

.pointer-arrow {
  width: 0;
  height: 0;
  border-left: 15px solid transparent;
  border-right: 15px solid transparent;
  border-top: 30px solid #fbbf24;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.3));
}

.wheel {
  width: 100%;
  height: 100%;
  border-radius: 50%;
  box-shadow: 
    0 0 0 8px #1f2937,
    0 0 0 12px #374151,
    0 0 20px rgba(0, 0, 0, 0.5);
}

.wheel-svg {
  width: 100%;
  height: 100%;
}

.segment-label {
  text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.5);
  pointer-events: none;
}

.spinning {
  animation: pulse 0.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% {
    filter: brightness(1);
  }
  50% {
    filter: brightness(1.1);
  }
}

/* Responsive adjustments */
@media (max-width: 768px) {
  .wheel-container {
    padding: 15px;
  }

  .wheel-wrapper {
    width: 260px;
    height: 260px;
  }
  
  .pointer-arrow {
    border-left: 12px solid transparent;
    border-right: 12px solid transparent;
    border-top: 24px solid #fbbf24;
  }

  .wheel {
    box-shadow: 
      0 0 0 6px #1f2937,
      0 0 0 9px #374151,
      0 0 15px rgba(0, 0, 0, 0.5);
  }
}

@media (max-width: 480px) {
  .wheel-container {
    padding: 10px;
  }

  .wheel-wrapper {
    width: 220px;
    height: 220px;
  }
  
  .pointer-arrow {
    border-left: 10px solid transparent;
    border-right: 10px solid transparent;
    border-top: 20px solid #fbbf24;
  }

  .wheel-pointer {
    top: -8px;
  }

  .wheel {
    box-shadow: 
      0 0 0 5px #1f2937,
      0 0 0 8px #374151,
      0 0 12px rgba(0, 0, 0, 0.5);
  }

  .segment-label {
    font-size: 8px;
  }
}

@media (max-width: 360px) {
  .wheel-wrapper {
    width: 180px;
    height: 180px;
  }

  .pointer-arrow {
    border-left: 8px solid transparent;
    border-right: 8px solid transparent;
    border-top: 16px solid #fbbf24;
  }

  .segment-label {
    font-size: 7px;
  }
}
</style>
