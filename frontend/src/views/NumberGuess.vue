<template>
  <div class="number-guess-page">
    <!-- Header -->
    <div class="page-header">
      <n-button quaternary @click="$router.push('/games')">
        <template #icon>
          <n-icon><ArrowBackOutline /></n-icon>
        </template>
        返回游戏中心
      </n-button>
      <h1 class="page-title">🎲 猜大小</h1>
      <div class="balance-display">
        <n-icon><WalletOutline /></n-icon>
        <span>{{ gameStore.gameCoins.toFixed(2) }} 游戏币</span>
      </div>
    </div>

    <!-- Main Content -->
    <div class="game-content">
      <!-- Game Section -->
      <div class="game-section">
        <!-- Current Number Display -->
        <div class="number-display-area">
          <div class="number-box" :class="{ revealing: isRevealing, win: showResult && lastResult?.isWin, lose: showResult && !lastResult?.isWin }">
            <span v-if="!isPlaying && !showResult" class="question-mark">?</span>
            <span v-else-if="isRevealing" class="rolling-number">{{ rollingNumber }}</span>
            <span v-else class="final-number">{{ lastResult?.target }}</span>
          </div>
          <div class="midpoint-hint">
            中间值: <strong>{{ numberGuessConfig.midPoint }}</strong>
          </div>
        </div>

        <!-- Choice Buttons -->
        <div class="choice-section">
          <h3 class="section-label">猜下一个数字是大还是小？</h3>
          <div class="choice-buttons">
            <button
              class="choice-btn small-btn"
              :class="{ selected: selectedChoice === 'small', disabled: isPlaying }"
              :disabled="isPlaying"
              @click="selectedChoice = 'small'"
            >
              <span class="choice-icon">⬇️</span>
              <span class="choice-text">小</span>
              <span class="choice-range">1 - {{ numberGuessConfig.midPoint - 1 }}</span>
            </button>
            <button
              class="choice-btn big-btn"
              :class="{ selected: selectedChoice === 'big', disabled: isPlaying }"
              :disabled="isPlaying"
              @click="selectedChoice = 'big'"
            >
              <span class="choice-icon">⬆️</span>
              <span class="choice-text">大</span>
              <span class="choice-range">{{ numberGuessConfig.midPoint + 1 }} - {{ numberGuessConfig.range.max }}</span>
            </button>
          </div>
          <div class="special-hint">
            💡 如果刚好是 {{ numberGuessConfig.midPoint }}，算你赢！赔率 {{ specialPayoutMultiplier }}x
          </div>
        </div>

        <!-- Result Display -->
        <div v-if="showResult" class="result-display" :class="{ win: lastResult?.isWin, lose: !lastResult?.isWin }">
          <div class="result-message">
            <span v-if="lastResult?.isExactMid">🎯 正中靶心！</span>
            <span v-else-if="lastResult?.isWin">🎉 猜对了！</span>
            <span v-else>😢 猜错了，再来一次！</span>
          </div>
          <div class="result-details">
            <span>数字 {{ lastResult?.target }} 是</span>
            <strong :class="lastResult?.actualSide">
              {{ lastResult?.target === numberGuessConfig.midPoint ? '中间值' : (lastResult?.actualSide === 'small' ? '小' : '大') }}
            </strong>
          </div>
          <div class="result-payout" :class="{ win: lastResult?.isWin }">
            {{ lastResult?.isWin ? '+' : '-' }}{{ (lastResult?.isWin ? (lastResult?.payout ?? 0) - (lastResult?.bet ?? 0) : lastResult?.bet ?? 0).toFixed(2) }} 游戏币
          </div>
        </div>
      </div>

      <!-- Controls Section -->
      <div class="controls-section">
        <n-card class="control-card glass-card">
          <h3>下注金额</h3>
          <n-input-number
            v-model:value="betAmount"
            :min="numberGuessConfig.minBet"
            :max="Math.min(numberGuessConfig.maxBet, gameStore.gameCoins)"
            :disabled="isPlaying"
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
              :disabled="isPlaying || amount > gameStore.gameCoins"
              @click="betAmount = amount"
            >
              {{ amount }}
            </n-button>
            <n-button 
              size="small"
              :disabled="isPlaying || gameStore.gameCoins <= 0"
              @click="betAmount = Math.min(numberGuessConfig.maxBet, gameStore.gameCoins)"
            >
              最大
            </n-button>
          </div>

          <div class="potential-win" v-if="betAmount > 0">
            <span>猜中净赢:</span>
            <span class="win-amount">{{ (potentialWin - betAmount).toFixed(2) }} 游戏币</span>
          </div>

          <n-button
            type="primary"
            size="large"
            block
            :disabled="!canPlay"
            :loading="isPlaying"
            @click="playGame"
          >
            {{ isPlaying ? '开奖中...' : '确认下注' }}
          </n-button>

          <div class="bet-info">
            <p>最低下注: {{ numberGuessConfig.minBet }} 游戏币</p>
            <p>最高下注: {{ numberGuessConfig.maxBet }} 游戏币</p>
            <p>普通赔率: {{ numberGuessConfig.payoutMultiplier }}x</p>
            <p>中间值赔率: {{ specialPayoutMultiplier }}x</p>
          </div>
        </n-card>

        <!-- Game Rules -->
        <n-card class="rules-card glass-card">
          <h3>游戏规则</h3>
          <ul class="rules-list">
            <li>系统随机生成 1-{{ numberGuessConfig.range.max }} 的数字</li>
            <li>猜数字是大于还是小于 {{ numberGuessConfig.midPoint }}</li>
            <li>猜对获得 {{ numberGuessConfig.payoutMultiplier }}x 下注金额</li>
            <li>如果正好是 {{ numberGuessConfig.midPoint }}，获得 {{ specialPayoutMultiplier }}x</li>
            <li>获胜概率: ~50%</li>
          </ul>
        </n-card>
      </div>
    </div>
  </div>
