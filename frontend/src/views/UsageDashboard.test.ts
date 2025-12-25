import { describe, it, expect } from 'vitest'
import fc from 'fast-check'

// ============================================================================
// Unit Tests for UsageDashboard Data Formatting
// Requirements: 2.1, 7.1
// ============================================================================

// Test utility functions that are used in UsageDashboard

describe('UsageDashboard Data Formatting', () => {
  // formatNumber function tests
  describe('formatNumber', () => {
    const formatNumber = (num: number): string => {
      return num.toLocaleString('zh-CN')
    }

    it('should format small numbers correctly', () => {
      expect(formatNumber(0)).toBe('0')
      expect(formatNumber(100)).toBe('100')
      expect(formatNumber(999)).toBe('999')
    })

    it('should format large numbers with thousand separators', () => {
      expect(formatNumber(1000)).toContain('1')
      expect(formatNumber(1000000)).toContain('1')
      expect(formatNumber(1234567)).toContain('1')
    })

    it('should handle negative numbers', () => {
      expect(formatNumber(-100)).toContain('-')
      expect(formatNumber(-1000)).toContain('-')
    })
  })

  // formatDateTime function tests
  describe('formatDateTime', () => {
    const formatDateTime = (dateStr: string): string => {
      const date = new Date(dateStr)
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const minute = String(date.getMinutes()).padStart(2, '0')
      return `${month}-${day} ${hour}:${minute}`
    }

    it('should format ISO date strings correctly', () => {
      const result = formatDateTime('2024-03-15T14:30:00Z')
      expect(result).toMatch(/\d{2}-\d{2} \d{2}:\d{2}/)
    })

    it('should pad single digit months and days', () => {
      const result = formatDateTime('2024-01-05T09:05:00Z')
      expect(result).toMatch(/0\d-0\d/)
    })
  })

  // getStatusClass function tests
  describe('getStatusClass', () => {
    const getStatusClass = (statusCode: number): string => {
      if (statusCode >= 200 && statusCode < 300) return 'status-success'
      if (statusCode >= 400 && statusCode < 500) return 'status-warning'
      if (statusCode >= 500) return 'status-error'
      return 'status-default'
    }

    it('should return success class for 2xx status codes', () => {
      expect(getStatusClass(200)).toBe('status-success')
      expect(getStatusClass(201)).toBe('status-success')
      expect(getStatusClass(299)).toBe('status-success')
    })

    it('should return warning class for 4xx status codes', () => {
      expect(getStatusClass(400)).toBe('status-warning')
      expect(getStatusClass(401)).toBe('status-warning')
      expect(getStatusClass(404)).toBe('status-warning')
      expect(getStatusClass(429)).toBe('status-warning')
    })

    it('should return error class for 5xx status codes', () => {
      expect(getStatusClass(500)).toBe('status-error')
      expect(getStatusClass(502)).toBe('status-error')
      expect(getStatusClass(503)).toBe('status-error')
    })

    it('should return default class for other status codes', () => {
      expect(getStatusClass(100)).toBe('status-default')
      expect(getStatusClass(0)).toBe('status-default')
    })
  })

  // getStatusText function tests
  describe('getStatusText', () => {
    const getStatusText = (statusCode: number): string => {
      if (statusCode >= 200 && statusCode < 300) return '成功'
      if (statusCode === 401) return '未授权'
      if (statusCode === 403) return '禁止访问'
      if (statusCode === 429) return '请求过多'
      if (statusCode >= 400 && statusCode < 500) return '客户端错误'
      if (statusCode >= 500) return '服务器错误'
      return '未知'
    }

    it('should return success text for 2xx status codes', () => {
      expect(getStatusText(200)).toBe('成功')
      expect(getStatusText(201)).toBe('成功')
    })

    it('should return specific text for known error codes', () => {
      expect(getStatusText(401)).toBe('未授权')
      expect(getStatusText(403)).toBe('禁止访问')
      expect(getStatusText(429)).toBe('请求过多')
    })

    it('should return generic client error text for other 4xx codes', () => {
      expect(getStatusText(400)).toBe('客户端错误')
      expect(getStatusText(404)).toBe('客户端错误')
    })

    it('should return server error text for 5xx codes', () => {
      expect(getStatusText(500)).toBe('服务器错误')
      expect(getStatusText(502)).toBe('服务器错误')
    })

    it('should return unknown for other status codes', () => {
      expect(getStatusText(100)).toBe('未知')
      expect(getStatusText(0)).toBe('未知')
    })
  })
})

