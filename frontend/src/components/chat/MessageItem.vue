<template>
  <div class="message-item" :class="[message.role, { streaming: isStreaming }]">
    <!-- Avatar -->
    <div class="message-avatar">
      <n-icon v-if="message.role === 'user'" size="20">
        <PersonOutline />
      </n-icon>
      <n-icon v-else size="20">
        <SparklesOutline />
      </n-icon>
    </div>

    <!-- Content -->
    <div class="message-body">
      <div class="message-header">
        <span class="message-role">{{ roleLabel }}</span>
        <span class="message-time">{{ formattedTime }}</span>
      </div>
      <div
        class="message-content"
        :class="{ 'markdown-body': message.role === 'assistant' }"
        v-html="renderedContent"
      />
      <!-- Streaming cursor -->
      <span v-if="isStreaming" class="streaming-cursor">▊</span>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * MessageItem.vue - Single message display component
 * 单条消息显示组件
 * 
 * Renders user and assistant messages with markdown support,
 * code syntax highlighting, and copy functionality.
 * 
 * Requirements: 4.1, 4.2, 4.3, 4.4
 */

import { computed, onMounted, onUpdated } from 'vue'
import { PersonOutline, SparklesOutline } from '@vicons/ionicons5'
import type { Message } from '@/api/chat'
import { renderMarkdown } from '@/utils/markdown'
import dayjs from 'dayjs'

// ============================================================================
// Props
// ============================================================================

interface Props {
  /** Message to display */
  message: Message
  /** Whether this message is currently streaming */
  isStreaming?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  isStreaming: false
})

// ============================================================================
// Computed
// ============================================================================

/** Role label for display */
const roleLabel = computed(() => {
  switch (props.message.role) {
    case 'user':
      return '你'
    case 'assistant':
      return 'AI'
    case 'system':
      return '系统'
    default:
      return props.message.role
  }
})

/** Formatted timestamp */
const formattedTime = computed(() => {
  return dayjs(props.message.created_at).format('HH:mm')
})

/** Rendered markdown content */
const renderedContent = computed(() => {
  if (props.message.role === 'user') {
    // User messages: escape HTML and convert newlines to <br>
    return escapeHtml(props.message.content).replace(/\n/g, '<br>')
  }
  // Assistant messages: render markdown
  return renderMarkdown(props.message.content)
})

// ============================================================================
// Methods
// ============================================================================

/** Escape HTML special characters */
function escapeHtml(text: string): string {
  const div = document.createElement('div')
  div.textContent = text
  return div.innerHTML
}

/** Add copy buttons to code blocks after render */
function addCopyButtons() {
  const codeBlocks = document.querySelectorAll('.message-content pre code')
  codeBlocks.forEach((block) => {
    const pre = block.parentElement
    if (!pre || pre.querySelector('.copy-btn')) return

    const copyBtn = document.createElement('button')
    copyBtn.className = 'copy-btn'
    copyBtn.textContent = '复制'
    copyBtn.onclick = async () => {
      try {
        await navigator.clipboard.writeText(block.textContent || '')
        copyBtn.textContent = '已复制!'
        setTimeout(() => {
          copyBtn.textContent = '复制'
        }, 2000)
      } catch {
        copyBtn.textContent = '复制失败'
      }
    }
    pre.style.position = 'relative'
    pre.appendChild(copyBtn)
  })
}

// ============================================================================
// Lifecycle
// ============================================================================

onMounted(() => {
  addCopyButtons()
})

onUpdated(() => {
  addCopyButtons()
})
</script>

<style scoped>
.message-item {
  display: flex;
  gap: 0.75rem;
  max-width: 85%;
}

.message-item.user {
  align-self: flex-end;
  flex-direction: row-reverse;
}

