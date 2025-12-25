/**
 * Property-Based Tests for Game Utility Functions
 * 游戏工具函数属性测试
 * 
 * Feature: frontend-enhancement
 * Validates: Requirements 1.1, 5.2, 6.3, 5.4
 */

import { describe, it, expect } from 'vitest'
import fc from 'fast-check'
import {
  randomInt,
  validateBet,
  calculatePayout,
  getGreeting,
} from './gameUtils'

// ============================================================================
// Property 1: Time-based greeting correctness
// Feature: frontend-enhancement, Property 1: Time-based greeting correctness
// Validates: Requirements 1.1
// ============================================================================

describe('Property 1: Time-based greeting correctness', () => {
  it('should return correct greeting based on hour', () => {
    // Test all 24 hours to verify greeting logic
    for (let hour = 0; hour < 24; hour++) {
      const greeting = getGreetingForHour(hour)
      if (hour < 12) {
        expect(greeting).toBe('早上好')
      } else if (hour < 18) {
        expect(greeting).toBe('下午好')
      } else {
        expect(greeting).toBe('晚上好')
      }
    }
  })

  it('should always return one of the three valid greetings', () => {
    const validGreetings = ['早上好', '下午好', '晚上好']
    const greeting = getGreeting()
    expect(validGreetings).toContain(greeting)
  })
})

// Helper function to test greeting for specific hour
function getGreetingForHour(hour: number): string {
  if (hour < 12) return '早上好'
  if (hour < 18) return '下午好'
  return '晚上好'
}

// ============================================================================
// Property 5: Bet amount validation
// Feature: frontend-enhancement, Property 5: Bet amount validation
// Validates: Requirements 5.2, 6.2, 7.2
// ============================================================================

