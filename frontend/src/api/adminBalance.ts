import apiClient from './client'

// Types for admin balance management
export interface UserBalanceInfo {
  user_id: number
  username: string
  email: string
  balance: number
  status: 'active' | 'exhausted'
  referral_code: string
  total_consumed: number
  total_recharged: number
  created_at: string
  updated_at: string
}

export interface UserBalancesResponse {
  users: UserBalanceInfo[]
  total: number
  limit: number
  offset: number
}

export interface AdjustBalanceRequest {
  user_id: number
  amount: number
  reason: string
}

export interface AdjustBalanceResponse {
  message: string
  user_id: number
  amount: number
  balance_after: number
  transaction_id: number
}

// Get all user balances with pagination
// Requirements: 8.2
export const getAllUserBalances = (params?: { limit?: number; offset?: number }) =>
  apiClient.get<UserBalancesResponse>('/admin/balance/users', { params })

// Adjust user balance (add or deduct)
// Requirements: 8.1, 8.2
export const adjustUserBalance = (data: AdjustBalanceRequest) =>
  apiClient.post<AdjustBalanceResponse>('/admin/balance/adjust', data)
