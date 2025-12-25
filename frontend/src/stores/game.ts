/**
 * Game State Management Store
 * 游戏状态管理 Store
 * 
 * Implements backend API-based storage for game data
 * 实现基于后端 API 的游戏数据存储
 * 
 * Requirements: 1.1, 1.2, 1.3, 1.5, 2.1, 3.1, 7.1, 7.2, 7.3, 7.4, 7.5, 8.2
 */

import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'
import { useAuthStore } from './auth'
import { INITIAL_GAME_COINS } from '@/config/gameConfig'
import {
  getGameBalance,
  deductGameCoins,
  addGameCoins as apiAddGameCoins,
  resetGameCoins,
  migrateLocalStorage,
  createGameRecord,
  getGameRecords,
  getGameStats,
  getLeaderboard,
  type GameBalance,
  type GameRecord as ApiGameRecord,
  type CreateGameRecordRequest,
  type GameStatsResponse,
  type LeaderboardEntry,
  type LeaderboardResponse,
  type GameDetails
} from '@/api/gameCoin'

// Local game record interface (for backward compatibility with local cache)
export interface GameRecord {
  id: string
  gameType: 'wheel' | 'coin' | 'number'
  betAmount: number
  result: 'win' | 'lose'
  payout: number
  details: Record<string, unknown>
  timestamp: number
}

interface UserGameData {
  gameCoins: number
  history: GameRecord[]
  lastUpdated: number
}

// Re-export types for external use
export type { ApiGameRecord, LeaderboardEntry, LeaderboardResponse, GameStatsResponse, GameDetails, CreateGameRecordRequest }

/**
 * Generate user-specific storage key (for migration purposes)
 * 生成用户专属的存储键（用于迁移）
 * 
 * @param userId - User ID
 * @returns Storage key string
 */
export function getStorageKey(userId: number): string {
  return `curry2api_game_data_${userId}`
}

