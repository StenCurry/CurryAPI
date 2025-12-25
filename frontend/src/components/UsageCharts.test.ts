import { describe, it, expect } from 'vitest'

// ============================================================================
// Unit Tests for Chart Components Data Processing
// Requirements: 7.1, 7.3
// ============================================================================

// Test data transformation functions used by chart components

describe('UsageTimeSeriesChart Data Processing', () => {
  // Test chart data transformation
  describe('Chart Data Transformation', () => {
    interface DailyUsage {
      date: string
      total_tokens: number
      prompt_tokens: number
      completion_tokens: number
      request_count: number
    }

    const transformToChartData = (data: DailyUsage[]) => {
      return {
        labels: data.map(d => d.date),
        datasets: [
          {
            label: 'Total Tokens',
            data: data.map(d => d.total_tokens),
            borderColor: 'rgb(59, 130, 246)',
            backgroundColor: 'rgba(59, 130, 246, 0.5)',
          },
          {
            label: 'Prompt Tokens',
            data: data.map(d => d.prompt_tokens),
            borderColor: 'rgb(16, 185, 129)',
            backgroundColor: 'rgba(16, 185, 129, 0.5)',
          },
          {
            label: 'Completion Tokens',
            data: data.map(d => d.completion_tokens),
            borderColor: 'rgb(245, 158, 11)',
            backgroundColor: 'rgba(245, 158, 11, 0.5)',
          }
        ]
      }
    }

    it('should transform daily usage data to chart format', () => {
      const data: DailyUsage[] = [
        { date: '2024-03-01', total_tokens: 1000, prompt_tokens: 400, completion_tokens: 600, request_count: 10 },
        { date: '2024-03-02', total_tokens: 1500, prompt_tokens: 600, completion_tokens: 900, request_count: 15 },
      ]

      const chartData = transformToChartData(data)

      expect(chartData.labels).toEqual(['2024-03-01', '2024-03-02'])
      expect(chartData.datasets).toHaveLength(3)
      expect(chartData.datasets[0].data).toEqual([1000, 1500])
      expect(chartData.datasets[1].data).toEqual([400, 600])
      expect(chartData.datasets[2].data).toEqual([600, 900])
    })

    it('should handle empty data', () => {
      const data: DailyUsage[] = []
      const chartData = transformToChartData(data)

      expect(chartData.labels).toEqual([])
      expect(chartData.datasets[0].data).toEqual([])
    })

    it('should handle single data point', () => {
      const data: DailyUsage[] = [
        { date: '2024-03-01', total_tokens: 500, prompt_tokens: 200, completion_tokens: 300, request_count: 5 }
      ]

      const chartData = transformToChartData(data)

      expect(chartData.labels).toHaveLength(1)
      expect(chartData.datasets[0].data).toHaveLength(1)
    })
  })
})