</template>


<script setup lang="ts">
/**
 * Number Guess Game Page - 猜大小模式
 * 猜数字游戏页面
 * 
 * Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7
 */

import { ref, computed } from 'vue'
import { NButton, NCard, NInputNumber, NIcon } from 'naive-ui'
import { ArrowBackOutline, WalletOutline } from '@vicons/ionicons5'
import { useGameStore } from '@/stores/game'
import { numberGuessConfig } from '@/config/gameConfig'
import { validateBet, randomInt, calculatePayout } from '@/utils/gameUtils'

const gameStore = useGameStore()

// Special payout for hitting exact midpoint
const specialPayoutMultiplier = 5

// Game state
const betAmount = ref(numberGuessConfig.minBet)
const selectedChoice = ref<'big' | 'small' | null>(null)
const isPlaying = ref(false)
const isRevealing = ref(false)
const showResult = ref(false)
const rollingNumber = ref(0)
const lastResult = ref<{
  isWin: boolean
  isExactMid: boolean
  target: number
  actualSide: 'big' | 'small' | 'mid'
  bet: number
  payout: number
} | null>(null)

// Quick bet amounts
const quickBetAmounts = [1, 5, 10, 20, 50]

// Calculate potential win
const potentialWin = computed(() => {
  return calculatePayout(betAmount.value, numberGuessConfig.payoutMultiplier)
})

// Can play check
const canPlay = computed(() => {
  if (isPlaying.value) return false
  if (selectedChoice.value === null) return false
  const validation = validateBet(
    betAmount.value,
    gameStore.gameCoins,
    numberGuessConfig.minBet,
    numberGuessConfig.maxBet
  )
  return validation.valid
})

// Rolling number animation
function startRollingAnimation(finalNumber: number, callback: () => void) {
  isRevealing.value = true
  let count = 0
  const maxCount = 20
  const interval = setInterval(() => {
    rollingNumber.value = randomInt(1, numberGuessConfig.range.max)
    count++
    if (count >= maxCount) {
      clearInterval(interval)
      rollingNumber.value = finalNumber
      setTimeout(() => {
        isRevealing.value = false
        callback()
      }, 300)
    }
  }, 80)
}

