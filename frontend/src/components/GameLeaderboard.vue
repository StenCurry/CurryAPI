<template>
  <div class="leaderboard-container glass-card">
    <div class="leaderboard-header">
      <h3 class="leaderboard-title">
        <span class="title-icon">🏆</span>
        排行榜
      </h3>
      <div class="leaderboard-tabs">
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'winnings' }"
          @click="switchTab('winnings')"
          :disabled="isLoading"
        >
          总赢取
        </button>
        <button 
          class="tab-btn" 
          :class="{ active: activeTab === 'games' }"
          @click="switchTab('games')"
          :disabled="isLoading"
        >
          游戏次数
        </button>
      </div>
    </div>

    <div class="leaderboard-content">
      <!-- Loading State -->
      <div v-if="isLoading" class="loading-state">
        <div class="loading-spinner"></div>
        <p>加载排行榜...</p>
      </div>

      <!-- Error State -->
      <div v-else-if="hasError" class="error-state">
        <span class="error-icon">⚠️</span>
        <p>{{ errorMessage }}</p>
        <button class="retry-btn" @click="refreshLeaderboard">
          重试
        </button>
      </div>

      <!-- Empty State -->
      <div v-else-if="sortedEntries.length === 0" class="empty-leaderboard">
        <span class="empty-icon">📊</span>
        <p>暂无排行数据</p>
        <p class="empty-hint">快去玩游戏，成为第一名！</p>
      </div>

      <!-- Leaderboard List -->
      <div v-else class="leaderboard-list">
        <div 
          v-for="(entry, index) in sortedEntries" 
          :key="entry.user_id"
          class="leaderboard-item"
          :class="{ 
            'current-user': isCurrentUser(entry.user_id),
            'top-three': index < 3 
          }"
        >
          <div class="rank">
            <span v-if="entry.rank === 1" class="rank-medal">🥇</span>
            <span v-else-if="entry.rank === 2" class="rank-medal">🥈</span>
            <span v-else-if="entry.rank === 3" class="rank-medal">🥉</span>
            <span v-else class="rank-number">{{ entry.rank }}</span>
          </div>
          
          <div class="user-info">
            <span class="username">{{ entry.username }}</span>
            <span v-if="isCurrentUser(entry.user_id)" class="you-badge">你</span>
          </div>
          
          <div class="stats">
            <span class="stat-value">
              {{ activeTab === 'winnings' ? formatWinnings(entry.total_winnings) : entry.games_played }}
            </span>
            <span class="stat-label">
              {{ activeTab === 'winnings' ? '游戏币' : '局' }}
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Current User Section (if not in top entries) -->
    <div v-if="showCurrentUserSection" class="current-user-section">
      <div class="divider">
        <span>···</span>
      </div>
      <div class="leaderboard-item current-user highlight">
        <div class="rank">
          <span class="rank-number">{{ currentUserEntry?.rank }}</span>
        </div>
        <div class="user-info">
          <span class="username">{{ currentUserEntry?.username }}</span>
          <span class="you-badge">你</span>
        </div>
        <div class="stats">
          <span class="stat-value">
            {{ activeTab === 'winnings' ? formatWinnings(currentUserEntry?.total_winnings || 0) : currentUserEntry?.games_played || 0 }}
          </span>
          <span class="stat-label">
            {{ activeTab === 'winnings' ? '游戏币' : '局' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useGameStore } from '@/stores/game'
import { migrateLocalStorage } from '@/api/gameCoin'
import type { LeaderboardEntry } from '@/api/gameCoin'

// Re-export LeaderboardEntry for external use
export type { LeaderboardEntry }

// localStorage key for legacy leaderboard data
const LEGACY_LEADERBOARD_KEY = 'curry2api_game_leaderboard'

const authStore = useAuthStore()
const gameStore = useGameStore()

const activeTab = ref<'winnings' | 'games'>('winnings')
const isLoading = ref(false)
const hasError = ref(false)
const errorMessage = ref('')
const migrationAttempted = ref(false)

/**
 * Legacy leaderboard entry interface (from old localStorage format)
 * 旧版排行榜条目接口（来自旧的 localStorage 格式）
 */
interface LegacyLeaderboardEntry {
  id: string
  name: string
  totalWinnings: number
  gamesPlayed: number
  timestamp: number
}

/**
 * Check and migrate localStorage leaderboard data to backend
 * 检查并迁移 localStorage 排行榜数据到后端
 * Requirements: 5.1, 5.2, 5.3, 5.4
 */
async function migrateLeaderboardData(): Promise<void> {
  // Only attempt migration once per session
  if (migrationAttempted.value) return
  migrationAttempted.value = true
  
  // Check if user is authenticated
  const userId = authStore.user?.id
  if (!userId) return
  
  try {
    // Check for existing localStorage leaderboard data
    const storedData = localStorage.getItem(LEGACY_LEADERBOARD_KEY)
    if (!storedData) return
    
    const leaderboardData: LegacyLeaderboardEntry[] = JSON.parse(storedData)
    if (!Array.isArray(leaderboardData) || leaderboardData.length === 0) {
      // Invalid or empty data, clear it
      localStorage.removeItem(LEGACY_LEADERBOARD_KEY)
      return
    }
    
    // Find current user's entry in the legacy leaderboard
    // The legacy format stored entries with a string id that might be the username or a generated id
    // We'll look for an entry that matches the current user's username or id
    const currentUsername = authStore.user?.username || ''
    const userEntry = leaderboardData.find(entry => 
      entry.name === currentUsername || 
      entry.id === String(userId) ||
      entry.id === currentUsername
    )
    
    if (!userEntry) {
      // No entry for current user, clear the localStorage
      // (other users' data shouldn't be migrated by this user)
      console.log('No matching leaderboard entry found for current user, clearing localStorage')
      localStorage.removeItem(LEGACY_LEADERBOARD_KEY)
      return
    }
    
    // Attempt to sync with backend via migration endpoint
    // The migration endpoint updates user_game_balances which feeds into the leaderboard
    console.log('Migrating leaderboard data for user:', currentUsername, userEntry)
    
    await migrateLocalStorage({
      balance: 0, // Don't override balance, just sync stats
      total_won: userEntry.totalWinnings > 0 ? userEntry.totalWinnings : 0,
      total_lost: userEntry.totalWinnings < 0 ? Math.abs(userEntry.totalWinnings) : 0,
      games_played: userEntry.gamesPlayed || 0
    })
    
    // Migration successful, clear localStorage (Requirement 5.2)
    localStorage.removeItem(LEGACY_LEADERBOARD_KEY)
    console.log('Leaderboard migration successful, localStorage cleared')
    
  } catch (err) {
    // Migration failed, retain localStorage data for retry on next login (Requirement 5.3)
    console.error('Failed to migrate leaderboard data:', err)
    // Don't clear localStorage on failure - will retry on next mount
  }
}

/**
 * Load leaderboard data from backend API
 * 从后端 API 加载排行榜数据
 * Requirements: 3.1, 3.2
 */
async function loadLeaderboard(): Promise<void> {
  isLoading.value = true
  hasError.value = false
  errorMessage.value = ''
  
  try {
    await gameStore.loadLeaderboard(activeTab.value, 10)
  } catch (err: unknown) {
    hasError.value = true
    errorMessage.value = err instanceof Error ? err.message : '加载排行榜失败'
    console.error('Failed to load leaderboard:', err)
  } finally {
    isLoading.value = false
  }
}

/**
 * Refresh leaderboard data
 * 刷新排行榜数据
 * Requirements: 4.1
 */
async function refreshLeaderboard(): Promise<void> {
  await loadLeaderboard()
}

/**
 * Switch between tabs and reload data
 * 切换标签并重新加载数据
 */
async function switchTab(tab: 'winnings' | 'games'): Promise<void> {
  if (activeTab.value === tab || isLoading.value) return
  activeTab.value = tab
  await loadLeaderboard()
}

/**
 * Sorted leaderboard entries from backend
 * 从后端获取的排序后的排行榜条目
 * Requirements: 3.2
 */
const sortedEntries = computed((): LeaderboardEntry[] => {
  if (!gameStore.leaderboard) return []
  return gameStore.leaderboard.entries || []
})

/**
 * Current user's entry from backend (if not in top entries)
 * 当前用户的条目（如果不在前几名中）
 * Requirements: 3.4, 3.5
 */
const currentUserEntry = computed((): LeaderboardEntry | null => {
  if (!gameStore.leaderboard) return null
  return gameStore.leaderboard.current_user
})

/**
 * Whether to show the current user section below the main list
 * 是否显示当前用户区域（在主列表下方）
 * Requirements: 3.5
 */
const showCurrentUserSection = computed((): boolean => {
  if (!currentUserEntry.value) return false
  // Show if current user is not in the top entries
  const isInTopEntries = sortedEntries.value.some(
    entry => entry.user_id === currentUserEntry.value?.user_id
  )
  return !isInTopEntries && currentUserEntry.value.rank > 0
})

/**
 * Check if entry belongs to current user
 * 检查条目是否属于当前用户
 * Requirements: 3.4
 */
function isCurrentUser(userId: number): boolean {
  return authStore.user?.id === userId
}

/**
 * Format winnings display
 * 格式化赢取金额显示
 */
function formatWinnings(amount: number): string {
  if (amount >= 0) {
    return `+${amount.toFixed(2)}`
  }
  return amount.toFixed(2)
}

// Load leaderboard on mount and check for migration (Requirement 5.4)
onMounted(async () => {
  // Check for and trigger migration if needed (Requirement 5.4)
  await migrateLeaderboardData()
  // Load leaderboard data from backend
  loadLeaderboard()
})

// Watch for auth changes and reload (Requirement 5.1 - migrate on login)
watch(() => authStore.user?.id, async () => {
  if (authStore.user) {
    // Reset migration status for new user login
    migrationAttempted.value = false
    // Attempt migration for the new user
    await migrateLeaderboardData()
    // Load leaderboard data
    loadLeaderboard()
  }
})

// Expose for external use (e.g., refresh after game completion)
defineExpose({
  loadLeaderboard,
  refreshLeaderboard
})
</script>


<style scoped>
.leaderboard-container {
  padding: 1.5rem;
}

.leaderboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
  gap: 1rem;
}