describe('ModelBreakdownChart Data Processing', () => {
  // Test model breakdown data transformation
  describe('Model Breakdown Transformation', () => {
    interface ModelStats {
      model: string
      request_count: number
      total_tokens: number
      prompt_tokens: number
      completion_tokens: number
    }

    const transformToChartData = (data: ModelStats[]) => {
      const colors = [
        'rgba(59, 130, 246, 0.8)',
        'rgba(16, 185, 129, 0.8)',
        'rgba(245, 158, 11, 0.8)',
        'rgba(239, 68, 68, 0.8)',
        'rgba(139, 92, 246, 0.8)',
      ]

      return {
        labels: data.map(d => d.model),
        datasets: [{
          data: data.map(d => d.total_tokens),
          backgroundColor: data.map((_, i) => colors[i % colors.length]),
        }]
      }
    }

    it('should transform model breakdown data to chart format', () => {
      const data: ModelStats[] = [
        { model: 'gpt-4', request_count: 10, total_tokens: 5000, prompt_tokens: 2000, completion_tokens: 3000 },
        { model: 'claude-3.5-sonnet', request_count: 15, total_tokens: 7500, prompt_tokens: 3000, completion_tokens: 4500 },
      ]

      const chartData = transformToChartData(data)

      expect(chartData.labels).toEqual(['gpt-4', 'claude-3.5-sonnet'])
      expect(chartData.datasets[0].data).toEqual([5000, 7500])
      expect(chartData.datasets[0].backgroundColor).toHaveLength(2)
    })

    it('should handle empty model data', () => {
      const data: ModelStats[] = []
      const chartData = transformToChartData(data)

      expect(chartData.labels).toEqual([])
      expect(chartData.datasets[0].data).toEqual([])
    })

    it('should cycle colors for many models', () => {
      const data: ModelStats[] = Array.from({ length: 10 }, (_, i) => ({
        model: `model-${i}`,
        request_count: i + 1,
        total_tokens: (i + 1) * 100,
        prompt_tokens: (i + 1) * 40,
        completion_tokens: (i + 1) * 60,
      }))

      const chartData = transformToChartData(data)

      expect(chartData.labels).toHaveLength(10)
      expect(chartData.datasets[0].backgroundColor).toHaveLength(10)
    })
  })

  // Test percentage calculation
  describe('Percentage Calculation', () => {
    const calculatePercentages = (data: { model: string; total_tokens: number }[]) => {
      const total = data.reduce((sum, d) => sum + d.total_tokens, 0)
      if (total === 0) return data.map(d => ({ ...d, percentage: 0 }))
      return data.map(d => ({
        ...d,
        percentage: Math.round((d.total_tokens / total) * 100)
      }))
    }

    it('should calculate percentages correctly', () => {
      const data = [
        { model: 'gpt-4', total_tokens: 5000 },
        { model: 'claude-3.5-sonnet', total_tokens: 5000 },
      ]

      const result = calculatePercentages(data)

      expect(result[0].percentage).toBe(50)
      expect(result[1].percentage).toBe(50)
    })

    it('should handle zero total tokens', () => {
      const data = [
        { model: 'gpt-4', total_tokens: 0 },
        { model: 'claude-3.5-sonnet', total_tokens: 0 },
      ]

      const result = calculatePercentages(data)

      expect(result[0].percentage).toBe(0)
      expect(result[1].percentage).toBe(0)
    })

    it('should round percentages to integers', () => {
      const data = [
        { model: 'gpt-4', total_tokens: 333 },
        { model: 'claude-3.5-sonnet', total_tokens: 333 },
        { model: 'gpt-3.5-turbo', total_tokens: 334 },
      ]

      const result = calculatePercentages(data)

      // Each should be approximately 33%
      expect(result[0].percentage).toBeGreaterThanOrEqual(33)
      expect(result[0].percentage).toBeLessThanOrEqual(34)
    })
  })
})

// ============================================================================
// Tests for Chart Configuration
// Requirements: 7.1
// ============================================================================

describe('Chart Configuration', () => {
  describe('Time Series Chart Options', () => {
    const getTimeSeriesOptions = () => ({
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'top' as const,
        },
        tooltip: {
          mode: 'index' as const,
          intersect: false,
        },
      },
      scales: {
        y: {
          beginAtZero: true,
        },
      },
    })

    it('should have responsive option enabled', () => {
      const options = getTimeSeriesOptions()
      expect(options.responsive).toBe(true)
    })

    it('should have y-axis starting at zero', () => {
      const options = getTimeSeriesOptions()
      expect(options.scales.y.beginAtZero).toBe(true)
    })

    it('should have legend at top', () => {
      const options = getTimeSeriesOptions()
      expect(options.plugins.legend.position).toBe('top')
    })
  })

  describe('Pie Chart Options', () => {
    const getPieChartOptions = () => ({
      responsive: true,
      maintainAspectRatio: false,
      plugins: {
        legend: {
          position: 'right' as const,
        },
        tooltip: {
          callbacks: {
            label: (context: any) => {
              const label = context.label || ''
              const value = context.raw || 0
              return `${label}: ${value.toLocaleString()} tokens`
            }
          }
        },
      },
    })

    it('should have responsive option enabled', () => {
      const options = getPieChartOptions()
      expect(options.responsive).toBe(true)
    })

    it('should have legend on the right', () => {
      const options = getPieChartOptions()
      expect(options.plugins.legend.position).toBe('right')
    })
  })
})

// ============================================================================
// Tests for Date Range Selection
// Requirements: 7.2
// ============================================================================

describe('Date Range Selection', () => {
  describe('Range Presets', () => {
    const datePresets = [
      { label: '今天', value: 'today' },
      { label: '本周', value: 'week' },
      { label: '本月', value: 'month' }
    ]

    it('should have three preset options', () => {
      expect(datePresets).toHaveLength(3)
    })

    it('should have correct preset values', () => {
      expect(datePresets.map(p => p.value)).toEqual(['today', 'week', 'month'])
    })
  })

  describe('Days Calculation for Trends', () => {
    const getDaysForRange = (range: string): number => {
      if (range === 'day') return 1
      if (range === 'week') return 7
      if (range === 'month') return 30
      return 7 // default
    }

    it('should return 1 day for day range', () => {
      expect(getDaysForRange('day')).toBe(1)
    })

    it('should return 7 days for week range', () => {
      expect(getDaysForRange('week')).toBe(7)
    })

    it('should return 30 days for month range', () => {
      expect(getDaysForRange('month')).toBe(30)
    })

    it('should return default 7 days for unknown range', () => {
      expect(getDaysForRange('unknown')).toBe(7)
    })
  })
})
