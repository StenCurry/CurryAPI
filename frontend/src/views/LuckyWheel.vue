<template>
  <div class="lucky-wheel-page">
    <!-- Header -->
    <div class="page-header">
      <n-button quaternary @click="$router.push('/games')">
        <template #icon>
          <n-icon><ArrowBackOutline /></n-icon>
        </template>
        返回游戏中心
      </n-button>
      <h1 class="page-title">🎡 幸运转盘</h1>
      <div class="balance-display">
        <n-icon><WalletOutline /></n-icon>
        <span>{{ gameStore.gameCoins.toFixed(2) }} 游戏币</span>
      </div>
    </div>

    <!-- Main Content -->
    <div class="game-content">
      <!-- Wheel Section -->
      <div class="wheel-section">
        <WheelSpinner
          :segments="wheelConfig.segments"
          :spinning="isSpinning"
          :target-index="targetIndex"
          @spin-end="onSpinEnd"
        />
      </div>

      <!-- Controls Section -->
      <div class="controls-section">
        <n-card class="control-card glass-card">
          <h3>下注金额</h3>
          <n-input-number
            v-model:value="betAmount"
            :min="wheelConfig.minBet"
            :max="Math.min(wheelConfig.maxBet, gameStore.gameCoins)"
            :disabled="isSpinning"
            :precision="2"
            placeholder="输入下注金额"
          >
            <template #prefix>🪙</template>
          </n-input-number>
          
          <div class="quick-bet-buttons">
            <n-button 
              v-for="amount in quickBetAmounts" 
              :key="amount"
              size="small"
              :disabled="isSpinning || amount > gameStore.gameCoins"
              @click="betAmount = amount"
            >
              {{ amount }}
            </n-button>
            <n-button 
              size="small"
              :disabled="isSpinning || gameStore.gameCoins <= 0"
              @click="betAmount = Math.min(wheelConfig.maxBet, gameStore.gameCoins)"
            >
              最大
            </n-button>
          </div>

          <n-button
            type="primary"
            size="large"
            block
            :disabled="!canSpin"
            :loading="isSpinning"
            @click="spin"
          >
            {{ isSpinning ? '转动中...' : '开始转动' }}
          </n-button>

          <div class="bet-info">
            <p>最低下注: {{ wheelConfig.minBet }} 游戏币</p>
            <p>最高下注: {{ wheelConfig.maxBet }} 游戏币</p>
          </div>
        </n-card>

        <!-- Prize Table -->
        <n-card class="prize-card glass-card">
          <h3>奖励倍率</h3>
          <div class="prize-table">
            <div 
              v-for="segment in wheelConfig.segments" 
              :key="segment.label"
              class="prize-item"
              :style="{ borderLeftColor: segment.color }"
            >
              <span class="prize-label">{{ segment.label }}</span>
              <span class="prize-desc">
                {{ getPrizeDescription(segment.multiplier) }}
              </span>
            </div>
          </div>
        </n-card>
      </div>
    </div>

    <!-- Result Modal -->
    <n-modal v-model:show="showResult" :mask-closable="false">
      <n-card
        class="result-card"
        :class="{ 
          'win-card': (lastResult?.netProfit ?? 0) > 0, 
          'lose-card': (lastResult?.netProfit ?? 0) < 0,
          'even-card': lastResult?.netProfit === 0
        }"
        style="width: 400px"
      >
        <div class="result-content">
          <div class="result-icon">
            {{ (lastResult?.netProfit ?? 0) > 0 ? '🎉' : (lastResult?.netProfit === 0 ? '😐' : '😢') }}
          </div>
          <h2 class="result-title">
            {{ (lastResult?.netProfit ?? 0) > 0 ? '恭喜获胜！' : (lastResult?.netProfit === 0 ? '保本！' : '很遗憾...') }}
          </h2>
          <div class="result-details">
            <p>落在: <strong>{{ lastResult?.segment }}</strong></p>
            <p>下注: <strong>{{ lastResult?.bet }} 游戏币</strong></p>
            <p v-if="(lastResult?.netProfit ?? 0) > 0" class="win-amount">
              净赢: <strong>+{{ (lastResult?.netProfit ?? 0).toFixed(2) }} 游戏币</strong>
            </p>
            <p v-else-if="lastResult?.netProfit === 0" class="break-even">
              保本: <strong>0 游戏币</strong>
            </p>
            <p v-else class="lose-amount">
              损失: <strong>{{ (lastResult?.netProfit ?? 0).toFixed(2) }} 游戏币</strong>
            </p>
          </div>
          <n-button type="primary" block @click="closeResult">
            继续游戏
          </n-button>
        </div>
      </n-card>
    </n-modal>
  </div>
</template>


<script setup lang="ts">
/**
 * Lucky Wheel Game Page
 * 幸运转盘游戏页面
 * 
 * Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6
 */