.leaderboard-title {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin: 0;
  color: white;
  font-size: 1.25rem;
  font-weight: 700;
}

.title-icon {
  font-size: 1.5rem;
}

.leaderboard-tabs {
  display: flex;
  gap: 0.5rem;
  background: rgba(255, 255, 255, 0.05);
  padding: 0.25rem;
  border-radius: 10px;
}

.tab-btn {
  padding: 0.5rem 1rem;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.6);
  font-size: 0.85rem;
  font-weight: 600;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.3s ease;
}

.tab-btn:hover:not(:disabled) {
  color: white;
  background: rgba(255, 255, 255, 0.1);
}

.tab-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}

.tab-btn.active {
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  color: white;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

.leaderboard-content {
  min-height: 200px;
}

/* Loading State */
.loading-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 3rem 1rem;
  color: rgba(255, 255, 255, 0.6);
}

.loading-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: #3b82f6;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-bottom: 1rem;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

/* Error State */
.error-state {
  text-align: center;
  padding: 3rem 1rem;
  color: rgba(255, 255, 255, 0.6);
}

.error-icon {
  font-size: 2.5rem;
  display: block;
  margin-bottom: 1rem;
}

.error-state p {
  margin: 0.5rem 0;
}

.retry-btn {
  margin-top: 1rem;
  padding: 0.5rem 1.5rem;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  color: white;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
}