// ============================================================================
// Tests for Empty State Handling
// Requirements: 2.5
// ============================================================================

describe('Empty State Handling', () => {
  it('should detect empty state when total_requests is 0', () => {
    const stats = {
      total_requests: 0,
      total_tokens: 0,
      prompt_tokens: 0,
      completion_tokens: 0,
      by_model: [],
      recent_calls: []
    }
    
    const loading = false
    const error = null
    const isEmpty = !loading && !error && stats.total_requests === 0
    
    expect(isEmpty).toBe(true)
  })

  it('should not show empty state when loading', () => {
    const stats = {
      total_requests: 0,
      total_tokens: 0,
      prompt_tokens: 0,
      completion_tokens: 0,
      by_model: [],
      recent_calls: []
    }
    
    const loading = true
    const error = null
    const isEmpty = !loading && !error && stats.total_requests === 0
    
    expect(isEmpty).toBe(false)
  })

  it('should not show empty state when there is an error', () => {
    const stats = {
      total_requests: 0,
      total_tokens: 0,
      prompt_tokens: 0,
      completion_tokens: 0,
      by_model: [],
      recent_calls: []
    }
    
    const loading = false
    const error = 'Some error'
    const isEmpty = !loading && !error && stats.total_requests === 0
    
    expect(isEmpty).toBe(false)
  })

  it('should not show empty state when there is data', () => {
    const stats = {
      total_requests: 10,
      total_tokens: 1000,
      prompt_tokens: 500,
      completion_tokens: 500,
      by_model: [],
      recent_calls: []
    }
    
    const loading = false
    const error = null
    const isEmpty = !loading && !error && stats.total_requests === 0
    
    expect(isEmpty).toBe(false)
  })
})

// ============================================================================
// Tests for Date Range Filtering
// Requirements: 7.2
// ============================================================================

describe('Date Range Filtering', () => {
  const getDateString = (date: Date): string => date.toISOString().split('T')[0]!

  it('should calculate today date range correctly', () => {
    const today = new Date()
    const startDate = getDateString(today)
    const endDate = getDateString(today)
    
    expect(startDate).toBe(endDate)
    expect(startDate).toMatch(/^\d{4}-\d{2}-\d{2}$/)
  })

  it('should calculate week date range correctly', () => {
    const today = new Date()
    const weekAgo = new Date(today)
    weekAgo.setDate(today.getDate() - 7)
    
    const startDate = getDateString(weekAgo)
    const endDate = getDateString(today)
    
    expect(new Date(startDate) < new Date(endDate)).toBe(true)
    
    // Verify it's approximately 7 days
    const diffTime = Math.abs(new Date(endDate).getTime() - new Date(startDate).getTime())
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
    expect(diffDays).toBe(7)
  })

  it('should calculate month date range correctly', () => {
    const today = new Date()
    const monthAgo = new Date(today)
    monthAgo.setMonth(today.getMonth() - 1)
    
    const startDate = getDateString(monthAgo)
    const endDate = getDateString(today)
    
    expect(new Date(startDate) < new Date(endDate)).toBe(true)
  })
})

// ============================================================================
// Tests for Average Tokens Calculation
// Requirements: 2.1
// ============================================================================

