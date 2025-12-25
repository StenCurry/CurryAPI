<template>
  <div class="game-center">
    <!-- Balance Section -->
    <div class="balance-section">
      <div class="balance-card glass-card">
        <div class="balance-icon">🎮</div>
        <div class="balance-info">
          <span class="balance-label">游戏币余额</span>
          <span class="balance-value">{{ gameStore.gameCoins.toFixed(2) }}</span>
        </div>
        <div class="balance-actions">
          <n-button 
            v-if="gameStore.gameCoins < 10" 
            type="warning" 
            size="small"
            :loading="gameStore.loading"
            :disabled="gameStore.loading"
            @click="handleResetCoins"
          >
            重置游戏币
          </n-button>
        </div>
      </div>
      
      <!-- Low Balance Warning -->
      <div v-if="gameStore.gameCoins < 10" class="low-balance-warning">
        <n-alert type="warning" :bordered="false">
          <template #icon>
            <span>💡</span>
          </template>
          游戏币不足！您可以重置游戏币获得 {{ INITIAL_GAME_COINS }} 游戏币，或通过邀请好友获得更多奖励。
        </n-alert>
      </div>
    </div>

    <!-- Games Grid -->
    <div class="section-header">
      <h2 class="section-title">🎰 可用游戏</h2>
    </div>
    <div class="games-grid">
      <GameCard
        v-for="game in games"
        :key="game.route"
        :name="game.name"
        :description="game.description"
        :icon="game.icon"
        :min-bet="game.minBet"
        :route="game.route"
        :gradient="game.gradient"
      />
    </div>

    <!-- Two Column Layout: Stats and Leaderboard -->
    <div class="two-column-section">
      <!-- Stats Section -->
      <div class="stats-column">
        <div class="section-header">
          <h2 class="section-title">📊 游戏统计</h2>
        </div>
        <div class="stats-grid">
          <StatCard
            title="总游戏次数"
            :value="gameStore.stats.totalGames"
            icon="🎲"
            color="primary"
          />
          <StatCard
            title="胜利次数"
            :value="gameStore.stats.wins"
            icon="🏆"
            color="success"
          />
          <StatCard
            title="失败次数"
            :value="gameStore.stats.losses"
            icon="💔"
            color="error"
          />
          <StatCard
            title="胜率"
            :value="`${gameStore.stats.winRate}%`"
            icon="📈"
            color="warning"
          />
          <StatCard
            title="净收益"
            :value="gameStore.stats.netProfit"
            icon="💰"
            :color="Number(gameStore.stats.netProfit) >= 0 ? 'success' : 'error'"
          />
        </div>
      </div>

      <!-- Leaderboard Section -->
      <div class="leaderboard-column">
        <div class="section-header">
          <h2 class="section-title">🏆 排行榜</h2>
        </div>
        <GameLeaderboard ref="leaderboardRef" />
      </div>
    </div>

    <!-- Game History Section -->
    <div class="section-header">
      <h2 class="section-title">📜 最近游戏记录</h2>
    </div>
    <div class="history-section glass-card">
      <!-- Loading state -->
      <div v-if="historyLoading" class="empty-history">
        <span class="empty-icon">⏳</span>
        <p>加载游戏记录中...</p>
      </div>
      <!-- Empty state (Requirements: 6.3) -->
      <div v-else-if="recentHistory.length === 0" class="empty-history">
        <span class="empty-icon">🎮</span>
        <p>还没有游戏记录，快去玩一局吧！</p>
      </div>
      <div v-else class="history-list">
        <div 
          v-for="record in recentHistory" 
          :key="record.id" 
          class="history-item"
          :class="{ 'win': record.result === 'win', 'lose': record.result === 'lose' }"
        >
          <div class="history-game">
            <span class="game-type-icon">
              {{ record.gameType === 'wheel' ? '🎡' : record.gameType === 'coin' ? '🪙' : '🔢' }}
            </span>
            <span class="game-type-name">{{ getGameTypeName(record.gameType) }}</span>
          </div>
          <div class="history-bet">
            <span class="bet-label">下注</span>
            <span class="bet-amount">{{ record.betAmount }}</span>
          </div>
          <div class="history-result">
            <span class="result-badge" :class="record.result">
              {{ record.result === 'win' ? '赢' : '输' }}
            </span>
            <span class="payout-amount" :class="record.result">
              {{ record.result === 'win' ? '+' : '-' }}{{ record.result === 'win' ? record.payout.toFixed(2) : record.betAmount.toFixed(2) }}
            </span>
          </div>
          <div class="history-time">
            {{ formatTime(record.timestamp) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, onMounted } from 'vue'
import { useMessage } from 'naive-ui'
import { useGameStore, type ApiGameRecord } from '@/stores/game'
import { wheelConfig, numberGuessConfig, coinFlipConfig, INITIAL_GAME_COINS } from '@/config/gameConfig'
import GameCard from '@/components/GameCard.vue'
import StatCard from '@/components/StatCard.vue'
import GameLeaderboard from '@/components/GameLeaderboard.vue'

const gameStore = useGameStore()
const message = useMessage()
const leaderboardRef = ref<InstanceType<typeof GameLeaderboard> | null>(null)
const historyLoading = ref(false)

// Load game records from backend on mount (Requirements: 6.1)
onMounted(async () => {
  historyLoading.value = true
  try {
    await gameStore.loadGameRecords(10, 0)
    // Also load game stats to ensure stats are up-to-date
    await gameStore.loadGameStats()
  } catch (err) {
    console.error('Failed to load game records:', err)
  } finally {
    historyLoading.value = false
  }
})

// Game definitions
const games = computed(() => [
  {
    name: '幸运转盘',
    description: '转动转盘，赢取高达 5 倍的奖励！多种倍率等你来挑战。',
    icon: '🎡',
    minBet: wheelConfig.minBet,
    route: '/games/wheel',
    gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)'
  },
  {
    name: '硬币翻转',
    description: '经典的 50/50 游戏，猜对正反面即可获得 1.95 倍奖励！',
    icon: '🪙',
    minBet: coinFlipConfig.minBet,
    route: '/games/coin',
    gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)'
  },
  {
    name: '猜数字',
    description: '猜中 1-10 的数字，即可获得 9 倍奖励！考验你的直觉。',
    icon: '🔢',
    minBet: numberGuessConfig.minBet,
    route: '/games/number',
    gradient: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)'
  }
])

