/**
 * Chat State Management Store
 * 聊天状态管理 Store
 * 
 * Implements conversation and message state management with streaming support
 * 实现会话和消息状态管理，支持流式响应
 * 
 * Requirements: 1.1, 1.2, 2.1
 */

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import {
  chatApi,
  type Conversation,
  type Message,
  type ChatModel,
  type TokenUsage,
  type CreateConversationRequest,
  type UpdateConversationRequest
} from '@/api/chat'

// ============================================================================
// State Interface
// ============================================================================

export interface ChatState {
  // Conversation state
  conversations: Conversation[]
  currentConversation: Conversation | null
  conversationsTotal: number
  conversationsPage: number
  conversationsLoading: boolean
  
  // Message state
  messages: Message[]
  messagesTotal: number
  messagesPage: number
  messagesLoading: boolean
  
  // Streaming state
  isStreaming: boolean
  streamingContent: string
  streamingMessageId: number | null
  streamController: AbortController | null
  
  // Model state
  models: ChatModel[]
  modelsLoading: boolean
  selectedModel: string
  
  // Error state
  error: string | null
}

// ============================================================================
// Store Definition
// ============================================================================

export const useChatStore = defineStore('chat', () => {
  // ---------------------------------------------------------------------------
  // Conversation State
  // ---------------------------------------------------------------------------
  const conversations = ref<Conversation[]>([])
  const currentConversation = ref<Conversation | null>(null)
  const conversationsTotal = ref(0)
  const conversationsPage = ref(1)
  const conversationsLoading = ref(false)
  
  // ---------------------------------------------------------------------------
  // Message State
  // ---------------------------------------------------------------------------
  const messages = ref<Message[]>([])
  const messagesTotal = ref(0)
  const messagesPage = ref(1)
  const messagesLoading = ref(false)
  
  // ---------------------------------------------------------------------------
  // Streaming State
  // ---------------------------------------------------------------------------
  const isStreaming = ref(false)
  const streamingContent = ref('')
  const streamingMessageId = ref<number | null>(null)
  const streamController = ref<AbortController | null>(null)
  
  // ---------------------------------------------------------------------------
  // Model State
  // ---------------------------------------------------------------------------
  const models = ref<ChatModel[]>([])
  const modelsLoading = ref(false)
  const selectedModel = ref('gpt-4o')
  
  // ---------------------------------------------------------------------------
  // Error State
  // Requirements: 2.5 - Error handling and retry mechanism
  // ---------------------------------------------------------------------------
  const error = ref<string | null>(null)
  const errorType = ref<string | null>(null)
  const lastFailedMessage = ref<string | null>(null)
  
  // ---------------------------------------------------------------------------
  // Computed Properties
  // ---------------------------------------------------------------------------
  
  /** Check if there are more conversations to load */
  const hasMoreConversations = computed(() => {
    return conversations.value.length < conversationsTotal.value
  })
  
  /** Check if there are more messages to load */
  const hasMoreMessages = computed(() => {
    return messages.value.length < messagesTotal.value
  })
  
  /** Get current model info */
  const currentModel = computed(() => {
    return models.value.find(m => m.id === selectedModel.value) || null
  })
  
  // ---------------------------------------------------------------------------
  // Conversation Actions
  // Requirements: 1.1, 1.2
  // ---------------------------------------------------------------------------
  
  /**
   * Load conversations with pagination
   * 加载会话列表（分页）
   */
  async function loadConversations(page: number = 1, limit: number = 20): Promise<void> {
    conversationsLoading.value = true
    error.value = null
    
    try {
      const response = await chatApi.getConversations(page, limit)
      
      // Ensure conversations is always an array (defensive check for null/undefined)
      const loadedConversations = response.conversations || []
      
      if (page === 1) {
        conversations.value = loadedConversations
      } else {
        // Append for pagination
        conversations.value = [...conversations.value, ...loadedConversations]
      }
      
      conversationsTotal.value = response.total || 0
      conversationsPage.value = page
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load conversations'
      error.value = errorMessage
      console.error('Failed to load conversations:', err)
      // Ensure conversations is an empty array on error
      if (page === 1) {
        conversations.value = []
      }
    } finally {
      conversationsLoading.value = false
    }
  }
  
  /**
   * Load more conversations (next page)
   * 加载更多会话
   */
  async function loadMoreConversations(): Promise<void> {
    if (!hasMoreConversations.value || conversationsLoading.value) return
    await loadConversations(conversationsPage.value + 1)
  }
  
  /**
   * Create a new conversation
   * 创建新会话
   */
  async function createConversation(data?: Partial<CreateConversationRequest>): Promise<Conversation | null> {
    conversationsLoading.value = true
    error.value = null
    
    try {
      const conversation = await chatApi.createConversation({
        title: data?.title || '新对话',
        model: data?.model || selectedModel.value,
        system_prompt: data?.system_prompt
      })
      
      // Add to the beginning of the list
      conversations.value.unshift(conversation)
      conversationsTotal.value++
      
      // Set as current conversation
      currentConversation.value = conversation
      messages.value = []
      messagesTotal.value = 0
      
      return conversation
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to create conversation'
      error.value = errorMessage
      console.error('Failed to create conversation:', err)
      return null
    } finally {
      conversationsLoading.value = false
    }
  }
  
  /**
   * Select and load a conversation
   * 选择并加载会话
   */
  async function selectConversation(id: number): Promise<void> {
    // Check if already selected
    if (currentConversation.value?.id === id) return
    
    // Cancel any ongoing stream
    cancelStream()
    
    conversationsLoading.value = true
    error.value = null
    
    try {
      const conversation = await chatApi.getConversation(id)
      currentConversation.value = conversation
      selectedModel.value = conversation.model
      
      // Load messages for this conversation
      await loadMessages(id)
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load conversation'
      error.value = errorMessage
      console.error('Failed to load conversation:', err)
    } finally {
      conversationsLoading.value = false
    }
  }
  
  /**
   * Update a conversation (rename, change model, etc.)
   * 更新会话
   */
  async function updateConversation(id: number, data: UpdateConversationRequest): Promise<boolean> {
    error.value = null
    
    try {
      const updated = await chatApi.updateConversation(id, data)
      
      // Only update if we got a valid response
      if (updated) {
        // Update in list
        const index = conversations.value.findIndex(c => c.id === id)
        if (index !== -1) {
          conversations.value[index] = updated
        }
        
        // Update current if it's the same
        if (currentConversation.value?.id === id) {
          currentConversation.value = updated
        }
      }
      
      // Update selected model if provided (even if updated is null)
      if (data.model && currentConversation.value?.id === id) {
        selectedModel.value = data.model
      }
      
      return true
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to update conversation'
      error.value = errorMessage
      console.error('Failed to update conversation:', err)
      return false
    }
  }
  
  /**
   * Delete a conversation
   * 删除会话
   */
  async function deleteConversation(id: number): Promise<boolean> {
    error.value = null
    
    try {
      await chatApi.deleteConversation(id)
      
      // Remove from list
      conversations.value = conversations.value.filter(c => c.id !== id)
      conversationsTotal.value--
      
      // Clear current if it was deleted
      if (currentConversation.value?.id === id) {
        currentConversation.value = null
        messages.value = []
        messagesTotal.value = 0
      }
      
      return true
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to delete conversation'
      error.value = errorMessage
      console.error('Failed to delete conversation:', err)
      return false
    }
  }
  
  // ---------------------------------------------------------------------------
  // Message Actions
  // Requirements: 2.1
  // ---------------------------------------------------------------------------
  
  /**
   * Load messages for a conversation
   * 加载会话消息
   */
  async function loadMessages(conversationId: number, page: number = 1, limit: number = 50): Promise<void> {
    messagesLoading.value = true
    error.value = null
    
    try {
      const response = await chatApi.getMessages(conversationId, page, limit)
      
      // Ensure messages is always an array (defensive check for null/undefined)
      const loadedMessages = response.messages || []
      
      if (page === 1) {
        messages.value = loadedMessages
      } else {
        // Prepend for older messages (pagination loads older first)
        messages.value = [...loadedMessages, ...messages.value]
      }
      
      messagesTotal.value = response.total || 0
      messagesPage.value = page
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load messages'
      error.value = errorMessage
      console.error('Failed to load messages:', err)
      // Ensure messages is an empty array on error
      if (page === 1) {
        messages.value = []
      }
    } finally {
      messagesLoading.value = false
    }
  }
  
  /**
   * Load more messages (older messages)
   * 加载更多消息（历史消息）
   */
  async function loadMoreMessages(): Promise<void> {
    if (!hasMoreMessages.value || messagesLoading.value || !currentConversation.value) return
    await loadMessages(currentConversation.value.id, messagesPage.value + 1)
  }
  
  /**
   * Send a message and handle streaming response
   * 发送消息并处理流式响应
   * Requirements: 2.5 - Error handling with retry support
   */
  async function sendMessage(content: string): Promise<boolean> {
    if (!currentConversation.value || isStreaming.value || !content.trim()) {
      return false
    }
    
    error.value = null
    errorType.value = null
    lastFailedMessage.value = content.trim() // Store for potential retry
    isStreaming.value = true
    streamingContent.value = ''
    streamingMessageId.value = null
    
    // Add user message to the list immediately
    // Use negative ID to avoid collision with server IDs
    const tempId = -(Date.now())
    const userMessage: Message = {
      id: tempId,
      conversation_id: currentConversation.value.id,
      role: 'user',
      content: content.trim(),
      tokens: 0,
      cost: 0,
      created_at: new Date().toISOString()
    }
    messages.value.push(userMessage)
    messagesTotal.value++
    
    console.log('[Chat] User message added to list:', userMessage)
    
    try {
      // Start streaming with selected model
      streamController.value = chatApi.sendMessageStream(
        currentConversation.value.id,
        content.trim(),
        selectedModel.value, // Pass the selected model
        {
          onStart: (messageId) => {
            console.log('[Chat] Stream started, message ID:', messageId)
            streamingMessageId.value = messageId
            // Clear last failed message on successful start
            lastFailedMessage.value = null
          },
          onContent: (delta) => {
            streamingContent.value += delta
          },
          onDone: (tokens: TokenUsage, cost: number) => {
            console.log('[Chat] Stream done, tokens:', tokens, 'cost:', cost)
            // Add assistant message to the list
            const assistantMessage: Message = {
              id: streamingMessageId.value || Date.now(),
              conversation_id: currentConversation.value!.id,
              role: 'assistant',
              content: streamingContent.value,
              tokens: tokens.prompt + tokens.completion,
              cost: cost,
              created_at: new Date().toISOString()
            }
            messages.value.push(assistantMessage)
            messagesTotal.value++
            console.log('[Chat] Assistant message added to list:', assistantMessage)
            
            // Update conversation's updated_at and move to top
            if (currentConversation.value) {
              currentConversation.value.updated_at = new Date().toISOString()
              
              // Move to top of list
              const index = conversations.value.findIndex(c => c.id === currentConversation.value!.id)
              if (index > 0) {
                const [conv] = conversations.value.splice(index, 1)
                if (conv) {
                  conversations.value.unshift(conv)
                }
              }
            }
            
            // Reset streaming state and clear last failed message
            isStreaming.value = false
            streamingContent.value = ''
            streamingMessageId.value = null
            streamController.value = null
            lastFailedMessage.value = null
          },
          onError: (errorMsg) => {
            console.error('[Chat] Stream error:', errorMsg)
            // Requirements: 2.5, 2.6, 10.1-10.5 - Store error for display and retry
            error.value = errorMsg
            // Detect error type from message for provider-specific errors
            if (errorMsg.includes('余额不足') || errorMsg.includes('balance')) {
              errorType.value = 'INSUFFICIENT_BALANCE'
            } else if (errorMsg.includes('超时') || errorMsg.includes('timeout')) {
              errorType.value = 'SERVICE_TIMEOUT'
            } else if (errorMsg.includes('服务提供商未配置') || errorMsg.includes('PROVIDER_NOT_AVAILABLE')) {
              errorType.value = 'PROVIDER_NOT_AVAILABLE'
            } else if (errorMsg.includes('API 密钥无效') || errorMsg.includes('INVALID_API_KEY')) {
              errorType.value = 'INVALID_API_KEY'
            } else if (errorMsg.includes('请求过于频繁') || errorMsg.includes('RATE_LIMITED')) {
              errorType.value = 'RATE_LIMITED'
            } else if (errorMsg.includes('消息内容过长') || errorMsg.includes('CONTEXT_TOO_LONG')) {
              errorType.value = 'CONTEXT_TOO_LONG'
            } else if (errorMsg.includes('不可用') || errorMsg.includes('unavailable') || errorMsg.includes('PROVIDER_ERROR')) {
              errorType.value = 'PROVIDER_ERROR'
            } else {
              errorType.value = 'UNKNOWN_ERROR'
            }
            isStreaming.value = false
            streamingContent.value = ''
            streamingMessageId.value = null
            streamController.value = null
            // Keep lastFailedMessage for retry
          }
        }
      )
      
      return true
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to send message'
      error.value = errorMessage
      errorType.value = 'UNKNOWN_ERROR'
      isStreaming.value = false
      streamingContent.value = ''
      streamingMessageId.value = null
      streamController.value = null
      // Keep lastFailedMessage for retry
      return false
    }
  }
  
  /**
   * Cancel ongoing stream
   * 取消正在进行的流
   */
  function cancelStream(): void {
    if (streamController.value) {
      streamController.value.abort()
      streamController.value = null
    }
    
    // If there was streaming content, save it as a partial message
    if (streamingContent.value && currentConversation.value) {
      const partialMessage: Message = {
        id: streamingMessageId.value || Date.now(),
        conversation_id: currentConversation.value.id,
        role: 'assistant',
        content: streamingContent.value + '\n\n[已中断]',
        tokens: 0,
        cost: 0,
        created_at: new Date().toISOString()
      }
      messages.value.push(partialMessage)
      messagesTotal.value++
    }
    
    isStreaming.value = false
    streamingContent.value = ''
    streamingMessageId.value = null
  }
  
  // ---------------------------------------------------------------------------
  // Model Actions
  // ---------------------------------------------------------------------------
  
  /**
   * Load available models
   * 加载可用模型列表
   */
  async function loadModels(): Promise<void> {
    modelsLoading.value = true
    error.value = null
    
    try {
      models.value = await chatApi.getChatModels()
      
      // Set default model if not set
      if (!selectedModel.value && models.value.length > 0) {
        const firstModel = models.value[0]
        if (firstModel) {
          selectedModel.value = firstModel.id
        }
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to load models'
      error.value = errorMessage
      console.error('Failed to load models:', err)
    } finally {
      modelsLoading.value = false
    }
  }
  
  /**
   * Set selected model
   * 设置选中的模型
   */
  function setSelectedModel(modelId: string): void {
    selectedModel.value = modelId
    
    // Update current conversation's model if exists
    if (currentConversation.value) {
      updateConversation(currentConversation.value.id, { model: modelId })
    }
  }
  
  // ---------------------------------------------------------------------------
  // Utility Actions
  // ---------------------------------------------------------------------------
  
  /**
   * Clear error state
   * 清除错误状态
   */
  function clearError(): void {
    error.value = null
    errorType.value = null
  }
  
  /**
   * Retry the last failed message
   * 重试上次失败的消息
   * Requirements: 2.5 - Retry mechanism for failed requests
   */
  async function retryLastMessage(): Promise<boolean> {
    if (!lastFailedMessage.value || isStreaming.value) {
      return false
    }
    
    // Remove the last user message that failed (it will be re-added by sendMessage)
    if (messages.value.length > 0) {
      const lastMsg = messages.value[messages.value.length - 1]
      if (lastMsg && lastMsg.role === 'user' && lastMsg.content === lastFailedMessage.value) {
        messages.value.pop()
        messagesTotal.value--
      }
    }
    
    clearError()
    const messageToRetry = lastFailedMessage.value
    return sendMessage(messageToRetry)
  }
  
  /**
   * Check if retry is available
   * 检查是否可以重试
   */
  const canRetry = computed(() => {
    return !!lastFailedMessage.value && !isStreaming.value && !!error.value
  })
  
  /**
   * Check if error is due to insufficient balance
   * 检查是否是余额不足错误
   */
  const isInsufficientBalance = computed(() => {
    return errorType.value === 'INSUFFICIENT_BALANCE'
  })
  
  /**
   * Check if error is a provider-related error
   * 检查是否是提供商相关错误
   * Requirements: 2.6, 10.1-10.5
   */
  const isProviderError = computed(() => {
    return [
      'PROVIDER_NOT_AVAILABLE',
      'INVALID_API_KEY',
      'RATE_LIMITED',
      'PROVIDER_ERROR',
      'CONTEXT_TOO_LONG'
    ].includes(errorType.value || '')
  })
  
  /**
   * Get user-friendly error message based on error type
   * 根据错误类型获取用户友好的错误消息
   * Requirements: 2.6, 10.1-10.5
   */
  const errorMessage = computed(() => {
    if (!error.value) return null
    
    // Return the error message directly as it's already user-friendly
    return error.value
  })
  
  /**
   * Reset store state
   * 重置 store 状态
   */
  function reset(): void {
    cancelStream()
    
    conversations.value = []
    currentConversation.value = null
    conversationsTotal.value = 0
    conversationsPage.value = 1
    
    messages.value = []
    messagesTotal.value = 0
    messagesPage.value = 1
    
    error.value = null
  }
  
  // ---------------------------------------------------------------------------
  // Return Store
  // ---------------------------------------------------------------------------
  
  return {
    // Conversation state
    conversations,
    currentConversation,
    conversationsTotal,
    conversationsPage,
    conversationsLoading,
    
    // Message state
    messages,
    messagesTotal,
    messagesPage,
    messagesLoading,
    
    // Streaming state
    isStreaming,
    streamingContent,
    streamingMessageId,
    
    // Model state
    models,
    modelsLoading,
    selectedModel,
    
    // Error state
    error,
    errorType,
    lastFailedMessage,
    
    // Computed
    hasMoreConversations,
    hasMoreMessages,
    currentModel,
    canRetry,
    isInsufficientBalance,
    isProviderError,
    errorMessage,
    
    // Conversation actions
    loadConversations,
    loadMoreConversations,
    createConversation,
    selectConversation,
    updateConversation,
    deleteConversation,
    
    // Message actions
    loadMessages,
    loadMoreMessages,
    sendMessage,
    cancelStream,
    retryLastMessage,
    
    // Model actions
    loadModels,
    setSelectedModel,
    
    // Utility actions
    clearError,
    reset
  }
})

// Re-export types for convenience
export type { Conversation, Message, ChatModel, TokenUsage } from '@/api/chat'