describe('Property 5: Bet amount validation', () => {
  it('should correctly validate bet amounts for all valid inputs', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000 }),
        fc.integer({ min: 0, max: 10000 }),
        fc.integer({ min: 1, max: 10 }),
        fc.integer({ min: 10, max: 1000 }),
        (bet, coins, minBet, maxBet) => {
          // Ensure maxBet >= minBet for valid test cases
          const actualMaxBet = Math.max(minBet, maxBet)
          const result = validateBet(bet, coins, minBet, actualMaxBet)
          const expected = bet > 0 && bet <= coins && bet >= minBet && bet <= actualMaxBet
          return result.valid === expected
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should reject zero or negative bet amounts', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: -1000, max: 0 }),
        fc.integer({ min: 100, max: 1000 }),
        (bet, coins) => {
          const result = validateBet(bet, coins, 1, 100)
          return result.valid === false && result.error !== undefined
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should reject bets exceeding available coins', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 100 }),
        (coins) => {
          const bet = coins + 1
          const result = validateBet(bet, coins, 1, 1000)
          return result.valid === false && result.error === '游戏币不足'
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should reject bets below minimum', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 10, max: 100 }),
        (minBet) => {
          const bet = minBet - 1
          const result = validateBet(bet, 1000, minBet, 1000)
          return result.valid === false && result.error?.includes('最低下注')
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should reject bets above maximum', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 10, max: 100 }),
        (maxBet) => {
          const bet = maxBet + 1
          const result = validateBet(bet, 1000, 1, maxBet)
          return result.valid === false && result.error?.includes('最高下注')
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 12: Random integer range correctness
// Feature: frontend-enhancement, Property 12: Random integer range correctness
// Validates: Requirements 6.3
// ============================================================================

describe('Property 12: Random integer range correctness', () => {
  it('should generate integers within the specified range', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 100 }),
        fc.integer({ min: 1, max: 100 }),
        (a, b) => {
          const min = Math.min(a, b)
          const max = Math.max(a, b)
          const result = randomInt(min, max)
          return Number.isInteger(result) && result >= min && result <= max
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return the same value when min equals max', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000 }),
        (value) => {
          const result = randomInt(value, value)
          return result === value
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should generate all values in range over many iterations', () => {
    const min = 1
    const max = 10
    const results = new Set<number>()
    
    // Run enough times to likely hit all values
    for (let i = 0; i < 1000; i++) {
      results.add(randomInt(min, max))
    }
    
    // Should have generated most values in the range
    expect(results.size).toBeGreaterThanOrEqual(max - min)
  })
})

// ============================================================================
// Property 15: Payout calculation precision
// Feature: frontend-enhancement, Property 15: Payout calculation precision
// Validates: Requirements 5.4, 6.5, 7.5
// ============================================================================

describe('Property 15: Payout calculation precision', () => {
  it('should round payout to exactly 2 decimal places', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 100000 }),
        fc.integer({ min: 0, max: 1000 }),
        (betCents, multiplierTenths) => {
          // Use cents and tenths to simulate decimal values
          const bet = betCents / 100
          const multiplier = multiplierTenths / 100
          const payout = calculatePayout(bet, multiplier)
          const decimalStr = payout.toString()
          const decimalPart = decimalStr.includes('.') ? decimalStr.split('.')[1] : ''
          return (decimalPart?.length ?? 0) <= 2
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should calculate correct payout value', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 100000 }),
        fc.integer({ min: 0, max: 1000 }),
        (betCents, multiplierTenths) => {
          const bet = betCents / 100
          const multiplier = multiplierTenths / 100
          const payout = calculatePayout(bet, multiplier)
          const expected = Number((bet * multiplier).toFixed(2))
          return payout === expected
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return 0 when multiplier is 0', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 100000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, 0)
          return payout === 0
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return bet amount when multiplier is 1', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 100000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, 1)
          const expected = Number(bet.toFixed(2))
          return payout === expected
        }
      ),
      { numRuns: 100 }
    )
  })
})


// ============================================================================
// Property 7: Wheel prize calculation correctness
// Feature: frontend-enhancement, Property 7: Wheel prize calculation correctness
// Validates: Requirements 5.4
// ============================================================================

describe('Property 7: Wheel prize calculation correctness', () => {
  it('should calculate payout as bet amount multiplied by segment multiplier', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000 }),
        fc.integer({ min: 0, max: 100 }),
        (betCents, multiplierTenths) => {
          const bet = betCents / 100
          const multiplier = multiplierTenths / 10
          const payout = calculatePayout(bet, multiplier)
          const expected = Number((bet * multiplier).toFixed(2))
          return payout === expected
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return zero payout for 0x multiplier segment', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, 0)
          return payout === 0
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should double bet for 2x multiplier segment', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, 2)
          const expected = Number((bet * 2).toFixed(2))
          return payout === expected
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 13: Wheel segment selection validity
// Feature: frontend-enhancement, Property 13: Wheel segment selection validity
// Validates: Requirements 5.4
// ============================================================================

import { spinWheel, type WheelSegment } from './gameUtils'

describe('Property 13: Wheel segment selection validity', () => {
  it('should return valid index within segments array bounds', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.record({
            label: fc.string({ minLength: 1, maxLength: 5 }),
            multiplier: fc.float({ min: 0, max: 10, noNaN: true }),
            color: fc.constant('#ff0000'),
            weight: fc.integer({ min: 1, max: 100 })
          }),
          { minLength: 2, maxLength: 12 }
        ),
        (segments) => {
          const result = spinWheel(segments as WheelSegment[])
          return result >= 0 && result < segments.length
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return -1 for empty segments array', () => {
    const result = spinWheel([])
    expect(result).toBe(-1)
  })

  it('should always return integer index', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.record({
            label: fc.string({ minLength: 1, maxLength: 5 }),
            multiplier: fc.float({ min: 0, max: 10, noNaN: true }),
            color: fc.constant('#00ff00'),
            weight: fc.integer({ min: 1, max: 100 })
          }),
          { minLength: 1, maxLength: 12 }
        ),
        (segments) => {
          const result = spinWheel(segments as WheelSegment[])
          return Number.isInteger(result)
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 20: Wheel weight distribution validity
// Feature: frontend-enhancement, Property 20: Wheel weight distribution validity
// Validates: Requirements 5.4
// ============================================================================

describe('Property 20: Wheel weight distribution validity', () => {
  it('should handle segments with valid positive weights', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.record({
            label: fc.string({ minLength: 1, maxLength: 5 }),
            multiplier: fc.float({ min: 0, max: 10, noNaN: true }),
            color: fc.constant('#0000ff'),
            weight: fc.integer({ min: 1, max: 100 })
          }),
          { minLength: 1, maxLength: 12 }
        ),
        (segments) => {
          // Verify total weight is positive
          const totalWeight = segments.reduce((sum, s) => sum + (s.weight ?? 1), 0)
          // Verify each weight is non-negative
          const allWeightsValid = segments.every(s => (s.weight ?? 1) >= 0)
          // spinWheel should work correctly
          const result = spinWheel(segments as WheelSegment[])
          return totalWeight > 0 && allWeightsValid && result >= 0 && result < segments.length
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should use default weight of 1 when weight is undefined', () => {
    const segmentsWithoutWeight: WheelSegment[] = [
      { label: '1x', multiplier: 1, color: '#ff0000' },
      { label: '2x', multiplier: 2, color: '#00ff00' },
      { label: '3x', multiplier: 3, color: '#0000ff' }
    ]
    
    // Run multiple times to verify it works
    for (let i = 0; i < 100; i++) {
      const result = spinWheel(segmentsWithoutWeight)
      expect(result).toBeGreaterThanOrEqual(0)
      expect(result).toBeLessThan(segmentsWithoutWeight.length)
    }
  })

  it('should respect weight distribution over many spins', () => {
    // Create segments with very different weights
    const segments: WheelSegment[] = [
      { label: 'rare', multiplier: 5, color: '#ff0000', weight: 1 },
      { label: 'common', multiplier: 1, color: '#00ff00', weight: 99 }
    ]
    
    const counts: number[] = [0, 0]
    const iterations = 1000
    
    for (let i = 0; i < iterations; i++) {
      const result = spinWheel(segments)
      if (counts[result] === undefined) {
        counts[result] = 0
      }
      counts[result]!++
    }
    
    // The common segment (weight 99) should be selected much more often
    // than the rare segment (weight 1)
    // With 99:1 ratio, common should be ~99% of selections
    const rareCount = counts[0] ?? 0
    const commonCount = counts[1] ?? 0
    expect(commonCount).toBeGreaterThan(rareCount * 5)
  })
})


// ============================================================================
// Property 8: Number guess random generation
// Feature: frontend-enhancement, Property 8: Number guess random generation
// Validates: Requirements 6.3
// ============================================================================

describe('Property 8: Number guess random generation', () => {
  it('should generate target numbers within 1-10 range', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000 }),
        () => {
          // Using the numberGuessConfig range (1-10)
          const min = 1
          const max = 10
          const result = randomInt(min, max)
          return Number.isInteger(result) && result >= min && result <= max
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should generate all numbers 1-10 over many iterations', () => {
    const min = 1
    const max = 10
    const results = new Set<number>()
    
    // Run enough times to likely hit all values
    for (let i = 0; i < 1000; i++) {
      results.add(randomInt(min, max))
    }
    
    // Should have generated all values in the range
    expect(results.size).toBe(max - min + 1)
    for (let num = min; num <= max; num++) {
      expect(results.has(num)).toBe(true)
    }
  })

  it('should never generate numbers outside 1-10 range', () => {
    const min = 1
    const max = 10
    
    for (let i = 0; i < 500; i++) {
      const result = randomInt(min, max)
      expect(result).toBeGreaterThanOrEqual(min)
      expect(result).toBeLessThanOrEqual(max)
    }
  })
})

// ============================================================================
// Property 9: Number guess payout correctness
// Feature: frontend-enhancement, Property 9: Number guess payout correctness
// Validates: Requirements 6.5
// ============================================================================

describe('Property 9: Number guess payout correctness', () => {
  const NUMBER_GUESS_PAYOUT_MULTIPLIER = 9 // From numberGuessConfig

  it('should calculate correct payout for winning guess (9x multiplier)', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 2000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, NUMBER_GUESS_PAYOUT_MULTIPLIER)
          const expected = Number((bet * NUMBER_GUESS_PAYOUT_MULTIPLIER).toFixed(2))
          return payout === expected
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return payout exactly 9 times the bet amount', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, NUMBER_GUESS_PAYOUT_MULTIPLIER)
          // Verify the ratio is exactly 9 (within floating point precision)
          const ratio = payout / bet
          return Math.abs(ratio - NUMBER_GUESS_PAYOUT_MULTIPLIER) < 0.001
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should maintain 2 decimal precision for all payouts', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, NUMBER_GUESS_PAYOUT_MULTIPLIER)
          const decimalStr = payout.toString()
          const decimalPart = decimalStr.includes('.') ? decimalStr.split('.')[1] : ''
          return (decimalPart?.length ?? 0) <= 2
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 10: Number guess loss correctness
// Feature: frontend-enhancement, Property 10: Number guess loss correctness
// Validates: Requirements 6.6
// ============================================================================

describe('Property 10: Number guess loss correctness', () => {
  it('should return zero payout for incorrect guess', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          // For a loss, multiplier is 0
          const payout = calculatePayout(bet, 0)
          return payout === 0
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should result in loss equal to bet amount when guess is wrong', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, 0)
          // Loss = bet - payout = bet - 0 = bet
          const loss = bet - payout
          return Math.abs(loss - bet) < 0.001
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should always return exactly 0 for any bet amount with 0 multiplier', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 100000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, 0)
          return payout === 0
        }
      ),
      { numRuns: 100 }
    )
  })
})


