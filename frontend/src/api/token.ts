import apiClient from './client'
import type { TokenItem } from '@/types'

export interface CreateTokenPayload {
  name: string
}

export const listTokens = () => apiClient.get<TokenItem[]>('/tokens')

export const createToken = (payload: CreateTokenPayload) =>
  apiClient.post<TokenItem>('/tokens', payload)

export const deleteToken = (tokenId: string) => apiClient.delete(`/tokens/${tokenId}`)
