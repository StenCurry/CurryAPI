<template>
  <div class="message-list" ref="containerRef">
    <!-- Loading Skeleton State -->
    <!-- Requirements: 5.3 - Loading indicator for messages -->
    <template v-if="loading">
      <div class="message-skeleton" v-for="i in 3" :key="i">
        <div class="skeleton-avatar"></div>
        <div class="skeleton-content">
          <div class="skeleton-line skeleton-line-full"></div>
          <div class="skeleton-line skeleton-line-medium"></div>
          <div class="skeleton-line skeleton-line-short"></div>
        </div>
      </div>
    </template>

    <!-- Empty state -->
    <!-- Requirements: 5.1 - No messages in conversation -->
    <div v-else-if="messages.length === 0 && !streamingContent" class="no-messages">
      <div class="empty-messages-content">
        <div class="empty-icon">
          <n-icon size="64" color="rgba(59, 130, 246, 0.4)">
            <ChatboxEllipsesOutline />
          </n-icon>
        </div>
        <h3 class="empty-title">开始对话</h3>
        <p class="empty-description">在下方输入框中输入消息，开始与 AI 交流</p>
        <div class="empty-suggestions">
          <span class="suggestion-label">试试问：</span>
          <div class="suggestion-chips">
            <span class="suggestion-chip">帮我写一段代码</span>
            <span class="suggestion-chip">解释一个概念</span>
            <span class="suggestion-chip">翻译一段文字</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Messages -->
    <template v-else>
      <!-- Load more button for older messages -->
      <div v-if="hasMore" class="load-more">
        <n-button text :loading="loadingMore" @click="handleLoadMore">
          加载更多历史消息
        </n-button>
      </div>

      <!-- Message items -->
      <MessageItem
        v-for="msg in messages"
        :key="msg.id"
        :message="msg"
      />

      <!-- Streaming message -->
      <MessageItem
        v-if="isStreaming && streamingContent"
        :message="streamingMessage"
        :is-streaming="true"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
/**
 * MessageList.vue - Message list component
 * 消息列表组件
 * 
 * Displays messages in a conversation with auto-scroll to bottom
 * and loading states.
 * 
 * Requirements: 1.3, 5.4
 */

import { ref, computed, watch, nextTick, onMounted } from 'vue'
import { ChatboxEllipsesOutline } from '@vicons/ionicons5'
import type { Message } from '@/api/chat'
import MessageItem from './MessageItem.vue'

// ============================================================================
// Props
// ============================================================================

interface Props {
  /** List of messages to display */
  messages: Message[]
  /** Whether messages are loading */
  loading?: boolean
  /** Whether more messages are being loaded */
  loadingMore?: boolean
  /** Whether there are more messages to load */
  hasMore?: boolean
  /** Whether currently streaming a response */
  isStreaming?: boolean
  /** Current streaming content */
  streamingContent?: string
  /** Conversation ID for streaming message */
  conversationId?: number
}

const props = withDefaults(defineProps<Props>(), {
  messages: () => [],
  loading: false,
  loadingMore: false,
  hasMore: false,
  isStreaming: false,
  streamingContent: '',
  conversationId: 0
})

// ============================================================================
// Emits
// ============================================================================

const emit = defineEmits<{
  /** Emitted when user wants to load more messages */
  (e: 'load-more'): void
}>()

// ============================================================================
// Refs
// ============================================================================

const containerRef = ref<HTMLElement | null>(null)

// ============================================================================
// Computed
// ============================================================================

/** Create a temporary message object for streaming content */
const streamingMessage = computed<Message>(() => ({
  id: -1,
  conversation_id: props.conversationId,
  role: 'assistant',
  content: props.streamingContent,
  tokens: 0,
  cost: 0,
  created_at: new Date().toISOString()
}))

// ============================================================================
// Methods
// ============================================================================

/** Scroll to the bottom of the message list */
function scrollToBottom(smooth = true) {
  nextTick(() => {
    if (containerRef.value) {
      containerRef.value.scrollTo({
        top: containerRef.value.scrollHeight,
        behavior: smooth ? 'smooth' : 'auto'
      })
    }
  })
}

/** Handle load more button click */
function handleLoadMore() {
  emit('load-more')
}

// ============================================================================
// Watchers
// ============================================================================

// Auto-scroll when new messages are added
watch(
  () => props.messages.length,
  (newLen, oldLen) => {
    if (newLen > oldLen) {
      scrollToBottom()
    }
  }
)

// Auto-scroll when streaming content updates
watch(
  () => props.streamingContent,
  () => {
    if (props.isStreaming) {
      scrollToBottom(false)
    }
  }
)

// Scroll to bottom on initial load
onMounted(() => {
  if (props.messages.length > 0) {
    scrollToBottom(false)
  }
})

// ============================================================================
// Expose
// ============================================================================

defineExpose({
  scrollToBottom
})
</script>

<style scoped>
.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 1rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

/* Empty State Styles */
/* Requirements: 5.1 - No messages in conversation */
.no-messages {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-messages-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  padding: 2rem;
  max-width: 400px;
}

.empty-icon {
  margin-bottom: 1rem;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-10px);
  }
}

.empty-title {
  margin: 0 0 0.5rem;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-description {
  margin: 0 0 1.5rem;
  font-size: 0.875rem;
  color: var(--text-secondary);
  line-height: 1.5;
}

.empty-suggestions {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.75rem;
}

.suggestion-label {
  font-size: 0.75rem;
  color: var(--text-muted);
}

.suggestion-chips {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 0.5rem;
}

.suggestion-chip {
  padding: 0.375rem 0.75rem;
  background: var(--color-primary-light);
  border: 1px solid var(--color-primary);
  border-radius: 16px;
  font-size: 0.75rem;
  color: var(--color-primary);
  cursor: default;
  transition: all var(--transition-fast);
}

.suggestion-chip:hover {
  background: var(--bg-hover);
}

.load-more {
  text-align: center;
  padding: 0.5rem;
  margin-bottom: 0.5rem;
}

/* Message Skeleton Loading Styles */
/* Requirements: 5.3 - Loading indicator */
.message-skeleton {
  display: flex;
  gap: 0.75rem;
  padding: 0.75rem;
  animation: fade-in 0.3s ease-out;
}

.message-skeleton:nth-child(even) {
  flex-direction: row-reverse;
}

.message-skeleton:nth-child(even) .skeleton-content {
  align-items: flex-end;
}

.skeleton-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(
    90deg,
    var(--color-primary-light) 0%,
    var(--bg-hover) 50%,
    var(--color-primary-light) 100%
  );
  background-size: 200% 100%;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
  flex-shrink: 0;
}

.skeleton-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  flex: 1;
  max-width: 70%;
}

.skeleton-line {
  height: 14px;
  border-radius: var(--border-radius-sm);
  background: linear-gradient(
    90deg,
    var(--color-primary-light) 0%,
    var(--bg-hover) 50%,
    var(--color-primary-light) 100%
  );
  background-size: 200% 100%;
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-line-full {
  width: 100%;
}

.skeleton-line-medium {
  width: 75%;
}

.skeleton-line-short {
  width: 40%;
}

@keyframes skeleton-pulse {
  0%, 100% {
    background-position: 200% 0;
  }
  50% {
    background-position: 0% 0;
  }
}

@keyframes fade-in {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