import { ref, computed } from 'vue'
import { NButton, NCard, NInputNumber, NModal, NIcon } from 'naive-ui'
import { ArrowBackOutline, WalletOutline } from '@vicons/ionicons5'
import WheelSpinner from '@/components/WheelSpinner.vue'
import { useGameStore } from '@/stores/game'
import { wheelConfig } from '@/config/gameConfig'
import { validateBet, spinWheel, calculatePayout } from '@/utils/gameUtils'

const gameStore = useGameStore()

// Game state
const betAmount = ref(wheelConfig.minBet)
const isSpinning = ref(false)
const targetIndex = ref(0)
const showResult = ref(false)
const lastResult = ref<{
  result: 'win' | 'lose'
  segment: string
  multiplier: number
  bet: number
  payout: number
  netProfit: number
} | null>(null)

// Quick bet amounts
const quickBetAmounts = [1, 5, 10, 20, 50]

// Can spin check
const canSpin = computed(() => {
  if (isSpinning.value) return false
  const validation = validateBet(
    betAmount.value,
    gameStore.gameCoins,
    wheelConfig.minBet,
    wheelConfig.maxBet
  )
  return validation.valid
})

// Spin the wheel
async function spin() {
  if (!canSpin.value) return

  // Validate bet
  const validation = validateBet(
    betAmount.value,
    gameStore.gameCoins,
    wheelConfig.minBet,
    wheelConfig.maxBet
  )
  if (!validation.valid) {
    return
  }

  // Deduct bet amount via API
  const deducted = await gameStore.deductCoins(betAmount.value, 'wheel', '幸运转盘下注')
  if (!deducted) {
    return
  }

  // Start spinning
  isSpinning.value = true
  
  // Determine result
  targetIndex.value = spinWheel(wheelConfig.segments)
  
  // Wait for animation to complete (4 seconds), then stop spinning
  // The WheelSpinner component will emit 'spinEnd' when isSpinning changes to false
  setTimeout(() => {
    isSpinning.value = false
  }, 4000)
}

// Handle spin end
async function onSpinEnd() {
  isSpinning.value = false
  
  const segment = wheelConfig.segments[targetIndex.value]
  if (!segment) return
  
  // Payout = bet * multiplier (total return)
  // e.g., 2x means total return is 2 times your bet
  // Net profit = payout - bet
  const payout = calculatePayout(betAmount.value, segment.multiplier)
  const netProfit = payout - betAmount.value
  // Win if net profit > 0
  const isWin = netProfit > 0
  
  // Add payout if > 0 via API
  if (payout > 0) {
    await gameStore.addCoins(payout, 'wheel', `幸运转盘 - ${segment.label}`)
  }
  
  // Record game result to backend API (Requirements: 1.2)
  // Include segment and multiplier in details
  await gameStore.recordGameResult({
    game_type: 'wheel',
    bet_amount: betAmount.value,
    result: isWin ? 'win' : 'lose',
    payout: payout,
    details: {
      segment: segment.label,
      multiplier: segment.multiplier
    }
  })
  
  // Refresh leaderboard after game completion (Requirements: 4.1)
  gameStore.loadLeaderboard()
  
  // Show result
  lastResult.value = {
    result: isWin ? 'win' : 'lose',
    segment: segment.label,
    multiplier: segment.multiplier,
    bet: betAmount.value,
    payout: payout,
    netProfit: netProfit
  }
  showResult.value = true
}

// Close result modal
function closeResult() {
  showResult.value = false
  lastResult.value = null
}

// Get prize description based on multiplier
function getPrizeDescription(multiplier: number): string {
  if (multiplier === 0) return '输掉全部'
  if (multiplier < 1) return `返还 ${(multiplier * 100).toFixed(0)}%`
  if (multiplier === 1) return '保本'
  // For multiplier > 1, net profit = (multiplier - 1) * bet
  const netMultiplier = multiplier - 1
  return `净赢 ${netMultiplier}x 下注金额`
}
</script>


<style scoped>
.lucky-wheel-page {
  padding: 20px;
  min-height: 100vh;
  position: relative;
  background: 
    radial-gradient(ellipse at 30% 20%, rgba(251, 191, 36, 0.15) 0%, transparent 50%),
    radial-gradient(ellipse at 70% 80%, rgba(139, 92, 246, 0.12) 0%, transparent 50%),
    radial-gradient(ellipse at 50% 50%, rgba(236, 72, 153, 0.08) 0%, transparent 60%);
}

.lucky-wheel-page::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 300px;
  background: linear-gradient(180deg, rgba(251, 191, 36, 0.1) 0%, transparent 100%);
  pointer-events: none;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 30px;
  flex-wrap: wrap;
  gap: 15px;
}

.page-title {
  font-size: 28px;
  font-weight: bold;
  margin: 0;
  background: linear-gradient(135deg, #fbbf24, #f59e0b, #d97706);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 0 30px rgba(251, 191, 36, 0.3);
}

.balance-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.2), rgba(245, 158, 11, 0.15));
  border-radius: 20px;
  border: 1px solid rgba(251, 191, 36, 0.4);
  font-weight: 600;
  color: #fbbf24;
  box-shadow: 0 4px 15px rgba(251, 191, 36, 0.2);
}