// Play the game
async function playGame() {
  if (!canPlay.value || selectedChoice.value === null) return

  // Validate bet
  const validation = validateBet(
    betAmount.value,
    gameStore.gameCoins,
    numberGuessConfig.minBet,
    numberGuessConfig.maxBet
  )
  if (!validation.valid) {
    return
  }

  // Deduct bet amount via API
  const choiceText = selectedChoice.value === 'big' ? '大' : '小'
  const deducted = await gameStore.deductCoins(betAmount.value, 'number', `猜大小下注 - ${choiceText}`)
  if (!deducted) {
    return
  }

  // Start playing
  isPlaying.value = true
  showResult.value = false
  
  // Generate random target number
  const targetNumber = randomInt(numberGuessConfig.range.min, numberGuessConfig.range.max)
  
  // Start rolling animation
  startRollingAnimation(targetNumber, async () => {
    // Determine actual side
    let actualSide: 'big' | 'small' | 'mid'
    if (targetNumber === numberGuessConfig.midPoint) {
      actualSide = 'mid'
    } else if (targetNumber > numberGuessConfig.midPoint) {
      actualSide = 'big'
    } else {
      actualSide = 'small'
    }
    
    // Check if won
    const isExactMid = targetNumber === numberGuessConfig.midPoint
    const isWin = isExactMid || actualSide === selectedChoice.value
    
    // Calculate payout
    let payout = 0
    if (isWin) {
      if (isExactMid) {
        payout = calculatePayout(betAmount.value, specialPayoutMultiplier)
      } else {
        payout = calculatePayout(betAmount.value, numberGuessConfig.payoutMultiplier)
      }
      const resultText = isExactMid ? '正中靶心' : `猜中${actualSide === 'big' ? '大' : '小'}`
      await gameStore.addCoins(payout, 'number', `猜大小获胜 - ${resultText}`)
    }
    
    // Record game result to backend API (Requirements: 1.4)
    // Include choice, target, actual_side, is_exact_mid in details
    if (selectedChoice.value) {
      await gameStore.recordGameResult({
        game_type: 'number',
        bet_amount: betAmount.value,
        result: isWin ? 'win' : 'lose',
        payout: payout,
        details: {
          choice: selectedChoice.value,
          target: targetNumber,
          actual_side: actualSide,
          is_exact_mid: isExactMid
        }
      })
      
      // Refresh leaderboard after game completion (Requirements: 4.1)
      gameStore.loadLeaderboard()
    }
    
    // Show result
    lastResult.value = {
      isWin,
      isExactMid,
      target: targetNumber,
      actualSide,
      bet: betAmount.value,
      payout
    }
    showResult.value = true
    isPlaying.value = false
  })
}
</script>


<style scoped>
.number-guess-page {
  padding: 20px;
  min-height: 100vh;
  position: relative;
  background: 
    radial-gradient(ellipse at 25% 25%, rgba(59, 130, 246, 0.15) 0%, transparent 50%),
    radial-gradient(ellipse at 75% 75%, rgba(79, 172, 254, 0.12) 0%, transparent 50%),
    radial-gradient(ellipse at 50% 50%, rgba(0, 242, 254, 0.08) 0%, transparent 60%);
}

.number-guess-page::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 300px;
  background: linear-gradient(180deg, rgba(59, 130, 246, 0.1) 0%, transparent 100%);
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
  background: linear-gradient(135deg, #4facfe, #00f2fe, #3b82f6);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 0 30px rgba(59, 130, 246, 0.3);
}

.balance-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(79, 172, 254, 0.15));
  border-radius: 20px;
  border: 1px solid rgba(59, 130, 246, 0.4);
  font-weight: 600;
  color: #3b82f6;
  box-shadow: 0 4px 15px rgba(59, 130, 246, 0.2);
}

.game-content {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 30px;
  max-width: 1200px;
  margin: 0 auto;
}

