<template>
  <div class="token-management">
    <!-- Header Section -->
    <div class="page-header glass-card">
      <div class="header-content">
        <div class="header-title">
          <div class="icon-wrapper">
            <n-icon size="40" class="header-icon">
              <KeyOutline />
            </n-icon>
          </div>
          <div class="title-text">
            <n-h2 style="margin: 0">API 令牌管理</n-h2>
            <n-text depth="3">安全管理您的 API 访问令牌</n-text>
          </div>
        </div>
        <n-space class="header-actions">
          <n-button 
            type="primary" 
            size="large"
            @click="handleAutoGenerateToken"
            class="action-button primary-button glass-button"
          >
            <template #icon>
              <n-icon><FlashOutline /></n-icon>
            </template>
            <span class="button-text">自动生成</span>
          </n-button>
          <n-button 
            type="info" 
            size="large"
            @click="openAddDialog"
            class="action-button secondary-button glass-button"
          >
            <template #icon>
              <n-icon><CreateOutline /></n-icon>
            </template>
            <span class="button-text">自定义令牌</span>
          </n-button>
        </n-space>
      </div>
    </div>

    <!-- Stats Cards with 3D Effect -->
    <n-grid :x-gap="20" :y-gap="20" :cols="3" class="stats-grid">
      <n-grid-item>
        <div class="stat-card-wrapper">
          <n-card :bordered="false" class="stat-card stat-card-1 glass-card">
            <div class="stat-icon-wrapper icon-purple">
              <n-icon size="32">
                <KeyOutline />
              </n-icon>
            </div>
            <div class="stat-content">
              <div class="stat-label">总令牌数</div>
              <div class="stat-value">{{ tokens.length }}</div>
              <div class="stat-trend">
                <n-icon size="16" color="#10b981"><TrendingUpOutline /></n-icon>
                <span>活跃管理中</span>
              </div>
            </div>
          </n-card>
        </div>
      </n-grid-item>
      <n-grid-item>
        <div class="stat-card-wrapper">
          <n-card :bordered="false" class="stat-card stat-card-2 glass-card">
            <div class="stat-icon-wrapper icon-green">
              <n-icon size="32">
                <StatsChartOutline />
              </n-icon>
            </div>
            <div class="stat-content">
              <div class="stat-label">总使用次数</div>
              <div class="stat-value">{{ totalUsage.toLocaleString() }}</div>
              <div class="stat-trend">
                <n-icon size="16" color="#10b981"><TrendingUpOutline /></n-icon>
                <span>持续增长</span>
              </div>
            </div>
          </n-card>
        </div>
      </n-grid-item>
      <n-grid-item>
        <div class="stat-card-wrapper">
          <n-card :bordered="false" class="stat-card stat-card-3 glass-card">
            <div class="stat-icon-wrapper icon-orange">
              <n-icon size="32">
                <CheckmarkCircleOutline />
              </n-icon>
            </div>
            <div class="stat-content">
              <div class="stat-label">活跃令牌</div>
              <div class="stat-value">{{ activeTokens }}</div>
              <div class="stat-trend">
                <n-icon size="16" color="#10b981"><TrendingUpOutline /></n-icon>
                <span>运行正常</span>
              </div>
            </div>
          </n-card>
        </div>
      </n-grid-item>
    </n-grid>

    <!-- Tokens Table with Glass Effect -->
    <n-card :bordered="false" class="tokens-card glass-card">
      <template #header>
        <div class="table-header">
          <div class="table-title">
            <n-icon size="20" color="#667eea"><ListOutline /></n-icon>
            <n-text strong style="font-size: 18px; margin-left: 8px">令牌列表</n-text>
          </div>
          <n-tag v-if="tokens.length > 0" type="info" round class="count-badge">
            <template #icon>
              <n-icon><CheckmarkCircleOutline /></n-icon>
            </template>
            {{ tokens.length }} 个令牌
          </n-tag>
        </div>
      </template>
      <n-spin :show="loading">
        <div v-if="!loading && tokens.length === 0" class="empty-state-modern">
          <div class="empty-icon-wrapper">
            <n-icon size="80" color="#d1d5db">
              <KeyOutline />
            </n-icon>
          </div>
          <n-h3 style="margin: 16px 0 8px">暂无令牌</n-h3>
          <n-text depth="3">开始创建您的第一个 API 令牌</n-text>
          <n-space style="margin-top: 24px">
            <n-button type="primary" size="large" @click="handleAutoGenerateToken" class="glass-button">
              <template #icon>
                <n-icon><FlashOutline /></n-icon>
              </template>
              快速生成令牌
            </n-button>
            <n-button size="large" @click="openAddDialog" class="glass-button">
              <template #icon>
                <n-icon><CreateOutline /></n-icon>
              </template>
              自定义令牌
            </n-button>
          </n-space>
        </div>
        <n-data-table 
          v-else 
          :columns="columns" 
          :data="tokens" 
          :pagination="false"
          :bordered="false"
          class="tokens-table modern-table"
        />
      </n-spin>
    </n-card>

    <!-- 自定义令牌对话框 -->
    <n-modal v-model:show="showAddDialog" preset="card" title="自定义令牌" style="width: 600px;">
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" label-width="auto">
        <n-form-item label="令牌名称" path="tokenName">
          <n-input 
            v-model:value="formData.tokenName" 
            placeholder="为令牌设置一个易于识别的名称" 
            type="text"
          />
        </n-form-item>
        <n-form-item label="自定义令牌" path="tokenKey">
          <n-input 
            v-model:value="formData.tokenKey" 
            placeholder="请输入自定义令牌（例如：sk-mytoken123）" 
            type="text"
          />
        </n-form-item>
        
        <n-divider>额度与有效期设置</n-divider>
        
        <!-- Quota Limit -->
        <n-form-item label="额度限制" path="quotaType">
          <n-space vertical style="width: 100%">
            <n-radio-group v-model:value="formData.quotaType">
              <n-space>
                <n-radio value="unlimited">无限制</n-radio>
                <n-radio value="limited">设置额度</n-radio>
              </n-space>
            </n-radio-group>
            <n-input-number 
              v-if="formData.quotaType === 'limited'"
              v-model:value="formData.quotaLimit" 
              :min="0.01"
              :precision="2"
              placeholder="输入额度限制（美元）"
              style="width: 100%"
            >
              <template #prefix>$</template>
            </n-input-number>
          </n-space>
        </n-form-item>
        
        <!-- Expiration Date -->
        <n-form-item label="有效期" path="expirationType">
          <n-space vertical style="width: 100%">
            <n-radio-group v-model:value="formData.expirationType">
              <n-space>
                <n-radio value="never">永不过期</n-radio>
                <n-radio value="date">设置过期日期</n-radio>
              </n-space>
            </n-radio-group>
            <n-date-picker 
              v-if="formData.expirationType === 'date'"
              v-model:value="formData.expiresAt" 
              type="datetime"
              placeholder="选择过期时间"
              style="width: 100%"
              :is-date-disabled="(ts: number) => ts < Date.now()"
            />
          </n-space>
        </n-form-item>
        
        <!-- Model Restrictions -->
        <n-form-item label="模型限制" path="modelRestriction">
          <n-space vertical style="width: 100%">
            <n-radio-group v-model:value="formData.modelRestriction">
              <n-space>
                <n-radio value="all">所有模型</n-radio>
                <n-radio value="selected">指定模型</n-radio>
              </n-space>
            </n-radio-group>
            <n-select
              v-if="formData.modelRestriction === 'selected'"
              v-model:value="formData.allowedModels"
              multiple
              :options="modelOptions"
              placeholder="选择允许访问的模型"
              style="width: 100%"
              filterable
            />
          </n-space>
        </n-form-item>
        
        <n-alert type="info" style="margin-top: 12px;">
          建议使用 sk- 开头的格式，例如：sk-mytoken123
        </n-alert>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showAddDialog = false">取消</n-button>
          <n-button type="primary" @click="handleAddToken" :loading="adding">确定</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 自动生成令牌设置对话框 -->
    <n-modal v-model:show="showAutoGenerateDialog" preset="card" title="自动生成令牌" style="width: 600px;">
      <n-form :model="autoGenFormData" label-placement="left" label-width="auto">
        <n-form-item label="令牌名称">
          <n-input 
            v-model:value="autoGenFormData.tokenName" 
            placeholder="为令牌设置一个易于识别的名称（可选）" 
            type="text"
          />
        </n-form-item>
        
        <n-divider>额度与有效期设置</n-divider>
        
        <!-- Quota Limit -->
        <n-form-item label="额度限制">
          <n-space vertical style="width: 100%">
            <n-radio-group v-model:value="autoGenFormData.quotaType">
              <n-space>
                <n-radio value="unlimited">无限制</n-radio>
                <n-radio value="limited">设置额度</n-radio>
              </n-space>
            </n-radio-group>
            <n-input-number 
              v-if="autoGenFormData.quotaType === 'limited'"
              v-model:value="autoGenFormData.quotaLimit" 
              :min="0.01"
              :precision="2"
              placeholder="输入额度限制（美元）"
              style="width: 100%"
            >
              <template #prefix>$</template>
            </n-input-number>
          </n-space>
        </n-form-item>
        
        <!-- Expiration Date -->
        <n-form-item label="有效期">
          <n-space vertical style="width: 100%">
            <n-radio-group v-model:value="autoGenFormData.expirationType">
              <n-space>
                <n-radio value="never">永不过期</n-radio>
                <n-radio value="date">设置过期日期</n-radio>
              </n-space>
            </n-radio-group>
            <n-date-picker 
              v-if="autoGenFormData.expirationType === 'date'"
              v-model:value="autoGenFormData.expiresAt" 
              type="datetime"
              placeholder="选择过期时间"
              style="width: 100%"
              :is-date-disabled="(ts: number) => ts < Date.now()"
            />
          </n-space>
        </n-form-item>
        
        <!-- Model Restrictions -->
        <n-form-item label="模型限制">
          <n-space vertical style="width: 100%">
            <n-radio-group v-model:value="autoGenFormData.modelRestriction">
              <n-space>
                <n-radio value="all">所有模型</n-radio>
                <n-radio value="selected">指定模型</n-radio>
              </n-space>
            </n-radio-group>
            <n-select
              v-if="autoGenFormData.modelRestriction === 'selected'"
              v-model:value="autoGenFormData.allowedModels"
              multiple
              :options="modelOptions"
              placeholder="选择允许访问的模型"
              style="width: 100%"
              filterable
            />
          </n-space>
        </n-form-item>
      </n-form>
      <template #footer>
        <n-space justify="end">
          <n-button @click="showAutoGenerateDialog = false">取消</n-button>
          <n-button type="primary" @click="confirmAutoGenerateToken" :loading="adding">生成令牌</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 自动生成令牌确认对话框 -->
    <n-modal v-model:show="showGeneratedDialog" preset="dialog" title="令牌生成成功">
      <n-space vertical>
        <n-alert type="success" title="令牌已生成">
          请妥善保管您的令牌，关闭后将无法再次查看完整令牌
        </n-alert>
        <n-form-item label="生成的令牌">
          <n-input 
            :value="generatedToken" 
            type="textarea" 
            :rows="3" 
            readonly
          />
        </n-form-item>
      </n-space>
      <template #action>
        <n-space>
          <n-button @click="handleCopyGenerated">
            <template #icon>
              <n-icon><CopyOutline /></n-icon>
            </template>
            复制令牌
          </n-button>
          <n-button type="primary" @click="showGeneratedDialog = false">关闭</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- Token Details Modal -->
    <n-modal v-model:show="showDetailsModal" preset="card" title="令牌详情" style="width: 500px;">
      <template v-if="selectedToken">
        <n-descriptions :column="1" bordered>
          <n-descriptions-item label="令牌名称">
            {{ selectedToken.token_name || '未命名' }}
          </n-descriptions-item>
          <n-descriptions-item label="令牌">
            <code>{{ selectedToken.masked_key }}</code>
          </n-descriptions-item>
          <n-descriptions-item label="状态">
            <n-tag :type="selectedToken.is_active ? 'success' : 'default'" size="small">
              {{ selectedToken.is_active ? '活跃' : '禁用' }}
            </n-tag>
          </n-descriptions-item>
          <n-descriptions-item label="额度限制">
            <template v-if="selectedToken.quota_limit != null">
              ${{ selectedToken.quota_used?.toFixed(2) || '0.00' }} / ${{ selectedToken.quota_limit.toFixed(2) }}
              <n-progress 
                type="line" 
                :percentage="getQuotaPercentage(selectedToken)" 
                :status="getQuotaStatus(selectedToken)"
                style="margin-top: 4px"
              />
            </template>
            <span v-else class="unlimited-badge">无限制</span>
          </n-descriptions-item>
          <n-descriptions-item label="有效期">
            <template v-if="selectedToken.expires_at">
              <n-tag :type="isExpired(selectedToken.expires_at) ? 'error' : 'info'" size="small">
                {{ isExpired(selectedToken.expires_at) ? '已过期' : '有效' }}
              </n-tag>
              {{ formatDate(selectedToken.expires_at) }}
            </template>
            <span v-else class="unlimited-badge">永不过期</span>
          </n-descriptions-item>
          <n-descriptions-item label="允许的模型">
            <template v-if="selectedToken.allowed_models && selectedToken.allowed_models.length > 0">
              <n-space>
                <n-tag v-for="model in selectedToken.allowed_models" :key="model" size="small" type="info">
                  {{ model }}
                </n-tag>
              </n-space>
            </template>
            <span v-else class="unlimited-badge">所有模型</span>
          </n-descriptions-item>
          <n-descriptions-item label="使用次数">
            {{ selectedToken.usage_count || 0 }}
          </n-descriptions-item>
          <n-descriptions-item label="创建时间">
            {{ formatDate(selectedToken.created_at) }}
          </n-descriptions-item>
          <n-descriptions-item label="最后使用">
            {{ selectedToken.last_used_at ? formatDate(selectedToken.last_used_at) : '从未使用' }}
          </n-descriptions-item>
        </n-descriptions>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, h, onMounted, computed } from 'vue'
