import apiClient from './client'

export interface ReferralCode {
  referral_code: string
  referral_link: string
}

export interface ReferralStats {
  total_referrals: number
  total_bonus: number
}

export interface ReferredUser {
  user_id: number
  username: string
  email: string
  registered_at: string
  bonus_amount: number
}

export interface ReferralListResponse {
  referrals: ReferredUser[]
  total: number
  limit: number
  offset: number
}

export const getReferralCode = () => 
  apiClient.get<ReferralCode>('/api/referral/code')

export const getReferralStats = () => 
  apiClient.get<ReferralStats>('/api/referral/stats')

export const getReferralList = (limit = 20, offset = 0) =>
  apiClient.get<ReferralListResponse>('/api/referral/list', {
    params: { limit, offset }
  })
