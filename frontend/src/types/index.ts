export interface User {
  id: number
  username: string
  email: string
  role: 'admin' | 'user'
  created_at: string
  last_login?: string
}

export interface LoginRequest {
  username_or_email: string
  password: string
}

export interface RegisterRequest {
  username: string
  email: string
  password: string
  code: string
  turnstile_token?: string
  referral_code?: string  // Optional referral code for bonus
}

export interface SendCodeRequest {
  email: string
  turnstile_token?: string
}

export interface ApiResponse<T = any> {
  data?: T
  user?: User
  session_id?: string
  error?: {
    code: string
    message: string
  }
}

export interface ApiKey {
  key: string
  name?: string
  created_at?: string
  last_used?: string
}

export interface AdminKey {
  key: string
  masked_key: string
  token_name?: string
  user_id?: number
  username?: string
  created_at: string
  usage_count: number
  last_used_at?: string
  is_active: boolean
  // Balance system extension fields
  quota_limit?: number | null    // Quota limit in USD, null means unlimited
  quota_used: number             // Quota used in USD
  expires_at?: string | null     // Expiration time, null means never expires
  allowed_models?: string[]      // Allowed models, empty means all models
}

export interface ManagedUser {
  id: number
  username: string
  email: string
  role: string
  created_at: string
}

export interface CursorSession {
  email: string
  cookies: string
  status?: string
  created_at?: string
  last_used?: string
}

export interface Announcement {
  id: number
  title: string
  content: string
  created_by?: number
  created_at: string
  updated_at?: string
  is_read?: boolean
  read_count?: number
}

export interface ModelStats {
  model: string
  request_count: number
  total_tokens: number
  prompt_tokens: number
  completion_tokens: number
}

export interface UsageStats {
  total_requests: number
  total_tokens: number
  prompt_tokens: number
  completion_tokens: number
  by_model: ModelStats[]
  recent_calls: RecentCall[]
  message?: string
}

export interface RecentCall {
  id: number
  model: string
  prompt_tokens: number
  completion_tokens: number
  total_tokens: number
  status: number
  timestamp: string
  duration_ms: number
  error?: string
  token_name?: string
}

export interface RecentCallsResponse {
  calls: RecentCall[]
  total: number
  limit: number
  offset: number
  message?: string
}