import { NButton, NIcon, useMessage, useDialog, NTag, NProgress, NTooltip } from 'naive-ui'
import { 
  FlashOutline, 
  CreateOutline, 
  CopyOutline, 
  KeyOutline,
  StatsChartOutline,
  CheckmarkCircleOutline,
  TrashOutline,
  TrendingUpOutline,
  ListOutline,
  InformationCircleOutline
} from '@vicons/ionicons5'
import { listKeys, addKey, removeKey, updateKeyName } from '@/api/admin'
import type { AdminKey } from '@/types'

const message = useMessage()
const dialog = useDialog()
const tokens = ref<AdminKey[]>([])
const showAddDialog = ref(false)
const showGeneratedDialog = ref(false)
const showAutoGenerateDialog = ref(false)
const showDetailsModal = ref(false)
const selectedToken = ref<AdminKey | null>(null)
const generatedToken = ref('')
const loading = ref(false)
const adding = ref(false)
const editingTokenKey = ref<string | null>(null)
const editingTokenName = ref('')

// Form data for new token
const formData = ref({
  tokenName: '',
  tokenKey: '',
  quotaType: 'unlimited' as 'unlimited' | 'limited',
  quotaLimit: 10 as number,
  expirationType: 'never' as 'never' | 'date',
  expiresAt: null as number | null,
  modelRestriction: 'all' as 'all' | 'selected',
  allowedModels: [] as string[]
})

