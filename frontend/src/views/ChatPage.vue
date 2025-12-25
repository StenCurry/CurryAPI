<template>
  <div class="chat-page">
    <!-- Mobile sidebar toggle -->
    <n-button
      v-if="isMobile"
      class="mobile-sidebar-toggle"
      circle
      @click="showSidebar = !showSidebar"
    >
      <template #icon>
        <n-icon><MenuOutline /></n-icon>
      </template>
    </n-button>

    <!-- Sidebar overlay for mobile -->
    <div
      v-if="isMobile && showSidebar"
      class="sidebar-overlay"
      @click="showSidebar = false"
    />

    <!-- Conversation Sidebar -->
    <ChatSidebar
      :conversations="chatStore.conversations"
      :current-conversation-id="chatStore.currentConversation?.id"
      :visible="showSidebar || !isMobile"
      :loading="chatStore.conversationsLoading"
      :has-more="chatStore.hasMoreConversations"
      @new-conversation="handleNewConversation"
      @select-conversation="handleSelectConversation"
      @delete-conversation="handleDeleteConversation"
      @load-more="chatStore.loadMoreConversations()"
    />

    <!-- Main Chat Area -->
    <main class="chat-main">
      <!-- No conversation selected -->
      <!-- Requirements: 5.1 - Empty state when no conversation selected -->
      <div v-if="!chatStore.currentConversation" class="no-conversation">
        <div class="welcome-content">
          <div class="welcome-icon">
            <n-icon size="72" color="rgba(59, 130, 246, 0.5)">
              <ChatbubblesOutline />
            </n-icon>
          </div>
          <h2 class="welcome-title">欢迎使用 AI 聊天</h2>
          <p class="welcome-description">
            选择左侧的对话继续聊天，或创建一个新对话开始与 AI 交流
          </p>
          <n-button type="primary" size="large" @click="handleNewConversation">
            <template #icon>
              <n-icon><AddOutline /></n-icon>
            </template>
            开始新对话
          </n-button>
          <div class="welcome-features">
            <div class="feature-item">
              <n-icon size="20" color="#10b981"><SparklesOutline /></n-icon>
              <span>多种 AI 模型可选</span>
            </div>
            <div class="feature-item">
              <n-icon size="20" color="#8b5cf6"><TimeOutline /></n-icon>
              <span>对话历史自动保存</span>
            </div>
            <div class="feature-item">
              <n-icon size="20" color="#f59e0b"><CodeSlashOutline /></n-icon>
              <span>支持代码高亮显示</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Conversation content placeholder -->
      <div v-else class="conversation-content">
        <div class="chat-header">
          <div class="chat-title-section">
            <ModelSelector
              v-model="chatStore.selectedModel"
              :models="chatStore.models"
              :loading="chatStore.modelsLoading"
              :disabled="chatStore.isStreaming"
              @change="handleModelChange"
            />
          </div>
          <div class="chat-title-wrapper">
            <n-input
              v-if="isEditingTitle"
              v-model:value="editingTitle"
              size="small"
              @blur="saveTitle"
              @keyup.enter="saveTitle"
              ref="titleInputRef"
            />
            <span v-else class="chat-title" @click="startEditTitle">
              {{ chatStore.currentConversation.title }}
              <n-icon size="12" class="edit-icon"><CreateOutline /></n-icon>
            </span>
          </div>
        </div>

        <!-- Messages area -->
        <MessageList
          :messages="chatStore.messages"
          :loading="chatStore.messagesLoading"
          :has-more="chatStore.hasMoreMessages"
          :is-streaming="chatStore.isStreaming"
          :streaming-content="chatStore.streamingContent"
          :conversation-id="chatStore.currentConversation?.id"
          @load-more="chatStore.loadMoreMessages()"
        />

        <!-- Error display with retry -->
        <!-- Requirements: 2.5, 6.2 - Network error display and balance insufficient warning -->
        <div v-if="chatStore.error && !chatStore.isStreaming" class="error-banner">
          <n-alert 
            :type="chatStore.isInsufficientBalance ? 'warning' : 'error'" 
            closable 
            @close="chatStore.clearError()"
          >
            <template #header>
              {{ chatStore.isInsufficientBalance ? '余额不足' : '发送失败' }}
            </template>
            {{ chatStore.error }}
            <template #action>
              <n-space>
                <n-button 
                  v-if="chatStore.isInsufficientBalance" 
                  size="small" 
                  type="primary"
                  @click="goToRecharge"
                >
                  <template #icon>
                    <n-icon><WalletOutline /></n-icon>
                  </template>
                  去充值
                </n-button>
                <n-button 
                  v-if="chatStore.canRetry && !chatStore.isInsufficientBalance" 
                  size="small" 
                  @click="handleRetry"
                >
                  <template #icon>
                    <n-icon><RefreshOutline /></n-icon>
                  </template>
                  重试
                </n-button>
              </n-space>
            </template>
          </n-alert>
        </div>

        <!-- Streaming controls with typing indicator -->
        <!-- Requirements: 5.3 - Streaming indicator -->
        <div v-if="chatStore.isStreaming" class="streaming-controls">
          <div class="typing-indicator">
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-dot"></span>
            <span class="typing-text">AI 正在思考...</span>
          </div>
          <n-button size="small" type="warning" @click="handleCancelStream">
            <template #icon>
              <n-icon><StopOutline /></n-icon>
            </template>
            停止生成
          </n-button>
        </div>

        <!-- Message input component -->
        <MessageInput
          ref="messageInputRef"
          v-model="inputMessage"
          :is-streaming="chatStore.isStreaming"
          @send="handleSendMessage"
        />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
