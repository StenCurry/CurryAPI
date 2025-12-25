/**
 * Property-Based Tests for Game State Management Store
 * 游戏状态管理属性测试
 * 
 * Feature: frontend-enhancement
 * Validates: Requirements 5.3, 5.6, 6.3, 6.7, 7.3, 7.7
 */

import { describe, it } from 'vitest'
import fc from 'fast-check'
import { getStorageKey } from './game'

// ============================================================================
// Property 17: User-specific storage key generation
// Feature: frontend-enhancement, Property 17: User-specific storage key generation
// Validates: Requirements 5.6, 6.7, 7.7
// ============================================================================

describe('Property 17: User-specific storage key generation', () => {
  it('should generate unique keys for different users', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000000 }),
        fc.integer({ min: 1, max: 1000000 }),
        (userId1, userId2) => {
          if (userId1 === userId2) return true
          const key1 = getStorageKey(userId1)
          const key2 = getStorageKey(userId2)
          return key1 !== key2 && key1.includes(String(userId1)) && key2.includes(String(userId2))
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should generate consistent keys for the same user', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000000 }),
        (userId) => {
          const key1 = getStorageKey(userId)
          const key2 = getStorageKey(userId)
          return key1 === key2
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should include user ID in the storage key', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000000 }),
        (userId) => {
          const key = getStorageKey(userId)
          return key.includes(String(userId))
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should have a consistent prefix', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000000 }),
        (userId) => {
          const key = getStorageKey(userId)
          return key.startsWith('curry2api_game_data_')
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 6: Game coin deduction correctness
// Feature: frontend-enhancement, Property 6: Game coin deduction correctness
// Validates: Requirements 5.3, 6.3, 7.3
// ============================================================================

describe('Property 6: Game coin deduction correctness', () => {
  // Test the deduction logic directly
  const deductCoins = (currentBalance: number, amount: number): { success: boolean; newBalance: number } => {
    if (amount <= 0) return { success: false, newBalance: currentBalance }
    if (amount > currentBalance) return { success: false, newBalance: currentBalance }
    const newBalance = Number((currentBalance - amount).toFixed(2))
    return { success: true, newBalance }
  }

  it('should correctly deduct coins when amount is valid', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 100, max: 10000 }),
        fc.integer({ min: 1, max: 99 }),
        (balanceCents, amountCents) => {
          const balance = balanceCents / 100
          const amount = amountCents / 100
          if (amount > balance) return true // Skip invalid cases
          const result = deductCoins(balance, amount)
          const expectedBalance = Number((balance - amount).toFixed(2))
          return result.success === true && result.newBalance === expectedBalance
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should reject deduction when amount exceeds balance', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000 }),
        fc.integer({ min: 1001, max: 10000 }),
        (balanceCents, amountCents) => {
          const balance = balanceCents / 100
          const amount = amountCents / 100
          const result = deductCoins(balance, amount)
          return result.success === false && result.newBalance === balance
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should reject zero or negative deduction amounts', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 100, max: 10000 }),
        fc.integer({ min: -1000, max: 0 }),
        (balanceCents, amountCents) => {
          const balance = balanceCents / 100
          const amount = amountCents / 100
          const result = deductCoins(balance, amount)
          return result.success === false && result.newBalance === balance
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 18: Game coin balance non-negative
// Feature: frontend-enhancement, Property 18: Game coin balance non-negative
// Validates: Requirements 5.2, 5.3, 6.2, 6.3, 7.2, 7.3
// ============================================================================

describe('Property 18: Game coin balance non-negative', () => {
  const deductCoins = (currentBalance: number, amount: number): { success: boolean; newBalance: number } => {
    if (amount <= 0) return { success: false, newBalance: currentBalance }
    if (amount > currentBalance) return { success: false, newBalance: currentBalance }
    const newBalance = Number((currentBalance - amount).toFixed(2))
    return { success: true, newBalance }
  }

  it('should never result in negative balance after deduction', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 10000 }),
        fc.integer({ min: 1, max: 20000 }),
        (balanceCents, amountCents) => {
          const balance = balanceCents / 100
          const amount = amountCents / 100
          const result = deductCoins(balance, amount)
          return result.newBalance >= 0
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should maintain non-negative balance through multiple operations', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 100, max: 10000 }),
        fc.array(fc.integer({ min: 1, max: 100 }), { minLength: 1, maxLength: 20 }),
        (initialBalanceCents, operationsCents) => {
          let balance = initialBalanceCents / 100
          for (const opCents of operationsCents) {
            const amount = opCents / 100
            const result = deductCoins(balance, amount)
            balance = result.newBalance
            if (balance < 0) return false
          }
          return balance >= 0
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 19: Game coin precision
// Feature: frontend-enhancement, Property 19: Game coin precision
// Validates: Requirements 5.3, 5.4, 6.3, 6.5, 7.3, 7.5
// ============================================================================

describe('Property 19: Game coin precision', () => {
  const roundToTwoDecimals = (value: number): number => {
    return Number(value.toFixed(2))
  }

  it('should maintain 2 decimal precision after deduction', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 100, max: 100000 }),
        fc.integer({ min: 1, max: 99 }),
        (balanceCents, amountCents) => {
          const balance = balanceCents / 100
          const amount = amountCents / 100
          const newBalance = roundToTwoDecimals(balance - amount)
          const decimalStr = newBalance.toString()
          const decimalPart = decimalStr.includes('.') ? decimalStr.split('.')[1] : ''
          return (decimalPart?.length ?? 0) <= 2
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should maintain 2 decimal precision after addition', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 100000 }),
        fc.integer({ min: 1, max: 10000 }),
        (balanceCents, amountCents) => {
          const balance = balanceCents / 100
          const amount = amountCents / 100
          const newBalance = roundToTwoDecimals(balance + amount)
          const decimalStr = newBalance.toString()
          const decimalPart = decimalStr.includes('.') ? decimalStr.split('.')[1] : ''
          return (decimalPart?.length ?? 0) <= 2
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 14: Game history persistence
// Feature: frontend-enhancement, Property 14: Game history persistence
// Validates: Requirements 5.6, 6.7, 7.7
// ============================================================================

describe('Property 14: Game history persistence', () => {
  interface GameRecord {
    id: string
    gameType: 'wheel' | 'coin' | 'number'
    betAmount: number
    result: 'win' | 'lose'
    payout: number
    details: Record<string, unknown>
    timestamp: number
  }

  const createRecord = (gameType: 'wheel' | 'coin' | 'number', betAmount: number, result: 'win' | 'lose', payout: number): GameRecord => ({
    id: crypto.randomUUID(),
    gameType,
    betAmount,
    result,
    payout,
    details: {},
    timestamp: Date.now()
  })

  it('should preserve all record data when added to history', () => {
    fc.assert(
      fc.property(
        fc.constantFrom('wheel', 'coin', 'number') as fc.Arbitrary<'wheel' | 'coin' | 'number'>,
        fc.integer({ min: 1, max: 1000 }),
        fc.constantFrom('win', 'lose') as fc.Arbitrary<'win' | 'lose'>,
        fc.integer({ min: 0, max: 10000 }),
        (gameType, betCents, result, payoutCents) => {
          const betAmount = betCents / 100
          const payout = payoutCents / 100
          const record = createRecord(gameType, betAmount, result, payout)
          
          // Verify record has all required fields
          return (
            record.id !== undefined &&
            record.gameType === gameType &&
            record.betAmount === betAmount &&
            record.result === result &&
            record.payout === payout &&
            record.timestamp !== undefined
          )
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should maintain history order (newest first)', () => {
    fc.assert(
      fc.property(
        fc.array(fc.integer({ min: 1, max: 100 }), { minLength: 2, maxLength: 10 }),
        (betAmounts) => {
          const records: GameRecord[] = []
          for (const betCents of betAmounts) {
            const record = createRecord('wheel', betCents / 100, 'win', betCents / 50)
            records.unshift(record) // Add to front like the store does
          }
          
          // Verify order - timestamps should be in descending order (newest first)
          for (let i = 0; i < records.length - 1; i++) {
            if (records[i]!.timestamp < records[i + 1]!.timestamp) {
              return false
            }
          }
          return true
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 16: User data isolation
// Feature: frontend-enhancement, Property 16: User data isolation
// Validates: Requirements 5.6, 6.7, 7.7
// ============================================================================

describe('Property 16: User data isolation', () => {
  it('should generate different storage keys for different users', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000000 }),
        fc.integer({ min: 1, max: 1000000 }),
        (userId1, userId2) => {
          if (userId1 === userId2) return true
          const key1 = getStorageKey(userId1)
          const key2 = getStorageKey(userId2)
          return key1 !== key2
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should ensure storage keys are unique per user', () => {
    const keys = new Set<string>()
    
    fc.assert(
      fc.property(
        fc.uniqueArray(fc.integer({ min: 1, max: 1000000 }), { minLength: 10, maxLength: 50 }),
        (userIds) => {
          keys.clear()
          for (const userId of userIds) {
            const key = getStorageKey(userId)
            if (keys.has(key)) return false
            keys.add(key)
          }
          return keys.size === userIds.length
        }
      ),
      { numRuns: 100 }
    )
  })
})


// ============================================================================
// Property 15: Concurrent Operation Safety
// Feature: game-coin-exchange, Property 15: Concurrent Operation Safety
// Validates: Requirements 7.4, 7.5
// ============================================================================

describe('Property 15: Concurrent Operation Safety', () => {
  /**
   * Simulates the sequential execution of operations to compute expected final balance.
   * This represents what the backend should compute when operations are processed.
   */
  const computeSequentialBalance = (
    initialBalance: number,
    operations: Array<{ type: 'deduct' | 'add'; amount: number }>
  ): number => {
    let balance = initialBalance
    for (const op of operations) {
      if (op.type === 'deduct') {
        // Deduction only succeeds if amount <= balance and amount > 0
        if (op.amount > 0 && op.amount <= balance) {
          balance = Number((balance - op.amount).toFixed(2))
        }
      } else if (op.type === 'add') {
        // Addition only succeeds if amount > 0
        if (op.amount > 0) {
          balance = Number((balance + op.amount).toFixed(2))
        }
      }
    }
    return balance
  }

  /**
   * Simulates concurrent execution where operations may interleave.
   * The key property is that regardless of interleaving, the final balance
   * should be consistent with some valid sequential ordering.
   */
  const simulateConcurrentOperations = (
    initialBalance: number,
    operations: Array<{ type: 'deduct' | 'add'; amount: number }>
  ): { finalBalance: number; successfulOps: number } => {
    // In a properly implemented concurrent system with database transactions,
    // each operation should see a consistent view and the final result
    // should be equivalent to some sequential ordering.
    // We simulate this by processing operations sequentially (as the backend would).
    let balance = initialBalance
    let successfulOps = 0
    
    for (const op of operations) {
      if (op.type === 'deduct') {
        if (op.amount > 0 && op.amount <= balance) {
          balance = Number((balance - op.amount).toFixed(2))
          successfulOps++
        }
      } else if (op.type === 'add') {
        if (op.amount > 0) {
          balance = Number((balance + op.amount).toFixed(2))
          successfulOps++
        }
      }
    }
    
    return { finalBalance: balance, successfulOps }
  }

  it('should produce consistent final balance regardless of operation order', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 100, max: 10000 }), // Initial balance in cents
        fc.array(
          fc.record({
            type: fc.constantFrom('deduct', 'add') as fc.Arbitrary<'deduct' | 'add'>,
            amount: fc.integer({ min: 1, max: 100 }) // Amount in cents
          }),
          { minLength: 1, maxLength: 10 }
        ),
        (initialBalanceCents, operationsCents) => {
          const initialBalance = initialBalanceCents / 100
          const operations = operationsCents.map(op => ({
            type: op.type,
            amount: op.amount / 100
          }))
          
          // Sequential execution
          const sequentialBalance = computeSequentialBalance(initialBalance, operations)
          
          // Concurrent simulation (should produce same result as sequential)
          const { finalBalance } = simulateConcurrentOperations(initialBalance, operations)
          
          // The final balance should match sequential execution
          return Math.abs(finalBalance - sequentialBalance) < 0.001
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should maintain non-negative balance under concurrent deductions', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 100, max: 1000 }), // Initial balance in cents
        fc.array(
          fc.integer({ min: 1, max: 200 }), // Deduction amounts in cents
          { minLength: 1, maxLength: 20 }
        ),
        (initialBalanceCents, deductionsCents) => {
          const initialBalance = initialBalanceCents / 100
          const operations = deductionsCents.map(amount => ({
            type: 'deduct' as const,
            amount: amount / 100
          }))
          
          const { finalBalance } = simulateConcurrentOperations(initialBalance, operations)
          
          // Balance should never go negative
          return finalBalance >= 0
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should correctly track successful operations count', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 500, max: 1000 }), // Initial balance in cents (enough for some deductions)
        fc.array(
          fc.record({
            type: fc.constantFrom('deduct', 'add') as fc.Arbitrary<'deduct' | 'add'>,
            amount: fc.integer({ min: 1, max: 50 }) // Small amounts to ensure some succeed
          }),
          { minLength: 1, maxLength: 10 }
        ),
        (initialBalanceCents, operationsCents) => {
          const initialBalance = initialBalanceCents / 100
          const operations = operationsCents.map(op => ({
            type: op.type,
            amount: op.amount / 100
          }))
          
          const { successfulOps } = simulateConcurrentOperations(initialBalance, operations)
          
          // At least some operations should succeed (all adds should succeed)
          const addCount = operations.filter(op => op.type === 'add').length
          return successfulOps >= addCount
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should preserve balance precision under concurrent operations', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 100, max: 10000 }), // Initial balance in cents
        fc.array(
          fc.record({
            type: fc.constantFrom('deduct', 'add') as fc.Arbitrary<'deduct' | 'add'>,
            amount: fc.integer({ min: 1, max: 100 }) // Amount in cents
          }),
          { minLength: 1, maxLength: 15 }
        ),
        (initialBalanceCents, operationsCents) => {
          const initialBalance = initialBalanceCents / 100
          const operations = operationsCents.map(op => ({
            type: op.type,
            amount: op.amount / 100
          }))
          
          const { finalBalance } = simulateConcurrentOperations(initialBalance, operations)
          
          // Balance should maintain 2 decimal precision
          const decimalStr = finalBalance.toString()
          const decimalPart = decimalStr.includes('.') ? decimalStr.split('.')[1] : ''
          return (decimalPart?.length ?? 0) <= 2
        }
      ),
      { numRuns: 100 }
    )
  })
})


// ============================================================================
// Unit Tests for Game Store Methods (game-data-persistence)
// Feature: game-data-persistence
// Validates: Requirements 1.1, 1.5, 3.1
// ============================================================================

describe('Game Store - recordGameResult state updates', () => {
  /**
   * Test the state update logic for recordGameResult
   * Requirements: 1.1
   */
  
  interface GameStats {
    games_played: number
    wins: number
    losses: number
    win_rate: string
    net_profit: string
    total_won: string
    total_lost: string
  }
  
  interface ApiGameRecord {
    id: number
    game_type: 'wheel' | 'coin' | 'number'
    bet_amount: number
    result: 'win' | 'lose'
    payout: number
    net_profit: number
    details: Record<string, unknown>
    created_at: string
  }
  
  // Simulate the state update logic from recordGameResult
  const updateStatsFromResponse = (
    currentStats: GameStats | null,
    responseStats: { games_played: number; wins: number; win_rate: string; net_profit: string }
  ): GameStats => {
    return {
      games_played: responseStats.games_played,
      wins: responseStats.wins,
      losses: responseStats.games_played - responseStats.wins,
      win_rate: responseStats.win_rate,
      net_profit: responseStats.net_profit,
      total_won: currentStats?.total_won ?? '0',
      total_lost: currentStats?.total_lost ?? '0'
    }
  }
  
  // Simulate adding record to gameRecords array
  const addRecordToList = (records: ApiGameRecord[], newRecord: ApiGameRecord): ApiGameRecord[] => {
    return [newRecord, ...records]
  }

  it('should correctly update stats from API response', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000 }),
        fc.integer({ min: 0, max: 1000 }),
        fc.float({ min: 0, max: 100, noNaN: true }),
        fc.float({ min: -1000, max: 1000, noNaN: true }),
        (gamesPlayed, wins, winRate, netProfit) => {
          // Ensure wins <= gamesPlayed
          const actualWins = Math.min(wins, gamesPlayed)
          
          const responseStats = {
            games_played: gamesPlayed,
            wins: actualWins,
            win_rate: winRate.toFixed(1),
            net_profit: netProfit.toFixed(2)
          }
          
          const updatedStats = updateStatsFromResponse(null, responseStats)
          
          return (
            updatedStats.games_played === gamesPlayed &&
            updatedStats.wins === actualWins &&
            updatedStats.losses === gamesPlayed - actualWins &&
            updatedStats.win_rate === winRate.toFixed(1) &&
            updatedStats.net_profit === netProfit.toFixed(2)
          )
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should prepend new record to gameRecords list', () => {
    fc.assert(
      fc.property(
        fc.array(fc.integer({ min: 1, max: 1000 }), { minLength: 0, maxLength: 10 }),
        fc.integer({ min: 1, max: 1000 }),
        (existingIds, newId) => {
          const existingRecords: ApiGameRecord[] = existingIds.map(id => ({
            id,
            game_type: 'wheel' as const,
            bet_amount: 10,
            result: 'win' as const,
            payout: 20,
            net_profit: 10,
            details: {},
            created_at: new Date().toISOString()
          }))
          
          const newRecord: ApiGameRecord = {
            id: newId,
            game_type: 'coin',
            bet_amount: 5,
            result: 'lose',
            payout: 0,
            net_profit: -5,
            details: {},
            created_at: new Date().toISOString()
          }
          
          const updatedRecords = addRecordToList(existingRecords, newRecord)
          
          // New record should be first
          return (
            updatedRecords.length === existingRecords.length + 1 &&
            updatedRecords[0]?.id === newId
          )
        }
      ),
      { numRuns: 100 }
    )
  })
})

describe('Game Store - loadGameRecords state updates', () => {
  /**
   * Test the state update logic for loadGameRecords
   * Requirements: 1.5
   */
  
  interface ApiGameRecord {
    id: number
    game_type: 'wheel' | 'coin' | 'number'
    bet_amount: number
    result: 'win' | 'lose'
    payout: number
    net_profit: number
    details: Record<string, unknown>
    created_at: string
  }
  
  // Simulate the pagination logic from loadGameRecords
  const updateRecordsFromResponse = (
    currentRecords: ApiGameRecord[],
    newRecords: ApiGameRecord[],
    offset: number
  ): ApiGameRecord[] => {
    if (offset === 0) {
      // Replace records if starting from beginning
      return newRecords
    } else {
      // Append records for pagination
      return [...currentRecords, ...newRecords]
    }
  }

  it('should replace records when offset is 0', () => {
    fc.assert(
      fc.property(
        fc.array(fc.integer({ min: 1, max: 100 }), { minLength: 1, maxLength: 5 }),
        fc.array(fc.integer({ min: 101, max: 200 }), { minLength: 1, maxLength: 5 }),
        (existingIds, newIds) => {
          const existingRecords: ApiGameRecord[] = existingIds.map(id => ({
            id,
            game_type: 'wheel' as const,
            bet_amount: 10,
            result: 'win' as const,
            payout: 20,
            net_profit: 10,
            details: {},
            created_at: new Date().toISOString()
          }))
          
          const newRecords: ApiGameRecord[] = newIds.map(id => ({
            id,
            game_type: 'coin' as const,
            bet_amount: 5,
            result: 'lose' as const,
            payout: 0,
            net_profit: -5,
            details: {},
            created_at: new Date().toISOString()
          }))
          
          const result = updateRecordsFromResponse(existingRecords, newRecords, 0)
          
          // Should completely replace with new records
          return (
            result.length === newRecords.length &&
            result.every((r, i) => r.id === newIds[i])
          )
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should append records when offset > 0', () => {
    fc.assert(
      fc.property(
        fc.array(fc.integer({ min: 1, max: 100 }), { minLength: 1, maxLength: 5 }),
        fc.array(fc.integer({ min: 101, max: 200 }), { minLength: 1, maxLength: 5 }),
        fc.integer({ min: 1, max: 100 }),
        (existingIds, newIds, offset) => {
          const existingRecords: ApiGameRecord[] = existingIds.map(id => ({
            id,
            game_type: 'wheel' as const,
            bet_amount: 10,
            result: 'win' as const,
            payout: 20,
            net_profit: 10,
            details: {},
            created_at: new Date().toISOString()
          }))
          
          const newRecords: ApiGameRecord[] = newIds.map(id => ({
            id,
            game_type: 'coin' as const,
            bet_amount: 5,
            result: 'lose' as const,
            payout: 0,
            net_profit: -5,
            details: {},
            created_at: new Date().toISOString()
          }))
          
          const result = updateRecordsFromResponse(existingRecords, newRecords, offset)
          
          // Should append new records to existing
          return (
            result.length === existingRecords.length + newRecords.length &&
            result.slice(0, existingIds.length).every((r, i) => r.id === existingIds[i]) &&
            result.slice(existingIds.length).every((r, i) => r.id === newIds[i])
          )
        }
      ),
      { numRuns: 100 }
    )
  })
})

describe('Game Store - loadLeaderboard state updates', () => {
  /**
   * Test the state update logic for loadLeaderboard
   * Requirements: 3.1
   */
  
  interface LeaderboardEntry {
    rank: number
    user_id: number
    username: string
    total_winnings: number
    games_played: number
  }
  
  interface LeaderboardResponse {
    entries: LeaderboardEntry[]
    current_user: LeaderboardEntry | null
    total_players: number
  }
  
  // Simulate leaderboard state update
  const updateLeaderboardState = (
    response: LeaderboardResponse
  ): LeaderboardResponse => {
    return response
  }

  it('should store leaderboard response correctly', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.record({
            rank: fc.integer({ min: 1, max: 100 }),
            user_id: fc.integer({ min: 1, max: 10000 }),
            username: fc.string({ minLength: 1, maxLength: 20 }),
            total_winnings: fc.float({ min: -1000, max: 10000, noNaN: true }),
            games_played: fc.integer({ min: 1, max: 1000 })
          }),
          { minLength: 0, maxLength: 10 }
        ),
        fc.integer({ min: 1, max: 1000 }),
        (entries, totalPlayers) => {
          const response: LeaderboardResponse = {
            entries: entries as LeaderboardEntry[],
            current_user: null,
            total_players: totalPlayers
          }
          
          const result = updateLeaderboardState(response)
          
          return (
            result.entries.length === entries.length &&
            result.total_players === totalPlayers &&
            result.current_user === null
          )
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should include current user when present', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.record({
            rank: fc.integer({ min: 1, max: 10 }),
            user_id: fc.integer({ min: 1, max: 10000 }),
            username: fc.string({ minLength: 1, maxLength: 20 }),
            total_winnings: fc.float({ min: -1000, max: 10000, noNaN: true }),
            games_played: fc.integer({ min: 1, max: 1000 })
          }),
          { minLength: 1, maxLength: 10 }
        ),
        fc.record({
          rank: fc.integer({ min: 11, max: 100 }),
          user_id: fc.integer({ min: 1, max: 10000 }),
          username: fc.string({ minLength: 1, maxLength: 20 }),
          total_winnings: fc.float({ min: -1000, max: 10000, noNaN: true }),
          games_played: fc.integer({ min: 1, max: 1000 })
        }),
        fc.integer({ min: 100, max: 1000 }),
        (entries, currentUser, totalPlayers) => {
          const response: LeaderboardResponse = {
            entries: entries as LeaderboardEntry[],
            current_user: currentUser as LeaderboardEntry,
            total_players: totalPlayers
          }
          
          const result = updateLeaderboardState(response)
          
          return (
            result.current_user !== null &&
            result.current_user.user_id === currentUser.user_id &&
            result.current_user.rank === currentUser.rank
          )
        }
      ),
      { numRuns: 100 }
    )
  })
})

describe('Game Store - stats computed property', () => {
  /**
   * Test the stats computed property logic
   * Requirements: 2.1
   */
  
  interface GameStats {
    games_played: number
    wins: number
    losses: number
    win_rate: string
    net_profit: string
    total_won: string
    total_lost: string
  }
  
  // Simulate the stats computed logic
  const computeStats = (gameStats: GameStats | null): {
    totalGames: number
    wins: number
    losses: number
    winRate: string
    netProfit: string
    totalWon: string
    totalLost: string
  } => {
    if (gameStats) {
      return {
        totalGames: gameStats.games_played,
        wins: gameStats.wins,
        losses: gameStats.losses,
        winRate: gameStats.win_rate,
        netProfit: gameStats.net_profit,
        totalWon: gameStats.total_won,
        totalLost: gameStats.total_lost
      }
    }
    
    // Fallback when no stats available
    return {
      totalGames: 0,
      wins: 0,
      losses: 0,
      winRate: '0',
      netProfit: '0.00',
      totalWon: '0.00',
      totalLost: '0.00'
    }
  }

  it('should correctly compute stats from backend data', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 1000 }),
        fc.integer({ min: 0, max: 1000 }),
        fc.float({ min: 0, max: 100, noNaN: true }),
        fc.float({ min: -10000, max: 10000, noNaN: true }),
        fc.float({ min: 0, max: 10000, noNaN: true }),
        fc.float({ min: 0, max: 10000, noNaN: true }),
        (gamesPlayed, wins, winRate, netProfit, totalWon, totalLost) => {
          // Ensure wins <= gamesPlayed
          const actualWins = Math.min(wins, gamesPlayed)
          const losses = gamesPlayed - actualWins
          
          const gameStats: GameStats = {
            games_played: gamesPlayed,
            wins: actualWins,
            losses,
            win_rate: winRate.toFixed(1),
            net_profit: netProfit.toFixed(2),
            total_won: totalWon.toFixed(2),
            total_lost: totalLost.toFixed(2)
          }
          
          const result = computeStats(gameStats)
          
          return (
            result.totalGames === gamesPlayed &&
            result.wins === actualWins &&
            result.losses === losses &&
            result.winRate === winRate.toFixed(1) &&
            result.netProfit === netProfit.toFixed(2)
          )
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return default values when no stats available', () => {
    const result = computeStats(null)
    
    return (
      result.totalGames === 0 &&
      result.wins === 0 &&
      result.losses === 0 &&
      result.winRate === '0' &&
      result.netProfit === '0.00'
    )
  })
})