.retry-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.4);
}

/* Empty State */
.empty-leaderboard {
  text-align: center;
  padding: 3rem 1rem;
  color: rgba(255, 255, 255, 0.6);
}

.empty-icon {
  font-size: 3rem;
  display: block;
  margin-bottom: 1rem;
  opacity: 0.5;
}

.empty-leaderboard p {
  margin: 0.5rem 0;
}

.empty-hint {
  font-size: 0.85rem;
  opacity: 0.7;
}

.leaderboard-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.leaderboard-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.875rem 1rem;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 12px;
  transition: all 0.3s ease;
}

.leaderboard-item:hover {
  background: rgba(255, 255, 255, 0.08);
  transform: translateX(4px);
}

.leaderboard-item.top-three {
  background: rgba(255, 255, 255, 0.05);
}

.leaderboard-item.current-user {
  background: rgba(59, 130, 246, 0.15);
  border: 1px solid rgba(59, 130, 246, 0.3);
}

.leaderboard-item.highlight {
  background: rgba(59, 130, 246, 0.2);
}

.rank {
  width: 40px;
  text-align: center;
}

.rank-medal {
  font-size: 1.5rem;
}

.rank-number {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  color: rgba(255, 255, 255, 0.7);
  font-size: 0.85rem;
  font-weight: 600;
}

.user-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.username {
  color: white;
  font-weight: 600;
  font-size: 0.95rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.you-badge {
  padding: 0.125rem 0.5rem;
  background: linear-gradient(135deg, #3b82f6 0%, #8b5cf6 100%);
  color: white;
  font-size: 0.7rem;
  font-weight: 700;
  border-radius: 10px;
  text-transform: uppercase;
  flex-shrink: 0;
}

.stats {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.125rem;
}

.stat-value {
  color: #60a5fa;
  font-size: 1rem;
  font-weight: 700;
}

.stat-label {
  color: rgba(255, 255, 255, 0.5);
  font-size: 0.7rem;
  text-transform: uppercase;
}

.current-user-section {
  margin-top: 1rem;
}

.divider {
  text-align: center;
  color: rgba(255, 255, 255, 0.3);
  margin-bottom: 0.5rem;
  font-size: 1.25rem;
  letter-spacing: 4px;
}

/* Glass Card */
.glass-card {
  background: rgba(15, 23, 42, 0.85);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(59, 130, 246, 0.3);
  border-radius: 20px;
  box-shadow: 
    0 8px 32px rgba(0, 0, 0, 0.3),
    0 0 20px rgba(59, 130, 246, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

/* Responsive */
@media (max-width: 480px) {
  .leaderboard-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .leaderboard-tabs {
    width: 100%;
  }

  .tab-btn {
    flex: 1;
    text-align: center;
  }

  .leaderboard-item {
    padding: 0.75rem;
    gap: 0.75rem;
  }

  .rank {
    width: 32px;
  }

  .rank-medal {
    font-size: 1.25rem;
  }

  .username {
    font-size: 0.85rem;
  }

  .stat-value {
    font-size: 0.9rem;
  }
}
</style>