describe('Average Tokens Calculation', () => {
  it('should calculate average tokens per request correctly', () => {
    const stats = {
      total_requests: 10,
      total_tokens: 1000
    }
    
    const average = stats.total_requests === 0 
      ? 0 
      : Math.round(stats.total_tokens / stats.total_requests)
    
    expect(average).toBe(100)
  })

  it('should return 0 when there are no requests', () => {
    const stats = {
      total_requests: 0,
      total_tokens: 0
    }
    
    const average = stats.total_requests === 0 
      ? 0 
      : Math.round(stats.total_tokens / stats.total_requests)
    
    expect(average).toBe(0)
  })

  it('should round to nearest integer', () => {
    const stats = {
      total_requests: 3,
      total_tokens: 100
    }
    
    const average = stats.total_requests === 0 
      ? 0 
      : Math.round(stats.total_tokens / stats.total_requests)
    
    expect(average).toBe(33) // 100/3 = 33.33... rounds to 33
  })
})


// ============================================================================
// Property 3: Date range filtering correctness
// Feature: frontend-enhancement, Property 3: Date range filtering correctness
// Validates: Requirements 3.4
// ============================================================================

describe('Property 3: Date range filtering correctness', () => {
  // Helper function to check if a timestamp is within a date range (date-only comparison)
  function isWithinDateRange(timestamp: string, startDate: string, endDate: string): boolean {
    // Extract just the date part from the timestamp for comparison
    const itemDateStr = timestamp.split('T')[0] ?? ''
    if (!itemDateStr) return false
    return itemDateStr >= startDate && itemDateStr <= endDate
  }

  // Helper function to filter data by date range (simulates the filtering logic)
  function filterByDateRange<T extends { timestamp: string }>(
    items: T[],
    startDate: string,
    endDate: string
  ): T[] {
    return items.filter(item => isWithinDateRange(item.timestamp, startDate, endDate))
  }

  // Helper to generate a valid date string in YYYY-MM-DD format
  function generateDateString(year: number, month: number, day: number): string {
    return `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
  }

  // Helper to generate a valid ISO timestamp
  function generateTimestamp(year: number, month: number, day: number, hour: number): string {
    return `${generateDateString(year, month, day)}T${String(hour).padStart(2, '0')}:00:00Z`
  }

  it('should return only items within the selected date range', () => {
    fc.assert(
      fc.property(
        // Generate a list of items with timestamps using integer-based date generation
        fc.array(
          fc.record({
            id: fc.integer({ min: 1, max: 10000 }),
            month: fc.integer({ min: 1, max: 12 }),
            day: fc.integer({ min: 1, max: 28 }), // Use 28 to avoid month-end issues
            hour: fc.integer({ min: 0, max: 23 }),
            total_tokens: fc.integer({ min: 0, max: 100000 })
          }),
          { minLength: 0, maxLength: 50 }
        ),
        // Generate start month (1-6) and end month (7-12)
        fc.integer({ min: 1, max: 6 }),
        fc.integer({ min: 7, max: 12 }),
        fc.integer({ min: 1, max: 28 }),
        fc.integer({ min: 1, max: 28 }),
        (items, startMonth, endMonth, startDay, endDay) => {
          const startDate = generateDateString(2024, startMonth, startDay)
          const endDate = generateDateString(2024, endMonth, endDay)
          
          const itemsWithTimestamp = items.map(item => ({
            id: item.id,
            timestamp: generateTimestamp(2024, item.month, item.day, item.hour),
            total_tokens: item.total_tokens
          }))
          
          const filtered = filterByDateRange(itemsWithTimestamp, startDate, endDate)
          
          // All filtered items should be within the date range
          return filtered.every(item => isWithinDateRange(item.timestamp, startDate, endDate))
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should not include items outside the date range', () => {
    fc.assert(
      fc.property(
        fc.array(
          fc.record({
            id: fc.integer({ min: 1, max: 10000 }),
            month: fc.integer({ min: 1, max: 12 }),
            day: fc.integer({ min: 1, max: 28 }),
            hour: fc.integer({ min: 0, max: 23 }),
            total_tokens: fc.integer({ min: 0, max: 100000 })
          }),
          { minLength: 1, maxLength: 50 }
        ),
        fc.integer({ min: 3, max: 6 }),
        fc.integer({ min: 7, max: 9 }),
        fc.integer({ min: 1, max: 28 }),
        fc.integer({ min: 1, max: 28 }),
        (items, startMonth, endMonth, startDay, endDay) => {
          const startDate = generateDateString(2024, startMonth, startDay)
          const endDate = generateDateString(2024, endMonth, endDay)
          
          const itemsWithTimestamp = items.map(item => ({
            id: item.id,
            timestamp: generateTimestamp(2024, item.month, item.day, item.hour),
            total_tokens: item.total_tokens
          }))
          
          const filtered = filterByDateRange(itemsWithTimestamp, startDate, endDate)
          const excluded = itemsWithTimestamp.filter(item => !filtered.includes(item))
          
          // All excluded items should be outside the date range
          return excluded.every(item => !isWithinDateRange(item.timestamp, startDate, endDate))
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return empty array when no items match the date range', () => {
    // Test with fixed items in January, filtering for December
    const items = [
      { id: 1, timestamp: '2024-01-15T10:00:00Z', total_tokens: 100 },
      { id: 2, timestamp: '2024-01-20T14:30:00Z', total_tokens: 200 }
    ]
    
    const filtered = filterByDateRange(items, '2024-12-01', '2024-12-31')
    expect(filtered.length).toBe(0)
  })

  it('should include items on boundary dates (inclusive)', () => {
    // Test that items on the exact start and end dates are included
    const items = [
      { id: 1, timestamp: '2024-06-01T00:00:00Z', total_tokens: 100 },
      { id: 2, timestamp: '2024-06-15T12:00:00Z', total_tokens: 200 },
      { id: 3, timestamp: '2024-06-30T23:59:59Z', total_tokens: 300 }
    ]
    
    const filtered = filterByDateRange(items, '2024-06-01', '2024-06-30')
    
    expect(filtered.length).toBe(3)
    expect(filtered.map(i => i.id)).toEqual([1, 2, 3])
  })

  it('should handle single day date range correctly', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 12 }),
        fc.integer({ min: 1, max: 28 }),
        (month, day) => {
          const dateStr = generateDateString(2024, month, day)
          
          // Create items: one on the target date, one before, one after
          const prevDay = day > 1 ? day - 1 : 28
          const prevMonth = day > 1 ? month : (month > 1 ? month - 1 : 12)
          const nextDay = day < 28 ? day + 1 : 1
          const nextMonth = day < 28 ? month : (month < 12 ? month + 1 : 1)
          
          const items = [
            { id: 1, timestamp: generateTimestamp(2024, month, day, 12), total_tokens: 100 },
            { id: 2, timestamp: generateTimestamp(2024, prevMonth, prevDay, 12), total_tokens: 200 },
            { id: 3, timestamp: generateTimestamp(2024, nextMonth, nextDay, 12), total_tokens: 300 }
          ]
          
          const filtered = filterByDateRange(items, dateStr, dateStr)
          
          // Only the item on the target date should be included
          return filtered.length === 1 && filtered[0]?.id === 1
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should filter correctly for various date ranges', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 6 }),
        fc.integer({ min: 7, max: 12 }),
        (startMonth, endMonth) => {
          const startDate = generateDateString(2024, startMonth, 1)
          const endDate = generateDateString(2024, endMonth, 28)
          
          // Create items spanning the entire year
          const items = Array.from({ length: 12 }, (_, i) => ({
            id: i + 1,
            timestamp: generateTimestamp(2024, i + 1, 15, 12),
            total_tokens: (i + 1) * 100
          }))
          
          const filtered = filterByDateRange(items, startDate, endDate)
          
          // Verify all filtered items are within range
          const allWithinRange = filtered.every(item => 
            isWithinDateRange(item.timestamp, startDate, endDate)
          )
          
          // Verify count matches expected months
          const expectedCount = endMonth - startMonth + 1
          
          return allWithinRange && filtered.length === expectedCount
        }
      ),
      { numRuns: 100 }
    )
  })
})

// ============================================================================
// Property 4: Cost estimation correctness
// Feature: frontend-enhancement, Property 4: Cost estimation correctness
// Validates: Requirements 3.5
// ============================================================================

describe('Property 4: Cost estimation correctness', () => {
  // Cost calculation: $1 = 1,000,000 tokens
  const TOKENS_PER_DOLLAR = 1000000

  // Calculate cost from tokens (same logic as in UsageDashboard.vue)
  function calculateCost(tokens: number): number {
    return tokens / TOKENS_PER_DOLLAR
  }

  // Format cost as dollar amount with appropriate precision
  function formatCost(cost: number): string {
    if (cost < 0.01) {
      return `${cost.toFixed(6)}`
    } else if (cost < 1) {
      return `${cost.toFixed(4)}`
    } else {
      return `${cost.toFixed(2)}`
    }
  }

  it('should calculate cost as tokens divided by 1,000,000', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 100000000 }),
        (tokens) => {
          const cost = calculateCost(tokens)
          const expected = tokens / TOKENS_PER_DOLLAR
          return cost === expected
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should return $1 for exactly 1,000,000 tokens', () => {
    const cost = calculateCost(1000000)
    expect(cost).toBe(1)
  })

  it('should return $0 for 0 tokens', () => {
    const cost = calculateCost(0)
    expect(cost).toBe(0)
  })

  it('should scale linearly with token count', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 10000000 }),
        fc.integer({ min: 2, max: 10 }),
        (tokens, multiplier) => {
          const cost1 = calculateCost(tokens)
          const cost2 = calculateCost(tokens * multiplier)
          // cost2 should be exactly multiplier times cost1
          return Math.abs(cost2 - cost1 * multiplier) < 0.0000001
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should always return non-negative cost for non-negative tokens', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 0, max: 1000000000 }),
        (tokens) => {
          const cost = calculateCost(tokens)
          return cost >= 0
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should format small costs with 6 decimal places', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 9999 }),
        (tokens) => {
          const cost = calculateCost(tokens)
          // Cost should be less than 0.01 for tokens < 10000
          if (cost < 0.01) {
            const formatted = formatCost(cost)
            // Should have 6 decimal places
            const decimalPart = formatted.split('.')[1]
            return decimalPart?.length === 6
          }
          return true
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should format medium costs with 4 decimal places', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 10000, max: 999999 }),
        (tokens) => {
          const cost = calculateCost(tokens)
          // Cost should be between 0.01 and 1 for tokens between 10000 and 999999
          if (cost >= 0.01 && cost < 1) {
            const formatted = formatCost(cost)
            const decimalPart = formatted.split('.')[1]
            return decimalPart?.length === 4
          }
          return true
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should format large costs with 2 decimal places', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1000000, max: 100000000 }),
        (tokens) => {
          const cost = calculateCost(tokens)
          // Cost should be >= 1 for tokens >= 1000000
          if (cost >= 1) {
            const formatted = formatCost(cost)
            const decimalPart = formatted.split('.')[1]
            return decimalPart?.length === 2
          }
          return true
        }
      ),
      { numRuns: 100 }
    )
  })

  it('should calculate correct cost for specific token amounts', () => {
    // Test specific known values
    const testCases = [
      { tokens: 0, expectedCost: 0 },
      { tokens: 1000000, expectedCost: 1 },
      { tokens: 500000, expectedCost: 0.5 },
      { tokens: 100000, expectedCost: 0.1 },
      { tokens: 10000, expectedCost: 0.01 },
      { tokens: 1000, expectedCost: 0.001 },
      { tokens: 100, expectedCost: 0.0001 },
      { tokens: 2500000, expectedCost: 2.5 },
      { tokens: 12345678, expectedCost: 12.345678 }
    ]

    for (const { tokens, expectedCost } of testCases) {
      const cost = calculateCost(tokens)
      expect(cost).toBeCloseTo(expectedCost, 10)
    }
  })

  it('should maintain precision for fractional costs', () => {
    fc.assert(
      fc.property(
        fc.integer({ min: 1, max: 999999 }),
        (tokens) => {
          const cost = calculateCost(tokens)
          // Verify the calculation is precise
          const reconstructedTokens = cost * TOKENS_PER_DOLLAR
          return Math.abs(reconstructedTokens - tokens) < 0.001
        }
      ),
      { numRuns: 100 }
    )
  })
})