export const useGameStore = defineStore('game', () => {
  const authStore = useAuthStore()
  
  // Game coin balance
  const gameCoins = ref(0)
  // Game history (local cache)
  const gameHistory = ref<GameRecord[]>([])
  // Loading state
  const loading = ref(false)
  // Error state
  const error = ref<string | null>(null)
  // Migration status
  const migrated = ref(false)
  // Full balance data from backend
  const balanceData = ref<GameBalance | null>(null)
  
  // Game records from backend API (Requirements: 1.5)
  const gameRecords = ref<ApiGameRecord[]>([])
  // Game records total count for pagination
  const gameRecordsTotal = ref(0)
  // Leaderboard data from backend API (Requirements: 3.1)
  const leaderboard = ref<LeaderboardResponse | null>(null)
  // Game stats from backend API (Requirements: 2.1)
  const gameStats = ref<GameStatsResponse | null>(null)
  
  /**
   * Load user game balance from backend API
   * 从后端 API 加载用户游戏币余额
   * Requirements: 1.3, 7.3
   */
  async function loadBalance(): Promise<void> {
    const userId = authStore.user?.id
    if (!userId) {
      gameCoins.value = 0
      gameHistory.value = []
      balanceData.value = null
      return
    }
    
    loading.value = true
    error.value = null
    
    try {
      // First check and migrate localStorage data if needed
      await checkAndMigrate()
      
      // Fetch balance from backend
      const response = await getGameBalance()
      balanceData.value = response.data
      gameCoins.value = response.data.balance
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load game balance'
      error.value = errorMessage
      console.error('Failed to load game balance:', err)
      // Fallback to initial coins on error
      gameCoins.value = INITIAL_GAME_COINS
    } finally {
      loading.value = false
    }
  }
  
  /**
   * Check and migrate localStorage data to backend
   * 检查并迁移 localStorage 数据到后端
   * Requirements: 1.5
   */
  async function checkAndMigrate(): Promise<void> {
    const userId = authStore.user?.id
    if (!userId || migrated.value) return
    
    const key = getStorageKey(userId)
    const stored = localStorage.getItem(key)
    
    if (stored) {
      try {
        const data: UserGameData = JSON.parse(stored)
        
        // Calculate stats from history
        const totalWon = data.history
          .filter(r => r.result === 'win')
          .reduce((sum, r) => sum + r.payout, 0)
        const totalLost = data.history
          .filter(r => r.result === 'lose')
          .reduce((sum, r) => sum + r.betAmount, 0)
        const gamesPlayed = data.history.length
        
        // Migrate to backend
        await migrateLocalStorage({
          balance: data.gameCoins,
          total_won: totalWon,
          total_lost: totalLost,
          games_played: gamesPlayed
        })
        
        // Clear localStorage after successful migration
        localStorage.removeItem(key)
        migrated.value = true
      } catch (err) {
        console.error('Failed to migrate localStorage data:', err)
        // Don't block on migration failure, just log it
      }
    }
    
    migrated.value = true
  }
  
  // Watch for user changes and auto-load balance
  watch(() => authStore.user?.id, () => {
    migrated.value = false // Reset migration status for new user
    loadBalance()
  }, { immediate: true })
  
  /**
   * Deduct game coins (for betting)
   * 扣除游戏币（下注）
   * Requirements: 1.2, 7.1, 7.4, 7.5
   * 
   * @param amount - Amount to deduct
   * @param gameType - Type of game (wheel, coin, number)
   * @param description - Optional description
   * @returns true if successful, false if failed
   */
  async function deductCoins(amount: number, gameType: string = 'game', description?: string): Promise<boolean> {
    if (amount <= 0) return false
    if (amount > gameCoins.value) return false
    
    loading.value = true
    error.value = null
    
    try {
      const response = await deductGameCoins({
        amount,
        game_type: gameType,
        description
      })
      
      if (response.data.success) {
        gameCoins.value = response.data.balance_after
        return true
      }
      return false
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to deduct game coins'
      error.value = errorMessage
      console.error('Failed to deduct game coins:', err)
      return false
    } finally {
      loading.value = false
    }
  }
  
  /**
   * Add game coins (for rewards)
   * 增加游戏币（奖励）
   * Requirements: 1.2, 7.2, 7.4, 7.5
   * 
   * @param amount - Amount to add
   * @param gameType - Type of game (wheel, coin, number)
   * @param description - Optional description
   */
  async function addCoins(amount: number, gameType: string = 'game', description?: string): Promise<void> {
    if (amount <= 0) return
    
    loading.value = true
    error.value = null
    
    try {
      const response = await apiAddGameCoins({
        amount,
        game_type: gameType,
        description
      })
      
      if (response.data.success) {
        gameCoins.value = response.data.balance_after
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to add game coins'
      error.value = errorMessage
      console.error('Failed to add game coins:', err)
    } finally {
      loading.value = false
    }
  }
  
  /**
   * Reset game data (for testing or user request)
   * 重置游戏数据（用于测试或用户请求）
   * Requirements: 8.2
   */
  async function resetGameData(): Promise<void> {
    const userId = authStore.user?.id
    if (!userId) return
    
    loading.value = true
    error.value = null
    
    try {
      const response = await resetGameCoins()
      
      if (response.data.success) {
        gameCoins.value = response.data.balance
        balanceData.value = {
          balance: response.data.balance,
          total_won: response.data.total_won,
          total_lost: response.data.total_lost,
          total_exchanged: response.data.total_exchanged,
          games_played: response.data.games_played
        }
        // Clear local history cache
        gameHistory.value = []
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to reset game data'
      error.value = errorMessage
      console.error('Failed to reset game data:', err)
    } finally {
      loading.value = false
    }
  }
  
  /**
   * Add a game record to local history cache
   * 添加游戏记录到本地历史缓存
   * 
   * @param record - Game record without id and timestamp
   */
  function addGameRecord(record: Omit<GameRecord, 'id' | 'timestamp'>): void {
    const newRecord: GameRecord = {
      ...record,
      id: crypto.randomUUID(),
      timestamp: Date.now(),
    }
    gameHistory.value.unshift(newRecord)
    
    // Keep only the last 100 records in local cache
    if (gameHistory.value.length > 100) {
      gameHistory.value = gameHistory.value.slice(0, 100)
    }
  }
  
  /**
   * Get recent game history from local cache
   * 获取最近 N 条记录（本地缓存）
   * 
   * @param limit - Maximum number of records to return
   * @returns Array of recent game records
   */
  function getRecentHistory(limit: number = 10): GameRecord[] {
    return gameHistory.value.slice(0, limit)
  }
  
  /**
   * Record a game result to the backend API
   * 记录游戏结果到后端 API
   * Requirements: 1.1, 1.2, 1.3, 1.4
   * 
   * @param data - Game record data to save
   * @returns The created record and updated stats, or null on failure
   */
  async function recordGameResult(data: CreateGameRecordRequest): Promise<{ record: ApiGameRecord; stats: GameStatsResponse } | null> {
    loading.value = true
    error.value = null
    
    try {
      const response = await createGameRecord(data)
      
      if (response.data.success) {
        // Update local game stats
        gameStats.value = {
          games_played: response.data.stats.games_played,
          wins: response.data.stats.wins,
          losses: response.data.stats.games_played - response.data.stats.wins,
          win_rate: response.data.stats.win_rate,
          net_profit: response.data.stats.net_profit,
          total_won: '0', // Not returned in this response
          total_lost: '0'  // Not returned in this response
        }
        
        // Add to local game records cache (prepend to maintain newest first order)
        gameRecords.value.unshift(response.data.record)
        
        // Also add to legacy local history for backward compatibility
        addGameRecord({
          gameType: data.game_type,
          betAmount: data.bet_amount,
          result: data.result,
          payout: data.payout,
          details: data.details as unknown as Record<string, unknown>
        })
        
        return {
          record: response.data.record,
          stats: gameStats.value
        }
      }
      return null
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to record game result'
      error.value = errorMessage
      console.error('Failed to record game result:', err)
      // Don't block game flow on record failure (Requirement 1.7)
      return null
    } finally {
      loading.value = false
    }
  }
  
  /**
   * Load game records from backend API with pagination
   * 从后端 API 加载游戏记录（分页）
   * Requirements: 1.5, 1.6
   * 
   * @param limit - Maximum number of records to return (default 10, max 100)
   * @param offset - Number of records to skip (default 0)
   * @returns Array of game records
   */
  async function loadGameRecords(limit: number = 10, offset: number = 0): Promise<ApiGameRecord[]> {
    loading.value = true
    error.value = null
    
    try {
      const response = await getGameRecords(limit, offset)
      
      if (offset === 0) {
        // Replace records if starting from beginning
        gameRecords.value = response.data.records
      } else {
        // Append records for pagination
        gameRecords.value = [...gameRecords.value, ...response.data.records]
      }
      gameRecordsTotal.value = response.data.total
      
      return response.data.records
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load game records'
      error.value = errorMessage
      console.error('Failed to load game records:', err)
      return []
    } finally {
      loading.value = false
    }
  }
  
  /**
   * Load game statistics from backend API
   * 从后端 API 加载游戏统计
   * Requirements: 2.1
   * 
   * @returns Game statistics or null on failure
   */
  async function loadGameStats(): Promise<GameStatsResponse | null> {
    loading.value = true
    error.value = null
    
    try {
      const response = await getGameStats()
      gameStats.value = response.data
      return response.data
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load game stats'
      error.value = errorMessage
      console.error('Failed to load game stats:', err)
      return null
    } finally {
      loading.value = false
    }
  }
  
  /**
   * Load leaderboard from backend API
   * 从后端 API 加载排行榜
   * Requirements: 3.1, 3.2, 4.2
   * 
   * @param sort - Sort by 'winnings' (net profit) or 'games' (games played)
   * @param limit - Maximum number of entries to return (default 10)
   * @returns Leaderboard response or null on failure
   */
  async function loadLeaderboard(sort: 'winnings' | 'games' = 'winnings', limit: number = 10): Promise<LeaderboardResponse | null> {
    loading.value = true
    error.value = null
    
    try {
      const response = await getLeaderboard(sort, limit)
      leaderboard.value = response.data
      return response.data
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load leaderboard'
      error.value = errorMessage
      console.error('Failed to load leaderboard:', err)
      return null
    } finally {
      loading.value = false
    }
  }
  
  /**
   * Computed statistics from backend data or fallback to local
   * 从后端数据计算统计信息，或回退到本地数据
   * Requirements: 2.1, 2.4, 2.5
   */
  const stats = computed(() => {
    // Use backend game stats if available (preferred)
    if (gameStats.value) {
      return {
        totalGames: gameStats.value.games_played,
        wins: gameStats.value.wins,
        losses: gameStats.value.losses,
        winRate: gameStats.value.win_rate,
        netProfit: gameStats.value.net_profit,
        totalWon: gameStats.value.total_won,
        totalLost: gameStats.value.total_lost,
        totalExchanged: balanceData.value?.total_exchanged.toFixed(2) ?? '0'
      }
    }
    
    // Fallback to balance data if game stats not loaded
    if (balanceData.value) {
      const total = balanceData.value.games_played
      const netProfit = balanceData.value.total_won - balanceData.value.total_lost
      return {
        totalGames: total,
        wins: 0, // Not tracked in balance data
        losses: 0, // Not tracked in balance data
        winRate: '0', // Would need game stats
        netProfit: netProfit.toFixed(2),
        totalWon: balanceData.value.total_won.toFixed(2),
        totalLost: balanceData.value.total_lost.toFixed(2),
        totalExchanged: balanceData.value.total_exchanged.toFixed(2)
      }
    }
    
    // Final fallback to local history stats
    const total = gameHistory.value.length
    const wins = gameHistory.value.filter(r => r.result === 'win').length
    const totalBet = gameHistory.value.reduce((sum, r) => sum + r.betAmount, 0)
    const totalPayout = gameHistory.value.reduce((sum, r) => sum + r.payout, 0)
    return {
      totalGames: total,
      wins,
      losses: total - wins,
      winRate: total > 0 ? (wins / total * 100).toFixed(1) : '0',
      netProfit: (totalPayout - totalBet).toFixed(2),
      totalWon: totalPayout.toFixed(2),
      totalLost: totalBet.toFixed(2),
      totalExchanged: '0'
    }
  })
  
  // Legacy functions for backward compatibility
  function loadUserData(): void {
    loadBalance()
  }
  
  function saveUserData(): void {
    // No-op: data is now saved via API calls
  }
  
  return { 
    // State
    gameCoins,
    gameHistory,
    loading,
    error,
    migrated,
    balanceData,
    gameRecords,        // New: backend game records (Requirements: 1.5)
    gameRecordsTotal,   // New: total count for pagination
    leaderboard,        // New: backend leaderboard data (Requirements: 3.1)
    gameStats,          // New: backend game stats (Requirements: 2.1)
    
    // Methods
    loadBalance,
    loadUserData, // Legacy alias
    saveUserData, // Legacy no-op
    checkAndMigrate,
    deductCoins,
    addCoins,
    addGameRecord, 
    getRecentHistory,
    resetGameData,
    recordGameResult,   // New: record game result to backend (Requirements: 1.1)
    loadGameRecords,    // New: load game records from backend (Requirements: 1.5)
    loadGameStats,      // New: load game stats from backend (Requirements: 2.1)
    loadLeaderboard,    // New: load leaderboard from backend (Requirements: 3.1)
    
    // Computed
    stats 
  }
})
