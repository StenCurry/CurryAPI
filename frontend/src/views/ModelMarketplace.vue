<template>
  <div class="model-marketplace">
    <div class="page-header">
      <h1 class="page-title">🤖 模型广场</h1>
      <p class="page-subtitle">浏览所有可用的 AI 模型</p>
    </div>

    <!-- Filters -->
    <div class="glass-card filters-card">
      <div class="filters-row">
        <div class="filter-group">
          <label class="filter-label">提供商</label>
          <n-select
            v-model:value="selectedProvider"
            :options="providerOptions"
            placeholder="全部提供商"
            clearable
            class="filter-select"
          />
        </div>
        <div class="filter-group">
          <label class="filter-label">标签</label>
          <n-select
            v-model:value="selectedTag"
            :options="tagOptions"
            placeholder="全部标签"
            clearable
            class="filter-select"
          />
        </div>
        <div class="filter-group">
          <label class="filter-label">端点类型</label>
          <n-select
            v-model:value="selectedEndpointType"
            :options="endpointTypeOptions"
            placeholder="全部类型"
            clearable
            class="filter-select"
          />
        </div>
        <n-button 
          type="primary" 
          @click="resetFilters"
          :disabled="!hasActiveFilters"
          class="reset-button"
        >
          重置筛选
        </n-button>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-container">
      <n-spin size="large" />
      <p class="loading-text">加载模型中...</p>
    </div>

    <!-- Error State -->
    <div v-else-if="error" class="error-container glass-card">
      <div class="error-icon">❌</div>
      <p class="error-text">{{ error }}</p>
      <n-button type="primary" @click="loadModels">重试</n-button>
    </div>

    <!-- Empty State -->
    <div v-else-if="filteredModels.length === 0" class="empty-container glass-card">
      <div class="empty-icon">🔍</div>
      <p class="empty-text">没有找到匹配的模型</p>
      <n-button type="primary" @click="resetFilters">清除筛选条件</n-button>
    </div>


    <!-- Models Grid -->
    <div v-else class="models-grid">
      <div
        v-for="model in filteredModels"
        :key="model.id"
        class="glass-card model-card"
        @click="showModelDetail(model)"
      >
        <div class="model-header">
          <div class="model-provider-badge" :class="getProviderClass(model.provider)">
            {{ model.provider }}
          </div>
          <div class="model-endpoint-badge">
            {{ model.endpoint_type }}
          </div>
        </div>
        
        <h3 class="model-name">{{ model.name }}</h3>
        
        <div class="model-tags">
          <span 
            v-for="tag in model.tags" 
            :key="tag" 
            class="model-tag"
            :class="getTagClass(tag)"
          >
            {{ tag }}
          </span>
        </div>
        
        <p class="model-description">{{ model.description }}</p>
        
        <div class="model-specs">
          <div class="spec-item">
            <span class="spec-label">上下文窗口</span>
            <span class="spec-value">{{ formatNumber(model.context_window) }}</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">最大输出</span>
            <span class="spec-value">{{ formatNumber(model.max_tokens) }}</span>
          </div>
          <div class="spec-item">
            <span class="spec-label">计费方式</span>
            <span class="spec-value">{{ formatBillingType(model.billing_type) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Model Detail Modal -->
    <n-modal v-model:show="showDetail" preset="card" class="model-detail-modal">
      <template #header>
        <div class="modal-header">
          <span class="modal-provider-badge" :class="getProviderClass(selectedModel?.provider || '')">
            {{ selectedModel?.provider }}
          </span>
          <span class="modal-title">{{ selectedModel?.name }}</span>
        </div>
      </template>
      
      <div v-if="selectedModel" class="modal-content">
        <div class="modal-tags">
          <span 
            v-for="tag in selectedModel.tags" 
            :key="tag" 
            class="model-tag"
            :class="getTagClass(tag)"
          >
            {{ tag }}
          </span>
        </div>
        
        <p class="modal-description">{{ selectedModel.description }}</p>
        
        <div class="modal-specs-grid">
          <div class="modal-spec-item">
            <div class="modal-spec-icon">📝</div>
            <div class="modal-spec-info">
              <span class="modal-spec-label">上下文窗口</span>
              <span class="modal-spec-value">{{ formatNumber(selectedModel.context_window) }} tokens</span>
            </div>
          </div>
          <div class="modal-spec-item">
            <div class="modal-spec-icon">📤</div>
            <div class="modal-spec-info">
              <span class="modal-spec-label">最大输出</span>
              <span class="modal-spec-value">{{ formatNumber(selectedModel.max_tokens) }} tokens</span>
            </div>
          </div>
          <div class="modal-spec-item">
            <div class="modal-spec-icon">💰</div>
            <div class="modal-spec-info">
              <span class="modal-spec-label">计费方式</span>
              <span class="modal-spec-value">{{ formatBillingType(selectedModel.billing_type) }}</span>
            </div>
          </div>
          <div class="modal-spec-item">
            <div class="modal-spec-icon">🔌</div>
            <div class="modal-spec-info">
              <span class="modal-spec-label">端点类型</span>
              <span class="modal-spec-value">{{ selectedModel.endpoint_type }}</span>
            </div>
          </div>
        </div>
        
        <div class="modal-model-id">
          <span class="model-id-label">模型 ID:</span>
          <code class="model-id-value">{{ selectedModel.id }}</code>
        </div>
      </div>
    </n-modal>
  </div>
</template>


<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NSpin, NSelect, NButton, NModal, useMessage } from 'naive-ui'
import { 
  getModelMarketplace, 
  type ModelMarketplaceInfo,
  type ModelFilters
} from '@/api/models'

const message = useMessage()

// State
const loading = ref(true)
const error = ref('')
const models = ref<ModelMarketplaceInfo[]>([])
const filters = ref<ModelFilters>({ providers: [], tags: [], endpoint_types: [] })

// Filter selections
const selectedProvider = ref<string | null>(null)
const selectedTag = ref<string | null>(null)
const selectedEndpointType = ref<string | null>(null)

// Modal state
const showDetail = ref(false)
const selectedModel = ref<ModelMarketplaceInfo | null>(null)

// Computed filter options
const providerOptions = computed(() => 
  filters.value.providers.map(p => ({ label: p, value: p }))
)

const tagOptions = computed(() => 
  filters.value.tags.map(t => ({ label: t, value: t }))
)

const endpointTypeOptions = computed(() => 
  filters.value.endpoint_types.map(e => ({ label: e, value: e }))
)

const hasActiveFilters = computed(() => 
  selectedProvider.value || selectedTag.value || selectedEndpointType.value
)

// Filtered models
const filteredModels = computed(() => {
  return models.value.filter(model => {
    if (selectedProvider.value && model.provider !== selectedProvider.value) {
      return false
    }
    if (selectedEndpointType.value && model.endpoint_type !== selectedEndpointType.value) {
      return false
    }
    if (selectedTag.value && !model.tags.includes(selectedTag.value)) {
      return false
    }
    return true
  })
})

// Helper functions
function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  }
  if (num >= 1000) {
    return (num / 1000).toFixed(0) + 'K'
  }
  return num.toString()
}