const formRules = {
  tokenKey: {
    required: true,
    message: '请输入令牌',
    trigger: 'blur'
  }
}

// Form data for auto-generated token
const autoGenFormData = ref({
  tokenName: '',
  quotaType: 'unlimited' as 'unlimited' | 'limited',
  quotaLimit: 10 as number,
  expirationType: 'never' as 'never' | 'date',
  expiresAt: null as number | null,
  modelRestriction: 'all' as 'all' | 'selected',
  allowedModels: [] as string[]
})

// Available models for selection
const modelOptions = [
  { label: 'GPT-5', value: 'gpt-5' },
  { label: 'GPT-5.1', value: 'gpt-5.1' },
  { label: 'GPT-5 Codex', value: 'gpt-5-codex' },
  { label: 'GPT-5.1 Codex', value: 'gpt-5.1-codex' },
  { label: 'GPT-5.1 Codex Max', value: 'gpt-5.1-codex-max' },
  { label: 'GPT-5 Mini', value: 'gpt-5-mini' },
  { label: 'GPT-5 Nano', value: 'gpt-5-nano' },
  { label: 'GPT-4.1', value: 'gpt-4.1' },
  { label: 'GPT-4o', value: 'gpt-4o' },
  { label: 'Claude 3.5 Sonnet', value: 'claude-3.5-sonnet' },
  { label: 'Claude 3.5 Haiku', value: 'claude-3.5-haiku' },
  { label: 'Claude 3.7 Sonnet', value: 'claude-3.7-sonnet' },
  { label: 'Claude 4 Sonnet', value: 'claude-4-sonnet' },
  { label: 'Claude 4.5 Sonnet', value: 'claude-4.5-sonnet' },
  { label: 'Claude 4 Opus', value: 'claude-4-opus' },
  { label: 'Claude 4.1 Opus', value: 'claude-4.1-opus' },
  { label: 'Claude 4.5 Opus', value: 'claude-4.5-opus' },
  { label: 'Claude 4.5 Haiku', value: 'claude-4.5-haiku' },
  { label: 'Claude Code 1M', value: 'claude-code-1m' },
  { label: 'Gemini 2.5 Pro', value: 'gemini-2.5-pro' },
  { label: 'Gemini 2.5 Flash', value: 'gemini-2.5-flash' },
  { label: 'Gemini 3 Pro Preview', value: 'gemini-3-pro-preview' },
  { label: 'O3', value: 'o3' },
  { label: 'O4 Mini', value: 'o4-mini' },
  { label: 'DeepSeek R1', value: 'deepseek-r1' },
  { label: 'DeepSeek V3.1', value: 'deepseek-v3.1' },
  { label: 'Kimi K2 Instruct', value: 'kimi-k2-instruct' },
  { label: 'Grok 3', value: 'grok-3' },
  { label: 'Grok 3 Mini', value: 'grok-3-mini' },
  { label: 'Grok 4', value: 'grok-4' },
  { label: 'Code Supernova 1M', value: 'code-supernova-1-million' }
]