.message-item.assistant,
.message-item.system {
  align-self: flex-start;
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.message-item.user .message-avatar {
  background: var(--color-primary);
  color: var(--text-inverse);
}

.message-item.assistant .message-avatar {
  background: var(--color-success);
  color: var(--text-inverse);
}

.message-item.system .message-avatar {
  background: var(--text-muted);
  color: var(--text-inverse);
}

.message-body {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.75rem;
}

.message-role {
  font-weight: 600;
  color: var(--text-primary);
}

.message-time {
  color: var(--text-muted);
}

.message-content {
  padding: 0.75rem 1rem;
  border-radius: var(--border-radius-md);
  line-height: 1.6;
  word-break: break-word;
}

.message-item.user .message-content {
  background: var(--color-primary);
  color: var(--text-inverse);
  border-bottom-right-radius: var(--border-radius-sm);
}

.message-item.assistant .message-content,
.message-item.system .message-content {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-bottom-left-radius: var(--border-radius-sm);
  color: var(--text-primary);
}

/* Streaming cursor animation */
.streaming-cursor {
  animation: blink 1s infinite;
  color: var(--color-primary);
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

/* Markdown styles for assistant messages */
.message-content.markdown-body :deep(h1),
.message-content.markdown-body :deep(h2),
.message-content.markdown-body :deep(h3),
.message-content.markdown-body :deep(h4),
.message-content.markdown-body :deep(h5),
.message-content.markdown-body :deep(h6) {
  margin-top: 1em;
  margin-bottom: 0.5em;
  font-weight: 600;
  line-height: 1.25;
}

.message-content.markdown-body :deep(h1) { font-size: 1.5em; }
.message-content.markdown-body :deep(h2) { font-size: 1.25em; }
.message-content.markdown-body :deep(h3) { font-size: 1.1em; }

.message-content.markdown-body :deep(p) {
  margin: 0.5em 0;
}

.message-content.markdown-body :deep(ul),
.message-content.markdown-body :deep(ol) {
  margin: 0.5em 0;
  padding-left: 1.5em;
}

.message-content.markdown-body :deep(li) {
  margin: 0.25em 0;
}

.message-content.markdown-body :deep(blockquote) {
  margin: 0.5em 0;
  padding-left: 1em;
  border-left: 3px solid var(--border-color);
  color: var(--text-secondary);
}

.message-content.markdown-body :deep(code) {
  font-family: 'Fira Code', 'Consolas', monospace;
  font-size: 0.9em;
}

.message-content.markdown-body :deep(:not(pre) > code) {
  background: var(--bg-secondary);
  padding: 0.2em 0.4em;
  border-radius: var(--border-radius-sm);
  color: var(--color-error);
}

.message-content.markdown-body :deep(pre) {
  margin: 0.75em 0;
  padding: 1em;
  background: var(--bg-primary);
  border-radius: var(--border-radius);
  overflow-x: auto;
  position: relative;
  border: 1px solid var(--border-color);
}

.message-content.markdown-body :deep(pre code) {
  background: transparent;
  padding: 0;
  color: var(--text-primary);
  font-size: 0.85em;
  line-height: 1.5;
}

/* Copy button styles */
.message-content.markdown-body :deep(.copy-btn) {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  padding: 0.25rem 0.5rem;
  font-size: 0.75rem;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-sm);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all var(--transition-fast);
  min-height: 28px;
  min-width: 44px;
}

.message-content.markdown-body :deep(.copy-btn:hover) {
  background: var(--bg-hover);
  color: var(--text-primary);
}

/* Table styles */
.message-content.markdown-body :deep(table) {
  border-collapse: collapse;
  margin: 0.75em 0;
  width: 100%;
}

.message-content.markdown-body :deep(th),
.message-content.markdown-body :deep(td) {
  border: 1px solid var(--border-color);
  padding: 0.5em 0.75em;
  text-align: left;
}

.message-content.markdown-body :deep(th) {
  background: var(--bg-secondary);
  font-weight: 600;
  color: var(--text-primary);
}

/* Link styles */
.message-content.markdown-body :deep(a) {
  color: var(--color-primary);
  text-decoration: none;
}

.message-content.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

/* Horizontal rule */
.message-content.markdown-body :deep(hr) {
  margin: 1em 0;
  border: none;
  border-top: 1px solid var(--border-color);
}

/* Mobile responsive */
@media (max-width: 768px) {
  .message-item {
    max-width: 95%;
  }
}
</style>