.game-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 30px;
  background: var(--bg-card);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(59, 130, 246, 0.25);
  border-radius: 20px;
  box-shadow: 
    0 4px 20px rgba(59, 130, 246, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

/* Number Display */
.number-display-area {
  text-align: center;
  margin-bottom: 30px;
}

.number-box {
  width: 150px;
  height: 150px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 64px;
  font-weight: bold;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.1), rgba(79, 172, 254, 0.05));
  border: 4px solid rgba(59, 130, 246, 0.3);
  border-radius: 20px;
  margin: 0 auto 15px;
  transition: all 0.3s ease;
  color: var(--text-primary);
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.number-box.revealing {
  animation: shake 0.1s infinite;
  border-color: #fbbf24;
  box-shadow: 
    0 0 30px rgba(251, 191, 36, 0.4),
    0 0 60px rgba(251, 191, 36, 0.2);
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.2), rgba(245, 158, 11, 0.1));
}

.number-box.win {
  border-color: #10b981;
  box-shadow: 
    0 0 30px rgba(16, 185, 129, 0.4),
    0 0 60px rgba(16, 185, 129, 0.2);
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.2), rgba(52, 211, 153, 0.1));
}

.number-box.lose {
  border-color: #ef4444;
  box-shadow: 
    0 0 30px rgba(239, 68, 68, 0.4),
    0 0 60px rgba(239, 68, 68, 0.2);
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.2), rgba(248, 113, 113, 0.1));
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-3px); }
  75% { transform: translateX(3px); }
}

.question-mark {
  color: rgba(59, 130, 246, 0.5);
}

.rolling-number {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.final-number {
  color: var(--text-primary);
}

.midpoint-hint {
  color: var(--text-muted);
  font-size: 14px;
}

.midpoint-hint strong {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

/* Choice Section */
.choice-section {
  width: 100%;
  max-width: 500px;
}

.section-label {
  color: var(--text-primary);
  font-size: 18px;
  margin: 0 0 20px 0;
  text-align: center;
}

.choice-buttons {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
  margin-bottom: 15px;
}

.choice-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 25px 20px;
  border: 3px solid transparent;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.3s ease;
  min-height: 120px;
}

.small-btn {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.2), rgba(79, 172, 254, 0.15));
  border-color: rgba(59, 130, 246, 0.4);
}

.small-btn:hover:not(.disabled) {
  border-color: #3b82f6;
  transform: translateY(-3px);
  box-shadow: 0 10px 30px rgba(59, 130, 246, 0.3);
}

.small-btn.selected {
  border-color: #3b82f6;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.35), rgba(79, 172, 254, 0.25));
  box-shadow: 
    0 0 30px rgba(59, 130, 246, 0.4),
    inset 0 0 20px rgba(59, 130, 246, 0.1);
}

.big-btn {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.2), rgba(248, 113, 113, 0.15));
  border-color: rgba(239, 68, 68, 0.4);
}

.big-btn:hover:not(.disabled) {
  border-color: #ef4444;
  transform: translateY(-3px);
  box-shadow: 0 10px 30px rgba(239, 68, 68, 0.3);
}

.big-btn.selected {
  border-color: #ef4444;
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.35), rgba(248, 113, 113, 0.25));
  box-shadow: 
    0 0 30px rgba(239, 68, 68, 0.4),
    inset 0 0 20px rgba(239, 68, 68, 0.1);
}

.choice-btn.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.choice-icon {
  font-size: 32px;
}

.choice-text {
  font-size: 24px;
  font-weight: bold;
  color: var(--text-primary);
}

.choice-range {
  font-size: 12px;
  color: var(--text-muted);
}

.special-hint {
  text-align: center;
  font-size: 13px;
  color: #fbbf24;
  padding: 12px;
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.15), rgba(245, 158, 11, 0.1));
  border-radius: 10px;
  border: 1px solid rgba(251, 191, 36, 0.3);
}