function formatBillingType(type: string): string {
  const typeMap: Record<string, string> = {
    'per_token': '按 Token',
    'per_request': '按请求'
  }
  return typeMap[type] || type
}

function getProviderClass(provider: string): string {
  const classMap: Record<string, string> = {
    'OpenAI': 'provider-openai',
    'Anthropic': 'provider-anthropic',
    'Google': 'provider-google',
    'DeepSeek': 'provider-deepseek',
    'Moonshot': 'provider-moonshot',
    'xAI': 'provider-xai',
    'Code Supernova': 'provider-supernova'
  }
  return classMap[provider] || 'provider-default'
}

function getTagClass(tag: string): string {
  const classMap: Record<string, string> = {
    'Fast': 'tag-fast',
    'Powerful': 'tag-powerful',
    'Code': 'tag-code',
    'Vision': 'tag-vision',
    'Multimodal': 'tag-multimodal',
    'Reasoning': 'tag-reasoning',
    'Latest': 'tag-latest',
    'Extended': 'tag-extended',
    'Premium': 'tag-premium',
    'Efficient': 'tag-efficient',
    'Lightweight': 'tag-lightweight',
    'Balanced': 'tag-balanced',
    'Preview': 'tag-preview'
  }
  return classMap[tag] || 'tag-default'
}