// Computed statistics
const totalUsage = computed(() => {
  return tokens.value.reduce((sum, token) => sum + (token.usage_count || 0), 0)
})

const activeTokens = computed(() => {
  return tokens.value.filter(token => token.is_active).length
})

// Helper functions
function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('zh-CN')
}

function isExpired(dateStr: string): boolean {
  return new Date(dateStr) < new Date()
}

function getQuotaPercentage(token: AdminKey): number {
  if (token.quota_limit == null || token.quota_limit === 0) return 0
  return Math.min(100, ((token.quota_used || 0) / token.quota_limit) * 100)
}

function getQuotaStatus(token: AdminKey): 'success' | 'warning' | 'error' {
  const percentage = getQuotaPercentage(token)
  if (percentage >= 100) return 'error'
  if (percentage >= 80) return 'warning'
  return 'success'
}

function openAddDialog() {
  // Reset form data
  formData.value = {
    tokenName: '',
    tokenKey: '',
    quotaType: 'unlimited',
    quotaLimit: 10,
    expirationType: 'never',
    expiresAt: null,
    modelRestriction: 'all',
    allowedModels: []
  }
  showAddDialog.value = true
}

function showTokenDetails(token: AdminKey) {
  selectedToken.value = token
  showDetailsModal.value = true
}