/* Result Display */
.result-display {
  margin-top: 25px;
  padding: 20px;
  border-radius: 16px;
  text-align: center;
  width: 100%;
  max-width: 400px;
}

.result-display.win {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.2), rgba(52, 211, 153, 0.15));
  border: 1px solid rgba(16, 185, 129, 0.4);
  box-shadow: 0 10px 40px rgba(16, 185, 129, 0.2);
}

.result-display.lose {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.2), rgba(248, 113, 113, 0.15));
  border: 1px solid rgba(239, 68, 68, 0.4);
  box-shadow: 0 10px 40px rgba(239, 68, 68, 0.2);
}

.result-message {
  font-size: 22px;
  font-weight: 600;
  margin-bottom: 10px;
}

.result-display.win .result-message {
  background: linear-gradient(135deg, #10b981, #34d399);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.result-display.lose .result-message {
  background: linear-gradient(135deg, #ef4444, #f87171);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.result-details {
  font-size: 16px;
  color: var(--text-muted);
  margin-bottom: 10px;
}

.result-details strong.small {
  color: #3b82f6;
}

.result-details strong.big {
  color: #ef4444;
}

.result-details strong.mid {
  background: linear-gradient(135deg, #fbbf24, #f59e0b);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.result-payout {
  font-size: 24px;
  font-weight: bold;
  color: #ef4444;
}

.result-payout.win {
  color: #10b981;
  text-shadow: 0 0 20px rgba(16, 185, 129, 0.4);
}

/* Controls Section */
.controls-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.glass-card {
  background: var(--bg-card);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(59, 130, 246, 0.25);
  border-radius: 16px;
  box-shadow: 
    0 4px 20px rgba(59, 130, 246, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.control-card h3,
.rules-card h3 {
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

.rules-card h3::before {
  content: '📋';
}

.quick-bet-buttons {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 15px 0;
}

.potential-win {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  margin: 15px 0;
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.2), rgba(52, 211, 153, 0.15));
  border-radius: 8px;
  border: 1px solid rgba(16, 185, 129, 0.4);
}

.potential-win .win-amount {
  color: #10b981;
  font-weight: 600;
  font-size: 18px;
  text-shadow: 0 0 10px rgba(16, 185, 129, 0.3);
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

.rules-list {
  margin: 0;
  padding-left: 20px;
  color: var(--text-muted);
  font-size: 14px;
  line-height: 1.8;
}

.rules-list li {
  margin: 5px 0;
}

/* Responsive */
@media (max-width: 900px) {
  .game-content {
    grid-template-columns: 1fr;
    gap: 20px;
  }
}

@media (max-width: 768px) {
  .number-guess-page {
    padding: 15px;
  }

  .number-box {
    width: 120px;
    height: 120px;
    font-size: 48px;
  }

  .choice-btn {
    padding: 20px 15px;
    min-height: 100px;
  }

  .choice-icon {
    font-size: 28px;
  }

  .choice-text {
    font-size: 20px;
  }

  .quick-bet-buttons :deep(.n-button) {
    min-height: 44px;
    min-width: 44px;
  }

  .control-card :deep(.n-button--primary) {
    min-height: 52px;
    font-size: 1.1rem;
  }
}

@media (max-width: 480px) {
  .number-guess-page {
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

  .balance-display {
    order: 2;
    width: 100%;
    justify-content: center;
  }

  .number-box {
    width: 100px;
    height: 100px;
    font-size: 40px;
  }

  .choice-buttons {
    gap: 12px;
  }

  .choice-btn {
    padding: 15px 10px;
    min-height: 90px;
  }

  .choice-icon {
    font-size: 24px;
  }

  .choice-text {
    font-size: 18px;
  }

  .choice-range {
    font-size: 11px;
  }

  .special-hint {
    font-size: 12px;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .choice-btn:hover:not(.disabled) {
    transform: none;
  }

  .choice-btn:active:not(.disabled) {
    transform: scale(0.98);
  }

  .control-card :deep(.n-input-number) {
    min-height: 48px;
  }
}
</style>
