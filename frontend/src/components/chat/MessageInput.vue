<template>
  <div class="message-input">
    <div class="input-wrapper">
      <n-input
        ref="inputRef"
        v-model:value="inputValue"
        type="textarea"
        :placeholder="placeholder"
        :autosize="{ minRows: 1, maxRows: 6 }"
        :disabled="disabled || isStreaming"
        @keydown="handleKeydown"
        @input="handleInput"
      />
    </div>
    <div class="input-actions">
      <n-button
        type="primary"
        :loading="isStreaming"
        :disabled="!canSend"
        @click="handleSend"
      >
        <template #icon>
          <n-icon><SendOutline /></n-icon>
        </template>
      </n-button>
    </div>
  </div>
</template>

<script setup lang="ts">
/**
 * MessageInput.vue - Message input component
 * 消息输入组件
 * 
 * Text input with auto-resize, send button, and keyboard shortcuts.
 * Enter to send, Shift+Enter for newline.
 * Disabled during streaming.
 * 
 * Requirements: 5.2, 5.3
 */

import { ref, computed, watch, nextTick } from 'vue'
import { SendOutline } from '@vicons/ionicons5'

// ============================================================================
// Props
// ============================================================================

interface Props {
  /** Initial value for the input */
  modelValue?: string
  /** Placeholder text */
  placeholder?: string
  /** Whether the input is disabled */
  disabled?: boolean
  /** Whether AI is currently streaming a response */
  isStreaming?: boolean
  /** Maximum character length */
  maxLength?: number
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: '',
  placeholder: '输入消息... (Enter 发送, Shift+Enter 换行)',
  disabled: false,
  isStreaming: false,
  maxLength: 10000
})

// ============================================================================
// Emits
// ============================================================================

const emit = defineEmits<{
  /** Emitted when input value changes */
  (e: 'update:modelValue', value: string): void
  /** Emitted when user sends a message */
  (e: 'send', content: string): void
}>()

// ============================================================================
// Refs
// ============================================================================

const inputRef = ref<InstanceType<typeof import('naive-ui').NInput> | null>(null)
const inputValue = ref(props.modelValue)

// ============================================================================
// Computed
// ============================================================================

/** Whether the send button should be enabled */
const canSend = computed(() => {
  return inputValue.value.trim().length > 0 && 
         !props.disabled && 
         !props.isStreaming &&
         inputValue.value.length <= props.maxLength
})

// ============================================================================
// Methods
// ============================================================================

/**
 * Handle keyboard events
 * Enter to send, Shift+Enter for newline
 */
function handleKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    handleSend()
  }
}

/**
 * Handle input changes
 */
function handleInput() {
  emit('update:modelValue', inputValue.value)
}

/**
 * Handle send button click
 */
function handleSend() {
  if (!canSend.value) return
  
  const content = inputValue.value.trim()
  
  // Clear input before emitting to avoid race conditions
  inputValue.value = ''
  emit('update:modelValue', '')
  
  // Emit send event
  emit('send', content)
  
  // Focus back on input after DOM update
  nextTick(() => {
    focus()
  })
}

/**
 * Focus the input element
 */
function focus() {
  inputRef.value?.focus()
}

/**
 * Clear the input
 */
function clear() {
  inputValue.value = ''
  emit('update:modelValue', '')
}

// ============================================================================
// Watchers
// ============================================================================

// Sync with external modelValue changes
watch(
  () => props.modelValue,
  (newValue) => {
    if (newValue !== inputValue.value) {
      inputValue.value = newValue
    }
  }
)

// ============================================================================
// Expose
// ============================================================================

defineExpose({
  focus,
  clear
})
</script>

<style scoped>
.message-input {
  padding: 1rem;
  border-top: 1px solid var(--border-color);
  display: flex;
  gap: 0.5rem;
  align-items: flex-end;
  background: var(--bg-card);
}

.input-wrapper {
  flex: 1;
  min-width: 0;
}

.input-wrapper :deep(.n-input) {
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
}

.input-wrapper :deep(.n-input__textarea-el) {
  resize: none;
  line-height: 1.5;
  color: var(--text-primary);
}

.input-actions {
  flex-shrink: 0;
  display: flex;
  align-items: flex-end;
}

/* Ensure send button meets minimum touch target size */
.input-actions .n-button {
  height: 44px;
  width: 44px;
  min-height: 44px;
  min-width: 44px;
}

/* Disabled state styling */
.message-input.disabled {
  opacity: 0.7;
}

/* Mobile responsive */
@media (max-width: 768px) {
  .message-input {
    padding: 0.75rem;
  }

  .input-wrapper :deep(.n-input) {
    font-size: 16px; /* Prevent iOS zoom on focus */
  }
}

/* Touch device optimizations */
@media (hover: none) and (pointer: coarse) {
  .input-actions .n-button:active {
    transform: scale(0.95);
  }
}
</style>