/**
 * ChatPage.vue - Main chat page layout
 * 聊天页面主布局
 * 
 * Two-column layout with sidebar and main chat area
 * Responsive design for mobile with collapsible sidebar
 * Handles SSE streaming, typing effect, stream completion and errors
 * 
 * Requirements: 2.2, 2.5, 5.1, 5.5
 */

import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, useDialog } from 'naive-ui'
import {
  MenuOutline,
  AddOutline,
  ChatbubblesOutline,
  CreateOutline,
  RefreshOutline,
  StopOutline,
  WalletOutline,
  SparklesOutline,
  TimeOutline,
  CodeSlashOutline
} from '@vicons/ionicons5'
import { useChatStore } from '@/stores/chat'
import ChatSidebar from '@/components/chat/ChatSidebar.vue'
import MessageList from '@/components/chat/MessageList.vue'
import MessageInput from '@/components/chat/MessageInput.vue'
import ModelSelector from '@/components/chat/ModelSelector.vue'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()
const chatStore = useChatStore()

// Responsive state
const isMobile = ref(window.innerWidth < 768)
const showSidebar = ref(false)

// Input state
const inputMessage = ref('')

// Title editing state
const isEditingTitle = ref(false)
const editingTitle = ref('')
const titleInputRef = ref<HTMLInputElement | null>(null)

// Message input ref
const messageInputRef = ref<InstanceType<typeof MessageInput> | null>(null)



// Handle window resize for responsive design
function handleResize() {
  isMobile.value = window.innerWidth < 768
  if (!isMobile.value) {
    showSidebar.value = false
  }
}

// Initialize data on mount
onMounted(async () => {
  window.addEventListener('resize', handleResize)
  
  // Load models and conversations
  await Promise.all([
    chatStore.loadModels(),
    chatStore.loadConversations()
  ])
  
  // Handle route parameter for specific conversation
  const conversationId = route.params.id
  if (conversationId) {
    await chatStore.selectConversation(Number(conversationId))
  }
})

// Cleanup on unmount - cancel any ongoing streams
onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chatStore.cancelStream()
})

// Watch for route changes to load specific conversation
watch(() => route.params.id, async (newId) => {
  if (newId) {
    await chatStore.selectConversation(Number(newId))
    if (isMobile.value) {
      showSidebar.value = false
    }
  }
})

