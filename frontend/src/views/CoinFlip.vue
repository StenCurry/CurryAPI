<template>
  <div class="coin-flip-page">
    <!-- Header -->
    <div class="page-header">
      <n-button quaternary @click="$router.push('/games')">
        <template #icon>
          <n-icon><ArrowBackOutline /></n-icon>
        </template>
        返回游戏中心
      </n-button>
      <h1 class="page-title">🪙 硬币翻转</h1>
      <div class="balance-display">
        <n-icon><WalletOutline /></n-icon>
        <span>{{ gameStore.gameCoins.toFixed(2) }} 游戏币</span>
      </div>
    </div>

    <!-- Main Content -->
    <div class="game-content">
      <!-- Coin Section -->
      <div class="coin-section">
        <CoinAnimation
          :flipping="isFlipping"
          :result="coinResult"
          @flip-end="onFlipEnd"
        />
        
        <!-- Choice Selection -->
        <div class="choice-section" v-if="!isFlipping">
          <h3>选择你的预测</h3>
          <div class="choice-buttons">
            <n-button
              :type="selectedChoice === 'heads' ? 'primary' : 'default'"
              size="large"
              :disabled="isFlipping"
              @click="selectedChoice = 'heads'"
              class="choice-btn"
            >
              <span class="choice-icon">👑</span>
              <span>正面</span>
            </n-button>
            <n-button
              :type="selectedChoice === 'tails' ? 'primary' : 'default'"
              size="large"
              :disabled="isFlipping"
              @click="selectedChoice = 'tails'"
              class="choice-btn"
            >
              <span class="choice-icon">🌙</span>
              <span>反面</span>
            </n-button>
          </div>
        </div>
      </div>

      <!-- Controls Section -->
      <div class="controls-section">
        <n-card class="control-card glass-card">
          <h3>下注金额</h3>
          <n-input-number
            v-model:value="betAmount"
            :min="coinFlipConfig.minBet"
            :max="Math.min(coinFlipConfig.maxBet, gameStore.gameCoins)"
            :disabled="isFlipping"
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
              :disabled="isFlipping || amount > gameStore.gameCoins"
              @click="betAmount = amount"
            >
              {{ amount }}
            </n-button>
            <n-button 
              size="small"
              :disabled="isFlipping || gameStore.gameCoins <= 0"
              @click="betAmount = Math.min(coinFlipConfig.maxBet, gameStore.gameCoins)"
            >
              最大
            </n-button>
          </div>

          <div class="potential-win" v-if="betAmount > 0">
            <span>潜在净赢:</span>
            <span class="win-amount">{{ (potentialWin - betAmount).toFixed(2) }} 游戏币</span>
          </div>

          <n-button
            type="primary"
            size="large"
            block
            :disabled="!canFlip"
            :loading="isFlipping"
            @click="flipCoin"
          >
            {{ isFlipping ? '翻转中...' : '开始翻转' }}
          </n-button>

          <div class="bet-info">
            <p>最低下注: {{ coinFlipConfig.minBet }} 游戏币</p>
            <p>最高下注: {{ coinFlipConfig.maxBet }} 游戏币</p>
            <p>赔率: {{ coinFlipConfig.payoutMultiplier }}x</p>
          </div>
        </n-card>

        <!-- Game Rules -->
        <n-card class="rules-card glass-card">
          <h3>游戏规则</h3>
          <ul class="rules-list">
            <li>选择正面或反面</li>
            <li>输入下注金额</li>
            <li>点击翻转硬币</li>
            <li>猜对获得 {{ coinFlipConfig.payoutMultiplier }}x 下注金额</li>
            <li>50/50 的获胜概率</li>
          </ul>
        </n-card>
      </div>
    </div>

    <!-- Result Modal -->
    <n-modal v-model:show="showResult" :mask-closable="false">
      <n-card
        class="result-card"
        :class="{ 'win-card': lastResult?.isWin, 'lose-card': !lastResult?.isWin }"
        style="width: 400px"
      >
        <div class="result-content">
          <div class="result-icon">
            {{ lastResult?.isWin ? '🎉' : '😢' }}
          </div>
          <h2 class="result-title">
            {{ lastResult?.isWin ? '恭喜获胜！' : '很遗憾...' }}
          </h2>
          <div class="result-details">
            <p>
              硬币结果: 
              <strong>{{ lastResult?.coinResult === 'heads' ? '👑 正面' : '🌙 反面' }}</strong>
            </p>
            <p>
              你的选择: 
              <strong>{{ lastResult?.choice === 'heads' ? '👑 正面' : '🌙 反面' }}</strong>
            </p>
            <p>下注: <strong>{{ lastResult?.bet }} 游戏币</strong></p>
            <p v-if="lastResult?.isWin" class="win-amount">
              净赢: <strong>+{{ (lastResult?.payout - lastResult?.bet).toFixed(2) }} 游戏币</strong>
            </p>
            <p v-else class="lose-amount">
              损失: <strong>-{{ lastResult?.bet }} 游戏币</strong>
            </p>
          </div>
          <p v-if="!lastResult?.isWin" class="consolation-message">
            再试一次，好运就在下一把！
          </p>
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
 * Coin Flip Game Page
 * 硬币翻转游戏页面
 * 
 * Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7
 */