// Recent game history from backend API (Requirements: 6.1, 6.2)
const recentHistory = computed(() => {
  // Use backend game records if available
  if (gameStore.gameRecords.length > 0) {
    return gameStore.gameRecords.slice(0, 10).map((record: ApiGameRecord) => ({
      id: String(record.id),
      gameType: record.game_type,
      betAmount: record.bet_amount,
      result: record.result,
      payout: record.payout,
      details: record.details,
      timestamp: new Date(record.created_at).getTime()
    }))
  }
  // Fallback to local cache if backend data not loaded
  return gameStore.getRecentHistory(10)
})

// Format timestamp (supports both number and ISO string)
function formatTime(timestamp: number | string): string {
  const date = typeof timestamp === 'string' ? new Date(timestamp) : new Date(timestamp)
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Get game type display name
function getGameTypeName(type: string): string {
  const names: Record<string, string> = {
    wheel: '幸运转盘',
    coin: '硬币翻转',
    number: '猜数字'
  }
  return names[type] || type
}

// Reset game coins
async function handleResetCoins() {
  try {
    await gameStore.resetGameData()
    message.success(`游戏币已重置为 ${INITIAL_GAME_COINS}`)
  } catch (err) {
    message.error('重置游戏币失败，请稍后重试')
    console.error('Failed to reset game coins:', err)
  }
}
</script>


<style scoped>
.game-center {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
  position: relative;
  background: 
    radial-gradient(ellipse at 20% 0%, rgba(139, 92, 246, 0.15) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 100%, rgba(236, 72, 153, 0.12) 0%, transparent 50%),
    radial-gradient(ellipse at 50% 50%, rgba(59, 130, 246, 0.08) 0%, transparent 60%);
  min-height: 100vh;
}

/* Balance Section */
.balance-section {
  margin-bottom: 2rem;
}

.balance-card {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  padding: 1.5rem 2rem;
  margin-bottom: 1rem;
  background: var(--bg-card);
  border: 1px solid rgba(139, 92, 246, 0.3);
  position: relative;
  overflow: hidden;
}

.balance-card::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -50%;
  width: 100%;
  height: 200%;
  background: linear-gradient(45deg, transparent, rgba(255, 255, 255, 0.05), transparent);
  animation: shimmer 3s infinite;
}

@keyframes shimmer {
  0% { transform: translateX(-100%) rotate(45deg); }
  100% { transform: translateX(100%) rotate(45deg); }
}

.balance-icon {
  font-size: 3rem;
  filter: drop-shadow(0 0 10px rgba(139, 92, 246, 0.5));
}

.balance-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.balance-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 1px;
}

.balance-value {
  background: linear-gradient(135deg, #8b5cf6, #ec4899, #f97316);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-size: 2.5rem;
  font-weight: 700;
}

.balance-actions {
  display: flex;
  gap: 0.5rem;
}

.low-balance-warning {
  margin-top: 0.5rem;
}

/* Section Headers */
.section-header {
  margin: 2rem 0 1rem;
}

.section-title {
  color: var(--text-primary);
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.15), rgba(236, 72, 153, 0.1));
  border-radius: 12px;
  border-left: 4px solid;
  border-image: linear-gradient(180deg, #8b5cf6, #ec4899) 1;
}

/* Games Grid */
.games-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 1.5rem;
}

/* Two Column Layout */
.two-column-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 2rem;
  margin-top: 2rem;
}

.stats-column,
.leaderboard-column {
  display: flex;
  flex-direction: column;
}