// Create new conversation
async function handleNewConversation() {
  const conv = await chatStore.createConversation()
  if (conv) {
    router.push({ name: 'ChatConversation', params: { id: conv.id } })
    if (isMobile.value) {
      showSidebar.value = false
    }
  }
}

// Select conversation
async function handleSelectConversation(id: number) {
  router.push({ name: 'ChatConversation', params: { id } })
  if (isMobile.value) {
    showSidebar.value = false
  }
}

// Delete conversation
function handleDeleteConversation(id: number) {
  dialog.warning({
    title: '确认删除',
    content: '删除后无法恢复，确定要删除这个对话吗？',
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const success = await chatStore.deleteConversation(id)
      if (success) {
        message.success('对话已删除')
        if (chatStore.currentConversation?.id === id) {
          router.push({ name: 'Chat' })
        }
      } else {
        message.error('删除失败')
      }
    }
  })
}

// Title editing
function startEditTitle() {
  if (chatStore.currentConversation) {
    editingTitle.value = chatStore.currentConversation.title
    isEditingTitle.value = true
    nextTick(() => {
      titleInputRef.value?.focus()
    })
  }
}

async function saveTitle() {
  if (chatStore.currentConversation && editingTitle.value.trim()) {
    await chatStore.updateConversation(chatStore.currentConversation.id, {
      title: editingTitle.value.trim()
    })
  }
  isEditingTitle.value = false
}

// Model change
// Requirements: 3.2, 3.3
async function handleModelChange(model: { id: string }) {
  chatStore.setSelectedModel(model.id)
}

// Send message with streaming response handling
// Requirements: 2.2, 2.5
async function handleSendMessage(content: string) {
  if (!content.trim() || chatStore.isStreaming) return
  
  // Note: MessageInput component handles clearing the input internally
  // Don't clear inputMessage.value here to avoid double update
  
  const success = await chatStore.sendMessage(content)
  if (!success && chatStore.error) {
    // Error is displayed in the error banner, also show a toast for visibility
    message.error(chatStore.error)
  }
}

// Retry last failed message
// Requirements: 2.5 - Retry mechanism for failed requests
async function handleRetry() {
  if (chatStore.canRetry) {
    const success = await chatStore.retryLastMessage()
    if (!success && chatStore.error) {
      message.error(chatStore.error)
    }
  }
}

// Navigate to recharge page
// Requirements: 6.2 - Balance insufficient warning with action
function goToRecharge() {
  chatStore.clearError()
  router.push({ name: 'Dashboard' })
}

// Cancel ongoing stream
// Requirements: 2.5
function handleCancelStream() {
  chatStore.cancelStream()
  message.info('已停止生成')
}
</script>

<style scoped>
.chat-page {
  display: flex;
  height: calc(100vh - 100px); /* 减去头部高度和外边距 */
  position: relative;
  background: var(--bg-card);
  backdrop-filter: var(--backdrop-blur);
  border-radius: var(--border-radius-lg);
  overflow: hidden;
  margin: 0;
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-lg);
}

/* 移动端侧边栏切换按钮 */
.mobile-sidebar-toggle {
  position: absolute;
  top: 12px;
  left: 12px;
  z-index: 100;
  width: 40px;
  height: 40px;
  box-shadow: var(--shadow-md);
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
}

/* 侧边栏遮罩 */
.sidebar-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  backdrop-filter: blur(4px);
  z-index: 199;
  animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 1; }
}

/* 主聊天区域 */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  position: relative;
  background: var(--bg-primary);
}

/* 空状态 - 欢迎界面 */
.no-conversation {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 2rem;
  background: radial-gradient(circle at center, var(--bg-secondary) 0%, var(--bg-primary) 100%);
}

.welcome-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  max-width: 480px;
  padding: 40px;
  background: var(--bg-card);
  border-radius: 24px;
  border: 1px solid var(--border-color);
  box-shadow: var(--shadow-xl);
}

