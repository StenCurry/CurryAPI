<template>
  <div class="model-selector">
    <n-select
      :value="modelValue"
      :options="modelOptions"
      :loading="loading"
      :disabled="disabled"
      size="small"
      filterable
      placeholder="选择模型"
      :consistent-menu-width="false"
      :render-label="renderLabel"
      :render-tag="renderTag"
      @update:value="handleModelChange"
    />
  </div>
</template>

<script setup lang="ts">
/**
 * ModelSelector.vue - Model selection dropdown component
 * 模型选择下拉组件
 * 
 * Displays provider name, pricing info, and indicates unavailable models.
 * Allows free model switching within a conversation.
 * 
 * Requirements: 3.1, 3.2, 3.3, 11.1-11.5
 */

import { computed, h } from 'vue'
import { NTag, NText } from 'naive-ui'
import type { SelectOption, SelectRenderLabel, SelectRenderTag } from 'naive-ui'
import type { ChatModel } from '@/api/chat'

// ============================================================================
// Props
// ============================================================================

interface Props {
  /** Currently selected model ID */
  modelValue: string
  /** Available models list */
  models: ChatModel[]
  /** Whether the selector is loading */
  loading?: boolean
  /** Whether the selector is disabled */
  disabled?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  disabled: false
})

// ============================================================================
// Emits
// ============================================================================

const emit = defineEmits<{
  /** Emitted when model selection changes */
  (e: 'update:modelValue', value: string): void
  /** Emitted when model is changed */
  (e: 'change', model: ChatModel): void
}>()

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Get provider display color
 * 获取提供商显示颜色
 */
function getProviderColor(provider: string): string {
  if (!provider) return '#6b7280'
  const colors: Record<string, string> = {
    openai: '#10a37f',
    anthropic: '#d97706',
    google: '#4285f4',
    deepseek: '#6366f1',
    cursor: '#8b5cf6',
    'openrouter-free': '#22c55e' // 免费模型使用绿色
  }
  return colors[provider.toLowerCase()] || '#6b7280'
}

/**
 * Format price for display
 * 格式化价格显示
 */
function formatPrice(price: number | undefined): string {
  if (price === undefined || price === null) return '-'
  if (price === 0) return '免费'
  if (price < 1) return `$${price.toFixed(3)}`
  return `$${price.toFixed(2)}`
}

/**
 * Get provider short name for tag display
 * 获取提供商简称用于标签显示
 */
function getProviderShortName(provider: string): string {
  if (!provider) return 'AI'
  const shortNames: Record<string, string> = {
    openai: 'GPT',
    anthropic: 'Claude',
    google: 'Gemini',
    deepseek: 'DeepSeek',
    cursor: 'Cursor',
    'openrouter-free': '🆓 免费'
  }
  return shortNames[provider.toLowerCase()] || provider
}

// ============================================================================
// Computed
// ============================================================================

/** Model options for the select - simple flat list without grouping */
const modelOptions = computed(() => {
  if (!props.models || !Array.isArray(props.models) || props.models.length === 0) {
    return []
  }
  
  return props.models
    .filter(model => model && model.id)
    .map(model => ({
      label: model.name || model.id,
      value: model.id,
      disabled: model.is_available === false,
      model: model
    }))
})

// ============================================================================
// Custom Render Functions
// ============================================================================

/**
 * Custom render function for the selected tag (显示在选择框中的内容)
 */
const renderTag: SelectRenderTag = ({ option }) => {
  if (!option) return ''
  
  const model = (option as SelectOption & { model?: ChatModel }).model
  if (!model) return String(option.label || option.value || '')
  
  return h(
    'div',
    {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: '6px',
        padding: '2px 0'
      }
    },
    [
      // Model name
      h('span', { 
        style: { 
          fontWeight: 500,
          fontSize: '13px',
          color: '#374151'
        } 
      }, model.name || model.id),
      // Provider badge (small)
      h(
        'span',
        {
          style: {
            fontSize: '10px',
            padding: '1px 6px',
            borderRadius: '4px',
            backgroundColor: getProviderColor(model.provider) + '15',
            color: getProviderColor(model.provider),
            fontWeight: 500
          }
        },
        getProviderShortName(model.provider)
      )
    ]
  )
}

/**
 * Custom render function for dropdown options
 * 自定义下拉选项渲染函数
 */