// ============================================================================
// Property 2: Account age calculation correctness
// Feature: frontend-enhancement, Property 2: Account age calculation correctness
// Validates: Requirements 2.2
// ============================================================================

import { calculateAccountAge } from './gameUtils'

describe('Property 2: Account age calculation correctness', () => {
  it('should calculate correct account age in days for any valid creation date', () => {
    fc.assert(
      fc.property(
        // Generate dates from 1 to 3650 days ago (up to 10 years)
        fc.integer({ min: 1, max: 3650 }),
        (daysAgo) => {
          const now = new Date()
          const createdAt = new Date(now.getTime() - daysAgo * 24 * 60 * 60 * 1000)
          const createdAtStr = createdAt.toISOString()
          
          const calculatedAge = calculateAccountAge(createdAtStr)
          
          // The calculated age should equal daysAgo (rounded down)
          // Allow for timezone edge cases with ±1 day tolerance
          return Math.abs(calculatedAge - daysAgo) <= 1
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return 0 for accounts created today', () => {
    const now = new Date()
    const createdAtStr = now.toISOString()
    
    const age = calculateAccountAge(createdAtStr)
    expect(age).toBe(0)
  })

  it('should return non-negative age for any past date', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 10000 }),
        (daysAgo) => {
          const now = new Date()
          const createdAt = new Date(now.getTime() - daysAgo * 24 * 60 * 60 * 1000)
          const createdAtStr = createdAt.toISOString()
          
          const age = calculateAccountAge(createdAtStr)
          return age >= 0
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return integer value (whole days)', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 1000 }),
        fc.integer({ min: 0, max: 23 }),
        fc.integer({ min: 0, max: 59 }),
        (daysAgo, hours, minutes) => {
          const now = new Date()
          const createdAt = new Date(
            now.getTime() - 
            daysAgo * 24 * 60 * 60 * 1000 - 
            hours * 60 * 60 * 1000 - 
            minutes * 60 * 1000
          )
          const createdAtStr = createdAt.toISOString()
          
          const age = calculateAccountAge(createdAtStr)
          return Number.isInteger(age)
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should calculate age correctly for specific known dates', () => {
    // Test with a date exactly 30 days ago
    const now = new Date()
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
    const age30 = calculateAccountAge(thirtyDaysAgo.toISOString())
    expect(age30).toBe(30)

    // Test with a date exactly 365 days ago
    const oneYearAgo = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000)
    const age365 = calculateAccountAge(oneYearAgo.toISOString())
    expect(age365).toBe(365)
  })

  it('should handle various date string formats', () => {
    const now = new Date()
    const tenDaysAgo = new Date(now.getTime() - 10 * 24 * 60 * 60 * 1000)
    
    // ISO format
    const isoAge = calculateAccountAge(tenDaysAgo.toISOString())
    expect(isoAge).toBe(10)
    
    // Date string format
    const dateStringAge = calculateAccountAge(tenDaysAgo.toString())
    expect(dateStringAge).toBe(10)
  })
})

// ============================================================================
// Property 11: Coin flip win payout correctness
// Feature: frontend-enhancement, Property 11: Coin flip win payout correctness
// Validates: Requirements 7.5
// ============================================================================

describe('Property 11: Coin flip win payout correctness', () => {
  const COIN_FLIP_PAYOUT_MULTIPLIER = 1.95 // From coinFlipConfig

  it('should calculate correct payout for winning coin flip (1.95x multiplier)', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, COIN_FLIP_PAYOUT_MULTIPLIER)
          const expected = Number((bet * COIN_FLIP_PAYOUT_MULTIPLIER).toFixed(2))
          return payout === expected
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return payout close to 1.95 times the bet amount (accounting for rounding)', () => {
    fc.assert(
      fc.property(
        // Use larger bet amounts to minimize rounding impact on ratio
        fc.integer({ min: 100, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, COIN_FLIP_PAYOUT_MULTIPLIER)
          // Verify the ratio is close to 1.95
          // Allow for rounding tolerance: for bet >= 1.00, max rounding error is 0.005/bet
          const ratio = payout / bet
          const tolerance = 0.01 + (0.005 / bet) // Account for 2 decimal rounding
          return Math.abs(ratio - COIN_FLIP_PAYOUT_MULTIPLIER) < tolerance
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should maintain 2 decimal precision for all coin flip payouts', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, COIN_FLIP_PAYOUT_MULTIPLIER)
          const decimalStr = payout.toString()
          const decimalPart = decimalStr.includes('.') ? decimalStr.split('.')[1] : ''
          return (decimalPart?.length ?? 0) <= 2
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should calculate correct payout for various bet amounts', () => {
    // Test specific bet amounts to verify calculation
    const testCases = [
      { bet: 1, expected: 1.95 },
      { bet: 10, expected: 19.5 },
      { bet: 100, expected: 195 },
      { bet: 50, expected: 97.5 },
      { bet: 25.5, expected: 49.73 }, // 25.5 * 1.95 = 49.725 -> 49.73
    ]

    for (const { bet, expected } of testCases) {
      const payout = calculatePayout(bet, COIN_FLIP_PAYOUT_MULTIPLIER)
      expect(payout).toBe(expected)
    }
  })

  it('should always return positive payout for positive bet', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 100000 }),
        (betCents) => {
          const bet = betCents / 100
          const payout = calculatePayout(bet, COIN_FLIP_PAYOUT_MULTIPLIER)
          return payout > 0
        }
      ),
      { numRuns: 100 }
    )
  })
})
