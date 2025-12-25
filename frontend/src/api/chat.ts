/**
 * Chat API Client
 * 聊天 API 客户端
 * 
 * Implements conversation CRUD, message operations, SSE streaming, and models API
 * 实现会话 CRUD、消息操作、SSE 流式传输和模型 API
 * 
 * Requirements: 1.1, 2.1, 2.2, 3.1
 */

import apiClient from './client'

// ============================================================================
// Type Definitions
// ============================================================================

/** Conversation model */
export interface Conversation {
  id: number
  user_id?: number
  title: string
  model: string
  system_prompt?: string
  created_at: string
  updated_at: string
}

/** Message model */
export interface Message {
  id: number
  conversation_id: number
  role: 'user' | 'assistant' | 'system'
  content: string
  tokens: number
  cost: number
  created_at: string
}

/** Token usage info */
export interface TokenUsage {
  prompt: number
  completion: number
}

/** SSE stream event types */
export interface StreamEvent {
  type: 'start' | 'content' | 'done' | 'error'
  message_id?: number
  delta?: string
  tokens?: TokenUsage
  cost?: number
  error?: string
}

/** Model info for selection */
export interface ChatModel {
  id: string
  name: string
  provider: string
  description?: string
  input_price?: number
  output_price?: number
  context_window?: number
  is_available?: boolean
}

// ============================================================================
// Request/Response Types
// ============================================================================

export interface CreateConversationRequest {
  title?: string
  model: string
  system_prompt?: string
}

export interface UpdateConversationRequest {
  title?: string
  model?: string
  system_prompt?: string
}

export interface SendMessageRequest {
  content: string
}

export interface ConversationListResponse {
  conversations: Conversation[]
  total: number
  page: number
  limit: number
}

export interface MessageListResponse {
  messages: Message[]
  total: number
  page: number
  limit: number
}

export interface ModelsResponse {
  models: ChatModel[]
}

// ============================================================================
// Conversation CRUD API
// Requirements: 1.1
// ============================================================================

/**
 * Create a new conversation
 * 创建新会话
 */
export async function createConversation(data: CreateConversationRequest): Promise<Conversation> {
  const response = await apiClient.post<{ success: boolean; data: Conversation }>(
    '/api/chat/conversations',
    data
  )
  return response.data.data
}

/**
 * Get conversation list with pagination
 * 获取会话列表（分页）
 */
export async function getConversations(
  page: number = 1,
  limit: number = 20
): Promise<ConversationListResponse> {
  const response = await apiClient.get<{ success: boolean; data: ConversationListResponse }>(
    '/api/chat/conversations',
    { params: { page, limit } }
  )
  return response.data.data
}

/**
 * Get a single conversation by ID
 * 获取单个会话
 */
export async function getConversation(id: number): Promise<Conversation> {
  const response = await apiClient.get<{ success: boolean; data: Conversation }>(
    `/api/chat/conversations/${id}`
  )
  return response.data.data
}

/**
 * Update a conversation
 * 更新会话
 */
export async function updateConversation(
  id: number,
  data: UpdateConversationRequest
): Promise<Conversation | null> {
  const response = await apiClient.put<{ success: boolean; data?: Conversation }>(
    `/api/chat/conversations/${id}`,
    data
  )
  // Return the updated conversation if available, otherwise null
  return response.data.data || null
}

/**
 * Delete a conversation
 * 删除会话
 */
export async function deleteConversation(id: number): Promise<void> {
  await apiClient.delete(`/api/chat/conversations/${id}`)
}

// ============================================================================
// Message API
// Requirements: 2.1
// ============================================================================

/**
 * Get messages for a conversation with pagination
 * 获取会话消息（分页）
 */
export async function getMessages(
  conversationId: number,
  page: number = 1,
  limit: number = 50
): Promise<MessageListResponse> {
  const response = await apiClient.get<{ success: boolean; data: MessageListResponse }>(
    `/api/chat/conversations/${conversationId}/messages`,
    { params: { page, limit } }
  )
  return response.data.data
}

// ============================================================================
// SSE Streaming Client
// Requirements: 2.2
// ============================================================================

export interface StreamCallbacks {
  onStart?: (messageId: number) => void
  onContent?: (delta: string) => void
  onDone?: (tokens: TokenUsage, cost: number) => void
  onError?: (error: string) => void
}

/**
 * Send a message and receive streaming response via SSE
 * 发送消息并通过 SSE 接收流式响应
 * 
 * @param conversationId - Conversation ID
 * @param content - Message content
 * @param model - Optional model to use (overrides conversation model)
 * @param callbacks - Event callbacks for streaming
 * @returns AbortController to cancel the stream
 */