const renderLabel: SelectRenderLabel = (option) => {
  if (!option) return ''
  
  if (option.type === 'group') {
    return String(option.label || '')
  }
  
  const model = (option as SelectOption & { model?: ChatModel }).model
  
  if (!model) {
    return String(option.label || option.value || '')
  }
  
  const isUnavailable = model.is_available === false
  
  return h(
    'div',
    {
      style: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        padding: '6px 0',
        opacity: isUnavailable ? 0.5 : 1,
        width: '100%'
      }
    },
    [
      // Left: Model info
      h(
        'div',
        {
          style: {
            display: 'flex',
            flexDirection: 'column',
            gap: '2px'
          }
        },
        [
          // Model name with provider tag
          h(
            'div',
            {
              style: {
                display: 'flex',
                alignItems: 'center',
                gap: '8px'
              }
            },
            [
              h('span', { 
                style: { 
                  fontWeight: 500,
                  fontSize: '13px'
                } 
              }, model.name || model.id),
              h(
                'span',
                {
                  style: {
                    fontSize: '10px',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    backgroundColor: getProviderColor(model.provider) + '15',
                    color: getProviderColor(model.provider),
                    fontWeight: 500
                  }
                },
                getProviderShortName(model.provider)
              ),
              isUnavailable
                ? h(
                    NTag,
                    {
                      size: 'tiny',
                      type: 'error',
                      bordered: false
                    },
                    { default: () => '不可用' }
                  )
                : null
            ]
          ),
          // Price info
          (model.input_price !== undefined || model.output_price !== undefined)
            ? h(
                'div',
                {
                  style: {
                    fontSize: '11px',
                    color: '#9ca3af'
                  }
                },
                `${formatPrice(model.input_price)} / ${formatPrice(model.output_price)} per 1M tokens`
              )
            : null
        ]
      ),
      // Right: Context window (if available)
      model.context_window
        ? h(
            'span',
            {
              style: {
                fontSize: '10px',
                color: '#9ca3af',
                marginLeft: '12px',
                whiteSpace: 'nowrap'
              }
            },
            `${Math.round(model.context_window / 1000)}K`
          )
        : null
    ]
  )
}

// ============================================================================
// Methods
// ============================================================================

/**
 * Handle model change
 * 处理模型切换
 */
function handleModelChange(modelId: string) {
  if (!props.models || !Array.isArray(props.models) || props.models.length === 0) {
    return
  }
  
  const model = props.models.find(m => m && m.id === modelId)
  if (model) {
    if (model.is_available === false) {
      return
    }
    emit('update:modelValue', modelId)
    emit('change', model)
  }
}
</script>

<style scoped>
.model-selector {
  display: inline-flex;
  align-items: center;
  min-width: 200px;
  max-width: 320px;
}

.model-selector :deep(.n-select) {
  --n-border: 1px solid var(--border-color);
  --n-border-hover: 1px solid var(--color-primary);
  --n-border-active: 1px solid var(--color-primary);
  --n-border-focus: 1px solid var(--color-primary);
}

.model-selector :deep(.n-base-selection) {
  background: var(--bg-card);
  border-radius: var(--border-radius);
  box-shadow: var(--shadow-xs);
  min-height: 36px;
}

.model-selector :deep(.n-base-selection:hover) {
  background: var(--bg-hover);
}

.model-selector :deep(.n-base-selection-label) {
  padding: 0 8px;
}

/* Dropdown menu styling */
.model-selector :deep(.n-base-select-menu) {
  min-width: 340px;
  border-radius: var(--border-radius-md);
  box-shadow: var(--shadow-lg);
  background: var(--bg-card);
}

.model-selector :deep(.n-base-select-option) {
  padding: 8px 12px;
  min-height: 44px;
}

.model-selector :deep(.n-base-select-option--selected) {
  background: var(--color-primary-light);
}

.model-selector :deep(.n-base-select-option:hover) {
  background: var(--bg-hover);
}

/* Mobile responsive */
@media (max-width: 768px) {
  .model-selector {
    min-width: 160px;
    max-width: 240px;
  }
  
  .model-selector :deep(.n-base-select-menu) {
    min-width: 280px;
  }

  /* Ensure touch-friendly size */
  .model-selector :deep(.n-base-selection) {
    min-height: 44px;
  }
}
</style>