const columns = [
  { 
    title: '令牌名称', 
    key: 'token_name',
    width: 150,
    render: (row: AdminKey) => {
      const isEditing = editingTokenKey.value === row.key
      if (isEditing) {
        return h('div', { class: 'name-edit-cell' }, [
          h('input', {
            class: 'name-edit-input',
            value: editingTokenName.value,
            onInput: (e: any) => { editingTokenName.value = e.target.value },
            onKeyup: (e: KeyboardEvent) => {
              if (e.key === 'Enter') handleSaveTokenName(row.key)
              if (e.key === 'Escape') handleCancelEdit()
            },
            placeholder: '输入名称'
          }),
          h('div', { class: 'name-edit-actions' }, [
            h(NButton, {
              size: 'tiny',
              type: 'primary',
              onClick: () => handleSaveTokenName(row.key)
            }, { default: () => '保存' }),
            h(NButton, {
              size: 'tiny',
              onClick: handleCancelEdit
            }, { default: () => '取消' })
          ])
        ])
      }
      return h('div', { 
        class: 'name-cell',
        onClick: () => handleEditTokenName(row.key, row.token_name || '')
      }, [
        h('span', { class: 'name-text' }, row.token_name || '未命名'),
        h(NIcon, { size: 14, class: 'edit-icon' }, { default: () => h(CreateOutline) })
      ])
    }
  },
  { 
    title: '令牌', 
    key: 'masked_key',
    width: 200,
    render: (row: AdminKey) => {
      return h('div', { class: 'token-cell' }, [
        h('code', { class: 'token-code' }, row.masked_key)
      ])
    }
  },
  { 
    title: '状态', 
    key: 'is_active',
    width: 80,
    render: (row: AdminKey) => {
      return h(
        NTag,
        {
          type: row.is_active ? 'success' : 'default',
          size: 'small',
          round: true
        },
        { default: () => row.is_active ? '活跃' : '禁用' }
      )
    }
  },

  { 
    title: '额度', 
    key: 'quota',
    width: 150,
    render: (row: AdminKey) => {
      if (row.quota_limit == null) {
        return h('span', { class: 'unlimited-badge' }, '无限制')
      }
      const percentage = getQuotaPercentage(row)
      const status = getQuotaStatus(row)
      return h('div', { class: 'quota-cell' }, [
        h('div', { class: 'quota-text' }, `$${(row.quota_used || 0).toFixed(2)} / $${row.quota_limit.toFixed(2)}`),
        h(NProgress, {
          type: 'line',
          percentage: percentage,
          status: status,
          showIndicator: false,
          height: 4,
          style: { marginTop: '4px' }
        })
      ])
    }
  },
  { 
    title: '有效期', 
    key: 'expires_at',
    width: 140,
    render: (row: AdminKey) => {
      if (!row.expires_at) {
        return h('span', { class: 'unlimited-badge' }, '永不过期')
      }
      const expired = isExpired(row.expires_at)
      return h('div', { class: 'expiry-cell' }, [
        h(NTag, {
          type: expired ? 'error' : 'info',
          size: 'small',
          round: true
        }, { default: () => expired ? '已过期' : '有效' }),
        h('div', { class: 'expiry-date' }, formatDate(row.expires_at))
      ])
    }
  },
  { 
    title: '模型限制', 
    key: 'allowed_models',
    width: 120,
    render: (row: AdminKey) => {
      if (!row.allowed_models || row.allowed_models.length === 0) {
        return h('span', { class: 'unlimited-badge' }, '所有模型')
      }
      return h(NTooltip, {
        trigger: 'hover'
      }, {
        trigger: () => h(NTag, {
          type: 'info',
          size: 'small',
          round: true
        }, { default: () => `${row.allowed_models!.length} 个模型` }),
        default: () => h('div', {}, row.allowed_models!.join(', '))
      })
    }
  },
  { 
    title: '使用次数', 
    key: 'usage_count',
    width: 100,
    render: (row: AdminKey) => {
      return h('div', { class: 'usage-cell' }, [
        h('span', { class: 'usage-number' }, row.usage_count || 0)
      ])
    }
  },

  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row: AdminKey) => {
      return h('div', { class: 'action-buttons' }, [
        h(
          NButton, 
          { 
            text: true,
            type: 'info',
            onClick: () => showTokenDetails(row),
            class: 'action-btn'
          }, 
          { 
            default: () => '详情',
            icon: () => h(NIcon, null, { default: () => h(InformationCircleOutline) })
          }
        ),
        h(
          NButton, 
          { 
            text: true,
            type: 'primary',
            onClick: () => handleCopy(row.key),
            class: 'action-btn'
          }, 
          { 
            default: () => '复制',
            icon: () => h(NIcon, null, { default: () => h(CopyOutline) })
          }
        ),
        h(
          NButton, 
          { 
            text: true, 
            type: 'error', 
            onClick: () => handleDelete(row),
            class: 'action-btn'
          }, 
          { 
            default: () => '删除',
            icon: () => h(NIcon, null, { default: () => h(TrashOutline) })
          }
        )
      ])
    }
  }
]

