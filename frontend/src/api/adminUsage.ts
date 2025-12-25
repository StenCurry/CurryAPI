import client from './client'

/**
 * Admin usage statistics types
 */
export interface UserUsageSummary {
  user_id: number
  username: string
  requests: number
  total_tokens: number
}

export interface AdminModelStats {
  model: string
  request_count: number
  total_tokens: number
  prompt_tokens: number
  completion_tokens: number
}

export interface AdminUsageStats {
  total_users: number
  total_requests: number
  total_tokens: number
  top_users: UserUsageSummary[]
  top_models: AdminModelStats[]
}

export interface TrendData {
  date: string
  requests: number
  total_tokens: number
}

export interface AdminUsageTrends {
  view: 'daily' | 'weekly' | 'monthly'
  days: number
  trends: TrendData[]
  growth_rate: number
}

export interface CursorSessionUsage {
  cursor_session: string
  requests: number
  total_tokens: number
}

export interface CursorSessionUsageResponse {
  sessions: CursorSessionUsage[]
  total: number
}

/**
 * Get admin usage statistics
 * @param params Query parameters for filtering
 * @returns Admin usage statistics
 */
export async function getAdminUsageStats(params?: {
  start_date?: string
  end_date?: string
  model?: string
}): Promise<AdminUsageStats> {
  const response = await client.get('/admin/usage/stats', { params })
  return response.data
}

/**
 * Get usage trends for administrators
 * @param params Query parameters for filtering
 * @returns Usage trends data for charts
 */
export async function getUsageTrends(params?: {
  days?: number
  view?: 'daily' | 'weekly' | 'monthly'
  user_id?: number
}): Promise<AdminUsageTrends> {
  const response = await client.get('/admin/usage/trends', { params })
  return response.data
}

/**
 * Get Cursor session usage statistics
 * @param params Query parameters for filtering
 * @returns Cursor session usage data
 */
export async function getCursorSessionUsage(params?: {
  start_date?: string
  end_date?: string
}): Promise<CursorSessionUsageResponse> {
  const response = await client.get('/admin/usage/sessions', { params })
  return response.data
}

/**
 * Export usage data as CSV
 * @param params Query parameters for filtering
 * @returns Blob containing CSV data
 */
export async function exportUsageData(params?: {
  start_date?: string
  end_date?: string
  user_id?: number
  model?: string
}): Promise<Blob> {
  const response = await client.get('/admin/usage/export', {
    params,
    responseType: 'blob'
  })
  return response.data
}

/**
 * Helper function to trigger CSV download
 * @param blob CSV blob data
 * @param filename Filename for download
 */
export function downloadCSV(blob: Blob, filename?: string): void {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename || `usage_export_${new Date().toISOString().split('T')[0]}.csv`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}