import { ref, computed } from 'vue'
import { NButton, NCard, NInputNumber, NModal, NIcon } from 'naive-ui'
import { ArrowBackOutline, WalletOutline } from '@vicons/ionicons5'
import CoinAnimation from '@/components/CoinAnimation.vue'
import type { CoinResult } from '@/components/CoinAnimation.vue'
import { useGameStore } from '@/stores/game'
import { coinFlipConfig } from '@/config/gameConfig'
import { validateBet, randomBoolean, calculatePayout } from '@/utils/gameUtils'

const gameStore = useGameStore()

// Game state
const betAmount = ref(coinFlipConfig.minBet)
const selectedChoice = ref<CoinResult>('heads')
const isFlipping = ref(false)
const coinResult = ref<CoinResult | undefined>(undefined)
const showResult = ref(false)
const lastResult = ref<{
  isWin: boolean
  choice: CoinResult
  coinResult: CoinResult
  bet: number
  payout: number
} | null>(null)

// Quick bet amounts
const quickBetAmounts = [1, 5, 10, 25, 50, 100]

// Calculate potential win
const potentialWin = computed(() => {
  return calculatePayout(betAmount.value, coinFlipConfig.payoutMultiplier)
})

// Can flip check
const canFlip = computed(() => {
  if (isFlipping.value) return false
  if (!selectedChoice.value) return false
  const validation = validateBet(
    betAmount.value,
    gameStore.gameCoins,
    coinFlipConfig.minBet,
    coinFlipConfig.maxBet
  )
  return validation.valid
})

// Flip the coin
async function flipCoin() {
  if (!canFlip.value) return

  // Validate bet
  const validation = validateBet(
    betAmount.value,
    gameStore.gameCoins,
    coinFlipConfig.minBet,
    coinFlipConfig.maxBet
  )
  if (!validation.valid) {
    return
  }

  // Deduct bet amount via API
  const deducted = await gameStore.deductCoins(betAmount.value, 'coin', '硬币翻转下注')
  if (!deducted) {
    return
  }

  // Start flipping
  isFlipping.value = true
  
  // Determine result (50/50 chance)
  coinResult.value = randomBoolean() ? 'heads' : 'tails'
  
  // Wait for animation to complete (2 seconds)
  setTimeout(() => {
    onFlipEnd()
  }, 2000)
}

// Handle flip end
async function onFlipEnd() {
  isFlipping.value = false
  
  if (!coinResult.value || !selectedChoice.value) return
  
  const isWin = coinResult.value === selectedChoice.value
  const payout = isWin ? calculatePayout(betAmount.value, coinFlipConfig.payoutMultiplier) : 0
  
  // Add winnings if won via API
  if (isWin && payout > 0) {
    const choiceText = selectedChoice.value === 'heads' ? '正面' : '反面'
    await gameStore.addCoins(payout, 'coin', `硬币翻转获胜 - ${choiceText}`)
  }
  
  // Record game result to backend API (Requirements: 1.3)
  // Include choice and coin_result in details
  await gameStore.recordGameResult({
    game_type: 'coin',
    bet_amount: betAmount.value,
    result: isWin ? 'win' : 'lose',
    payout: payout,
    details: {
      choice: selectedChoice.value,
      coin_result: coinResult.value
    }
  })
  
  // Refresh leaderboard after game completion (Requirements: 4.1)
  gameStore.loadLeaderboard()
  
  // Show result
  lastResult.value = {
    isWin,
    choice: selectedChoice.value,
    coinResult: coinResult.value,
    bet: betAmount.value,
    payout
  }
  showResult.value = true
}

// Close result modal
function closeResult() {
  showResult.value = false
  lastResult.value = null
  coinResult.value = undefined
}
</script>


<style scoped>
.coin-flip-page {
  padding: 20px;
  min-height: 100vh;
  position: relative;
  background: 
    radial-gradient(ellipse at 20% 30%, rgba(236, 72, 153, 0.15) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 70%, rgba(245, 87, 108, 0.12) 0%, transparent 50%),
    radial-gradient(ellipse at 50% 50%, rgba(240, 147, 251, 0.08) 0%, transparent 60%);
}

.coin-flip-page::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 300px;
  background: linear-gradient(180deg, rgba(236, 72, 153, 0.1) 0%, transparent 100%);
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
  background: linear-gradient(135deg, #f093fb, #f5576c, #ec4899);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  text-shadow: 0 0 30px rgba(236, 72, 153, 0.3);
}

