import apiClient from './client'

export interface ModelMarketplaceInfo {
  id: string
  name: string
  provider: string       // OpenAI, Anthropic, Google, etc.
  tags: string[]         // Fast, Powerful, Code, Vision
  billing_type: string   // per_token, per_request
  endpoint_type: string  // chat, completion, embedding
  max_tokens: number
  context_window: number
  description: string
}

export interface ModelFilters {
  providers: string[]
  tags: string[]
  endpoint_types: string[]
}

export interface ModelMarketplaceResponse {
  models: ModelMarketplaceInfo[]
  total: number
  filters: ModelFilters
}

export interface ModelMarketplaceParams {
  provider?: string
  tag?: string
  endpoint_type?: string
}

export const getModelMarketplace = (params?: ModelMarketplaceParams) =>
  apiClient.get<ModelMarketplaceResponse>('/api/models/marketplace', { params })

export const getModelDetail = (modelId: string) =>
  apiClient.get<ModelMarketplaceInfo>(`/api/models/marketplace/${modelId}`)