.stats-column .section-header,
.leaderboard-column .section-header {
  margin: 0 0 1rem 0;
}

/* Stats Grid */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

/* History Section */
.history-section {
  padding: 1.5rem;
}

.empty-history {
  text-align: center;
  padding: 3rem 1rem;
  color: var(--text-secondary);
}

.empty-icon {
  font-size: 4rem;
  display: block;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.empty-history p {
  margin: 0;
  font-size: 1rem;
}

.history-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.history-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 1rem 1.25rem;
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
  border-left: 3px solid transparent;
  transition: all var(--transition-fast);
}

.history-item.win {
  border-left-color: #10b981;
  background: linear-gradient(90deg, rgba(16, 185, 129, 0.1), transparent);
}

.history-item.lose {
  border-left-color: #ef4444;
  background: linear-gradient(90deg, rgba(239, 68, 68, 0.1), transparent);
}

.history-item:hover {
  background: var(--bg-hover);
}

.history-game {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 120px;
}

.game-type-icon {
  font-size: 1.25rem;
}

.game-type-name {
  color: var(--text-primary);
  font-weight: 600;
  font-size: 0.9rem;
}

.history-bet {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
  min-width: 80px;
}

.bet-label {
  color: var(--text-muted);
  font-size: 0.7rem;
  text-transform: uppercase;
}

.bet-amount {
  color: var(--text-primary);
  font-weight: 600;
}

.history-result {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;
}

.result-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 700;
  text-transform: uppercase;
}

.result-badge.win {
  background: linear-gradient(135deg, rgba(16, 185, 129, 0.3), rgba(52, 211, 153, 0.2));
  color: #10b981;
  border: 1px solid rgba(16, 185, 129, 0.4);
}

.result-badge.lose {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.3), rgba(248, 113, 113, 0.2));
  color: #ef4444;
  border: 1px solid rgba(239, 68, 68, 0.4);
}

.payout-amount {
  font-weight: 700;
  font-size: 1rem;
}

.payout-amount.win {
  color: var(--color-success);
}

.payout-amount.lose {
  color: var(--color-error);
}

.history-time {
  color: var(--text-muted);
  font-size: 0.8rem;
  min-width: 100px;
  text-align: right;
}

/* Glass Card - 适配亮色和暗色主题 */
.glass-card {
  background: var(--bg-card);
  border: 1px solid rgba(139, 92, 246, 0.25);
  border-radius: var(--border-radius-lg);
  box-shadow: 
    0 4px 20px rgba(139, 92, 246, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
}

/* Responsive Design */
@media (max-width: 1024px) {
  .two-column-section {
    grid-template-columns: 1fr;
    gap: 1.5rem;
  }
}

@media (max-width: 768px) {
  .game-center {
    padding: 1rem;
    /* Enable smooth scrolling on mobile */
    -webkit-overflow-scrolling: touch;
  }

  .balance-card {
    flex-direction: column;
    text-align: center;
    padding: 1.5rem 1rem;
  }

  .balance-info {
    align-items: center;
  }

  .balance-value {
    font-size: 2rem;
  }

  /* Touch-friendly button sizes (minimum 44px tap target) */
  .balance-actions :deep(.n-button) {
    min-height: 44px;
    min-width: 44px;
    padding: 0.75rem 1.25rem;
    font-size: 1rem;
  }

  .games-grid {
    grid-template-columns: 1fr;
    gap: 1rem;
  }

  .stats-grid {
    grid-template-columns: repeat(3, 1fr);
    gap: 0.75rem;
  }

  .history-item {
    flex-wrap: wrap;
    gap: 0.5rem;
    padding: 0.875rem 1rem;
    /* Touch-friendly tap target */
    min-height: 60px;
  }

  .history-game {
    min-width: auto;
    flex: 1;
  }

  .history-bet {
    min-width: auto;
  }

  .history-result {
    flex: auto;
    width: 100%;
    justify-content: space-between;
  }

  .history-time {
    width: 100%;
    text-align: left;
    margin-top: 0.25rem;
  }

  .section-header {
    margin: 1.5rem 0 0.75rem;
  }

  .history-section {
    padding: 1rem;
  }
}

@media (max-width: 480px) {
  .game-center {
    padding: 0.75rem;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
  }

  .section-title {
    font-size: 1.25rem;
  }

  .balance-icon {
    font-size: 2.5rem;
  }

  .balance-value {
    font-size: 1.75rem;
  }

  .balance-label {
    font-size: 0.8rem;
  }

  /* Smaller history items on very small screens */
  .history-item {
    padding: 0.75rem;
  }

  .game-type-name {
    font-size: 0.85rem;
  }

  .payout-amount {
    font-size: 0.9rem;
  }

  .result-badge {
    padding: 0.2rem 0.5rem;
    font-size: 0.7rem;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .history-item:hover {
    transform: none;
    background: rgba(255, 255, 255, 0.05);
  }

  .history-item:active {
    background: rgba(255, 255, 255, 0.1);
  }
}
</style>