.welcome-icon {
  margin-bottom: 24px;
  animation: float 4s ease-in-out infinite;
  background: var(--bg-secondary);
  width: 100px;
  height: 100px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-md);
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-12px); }
}

.welcome-title {
  margin: 0 0 12px;
  font-size: 1.75rem;
  font-weight: 800;
  background: linear-gradient(135deg, var(--color-primary) 0%, #8b5cf6 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
}

.welcome-description {
  margin: 0 0 32px;
  font-size: 1rem;
  color: var(--text-secondary);
  line-height: 1.6;
}

.welcome-features {
  margin-top: 32px;
  display: flex;
  flex-direction: column;
  gap: 16px;
  width: 100%;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 0.95rem;
  color: var(--text-primary);
  padding: 12px 16px;
  background: var(--bg-secondary);
  border-radius: 12px;
  transition: transform 0.2s;
}

.feature-item:hover {
  transform: translateX(4px);
  background: var(--bg-hover);
}

/* 聊天内容区域 */
.conversation-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.chat-header {
  padding: 16px 24px;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(12px);
  z-index: 10;
}

.dark-theme .chat-header {
  background: rgba(15, 23, 42, 0.8);
}

.chat-title-section {
  flex-shrink: 0;
}

.chat-title-wrapper {
  flex: 1;
  display: flex;
  justify-content: center;
  min-width: 0;
  padding: 0 20px;
}

.chat-title {
  font-size: 0.95rem;
  color: var(--text-primary);
  font-weight: 600;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  padding: 6px 12px;
  border-radius: 8px;
  transition: all 0.2s;
}

.chat-title:hover {
  background: var(--bg-hover);
}

.edit-icon {
  opacity: 0;
  transition: opacity 0.2s;
  color: var(--text-muted);
}

.chat-title:hover .edit-icon {
  opacity: 1;
}

/* 错误横幅 */
.error-banner {
  padding: 12px 24px;
  animation: slideDown 0.3s ease-out;
}

@keyframes slideDown {
  from { transform: translateY(-10px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}

.error-banner :deep(.n-alert) {
  background: var(--bg-card);
  border: 1px solid var(--color-error-light);
  box-shadow: var(--shadow-sm);
}

/* 流式传输控制栏 */
.streaming-controls {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 16px;
  padding: 8px 16px;
  background: var(--bg-card);
  border: 1px solid var(--color-primary-light);
  border-radius: 20px;
  margin: 0 auto 12px;
  box-shadow: var(--shadow-md);
  max-width: fit-content;
  animation: slideUp 0.3s ease-out;
}

.typing-indicator {
  display: flex;
  align-items: center;
  gap: 4px;
}

.typing-dot {
  width: 6px;
  height: 6px;
  background: var(--color-primary);
  border-radius: 50%;
  animation: typing-bounce 1.4s ease-in-out infinite;
}

.typing-dot:nth-child(1) { animation-delay: 0s; }
.typing-dot:nth-child(2) { animation-delay: 0.2s; }
.typing-dot:nth-child(3) { animation-delay: 0.4s; }

.typing-text {
  margin-left: 8px;
  font-size: 0.85rem;
  color: var(--text-secondary);
  font-weight: 500;
}

@keyframes typing-bounce {
  0%, 60%, 100% { transform: translateY(0); }
  30% { transform: translateY(-4px); }
}

/* 响应式适配 */
@media (max-width: 768px) {
  .chat-page {
    margin: 0;
    border-radius: 0;
    height: 100vh;
    border: none;
  }

  .chat-header {
    padding-left: 60px;
  }

  .welcome-content {
    padding: 24px;
    margin: 16px;
  }

  .welcome-title {
    font-size: 1.5rem;
  }

  .chat-title {
    max-width: 150px;
  }
}

@media (max-width: 480px) {
  .no-conversation {
    padding: 16px;
  }

  .welcome-features {
    gap: 12px;
  }
}

/* 触摸设备优化 */
@media (hover: none) and (pointer: coarse) {
  .chat-title:active {
    background: var(--bg-hover);
  }
}
</style>