.game-content {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 30px;
  max-width: 1200px;
  margin: 0 auto;
}

.wheel-section {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.controls-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.glass-card {
  background: var(--bg-card);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(251, 191, 36, 0.25);
  border-radius: 16px;
  box-shadow: 
    0 4px 20px rgba(251, 191, 36, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.control-card h3,
.prize-card h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
  color: var(--text-primary);
  display: flex;
  align-items: center;
  gap: 8px;
}

.control-card h3::before {
  content: '🎯';
}

.prize-card h3::before {
  content: '💎';
}

.quick-bet-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 15px 0;
}

.bet-info {
  margin-top: 15px;
  padding-top: 15px;
  border-top: 1px solid var(--border-color);
  font-size: 13px;
  color: var(--text-muted);
}

.bet-info p {
  margin: 5px 0;
}

.prize-table {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.prize-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.03), transparent);
  border-radius: 10px;
  border-left: 4px solid;
  transition: all 0.2s ease;
}

.prize-item:hover {
  background: linear-gradient(90deg, rgba(255, 255, 255, 0.08), rgba(255, 255, 255, 0.02));
  transform: translateX(5px);
}

.prize-label {
  font-weight: 600;
  font-size: 14px;
  color: var(--text-primary);
}

.prize-desc {
  font-size: 12px;
  color: var(--text-muted);
}

/* Result Modal */
.result-card {
  text-align: center;
  background: var(--bg-card) !important;
  border: 1px solid rgba(139, 92, 246, 0.3) !important;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.result-content {
  padding: 20px;
}

.result-icon {
  font-size: 64px;
  margin-bottom: 15px;
  animation: bounce 0.6s ease;
}

@keyframes bounce {
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.2); }
}

.result-title {
  font-size: 24px;
  margin: 0 0 20px 0;
  color: var(--text-primary);
}

.win-card .result-title {
  background: linear-gradient(135deg, #10b981, #34d399);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.lose-card .result-title {
  background: linear-gradient(135deg, #ef4444, #f87171);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.result-details {
  margin-bottom: 20px;
  font-size: 16px;
  color: var(--text-secondary);
}

.result-details p {
  margin: 8px 0;
}

.win-amount {
  color: #10b981;
  font-size: 20px;
  text-shadow: 0 0 20px rgba(16, 185, 129, 0.4);
}

.break-even {
  color: #fbbf24;
  font-size: 18px;
}

.lose-amount {
  color: #ef4444;
}

.even-card .result-title {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Responsive */
@media (max-width: 900px) {
  .game-content {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  
  .wheel-section {
    order: 1;
    min-height: 300px;
  }
  
  .controls-section {
    order: 2;
  }
}

@media (max-width: 768px) {
  .lucky-wheel-page {
    padding: 15px;
    -webkit-overflow-scrolling: touch;
  }

  .page-header {
    margin-bottom: 20px;
    gap: 12px;
  }

  .balance-display {
    padding: 8px 16px;
    font-size: 0.9rem;
  }

  /* Touch-friendly quick bet buttons (minimum 44px tap target) */
  .quick-bet-buttons {
    gap: 10px;
  }

  .quick-bet-buttons :deep(.n-button) {
    min-height: 44px;
    min-width: 44px;
    padding: 0.5rem 1rem;
    font-size: 1rem;
  }

  /* Touch-friendly main spin button */
  .control-card :deep(.n-button--primary) {
    min-height: 52px;
    font-size: 1.1rem;
  }

  .prize-item {
    padding: 10px 14px;
  }

  .prize-label {
    font-size: 15px;
  }

  .prize-desc {
    font-size: 13px;
  }
}

@media (max-width: 480px) {
  .lucky-wheel-page {
    padding: 12px;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .page-title {
    font-size: 22px;
    order: 1;
  }

  .page-header :deep(.n-button) {
    order: 0;
    min-height: 40px;
  }

  .balance-display {
    order: 2;
    width: 100%;
    justify-content: center;
  }
  
  .quick-bet-buttons {
    justify-content: center;
    flex-wrap: wrap;
  }

  .quick-bet-buttons :deep(.n-button) {
    flex: 1;
    min-width: 60px;
    max-width: 80px;
  }

  .control-card h3,
  .prize-card h3 {
    font-size: 16px;
  }

  .bet-info {
    font-size: 12px;
  }

  .wheel-section {
    min-height: 250px;
  }

  /* Result modal responsive */
  .result-card {
    width: 90vw !important;
    max-width: 360px;
  }

  .result-icon {
    font-size: 48px;
  }

  .result-title {
    font-size: 20px;
  }

  .result-details {
    font-size: 14px;
  }

  .win-amount,
  .lose-amount {
    font-size: 18px;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .prize-item:hover {
    background: rgba(255, 255, 255, 0.03);
  }

  .prize-item:active {
    background: rgba(255, 255, 255, 0.08);
  }

  /* Larger touch targets for input */
  .control-card :deep(.n-input-number) {
    min-height: 48px;
  }

  .control-card :deep(.n-input-number .n-input__input-el) {
    font-size: 18px;
  }
}
</style>