async function loadKeys(showSuccessMessage = false) {
  loading.value = true
  try {
    const response = await listKeys()
    if (response.data && response.data.keys) {
      tokens.value = response.data.keys.filter((key: AdminKey) => key.key !== '0000')
      if (showSuccessMessage && tokens.value.length > 0) {
        message.success(`成功加载 ${tokens.value.length} 个令牌`)
      }
    } else {
      tokens.value = []
    }
  } catch (error: any) {
    console.error('Failed to load keys:', error)
    if (error.type === 'UNAUTHORIZED') {
      message.error('会话已过期，请重新登录')
      window.location.href = '/login'
    } else {
      message.error(error.message || '加载密钥列表失败')
    }
  } finally {
    loading.value = false
  }
}

function handleCopy(key: string) {
  navigator.clipboard.writeText(key)
  message.success('完整密钥已复制到剪贴板')
}

function handleDelete(row: AdminKey) {
  dialog.warning({
    title: '确认删除',
    content: `确定要删除密钥 ${row.masked_key} 吗？此操作不可恢复。`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await removeKey(row.key)
        tokens.value = tokens.value.filter(token => token.key !== row.key)
        message.success('密钥删除成功')
        setTimeout(() => { loadKeys(false) }, 500)
      } catch (error: any) {
        console.error('Failed to delete key:', error)
        message.error(error.response?.data?.error?.message || '删除密钥失败')
        loadKeys(false)
      }
    }
  })
}

function generateToken(): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
  let token = 'sk-'
  for (let i = 0; i < 48; i++) {
    token += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  return token
}

function handleAutoGenerateToken() {
  // Reset auto-generate form data
  autoGenFormData.value = {
    tokenName: '',
    quotaType: 'unlimited',
    quotaLimit: 10,
    expirationType: 'never',
    expiresAt: null,
    modelRestriction: 'all',
    allowedModels: []
  }
  showAutoGenerateDialog.value = true
}

async function confirmAutoGenerateToken() {
  adding.value = true
  try {
    const token = generateToken()
    
    // Build the payload with optional fields
    const payload: any = {
      key: token,
      token_name: autoGenFormData.value.tokenName.trim() || undefined
    }

    // Add quota limit if set
    if (autoGenFormData.value.quotaType === 'limited' && autoGenFormData.value.quotaLimit > 0) {
      payload.quota_limit = autoGenFormData.value.quotaLimit
    }

    // Add expiration date if set
    if (autoGenFormData.value.expirationType === 'date' && autoGenFormData.value.expiresAt) {
      payload.expires_at = new Date(autoGenFormData.value.expiresAt).toISOString()
    }

    // Add allowed models if set
    if (autoGenFormData.value.modelRestriction === 'selected' && autoGenFormData.value.allowedModels.length > 0) {
      payload.allowed_models = autoGenFormData.value.allowedModels
    }

    await addKey(payload)
    generatedToken.value = token
    message.success('令牌生成成功')
    showAutoGenerateDialog.value = false
    showGeneratedDialog.value = true
    setTimeout(() => { loadKeys(false) }, 500)
  } catch (error: any) {
    console.error('Failed to generate token:', error)
    message.error(error.response?.data?.error?.message || '生成令牌失败')
  } finally {
    adding.value = false
  }
}

