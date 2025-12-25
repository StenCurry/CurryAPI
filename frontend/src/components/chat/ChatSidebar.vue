<template>
  <aside
    class="chat-sidebar"
    :class="{ 'sidebar-visible': visible }"
  >
    <!-- Sidebar Header -->
    <div class="sidebar-header">
      <h3 class="sidebar-title">对话列表</h3>
      <n-button
        type="primary"
        size="small"
        @click="handleNewConversation"
        :loading="loading"
      >
        <template #icon>
          <n-icon><AddOutline /></n-icon>
        </template>
        新对话
      </n-button>
    </div>

    <!-- Conversation List -->
    <div class="conversation-list">
      <!-- Loading Skeleton State -->
      <!-- Requirements: 5.3 - Loading indicator for conversation list -->
      <template v-if="loading && conversations.length === 0">
        <div class="conversation-skeleton" v-for="i in 5" :key="i">
          <div class="skeleton-content">
            <div class="skeleton-line skeleton-title"></div>
            <div class="skeleton-line skeleton-model"></div>
          </div>
        </div>
      </template>

      <!-- Empty State -->
      <!-- Requirements: 5.1 - No conversations message -->
      <div v-else-if="conversations.length === 0" class="empty-conversations">
        <div class="empty-icon">
          <n-icon size="48" color="rgba(59, 130, 246, 0.5)">
            <ChatbubblesOutline />
          </n-icon>
        </div>
        <p class="empty-title">暂无对话</p>
        <p class="empty-description">开始一个新对话，与 AI 交流吧</p>
        <n-button type="primary" size="small" @click="handleNewConversation">
          <template #icon>
            <n-icon><AddOutline /></n-icon>
          </template>
          创建第一个对话
        </n-button>
      </div>

      <!-- Conversation Items -->
      <template v-else>
        <div
          v-for="conv in conversations"
          :key="conv.id"
          class="conversation-item"
          :class="{ active: currentConversationId === conv.id }"
          @click="handleSelectConversation(conv.id)"
        >
          <div class="conversation-info">
            <span class="conversation-title">{{ conv.title }}</span>
          </div>
          <n-button
            text
            class="delete-btn"
            @click.stop="handleDeleteConversation(conv.id)"
          >
            <template #icon>
              <n-icon><TrashOutline /></n-icon>
            </template>
          </n-button>
        </div>

        <!-- Load More Button -->
        <div v-if="hasMore" class="load-more">
          <n-button
            text
            :loading="loading"
            @click="handleLoadMore"
          >
            加载更多
          </n-button>
        </div>
      </template>
    </div>
  </aside>
</template>

<script setup lang="ts">
/**
 * ChatSidebar.vue - Conversation sidebar component
 * 会话侧边栏组件
 * 
 * Displays conversation list with new conversation button,
 * conversation selection, and delete functionality.
 * 
 * Requirements: 1.1, 1.2, 1.4
 */

import { AddOutline, TrashOutline, ChatbubblesOutline } from '@vicons/ionicons5'
import type { Conversation } from '@/api/chat'

// ============================================================================
// Props
// ============================================================================

interface Props {
  /** List of conversations to display */
  conversations: Conversation[]
  /** Currently selected conversation ID */
  currentConversationId?: number | null
  /** Whether the sidebar is visible (for mobile) */
  visible?: boolean
  /** Loading state */
  loading?: boolean
  /** Whether there are more conversations to load */
  hasMore?: boolean
}

withDefaults(defineProps<Props>(), {
  conversations: () => [],
  currentConversationId: null,
  visible: true,
  loading: false,
  hasMore: false
})

// ============================================================================
// Emits
// ============================================================================

const emit = defineEmits<{
  /** Emitted when user clicks new conversation button */
  (e: 'new-conversation'): void
  /** Emitted when user selects a conversation */
  (e: 'select-conversation', id: number): void
  /** Emitted when user clicks delete on a conversation */
  (e: 'delete-conversation', id: number): void
  /** Emitted when user clicks load more */
  (e: 'load-more'): void
}>()

// ============================================================================
// Event Handlers
// ============================================================================

function handleNewConversation() {
  emit('new-conversation')
}

function handleSelectConversation(id: number) {
  emit('select-conversation', id)
}

function handleDeleteConversation(id: number) {
  emit('delete-conversation', id)
}

function handleLoadMore() {
  emit('load-more')
}
</script>

<style scoped>
.chat-sidebar {
  width: 280px;
  min-width: 280px;
  border-right: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  background: var(--bg-card);
}

.sidebar-header {
  padding: 1rem;
  border-bottom: 1px solid var(--border-color);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.sidebar-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.conversation-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
  -webkit-overflow-scrolling: touch;
}

/* Empty State Styles */
/* Requirements: 5.1 - No conversations message */
.empty-conversations {
  padding: 2rem 1rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 0.75rem;
}

.empty-icon {
  margin-bottom: 0.5rem;
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-8px);
  }
}

.empty-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-description {
  margin: 0;
  font-size: 0.875rem;
  color: var(--text-secondary);
}

.conversation-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-radius: var(--border-radius);
  cursor: pointer;
  transition: all var(--transition-fast);
  margin-bottom: 0.25rem;
  min-height: 44px;
}

.conversation-item:hover {
  background: var(--color-primary-light);
}

.conversation-item.active {
  background: var(--color-primary-light);
  border-left: 3px solid var(--color-primary);
}

.conversation-info {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.conversation-title {
  font-size: 0.875rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-primary);
}

.delete-btn {
  opacity: 0;
  transition: opacity var(--transition-fast);
  min-width: 32px;
  min-height: 32px;
}

.conversation-item:hover .delete-btn {
  opacity: 1;
}

.load-more {
  text-align: center;
  padding: 0.5rem;
}

/* Skeleton Loading Styles */
/* Requirements: 5.3 - Loading indicator */
.conversation-skeleton {
  padding: 0.75rem 1rem;
  margin-bottom: 0.25rem;
}

.skeleton-content {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.skeleton-line {
  background: linear-gradient(
    90deg,
    var(--color-primary-light) 0%,
    var(--bg-hover) 50%,
    var(--color-primary-light) 100%
  );
  background-size: 200% 100%;
  border-radius: var(--border-radius-sm);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-title {
  height: 14px;
  width: 80%;
}

.skeleton-model {
  height: 12px;
  width: 50%;
}

@keyframes skeleton-pulse {
  0%, 100% {
    background-position: 200% 0;
  }
  50% {
    background-position: 0% 0;
  }
}

/* Mobile responsive */
@media (max-width: 768px) {
  .chat-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: 200;
    transform: translateX(-100%);
    transition: transform var(--transition-normal);
  }

  .chat-sidebar.sidebar-visible {
    transform: translateX(0);
  }

  /* Ensure buttons have minimum touch target size */
  .sidebar-header :deep(.n-button) {
    min-height: 44px;
  }

  .empty-conversations :deep(.n-button) {
    min-height: 44px;
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .delete-btn {
    opacity: 0.7;
  }

  .conversation-item:active {
    background: var(--color-primary-light);
  }
}
</style>
