import { computed } from 'vue'
import type { ChartOptions } from 'chart.js'

/**
 * Composable for shared chart configuration and styling
 * Provides consistent theming across all chart components
 */
export function useChartConfig() {
  // Default color palette for charts
  const colorPalette = [
    'rgba(59, 130, 246, 0.8)',   // Blue
    'rgba(139, 92, 246, 0.8)',   // Purple
    'rgba(16, 185, 129, 0.8)',   // Green
    'rgba(245, 158, 11, 0.8)',   // Amber
    'rgba(239, 68, 68, 0.8)',    // Red
    'rgba(236, 72, 153, 0.8)',   // Pink
    'rgba(6, 182, 212, 0.8)',    // Cyan
    'rgba(132, 204, 22, 0.8)',   // Lime
  ]

  const borderColorPalette = [
    'rgba(59, 130, 246, 1)',
    'rgba(139, 92, 246, 1)',
    'rgba(16, 185, 129, 1)',
    'rgba(245, 158, 11, 1)',
    'rgba(239, 68, 68, 1)',
    'rgba(236, 72, 153, 1)',
    'rgba(6, 182, 212, 1)',
    'rgba(132, 204, 22, 1)',
  ]

  // Common plugin配置
  const basePlugins = {
    legend: {
      labels: {
        color: 'rgba(255, 255, 255, 0.8)',
        font: {
          family: 'system-ui, -apple-system, sans-serif',
          size: 12,
        },
        padding: 16,
      },
    },
    tooltip: {
      backgroundColor: 'rgba(15, 23, 42, 0.95)',
      titleColor: 'rgba(255, 255, 255, 0.9)',
      bodyColor: 'rgba(255, 255, 255, 0.8)',
      borderColor: 'rgba(59, 130, 246, 0.5)',
      borderWidth: 1,
      padding: 12,
      cornerRadius: 8,
      titleFont: {
        weight: 700,
      },
    },
  } satisfies NonNullable<ChartOptions<'bar'>['plugins']>

  // Line chart specific options
  const lineChartOptions = computed<ChartOptions<'line'>>(() => ({
    responsive: true,
    maintainAspectRatio: false,
    plugins: basePlugins as ChartOptions<'line'>['plugins'],
    scales: {
      x: {
        grid: {
          color: 'rgba(255, 255, 255, 0.1)',
        },
        ticks: {
          color: 'rgba(255, 255, 255, 0.7)',
          font: {
            size: 11,
          },
        },
      },
      y: {
        grid: {
          color: 'rgba(255, 255, 255, 0.1)',
        },
        ticks: {
          color: 'rgba(255, 255, 255, 0.7)',
          font: {
            size: 11,
          },
        },
        beginAtZero: true,
      },
    },
    interaction: {
      intersect: false,
      mode: 'index',
    },
  }))

  // Pie/Doughnut chart specific options
  const pieChartOptions = computed<ChartOptions<'pie'>>(() => ({
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      ...(basePlugins as ChartOptions<'pie'>['plugins']),
      legend: {
        ...basePlugins.legend,
        position: 'right',
      },
    },
  }))

  // Bar chart specific options
  const barChartOptions = computed<ChartOptions<'bar'>>(() => ({
    responsive: true,
    maintainAspectRatio: false,
    plugins: basePlugins as ChartOptions<'bar'>['plugins'],
    scales: {
      x: {
        grid: {
          color: 'rgba(255, 255, 255, 0.1)',
        },
        ticks: {
          color: 'rgba(255, 255, 255, 0.7)',
          font: {
            size: 11,
          },
        },
      },
      y: {
        grid: {
          color: 'rgba(255, 255, 255, 0.1)',
        },
        ticks: {
          color: 'rgba(255, 255, 255, 0.7)',
          font: {
            size: 11,
          },
        },
        beginAtZero: true,
      },
    },
  }))

  // Helper to get color by index
  function getColor(index: number, withAlpha = true): string {
    const palette = withAlpha ? colorPalette : borderColorPalette
    return (palette[(index % palette.length + palette.length) % palette.length] ?? palette[0]) as string
  }

  // Helper to format large numbers
  function formatNumber(num: number): string {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M'
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K'
    }
    return num.toString()
  }

  return {
    colorPalette,
    borderColorPalette,
    lineChartOptions,
    pieChartOptions,
    barChartOptions,
    getColor,
    formatNumber,
  }
}