function handleCopyGenerated() {
  navigator.clipboard.writeText(generatedToken.value)
  message.success('令牌已复制到剪贴板')
}

async function handleAddToken() {
  if (!formData.value.tokenKey.trim()) {
    message.warning('请输入自定义令牌')
    return
  }

  adding.value = true
  try {
    // Build the payload with optional fields
    const payload: any = {
      key: formData.value.tokenKey.trim(),
      token_name: formData.value.tokenName.trim() || undefined
    }

    // Add quota limit if set
    if (formData.value.quotaType === 'limited' && formData.value.quotaLimit > 0) {
      payload.quota_limit = formData.value.quotaLimit
    }

    // Add expiration date if set
    if (formData.value.expirationType === 'date' && formData.value.expiresAt) {
      payload.expires_at = new Date(formData.value.expiresAt).toISOString()
    }

    // Add allowed models if set
    if (formData.value.modelRestriction === 'selected' && formData.value.allowedModels.length > 0) {
      payload.allowed_models = formData.value.allowedModels
    }

    await addKey(payload)
    message.success('自定义令牌添加成功')
    showAddDialog.value = false
    
    // Reset form
    formData.value = {
      tokenName: '',
      tokenKey: '',
      quotaType: 'unlimited',
      quotaLimit: 10,
      expirationType: 'never',
      expiresAt: null,
      modelRestriction: 'all',
      allowedModels: []
    }
    
    setTimeout(() => { loadKeys(false) }, 500)
  } catch (error: any) {
    console.error('Failed to add key:', error)
    message.error(error.response?.data?.error?.message || '添加令牌失败')
  } finally {
    adding.value = false
  }
}

function handleEditTokenName(key: string, currentName: string) {
  editingTokenKey.value = key
  editingTokenName.value = currentName
}

function handleCancelEdit() {
  editingTokenKey.value = null
  editingTokenName.value = ''
}

async function handleSaveTokenName(key: string) {
  try {
    await updateKeyName(key, editingTokenName.value.trim())
    message.success('令牌名称更新成功')
    editingTokenKey.value = null
    editingTokenName.value = ''
    const token = tokens.value.find(t => t.key === key)
    if (token) {
      token.token_name = editingTokenName.value.trim()
    }
  } catch (error: any) {
    console.error('Failed to update key name:', error)
    message.error(error.response?.data?.error?.message || '更新令牌名称失败')
  }
}

onMounted(() => {
  loadKeys()
})
</script>

<style scoped>
/* ============================================
   Token Management - 简约设计
   ============================================ */

.token-management {
  max-width: 1400px;
  margin: 0 auto;
  position: relative;
  z-index: 1;
}

/* ============================================
   页面头部 - 简约设计
   ============================================ */

.page-header {
  border-radius: var(--border-radius-md);
  padding: var(--spacing-xl);
  margin-bottom: var(--spacing-xl);
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  flex-wrap: wrap;
  gap: var(--spacing-lg);
}

.header-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.icon-wrapper {
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-icon {
  background: var(--color-primary);
  color: white;
  padding: var(--spacing-md);
  border-radius: var(--border-radius-md);
}

.title-text h2 {
  color: var(--text-primary);
  font-weight: 600;
  margin-bottom: 4px;
}

.title-text .n-text {
  color: var(--text-secondary) !important;
  font-size: 14px;
}

.header-actions {
  display: flex;
  gap: var(--spacing-sm);
}

.glass-button {
  border-radius: var(--border-radius) !important;
  font-weight: 500;
  padding: 0 var(--spacing-lg) !important;
  height: 44px !important;
  transition: all var(--transition-normal);
}

.primary-button {
  background: var(--color-primary) !important;
  color: white !important;
  border: none !important;
}

.primary-button:hover {
  background: var(--color-primary-hover) !important;
}

.secondary-button {
  background: var(--bg-card) !important;
  color: var(--color-primary) !important;
  border: 1px solid var(--border-color) !important;
}

.secondary-button:hover {
  background: var(--bg-hover) !important;
  border-color: var(--color-primary) !important;
}

.button-text {
  font-size: 14px;
}

/* ============================================
   统计卡片 - 简约设计
   ============================================ */

.stats-grid {
  margin-bottom: var(--spacing-xl);
}

.stat-card-wrapper {
  height: 100%;
}

.stat-card {
  border-radius: var(--border-radius-md) !important;
  padding: var(--spacing-lg) !important;
  transition: box-shadow var(--transition-normal), transform var(--transition-normal);
  position: relative;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md) !important;
}

