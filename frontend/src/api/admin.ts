import apiClient from './client'
import type { AdminKey, CursorSession, ManagedUser } from '@/types'

export interface CreateKeyPayload {
  key: string
  token_name?: string
  quota_limit?: number | null    // Quota limit in USD, null means unlimited
  expires_at?: string | null     // ISO date string, null means never expires
  allowed_models?: string[]      // Allowed models, empty means all models
}

export interface UpdateKeyNamePayload {
  name: string
}

export interface CreateCursorSessionPayload {
  email: string
  session_token: string
  expires_at?: string
  extra_cookies?: Record<string, string>
}

export const listKeys = () => apiClient.get<{ total: number; keys: AdminKey[] }>('/admin/keys')

export const addKey = (payload: CreateKeyPayload) => apiClient.post('/admin/keys', payload)

export const removeKey = (key: string) => apiClient.delete(`/admin/keys/${key}`)

export const toggleKeyStatus = (key: string) => apiClient.put(`/admin/keys/${key}/toggle`)

export const updateKeyName = (key: string, name: string) => 
  apiClient.put(`/admin/keys/${key}/name`, { name })

export const listCursorSessions = () =>
  apiClient.get<{ sessions: CursorSession[]; stats?: any }>('/admin/cursor/sessions')

export const addCursorSession = (payload: CreateCursorSessionPayload) =>
  apiClient.post<CursorSession>('/admin/cursor/sessions', payload)

export const removeCursorSession = (email: string) =>
  apiClient.delete(`/admin/cursor/sessions/${email}`)

export const validateCursorSession = (email: string) =>
  apiClient.post<{ email: string; is_valid: boolean; message: string }>(
    '/admin/cursor/sessions/validate',
    { email }
  )

export const getCursorSessionStats = () =>
  apiClient.get<{
    total_sessions: number
    valid_sessions: number
    total_usage: number
    current_index: number
    fallback_active: boolean
  }>('/admin/cursor/sessions/stats')

export const reloadCursorSessions = () =>
  apiClient.post('/admin/cursor/sessions/reload')

export const listManagedUsers = () => apiClient.get<ManagedUser[]>('/admin/users')
