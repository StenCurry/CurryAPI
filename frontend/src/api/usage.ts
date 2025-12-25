import client, { handleApiResponse } from './client'
import type { ApiResponse, UsageStats, RecentCallsResponse } from '@/types'

/**
 * Get usage statistics for the authenticated user
 * @param params Query parameters for filtering
 * @returns Usage statistics
 */
export async function getUsageStats(params?: {
  start_date?: string
  end_date?: string
  model?: string
}): Promise<UsageStats> {
  const response = await client.get('/api/usage/stats', { params })
  return response.data
}

/**
 * Get recent API calls for the authenticated user
 * @param params Query parameters for pagination and filtering
 * @returns Recent calls with pagination info
 */
export async function getRecentCalls(params?: {
  limit?: number
  offset?: number
}): Promise<RecentCallsResponse> {
  const response = await client.get('/api/usage/recent', { params })
  return response.data
}

/**
 * Daily usage data for trends chart
 */
export interface DailyUsage {
  date: string
  total_tokens: number
  prompt_tokens: number
  completion_tokens: number
  request_count: number
}

/**
 * Response from usage trends endpoint
 */
export interface UsageTrendsResponse {
  days: number
  trends: DailyUsage[]
}

/**
 * Get usage trends for the authenticated user
 * @param params Query parameters for filtering
 * @returns Usage trends data for charts
 */
export async function getUsageTrends(params?: {
  days?: number
}): Promise<UsageTrendsResponse> {
  const response = await client.get('/api/usage/trends', { params })
  return response.data
}