export function sendMessageStream(
  conversationId: number,
  content: string,
  model: string | undefined,
  callbacks: StreamCallbacks
): AbortController {
  const controller = new AbortController()
  
  // Build the URL with credentials
  const baseUrl = import.meta.env.DEV
    ? ''
    : import.meta.env.VITE_API_BASE_URL || 'http://localhost:8002'
  const url = `${baseUrl}/api/chat/conversations/${conversationId}/messages`
  
  // Build request body with optional model
  const requestBody: { content: string; model?: string } = { content }
  if (model) {
    requestBody.model = model
  }
  
  console.log('[Chat API] Sending message to:', url, 'content:', content, 'model:', model)
  
  // Start the fetch request
  fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'text/event-stream',
      'Cache-Control': 'no-cache'
    },
    credentials: 'include',
    body: JSON.stringify(requestBody),
    signal: controller.signal
  })
    .then(async (response) => {
      console.log('[Chat API] Response status:', response.status, response.ok)
      if (!response.ok) {
        // Handle HTTP errors
        // Requirements: 2.5, 6.2 - Display error message and balance warning
        let errorMessage = 'Failed to send message'
        let errorType = 'UNKNOWN_ERROR'
        
        try {
          const errorData = await response.json()
          errorMessage = errorData.error?.message || errorMessage
          errorType = errorData.error?.type || errorType
        } catch {
          // Ignore JSON parse errors
        }
        
        // Map HTTP status codes and error types to user-friendly messages
        // Requirements: 2.6, 10.1-10.5 - Provider-specific error handling
        
        // Check for provider-specific error types from backend
        if (errorType === 'PROVIDER_NOT_AVAILABLE') {
          errorMessage = errorMessage || '该模型的 AI 服务提供商未配置，请选择其他模型'
        } else if (errorType === 'INVALID_API_KEY') {
          errorMessage = errorMessage || 'API 密钥无效或已过期，请联系管理员'
        } else if (errorType === 'RATE_LIMITED') {
          errorMessage = errorMessage || '请求过于频繁，请稍后重试'
        } else if (errorType === 'PROVIDER_ERROR') {
          errorMessage = errorMessage || 'AI 服务暂时不可用，请稍后重试'
        } else if (errorType === 'TIMEOUT') {
          errorMessage = errorMessage || '请求超时，请稍后重试'
        } else if (errorType === 'CONTEXT_TOO_LONG') {
          errorMessage = errorMessage || '消息内容过长，请缩短后重试'
        } else {
          // Fallback to HTTP status code mapping
          switch (response.status) {
            case 400:
              errorMessage = errorMessage || '请求参数错误'
              break
            case 401:
              errorMessage = '未登录或会话已过期，请重新登录'
              errorType = 'UNAUTHORIZED'
              break
            case 402:
              // Requirements: 6.2 - Insufficient balance warning
              errorMessage = '余额不足，请充值后再试'
              errorType = 'INSUFFICIENT_BALANCE'
              break
            case 403:
              errorMessage = '无权限访问此对话'
              errorType = 'FORBIDDEN'
              break
            case 404:
              errorMessage = '对话不存在或已被删除'
              errorType = 'NOT_FOUND'
              break
            case 429:
              errorMessage = '请求过于频繁，请稍后重试'
              errorType = 'RATE_LIMITED'
              break
            case 502:
              errorMessage = 'AI 服务暂时不可用，请稍后重试'
              errorType = 'SERVICE_UNAVAILABLE'
              break
            case 504:
              errorMessage = '请求超时，请稍后重试'
              errorType = 'SERVICE_TIMEOUT'
              break
            case 500:
            default:
              errorMessage = errorMessage || '服务器错误，请稍后重试'
              errorType = 'SERVER_ERROR'
              break
          }
        }
        
        callbacks.onError?.(errorMessage)
        return
      }
      
      // Read the SSE stream
      const reader = response.body?.getReader()
      if (!reader) {
        callbacks.onError?.('Failed to read response stream')
        return
      }
      
      const decoder = new TextDecoder()
      let buffer = ''
      
      while (true) {
        const { done, value } = await reader.read()
        
        if (done) break
        
        buffer += decoder.decode(value, { stream: true })
        
        // Process complete SSE events
        const lines = buffer.split('\n')
        buffer = lines.pop() || '' // Keep incomplete line in buffer
        
        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6).trim()
            if (!data || data === '[DONE]') continue
            
            try {
              const event: StreamEvent = JSON.parse(data)
              
              console.log('[Chat API] SSE event:', event.type, event)
              
              switch (event.type) {
                case 'start':
                  if (event.message_id) {
                    callbacks.onStart?.(event.message_id)
                  }
                  break
                case 'content':
                  if (event.delta) {
                    callbacks.onContent?.(event.delta)
                  }
                  break
                case 'done':
                  if (event.tokens) {
                    callbacks.onDone?.(event.tokens, event.cost || 0)
                  }
                  break
                case 'error':
                  callbacks.onError?.(event.error || 'Unknown error')
                  break
              }
            } catch (e) {
              console.error('Failed to parse SSE event:', e, data)
            }
          }
        }
      }
    })
    .catch((error) => {
      if (error.name === 'AbortError') {
        // Stream was cancelled, not an error
        return
      }
      callbacks.onError?.(error.message || 'Network error')
    })
  
  return controller
}

// ============================================================================
// Models API
// Requirements: 3.1
// ============================================================================

/**
 * Get available chat models
 * 获取可用的聊天模型列表
 */
export async function getChatModels(): Promise<ChatModel[]> {
  const response = await apiClient.get<{ success: boolean; data: ModelsResponse }>(
    '/api/chat/models'
  )
  return response.data.data.models
}

// ============================================================================
// Export all as chatApi object for convenience
// ============================================================================

export const chatApi = {
  // Conversations
  createConversation,
  getConversations,
  getConversation,
  updateConversation,
  deleteConversation,
  
  // Messages
  getMessages,
  sendMessageStream,
  
  // Models
  getChatModels
}

export default chatApi