function showModelDetail(model: ModelMarketplaceInfo) {
  selectedModel.value = model
  showDetail.value = true
}

function resetFilters() {
  selectedProvider.value = null
  selectedTag.value = null
  selectedEndpointType.value = null
}

async function loadModels() {
  try {
    loading.value = true
    error.value = ''
    const response = await getModelMarketplace()
    models.value = response.data.models
    filters.value = response.data.filters
  } catch (err: any) {
    error.value = err.message || '加载模型列表失败'
    message.error(error.value)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadModels()
})
</script>


<style scoped>
.model-marketplace {
  padding: 2rem;
  max-width: 1400px;
  margin: 0 auto;
}

.page-header {
  text-align: center;
  margin-bottom: 2rem;
}

.page-title {
  color: var(--text-primary);
  font-size: 2rem;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
}

.page-subtitle {
  color: var(--text-secondary);
  font-size: 1.1rem;
  margin: 0;
}

.glass-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-lg);
  box-shadow: var(--shadow-md);
  padding: 1.5rem;
  transition: all var(--transition-normal);
}

/* Filters */
.filters-card {
  margin-bottom: 2rem;
}

.filters-row {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  align-items: flex-end;
}

.filter-group {
  flex: 1;
  min-width: 180px;
}

.filter-label {
  display: block;
  color: var(--text-secondary);
  font-size: 0.85rem;
  margin-bottom: 0.5rem;
}

.filter-select {
  width: 100%;
}

.reset-button {
  flex-shrink: 0;
}

/* Loading, Error, Empty States */
.loading-container,
.error-container,
.empty-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem 2rem;
  text-align: center;
}

.loading-text {
  color: var(--text-secondary);
  margin-top: 1rem;
}

.error-icon,
.empty-icon {
  font-size: 4rem;
  margin-bottom: 1rem;
}

.error-text,
.empty-text {
  color: var(--text-secondary);
  font-size: 1.1rem;
  margin-bottom: 1rem;
}

/* Models Grid */
.models-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.5rem;
}

.model-card {
  cursor: pointer;
  display: flex;
  flex-direction: column;
}

.model-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-lg);
}

.model-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.model-provider-badge,
.modal-provider-badge {
  padding: 0.25rem 0.75rem;
  border-radius: 20px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
}

.provider-openai { background: var(--color-success-light); color: var(--color-success); }
.provider-anthropic { background: var(--color-warning-light); color: var(--color-warning); }
.provider-google { background: var(--color-primary-light); color: var(--color-primary); }
.provider-deepseek { background: var(--color-info-light); color: var(--color-info); }
.provider-moonshot { background: var(--color-error-light); color: var(--color-error); }
.provider-xai { background: var(--color-error-light); color: var(--color-error); }
.provider-supernova { background: var(--color-warning-light); color: var(--color-warning); }
.provider-default { background: var(--bg-tertiary); color: var(--text-muted); }

.model-endpoint-badge {
  padding: 0.25rem 0.5rem;
  border-radius: var(--border-radius-sm);
  font-size: 0.7rem;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  text-transform: uppercase;
}

.model-name {
  color: var(--text-primary);
  font-size: 1.25rem;
  font-weight: 600;
  margin: 0 0 0.75rem 0;
}

.model-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.model-tag {
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  font-size: 0.7rem;
  font-weight: 500;
}

