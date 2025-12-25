import apiClient from './client'

export interface UserBalance {
  balance: number
  status: string
  referral_code: string
  total_consumed: number
  total_recharged: number
  created_at: string
  updated_at: string
}

export interface BalanceTransaction {
  id: number
  type: string
  amount: number
  balance_after: number
  tokens?: number
  description?: string
  related_user_id?: number
  admin_id?: number
  api_token?: string
  model?: string
  created_at: string
}

export interface TransactionsResponse {
  transactions: BalanceTransaction[]
  total: number
  limit: number
  offset: number
}

export const getBalance = () => apiClient.get<UserBalance>('/api/balance')

export const getTransactions = (limit = 20, offset = 0) =>
  apiClient.get<TransactionsResponse>('/api/balance/transactions', {
    params: { limit, offset }
  })