.balance-display {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: linear-gradient(135deg, rgba(236, 72, 153, 0.2), rgba(245, 87, 108, 0.15));
  border-radius: 20px;
  border: 1px solid rgba(236, 72, 153, 0.4);
  font-weight: 600;
  color: #ec4899;
  box-shadow: 0 4px 15px rgba(236, 72, 153, 0.2);
}

.game-content {
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 30px;
  max-width: 1200px;
  margin: 0 auto;
}

.coin-section {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  min-height: 400px;
}

.choice-section {
  text-align: center;
  margin-top: 20px;
}

.choice-section h3 {
  margin: 0 0 15px 0;
  color: var(--text-primary);
  font-size: 18px;
}

.choice-buttons {
  display: flex;
  gap: 20px;
  justify-content: center;
}

.choice-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 20px 40px;
  min-width: 120px;
  border-radius: 16px;
  transition: all 0.3s ease;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.05), rgba(255, 255, 255, 0.02));
  border: 2px solid rgba(255, 255, 255, 0.1);
}

.choice-btn:hover:not(:disabled) {
  transform: translateY(-3px);
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.choice-btn:first-child {
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.15), rgba(245, 158, 11, 0.1));
  border-color: rgba(251, 191, 36, 0.3);
}

.choice-btn:first-child:hover:not(:disabled),
.choice-btn:first-child.n-button--primary-type {
  background: linear-gradient(135deg, rgba(251, 191, 36, 0.3), rgba(245, 158, 11, 0.2));
  border-color: #fbbf24;
  box-shadow: 0 10px 30px rgba(251, 191, 36, 0.3);
}

.choice-btn:last-child {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.15), rgba(124, 58, 237, 0.1));
  border-color: rgba(139, 92, 246, 0.3);
}

.choice-btn:last-child:hover:not(:disabled),
.choice-btn:last-child.n-button--primary-type {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.3), rgba(124, 58, 237, 0.2));
  border-color: #8b5cf6;
  box-shadow: 0 10px 30px rgba(139, 92, 246, 0.3);
}

.choice-icon {
  font-size: 32px;
}

.controls-section {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.glass-card {
  background: var(--bg-card);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(236, 72, 153, 0.25);
  border-radius: 16px;
  box-shadow: 
    0 4px 20px rgba(236, 72, 153, 0.1),
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
  content: '💰';
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

/* Result Modal */
.result-card {
  text-align: center;
  background: var(--bg-card) !important;
  border: 1px solid rgba(236, 72, 153, 0.3) !important;
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

.result-details .win-amount {
  color: #10b981;
  font-size: 20px;
  text-shadow: 0 0 15px rgba(16, 185, 129, 0.4);
}

.result-details .lose-amount {
  color: #ef4444;
}

.consolation-message {
  color: var(--text-muted);
  font-style: italic;
  margin-bottom: 20px;
}

/* Responsive */
@media (max-width: 900px) {
  .game-content {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  
  .coin-section {
    order: 1;
    min-height: 300px;
  }
  
  .controls-section {
    order: 2;
  }
}

@media (max-width: 768px) {
  .coin-flip-page {
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

  /* Touch-friendly choice buttons (minimum 44px tap target) */
  .choice-buttons {
    gap: 15px;
  }

  .choice-btn {
    min-height: 60px;
    padding: 16px 32px;
  }

  /* Touch-friendly quick bet buttons */
  .quick-bet-buttons {
    gap: 10px;
  }

  .quick-bet-buttons :deep(.n-button) {
    min-height: 44px;
    min-width: 44px;
    padding: 0.5rem 1rem;
    font-size: 1rem;
  }

  /* Touch-friendly main flip button */
  .control-card :deep(.n-button--primary) {
    min-height: 52px;
    font-size: 1.1rem;
  }

  .potential-win {
    padding: 14px;
  }

  .potential-win .win-amount {
    font-size: 16px;
  }

  .rules-list {
    font-size: 13px;
    line-height: 1.6;
  }
}

@media (max-width: 480px) {
  .coin-flip-page {
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
  
  .choice-buttons {
    flex-direction: column;
    gap: 10px;
    width: 100%;
  }
  
  .choice-btn {
    width: 100%;
    flex-direction: row;
    justify-content: center;
    padding: 15px 20px;
    min-height: 56px;
  }

  .choice-icon {
    font-size: 28px;
  }
  
  .quick-bet-buttons {
    justify-content: center;
    flex-wrap: wrap;
  }

  .quick-bet-buttons :deep(.n-button) {
    flex: 1;
    min-width: 50px;
    max-width: 70px;
  }

  .control-card h3,
  .rules-card h3 {
    font-size: 16px;
  }

  .bet-info {
    font-size: 12px;
  }

  .coin-section {
    min-height: 250px;
  }

  .choice-section h3 {
    font-size: 16px;
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

  .result-details .win-amount {
    font-size: 18px;
  }

  .consolation-message {
    font-size: 13px;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .choice-btn:hover {
    transform: none;
  }

  .choice-btn:active {
    transform: scale(0.98);
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