.tag-fast { background: rgba(34, 197, 94, 0.2); color: #22c55e; }
.tag-powerful { background: rgba(239, 68, 68, 0.2); color: #ef4444; }
.tag-code { background: rgba(99, 102, 241, 0.2); color: #6366f1; }
.tag-vision { background: rgba(236, 72, 153, 0.2); color: #ec4899; }
.tag-multimodal { background: rgba(168, 85, 247, 0.2); color: #a855f7; }
.tag-reasoning { background: rgba(245, 158, 11, 0.2); color: #f59e0b; }
.tag-latest { background: rgba(59, 130, 246, 0.2); color: #3b82f6; }
.tag-extended { background: rgba(20, 184, 166, 0.2); color: #14b8a6; }
.tag-premium { background: rgba(234, 179, 8, 0.2); color: #eab308; }
.tag-efficient { background: rgba(34, 197, 94, 0.2); color: #22c55e; }
.tag-lightweight { background: rgba(156, 163, 175, 0.2); color: #9ca3af; }
.tag-balanced { background: rgba(59, 130, 246, 0.2); color: #3b82f6; }
.tag-preview { background: rgba(249, 115, 22, 0.2); color: #f97316; }
.tag-default { background: rgba(107, 114, 128, 0.2); color: #6b7280; }

.model-description {
  color: var(--text-secondary);
  font-size: 0.9rem;
  line-height: 1.5;
  margin: 0 0 1rem 0;
  flex: 1;
}

.model-specs {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 0.5rem;
  padding-top: 1rem;
  border-top: 1px solid var(--border-color);
}

.spec-item {
  text-align: center;
}

.spec-label {
  display: block;
  color: var(--text-muted);
  font-size: 0.7rem;
  margin-bottom: 0.25rem;
}

.spec-value {
  color: var(--text-primary);
  font-size: 0.85rem;
  font-weight: 600;
}


/* Modal Styles */
:deep(.model-detail-modal) {
  max-width: 600px;
}

:deep(.model-detail-modal .n-card) {
  background: var(--bg-card) !important;
  border: 1px solid var(--border-color) !important;
  border-radius: var(--border-radius-lg) !important;
}

:deep(.model-detail-modal .n-card-header) {
  border-bottom: 1px solid var(--border-color) !important;
}

:deep(.model-detail-modal .n-card-header__main) {
  color: var(--text-primary) !important;
}

.modal-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.modal-title {
  color: var(--text-primary);
  font-size: 1.25rem;
  font-weight: 600;
}

.modal-content {
  padding: 0.5rem 0;
}

.modal-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.modal-description {
  color: var(--text-secondary);
  font-size: 1rem;
  line-height: 1.6;
  margin: 0 0 1.5rem 0;
}

.modal-specs-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.modal-spec-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
  border: 1px solid var(--border-color);
}

.modal-spec-icon {
  font-size: 1.5rem;
}

.modal-spec-info {
  display: flex;
  flex-direction: column;
}

.modal-spec-label {
  color: var(--text-muted);
  font-size: 0.75rem;
}

.modal-spec-value {
  color: var(--text-primary);
  font-size: 0.95rem;
  font-weight: 600;
}

.modal-model-id {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 1rem;
  background: var(--bg-secondary);
  border-radius: var(--border-radius);
}

.model-id-label {
  color: var(--text-muted);
  font-size: 0.85rem;
}

.model-id-value {
  color: var(--color-primary);
  font-family: monospace;
  font-size: 0.9rem;
  background: var(--color-primary-light);
  padding: 0.25rem 0.5rem;
  border-radius: var(--border-radius-sm);
}

/* Responsive */
@media (max-width: 768px) {
  .model-marketplace {
    padding: 1rem;
  }
  
  .page-title {
    font-size: 1.5rem;
  }
  
  .filters-row {
    flex-direction: column;
  }
  
  .filter-group {
    width: 100%;
  }
  
  .models-grid {
    grid-template-columns: 1fr;
  }
  
  .model-specs {
    grid-template-columns: 1fr;
    gap: 0.75rem;
  }
  
  .spec-item {
    display: flex;
    justify-content: space-between;
    text-align: left;
  }
  
  .modal-specs-grid {
    grid-template-columns: 1fr;
  }
}
</style>