.stat-icon-wrapper {
  width: 48px;
  height: 48px;
  border-radius: var(--border-radius);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: var(--spacing-md);
}

.icon-purple {
  background: var(--color-primary-light);
  color: var(--color-primary);
}

.icon-green {
  background: var(--color-success-light);
  color: var(--color-success);
}

.icon-orange {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.stat-content {
  position: relative;
  z-index: 1;
}

.stat-label {
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: var(--spacing-xs);
  font-weight: 500;
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: var(--spacing-sm);
  line-height: 1;
}

.stat-card-1 .stat-value {
  color: var(--color-primary);
}

.stat-card-2 .stat-value {
  color: var(--color-success);
}

.stat-card-3 .stat-value {
  color: var(--color-warning);
}

.stat-trend {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: 12px;
  color: var(--color-success);
  font-weight: 500;
}

/* ============================================
   令牌列表卡片 - 简约设计
   ============================================ */

.tokens-card {
  border-radius: var(--border-radius-md) !important;
  padding: var(--spacing-lg) !important;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: var(--spacing-md);
}

.table-title {
  display: flex;
  align-items: center;
}

.count-badge {
  background: var(--color-primary-light) !important;
  color: var(--color-primary) !important;
  border: none !important;
  padding: var(--spacing-xs) var(--spacing-md) !important;
  font-weight: 500;
}

/* ============================================
   表格样式 - 简约设计
   ============================================ */

.modern-table :deep(.n-data-table-th) {
  background: var(--bg-secondary) !important;
  font-weight: 500;
  color: var(--text-primary);
}

.modern-table :deep(.n-data-table-td) {
  padding: 16px;
}

/* 空状态样式 */
.empty-state-modern {
  padding: 60px 20px;
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.empty-icon-wrapper {
  width: 120px;
  height: 120px;
  background: var(--bg-secondary);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
  animation: pulse 3s infinite ease-in-out;
}

@keyframes pulse {
  0% { transform: scale(1); box-shadow: 0 0 0 0 rgba(209, 213, 219, 0.7); }
  70% { transform: scale(1.05); box-shadow: 0 0 0 20px rgba(209, 213, 219, 0); }
  100% { transform: scale(1); box-shadow: 0 0 0 0 rgba(209, 213, 219, 0); }
}

/* 令牌名称编辑 */
.name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background-color 0.2s;
  width: fit-content;
}

.name-cell:hover {
  background-color: var(--bg-hover);
}

.name-text {
  font-weight: 500;
  color: var(--text-primary);
}

.edit-icon {
  opacity: 0;
  color: var(--text-muted);
  transition: opacity 0.2s;
}

.name-cell:hover .edit-icon {
  opacity: 1;
}

.name-edit-input {
  border: 1px solid var(--color-primary);
  border-radius: 4px;
  padding: 4px 8px;
  font-size: inherit;
  outline: none;
  width: 100%;
}

.name-edit-actions {
  display: flex;
  gap: 4px;
  margin-top: 4px;
}

/* 令牌显示样式 */
.token-cell {
  background: var(--bg-secondary);
  padding: 6px 10px;
  border-radius: 6px;
  font-family: 'JetBrains Mono', monospace;
  font-size: 0.85rem;
  color: var(--text-secondary);
  border: 1px solid var(--border-color);
  width: fit-content;
}

/* 操作按钮 */
.action-buttons {
  display: flex;
  gap: 8px;
  opacity: 0.8;
  transition: opacity 0.2s;
}

.tokens-table :deep(.n-data-table-tr:hover) .action-buttons {
  opacity: 1;
}

/* 进度条 */
.quota-text {
  font-size: 0.8rem;
  color: var(--text-muted);
  margin-bottom: 2px;
  display: flex;
  justify-content: space-between;
}

/* 响应式适配 */
@media (max-width: 1024px) {
  .stats-grid {
    grid-template-columns: repeat(1, 1fr) !important;
  }
}

@media (max-width: 640px) {
  .header-content {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
  
  .action-button {
    flex: 1;
  }
  
  .stats-grid {
    grid-template-columns: 1fr !important;
  }
  
  .stat-value {
    font-size: 24px;
  }
  
  .tokens-card {
    padding: var(--spacing-md) !important;
  }
  
  .table-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-sm);
  }
  
  .modern-table :deep(.n-data-table-td),
  .modern-table :deep(.n-data-table-th) {
    padding: var(--spacing-sm) !important;
  }
  
  .token-code {
    font-size: 11px;
    padding: var(--spacing-sm);
  }
  
  .action-buttons {
    flex-direction: column;
    gap: var(--spacing-sm);
  }
}
</style>
