<template>
  <div class="admin-panel">
    <!-- 页面标题 -->
    <div class="page-header glass-card">
      <h1 class="gradient-text">🛠️ 管理后台</h1>
      <p class="subtitle">系统管理与监控中心</p>
    </div>

    <n-space vertical size="large">
      <n-tabs type="line" animated class="admin-tabs">
        <n-tab-pane name="users" tab="用户管理">
          <div class="tab-content glass-card">
            <n-space vertical size="large">
              <!-- 操作栏 -->
              <n-space justify="space-between" class="action-bar">
                <h3 class="section-title">👥 用户列表</h3>
                <n-button @click="loadUsers" :loading="usersLoading" class="refresh-btn">
                  <template #icon>
                    <n-icon><RefreshOutline /></n-icon>
                  </template>
                  刷新
                </n-button>
              </n-space>

              <!-- 用户列表 -->
              <n-data-table
                :columns="userColumns"
                :data="users"
                :loading="usersLoading"
                :pagination="false"
                class="modern-table"
              />
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="keys" tab="密钥管理">
          <div class="tab-content glass-card">
            <n-space vertical size="large">
              <!-- 操作栏 -->
              <n-space justify="space-between" class="action-bar">
                <h3 class="section-title">🔑 用户密钥列表</h3>
                <n-button @click="loadKeys" :loading="keysLoading" class="refresh-btn">
                  <template #icon>
                    <n-icon><RefreshOutline /></n-icon>
                  </template>
                  刷新
                </n-button>
              </n-space>

              <!-- 密钥列表 -->
              <n-data-table
                :columns="keyColumns"
                :data="keys"
                :loading="keysLoading"
                :pagination="false"
                class="modern-table"
              />
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="sessions" tab="Cursor Session 管理">
          <div class="tab-content glass-card">
            <n-space vertical size="large">
              <!-- 操作栏 -->
              <n-space justify="space-between" class="action-bar">
                <h3 class="section-title">💾 Session 列表</h3>
                <n-space>
                  <n-button @click="handleReloadSessions" :loading="loading" class="refresh-btn">
                    <template #icon>
                      <n-icon><RefreshOutline /></n-icon>
                    </template>
                    重新加载
                  </n-button>
                  <n-button type="primary" @click="showAddModal = true" class="add-btn">
                    <template #icon>
                      <n-icon><AddOutline /></n-icon>
                    </template>
                    添加 Session
                  </n-button>
                </n-space>
              </n-space>

              <!-- Session 列表 -->
              <n-data-table
                :columns="columns"
                :data="sessions"
                :loading="loading"
                :pagination="false"
                class="modern-table"
              />
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="announcements" tab="公告管理">
          <div class="tab-content glass-card">
            <n-space vertical size="large">
              <!-- 操作栏 -->
              <n-space justify="space-between" class="action-bar">
                <h3 class="section-title">📢 公告管理</h3>
                <n-space>
                  <n-button @click="loadAnnouncements" :loading="announcementsLoading" class="refresh-btn">
                    <template #icon>
                      <n-icon><RefreshOutline /></n-icon>
                    </template>
                    刷新
                  </n-button>
                  <n-button type="primary" @click="showAnnouncementModal = true" class="add-btn">
                    <template #icon>
                      <n-icon><AddOutline /></n-icon>
                    </template>
                    创建公告
                  </n-button>
                </n-space>
              </n-space>

              <!-- 公告列表 -->
              <n-data-table
                :columns="announcementColumns"
                :data="announcements"
                :loading="announcementsLoading"
                :pagination="false"
                class="modern-table"
              />
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="balances" tab="余额管理">
          <div class="tab-content glass-card">
            <n-space vertical size="large">
              <!-- 操作栏 -->
              <n-space justify="space-between" class="action-bar">
                <h3 class="section-title">💰 用户余额管理</h3>
                <n-button @click="loadUserBalances" :loading="balancesLoading" class="refresh-btn">
                  <template #icon>
                    <n-icon><RefreshOutline /></n-icon>
                  </template>
                  刷新
                </n-button>
              </n-space>

              <!-- 余额列表 -->
              <n-data-table
                :columns="balanceColumns"
                :data="userBalances"
                :loading="balancesLoading"
                :pagination="balancePagination"
                @update:page="handleBalancePageChange"
                class="modern-table"
              />
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="exchanges" tab="兑换记录">
          <div class="tab-content glass-card">
            <n-space vertical size="large">
              <!-- 统计卡片 -->
              <div class="exchange-stats-grid">
                <div class="stat-card">
                  <div class="stat-icon">🔄</div>
                  <div class="stat-content">
                    <div class="stat-value">{{ exchangeStats?.total_count || 0 }}</div>
                    <div class="stat-label">总兑换次数</div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-icon">💵</div>
                  <div class="stat-content">
                    <div class="stat-value">${{ (exchangeStats?.total_usd || 0).toFixed(2) }}</div>
                    <div class="stat-label">总兑换金额</div>
                  </div>
                </div>
              </div>

              <!-- 操作栏 -->
              <n-space justify="space-between" class="action-bar">
                <h3 class="section-title">💱 兑换记录列表</h3>
                <n-space>
                  <n-input-number
                    v-model:value="exchangeFilterUserId"
                    placeholder="用户ID筛选"
                    clearable
                    :min="1"
                    style="width: 140px"
                    @update:value="handleExchangeFilterChange"
                  />
                  <n-date-picker
                    v-model:value="exchangeDateRange"
                    type="daterange"
                    clearable
                    :shortcuts="dateRangeShortcuts"
                    @update:value="handleExchangeFilterChange"
                    style="width: 280px"
                  />
                  <n-button @click="loadExchangeRecords" :loading="exchangeRecordsLoading" class="refresh-btn">
                    <template #icon>
                      <n-icon><RefreshOutline /></n-icon>
                    </template>
                    刷新
                  </n-button>
                </n-space>
              </n-space>

              <!-- 兑换记录列表 -->
              <n-data-table
                :columns="exchangeColumns"
                :data="exchangeRecords"
                :loading="exchangeRecordsLoading"
                :pagination="exchangePagination"
                @update:page="handleExchangePageChange"
                class="modern-table"
              />
            </n-space>
          </div>
        </n-tab-pane>

        <n-tab-pane name="usage" tab="使用统计">
          <div class="tab-content glass-card">
            <n-space vertical size="large">
              <!-- 操作栏 -->
              <n-space justify="space-between" class="action-bar">
                <h3 class="section-title">📊 使用统计分析</h3>
                <n-space>
                  <n-date-picker
                    v-model:value="usageDateRange"
                    type="daterange"
                    clearable
                    :shortcuts="dateRangeShortcuts"
                    @update:value="handleDateRangeChange"
                    style="width: 280px"
                  />
                  <n-button @click="loadUsageStats" :loading="usageLoading" class="refresh-btn">
                    <template #icon>
                      <n-icon><RefreshOutline /></n-icon>
                    </template>
                    刷新
                  </n-button>
                  <n-button type="primary" @click="handleExportCSV" :loading="exportLoading" class="add-btn">
                    <template #icon>
                      <n-icon><DownloadOutline /></n-icon>
                    </template>
                    导出 CSV
                  </n-button>
                </n-space>
              </n-space>

              <!-- 系统概览统计卡片 -->
              <div class="stats-grid">
                <div class="stat-card">
                  <div class="stat-icon">👥</div>
                  <div class="stat-content">
                    <div class="stat-value">{{ usageStats?.total_users || 0 }}</div>
                    <div class="stat-label">活跃用户</div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-icon">📡</div>
                  <div class="stat-content">
                    <div class="stat-value">{{ formatNumber(usageStats?.total_requests || 0) }}</div>
                    <div class="stat-label">总请求数</div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-icon">🎯</div>
                  <div class="stat-content">
                    <div class="stat-value">{{ formatNumber(usageStats?.total_tokens || 0) }}</div>
                    <div class="stat-label">总 Token 消耗</div>
                  </div>
                </div>
                <div class="stat-card">
                  <div class="stat-icon">📈</div>
                  <div class="stat-content">
                    <div class="stat-value" :class="{ 'positive': (usageTrends?.growth_rate ?? 0) > 0, 'negative': (usageTrends?.growth_rate ?? 0) < 0 }">
                      {{ usageTrends?.growth_rate !== undefined ? ((usageTrends.growth_rate > 0 ? '+' : '') + usageTrends.growth_rate.toFixed(1) + '%') : '0%' }}
                    </div>
                    <div class="stat-label">增长率</div>
                  </div>
                </div>
              </div>

              <!-- 使用趋势图表 -->
              <div class="chart-section glass-card-inner">
                <div class="chart-header">
                  <h4 class="chart-title">📈 使用趋势</h4>
                  <n-radio-group v-model:value="trendView" @update:value="loadUsageTrends">
                    <n-radio-button value="daily">日</n-radio-button>
                    <n-radio-button value="weekly">周</n-radio-button>
                    <n-radio-button value="monthly">月</n-radio-button>
                  </n-radio-group>
                </div>
                <div class="chart-container">
                  <UsageTimeSeriesChart
                    v-if="usageTrends?.trends?.length"
                    :data="formatTrendsForChart(usageTrends.trends)"
                    :loading="trendsLoading"
                  />
                  <n-empty v-else description="暂无趋势数据" />
                </div>
              </div>

              <!-- 两列布局：Top 用户 和 模型统计 -->
              <div class="two-column-grid">
                <!-- Top 用户排行 -->
                <div class="column-card glass-card-inner">
                  <h4 class="column-title">🏆 Top 用户排行</h4>
                  <n-data-table
                    :columns="topUsersColumns"
                    :data="usageStats?.top_users || []"
                    :loading="usageLoading"
                    :pagination="false"
                    size="small"
                    class="inner-table"
                  />
                </div>

                <!-- 模型使用统计 -->
                <div class="column-card glass-card-inner">
                  <h4 class="column-title">🤖 模型使用统计</h4>
                  <ModelBreakdownChart
                    v-if="usageStats?.top_models?.length"
                    :data="usageStats.top_models"
                    :loading="usageLoading"
                  />
                  <n-empty v-else description="暂无模型数据" />
                </div>
              </div>

              <!-- Cursor Session 使用统计 -->
              <div class="session-section glass-card-inner">
                <h4 class="column-title">💾 Cursor Session 使用统计</h4>
                <n-data-table
                  :columns="sessionUsageColumns"
                  :data="sessionUsage?.sessions || []"
                  :loading="sessionUsageLoading"
                  :pagination="false"
                  class="inner-table"
                />
              </div>
            </n-space>
          </div>
        </n-tab-pane>
      </n-tabs>
    </n-space>

    <!-- 添加 Session 对话框 -->
    <n-modal v-model:show="showAddModal" preset="dialog" title="添加 Cursor Session">
      <n-form ref="formRef" :model="formData" :rules="rules" label-placement="left" label-width="120">
        <n-form-item label="邮箱" path="email">
          <n-input v-model:value="formData.email" placeholder="account@cursor.com" />
        </n-form-item>
        <n-form-item label="Session Token" path="session_token">
          <n-input
            v-model:value="formData.session_token"
            type="textarea"
            placeholder="粘贴 cursor_session cookie 值"
            :rows="3"
          />
        </n-form-item>
        <n-form-item label="过期时间" path="expires_at">
          <n-date-picker
            v-model:value="formData.expires_at"
            type="datetime"
            clearable
            style="width: 100%"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAddModal = false">取消</n-button>
          <n-button type="primary" @click="handleAddSession" :loading="submitting">
            添加
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 创建公告对话框 -->
    <n-modal v-model:show="showAnnouncementModal" preset="dialog" title="创建公告">
      <n-form ref="announcementFormRef" :model="announcementFormData" :rules="announcementRules" label-placement="left" label-width="80">
        <n-form-item label="标题" path="title">
          <n-input v-model:value="announcementFormData.title" placeholder="请输入公告标题" />
        </n-form-item>
        <n-form-item label="内容" path="content">
          <n-input
            v-model:value="announcementFormData.content"
            type="textarea"
            placeholder="请输入公告内容"
            :rows="5"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAnnouncementModal = false">取消</n-button>
          <n-button type="primary" @click="handleCreateAnnouncement" :loading="announcementSubmitting">
            创建
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 调整余额对话框 -->
    <n-modal v-model:show="showAdjustBalanceModal" preset="dialog" title="调整用户余额">
      <n-form ref="adjustBalanceFormRef" :model="adjustBalanceFormData" :rules="adjustBalanceRules" label-placement="left" label-width="100">
        <n-form-item label="用户">
          <n-input :value="selectedUserForBalance?.username + ' (' + selectedUserForBalance?.email + ')'" disabled />
        </n-form-item>
        <n-form-item label="当前余额">
          <n-input :value="'$' + (selectedUserForBalance?.balance?.toFixed(2) || '0.00')" disabled />
        </n-form-item>
        <n-form-item label="调整金额" path="amount">
          <n-input-number
            v-model:value="adjustBalanceFormData.amount"
            placeholder="正数增加，负数扣除"
            :precision="2"
            style="width: 100%"
          >
            <template #prefix>$</template>
          </n-input-number>
        </n-form-item>
        <n-form-item label="调整原因" path="reason">
          <n-input
            v-model:value="adjustBalanceFormData.reason"
            type="textarea"
            placeholder="请输入调整原因"
            :rows="3"
          />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showAdjustBalanceModal = false">取消</n-button>
          <n-button type="primary" @click="handleAdjustBalance" :loading="adjustBalanceSubmitting">
            确认调整
          </n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { useMessage, useDialog, type DataTableColumns, NButton, NTag, NSpace } from 'naive-ui'
import { AddOutline, RefreshOutline, TrashOutline, CheckmarkCircleOutline, DownloadOutline } from '@vicons/ionicons5'
import type { CursorSession, Announcement } from '@/types'
import {
  listCursorSessions,
  addCursorSession,
  removeCursorSession,
  validateCursorSession,
  reloadCursorSessions,
  listKeys,
  toggleKeyStatus,
  removeKey
} from '@/api/admin'
import { usersApi, type User } from '@/api/users'
import { announcementApi } from '@/api/announcement'
import type { AdminKey } from '@/types'
import {
  getAdminUsageStats,
  getUsageTrends,
  getCursorSessionUsage,
  exportUsageData,
  downloadCSV,
  type AdminUsageStats,
  type AdminUsageTrends,
  type CursorSessionUsageResponse,
  type UserUsageSummary,
  type CursorSessionUsage
} from '@/api/adminUsage'
import {
  getAllUserBalances,
  adjustUserBalance,
  type UserBalanceInfo
} from '@/api/adminBalance'
import {
  getAdminExchangeRecords,
  getAdminExchangeStats,
  type AdminExchangeRecord,
  type AdminExchangeStatsResponse
} from '@/api/gameCoin'
import UsageTimeSeriesChart from '@/components/UsageTimeSeriesChart.vue'
import ModelBreakdownChart from '@/components/ModelBreakdownChart.vue'

const message = useMessage()
const dialog = useDialog()
const loading = ref(false)
const submitting = ref(false)
const sessions = ref<CursorSession[]>([])
const showAddModal = ref(false)
const formRef = ref()

// 用户管理状态
const users = ref<User[]>([])
const usersLoading = ref(false)

// 密钥管理状态
const keys = ref<AdminKey[]>([])
const keysLoading = ref(false)

// 公告管理状态
const announcements = ref<Announcement[]>([])
const announcementsLoading = ref(false)
const showAnnouncementModal = ref(false)
const announcementSubmitting = ref(false)
const announcementFormRef = ref()

// 余额管理状态
const userBalances = ref<UserBalanceInfo[]>([])
const balancesLoading = ref(false)
const showAdjustBalanceModal = ref(false)
const adjustBalanceSubmitting = ref(false)
const adjustBalanceFormRef = ref()
const selectedUserForBalance = ref<UserBalanceInfo | null>(null)
const balancePagination = ref({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: false
})

// 兑换记录管理状态
const exchangeRecords = ref<AdminExchangeRecord[]>([])
const exchangeRecordsLoading = ref(false)
const exchangeStats = ref<AdminExchangeStatsResponse | null>(null)
const exchangeStatsLoading = ref(false)
const exchangeFilterUserId = ref<number | null>(null)
const exchangeDateRange = ref<[number, number] | null>(null)
const exchangePagination = ref({
  page: 1,
  pageSize: 20,
  itemCount: 0,
  showSizePicker: false
})

// 使用统计状态
const usageStats = ref<AdminUsageStats | null>(null)
const usageLoading = ref(false)
const usageTrends = ref<AdminUsageTrends | null>(null)
const trendsLoading = ref(false)
const sessionUsage = ref<CursorSessionUsageResponse | null>(null)
const sessionUsageLoading = ref(false)
const exportLoading = ref(false)
const trendView = ref<'daily' | 'weekly' | 'monthly'>('daily')
const usageDateRange = ref<[number, number] | null>(null)

// 日期范围快捷选项
const dateRangeShortcuts = {
  '今天': () => {
    const now = new Date()
    const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    return [start.getTime(), now.getTime()]
  },
  '最近7天': () => {
    const now = new Date()
    const start = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000)
    return [start.getTime(), now.getTime()]
  },
  '最近30天': () => {
    const now = new Date()
    const start = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000)
    return [start.getTime(), now.getTime()]
  },
  '最近90天': () => {
    const now = new Date()
    const start = new Date(now.getTime() - 90 * 24 * 60 * 60 * 1000)
    return [start.getTime(), now.getTime()]
  }
}

const formData = ref({
  email: '',
  session_token: '',
  expires_at: Date.now() + 30 * 24 * 60 * 60 * 1000 // 默认 30 天后
})

const announcementFormData = ref({
  title: '',
  content: ''
})

const rules = {
  email: {
    required: true,
    message: '请输入邮箱',
    trigger: 'blur'
  },
  session_token: {
    required: true,
    message: '请输入 Session Token',
    trigger: 'blur'
  }
}

const announcementRules = {
  title: {
    required: true,
    message: '请输入公告标题',
    trigger: 'blur'
  },
  content: {
    required: true,
    message: '请输入公告内容',
    trigger: 'blur'
  }
}

// 余额调整表单数据
const adjustBalanceFormData = ref({
  amount: 0,
  reason: ''
})

const adjustBalanceRules = {
  amount: {
    required: true,
    type: 'number' as const,
    message: '请输入调整金额',
    trigger: 'blur',
    validator: (_rule: any, value: number) => {
      if (value === 0) {
        return new Error('调整金额不能为0')
      }
      return true
    }
  },
  reason: {
    required: true,
    message: '请输入调整原因',
    trigger: 'blur'
  }
}

// 余额表格列定义
const balanceColumns: DataTableColumns<UserBalanceInfo> = [
  {
    title: 'ID',
    key: 'user_id',
    width: 80
  },
  {
    title: '用户名',
    key: 'username',
    width: 120
  },
  {
    title: '邮箱',
    key: 'email',
    width: 180,
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: '余额',
    key: 'balance',
    width: 120,
    render: (row) => {
      const balance = row.balance?.toFixed(2) || '0.00'
      const color = row.balance <= 0 ? '#ef4444' : row.balance < 10 ? '#f59e0b' : '#10b981'
      return h('span', { style: { color, fontWeight: '600' } }, `$${balance}`)
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      return h(
        NTag,
        {
          type: row.status === 'active' ? 'success' : 'error',
          size: 'small'
        },
        { default: () => (row.status === 'active' ? '正常' : '已耗尽') }
      )
    }
  },
  {
    title: '邀请码',
    key: 'referral_code',
    width: 100
  },
  {
    title: '累计消费',
    key: 'total_consumed',
    width: 120,
    render: (row) => `$${row.total_consumed?.toFixed(2) || '0.00'}`
  },
  {
    title: '累计充值',
    key: 'total_recharged',
    width: 120,
    render: (row) => `$${row.total_recharged?.toFixed(2) || '0.00'}`
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 160,
    render: (row) => {
      return new Date(row.created_at).toLocaleString('zh-CN')
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) => {
      return h(
        NButton,
        {
          size: 'small',
          type: 'primary',
          onClick: () => openAdjustBalanceModal(row)
        },
        { default: () => '调整余额' }
      )
    }
  }
]

// 兑换记录表格列定义
const exchangeColumns: DataTableColumns<AdminExchangeRecord> = [
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
  {
    title: '用户ID',
    key: 'user_id',
    width: 80
  },
  {
    title: '用户名',
    key: 'username',
    width: 120
  },
  {
    title: '邮箱',
    key: 'email',
    width: 180,
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: '游戏币数量',
    key: 'game_coins_amount',
    width: 120,
    render: (row) => {
      return h('span', { style: { color: '#f59e0b', fontWeight: '600' } }, `🪙 ${row.game_coins_amount.toFixed(2)}`)
    }
  },
  {
    title: '兑换USD',
    key: 'usd_amount',
    width: 120,
    render: (row) => {
      return h('span', { style: { color: '#10b981', fontWeight: '600' } }, `$${row.usd_amount.toFixed(2)}`)
    }
  },
  {
    title: '汇率',
    key: 'exchange_rate',
    width: 80,
    render: (row) => `1:${row.exchange_rate}`
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      return h(
        NTag,
        {
          type: row.status === 'completed' ? 'success' : 'error',
          size: 'small'
        },
        { default: () => (row.status === 'completed' ? '成功' : '失败') }
      )
    }
  },
  {
    title: '兑换时间',
    key: 'created_at',
    width: 180,
    render: (row) => {
      return new Date(row.created_at).toLocaleString('zh-CN')
    }
  }
]

// 公告表格列定义
const announcementColumns: DataTableColumns<Announcement> = [
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
  {
    title: '标题',
    key: 'title',
    width: 200
  },
  {
    title: '内容',
    key: 'content',
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: '阅读统计',
    key: 'read_count',
    width: 120,
    render: (row) => {
      return row.read_count !== undefined ? `${row.read_count} 人已读` : '-'
    }
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render: (row) => {
      return new Date(row.created_at).toLocaleString('zh-CN')
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    render: (row) => {
      return h(
        NButton,
        {
          size: 'small',
          type: 'error',
          onClick: () => handleDeleteAnnouncement(row)
        },
        { default: () => '删除' }
      )
    }
  }
]

// Top 用户排行表格列定义
const topUsersColumns: DataTableColumns<UserUsageSummary> = [
  {
    title: '排名',
    key: 'rank',
    width: 60,
    render: (_, index) => index + 1
  },
  {
    title: '用户名',
    key: 'username',
    width: 150
  },
  {
    title: '请求数',
    key: 'requests',
    width: 100,
    render: (row) => formatNumber(row.requests)
  },
  {
    title: 'Token 消耗',
    key: 'total_tokens',
    width: 120,
    render: (row) => formatNumber(row.total_tokens)
  }
]

// Cursor Session 使用统计表格列定义
const sessionUsageColumns: DataTableColumns<CursorSessionUsage> = [
  {
    title: 'Session',
    key: 'cursor_session',
    ellipsis: {
      tooltip: true
    },
    render: (row) => {
      // 显示 session 的前 20 个字符
      const session = row.cursor_session || '-'
      return session.length > 20 ? session.substring(0, 20) + '...' : session
    }
  },
  {
    title: '请求数',
    key: 'requests',
    width: 120,
    render: (row) => formatNumber(row.requests)
  },
  {
    title: 'Token 消耗',
    key: 'total_tokens',
    width: 150,
    render: (row) => formatNumber(row.total_tokens)
  },
  {
    title: '健康状态',
    key: 'health',
    width: 100,
    render: (row) => {
      // 根据请求数判断健康状态
      const isHealthy = row.requests > 0
      return h(
        NTag,
        {
          type: isHealthy ? 'success' : 'warning',
          size: 'small'
        },
        { default: () => (isHealthy ? '活跃' : '空闲') }
      )
    }
  }
]

// 密钥表格列定义
const keyColumns: DataTableColumns<AdminKey> = [
  {
    title: '掩码密钥',
    key: 'masked_key',
    width: 250
  },
  {
    title: '用户名',
    key: 'username',
    width: 150,
    render: (row) => {
      return row.username || h(NTag, { type: 'info', size: 'small' }, { default: () => '系统密钥' })
    }
  },
  {
    title: '状态',
    key: 'is_active',
    width: 100,
    render: (row) => {
      return h(
        NTag,
        {
          type: row.is_active ? 'success' : 'error',
          size: 'small'
        },
        { default: () => (row.is_active ? '启用' : '禁用') }
      )
    }
  },
  {
    title: '使用次数',
    key: 'usage_count',
    width: 120
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render: (row) => {
      return new Date(row.created_at).toLocaleString('zh-CN')
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 250,
    render: (row) => {
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                onClick: () => handleCopyKey(row.key)
              },
              { default: () => '复制完整密钥' }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: row.is_active ? 'warning' : 'success',
                onClick: () => handleToggleKey(row)
              },
              { default: () => (row.is_active ? '禁用' : '启用') }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: 'error',
                onClick: () => handleDeleteKey(row)
              },
              { default: () => '删除' }
            )
          ]
        }
      )
    }
  }
]

// 用户表格列定义
const userColumns: DataTableColumns<User> = [
  {
    title: 'ID',
    key: 'id',
    width: 80
  },
  {
    title: '用户名',
    key: 'username',
    width: 150
  },
  {
    title: '邮箱',
    key: 'email',
    width: 200
  },
  {
    title: '角色',
    key: 'role',
    width: 100,
    render: (row) => {
      return h(
        NTag,
        {
          type: row.role === 'admin' ? 'success' : 'default',
          size: 'small'
        },
        { default: () => (row.role === 'admin' ? '管理员' : '普通用户') }
      )
    }
  },
  {
    title: '状态',
    key: 'is_active',
    width: 100,
    render: (row) => {
      return h(
        NTag,
        {
          type: row.is_active ? 'success' : 'error',
          size: 'small'
        },
        { default: () => (row.is_active ? '正常' : '已禁用') }
      )
    }
  },
  {
    title: '注册时间',
    key: 'created_at',
    width: 180,
    render: (row) => {
      return new Date(row.created_at).toLocaleString('zh-CN')
    }
  },
  {
    title: '最后登录',
    key: 'last_login',
    width: 180,
    render: (row) => {
      if (!row.last_login) return '-'
      return new Date(row.last_login).toLocaleString('zh-CN')
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 250,
    render: (row) => {
      return h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                onClick: () => handleToggleRole(row)
              },
              { default: () => (row.role === 'admin' ? '设为用户' : '设为管理员') }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: row.is_active ? 'warning' : 'success',
                onClick: () => handleToggleStatus(row)
              },
              { default: () => (row.is_active ? '禁用' : '启用') }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: 'error',
                onClick: () => handleDeleteUser(row)
              },
              { default: () => '删除' }
            )
          ]
        }
      )
    }
  }
]

const columns: DataTableColumns<CursorSession> = [
  {
    title: '邮箱',
    key: 'email',
    width: 200
  },
  {
    title: 'Token',
    key: 'cookies',
    ellipsis: {
      tooltip: true
    }
  },
  {
    title: '状态',
    key: 'status',
    width: 100,
    render: (row) => {
      const isValid = (row as any).is_valid
      return h(
        'n-tag',
        {
          type: isValid ? 'success' : 'error',
          size: 'small'
        },
        { default: () => (isValid ? '有效' : '无效') }
      )
    }
  },
  {
    title: '使用次数',
    key: 'usage_count',
    width: 100,
    render: (row) => (row as any).usage_count || 0
  },
  {
    title: '失败次数',
    key: 'fail_count',
    width: 100,
    render: (row) => (row as any).fail_count || 0
  },
  {
    title: '创建时间',
    key: 'created_at',
    width: 180,
    render: (row) => {
      if (!row.created_at) return '-'
      return new Date(row.created_at).toLocaleString('zh-CN')
    }
  },
  {
    title: '最后使用',
    key: 'last_used',
    width: 180,
    render: (row) => {
      if (!row.last_used) return '-'
      return new Date(row.last_used).toLocaleString('zh-CN')
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 180,
    render: (row) => {
      return h(
        'n-space',
        {},
        {
          default: () => [
            h(
              'n-button',
              {
                size: 'small',
                onClick: () => handleValidate(row.email)
              },
              {
                default: () => '验证',
                icon: () => h('n-icon', {}, { default: () => h(CheckmarkCircleOutline) })
              }
            ),
            h(
              'n-button',
              {
                size: 'small',
                type: 'error',
                onClick: () => handleDelete(row.email)
              },
              {
                default: () => '删除',
                icon: () => h('n-icon', {}, { default: () => h(TrashOutline) })
              }
            )
          ]
        }
      )
    }
  }
]

const loadSessions = async () => {
  loading.value = true
  try {
    const response = await listCursorSessions()
    const data = response.data as any
    console.log('Cursor Sessions Response:', data)
    
    // 后端返回格式：{ sessions: [...], stats: {...} } 或直接数组兜底
    const sessionList: CursorSession[] = Array.isArray(data?.sessions)
      ? data.sessions
      : Array.isArray(data)
        ? data
        : []
    sessions.value = sessionList
    if (sessionList.length > 0) {
      message.success(`成功加载 ${sessionList.length} 个 Cursor Session`)
    } else {
      console.warn('Unexpected response format:', data)
    }
  } catch (error: any) {
    console.error('Failed to load sessions:', error)
    message.error(error.message || '加载 Session 列表失败')
  } finally {
    loading.value = false
  }
}

// 重新加载 sessions（从数据库）
async function handleReloadSessions() {
  loading.value = true
  try {
    const response = await reloadCursorSessions()
    message.success(response.data.message || 'Sessions 重新加载成功')
    console.log('Reload stats:', response.data.stats)
    await loadSessions()
  } catch (error: any) {
    console.error('Failed to reload sessions:', error)
    message.error(error.message || '重新加载失败')
  } finally {
    loading.value = false
  }
}

const handleAddSession = async () => {
  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  submitting.value = true
  try {
    const expiresAt = formData.value.expires_at
      ? new Date(formData.value.expires_at).toISOString()
      : undefined

    await addCursorSession({
      email: formData.value.email,
      session_token: formData.value.session_token,
      expires_at: expiresAt
    })

    message.success('Session 添加成功')
    showAddModal.value = false
    formData.value = {
      email: '',
      session_token: '',
      expires_at: Date.now() + 30 * 24 * 60 * 60 * 1000
    }
    await loadSessions()
  } catch (error: any) {
    message.error(error.response?.data?.error?.message || '添加 Session 失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (email: string) => {
  const confirmed = await new Promise((resolve) => {
    dialog.warning({
      title: '确认删除',
      content: `确定要删除 Session "${email}" 吗？`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: () => {
        resolve(true)
      },
      onNegativeClick: () => {
        resolve(false)
      }
    })
  })

  if (!confirmed) return

  try {
    await removeCursorSession(email)
    message.success('Session 删除成功')
    await loadSessions()
  } catch (error: any) {
    message.error(error.response?.data?.error?.message || '删除 Session 失败')
  }
}

const handleValidate = async (email: string) => {
  try {
    const response = await validateCursorSession(email)
    const result = response.data
    if (result.is_valid) {
      message.success(`Session "${email}" 验证成功：${result.message}`)
    } else {
      message.warning(`Session "${email}" 验证失败：${result.message}`)
    }
    await loadSessions()
  } catch (error: any) {
    message.error(error.response?.data?.error?.message || '验证 Session 失败')
  }
}

// 用户管理方法
async function loadUsers() {
  usersLoading.value = true
  try {
    const response = await usersApi.listUsers()
    users.value = response.users || []
    message.success(`成功加载 ${response.total} 个用户`)
  } catch (error: any) {
    console.error('Failed to load users:', error)
    message.error(error.message || '加载用户列表失败')
  } finally {
    usersLoading.value = false
  }
}

async function handleToggleRole(user: User) {
  const newRole = user.role === 'admin' ? 'user' : 'admin'
  const roleText = newRole === 'admin' ? '管理员' : '普通用户'
  
  dialog.warning({
    title: '确认修改',
    content: `确定要将用户 ${user.username} 的角色修改为 ${roleText} 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await usersApi.updateUserRole(user.id, newRole)
        message.success('用户角色修改成功')
        await loadUsers()
      } catch (error: any) {
        console.error('Failed to update user role:', error)
        message.error(error.message || '修改用户角色失败')
      }
    }
  })
}

async function handleToggleStatus(user: User) {
  const action = user.is_active ? '禁用' : '启用'
  
  dialog.warning({
    title: '确认操作',
    content: `确定要${action}用户 ${user.username} 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await usersApi.toggleUserStatus(user.id)
        message.success(`用户${action}成功`)
        await loadUsers()
      } catch (error: any) {
        console.error('Failed to toggle user status:', error)
        message.error(error.message || `${action}用户失败`)
      }
    }
  })
}

async function handleDeleteUser(user: User) {
  dialog.error({
    title: '确认删除',
    content: `确定要删除用户 ${user.username} 吗？此操作不可恢复。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await usersApi.deleteUser(user.id)
        message.success('用户删除成功')
        await loadUsers()
      } catch (error: any) {
        console.error('Failed to delete user:', error)
        message.error(error.message || '删除用户失败')
      }
    }
  })
}

// 密钥管理方法
async function loadKeys() {
  keysLoading.value = true
  try {
    const response = await listKeys()
    if (response.data && response.data.keys) {
      keys.value = response.data.keys
      message.success(`成功加载 ${response.data.total} 个密钥`)
    } else {
      keys.value = []
    }
  } catch (error: any) {
    console.error('Failed to load keys:', error)
    message.error(error.message || '加载密钥列表失败')
  } finally {
    keysLoading.value = false
  }
}

function handleCopyKey(key: string) {
  navigator.clipboard.writeText(key)
  message.success('完整密钥已复制到剪贴板')
}

async function handleToggleKey(key: AdminKey) {
  const action = key.is_active ? '禁用' : '启用'
  
  dialog.warning({
    title: '确认操作',
    content: `确定要${action}密钥 ${key.masked_key} 吗？${key.username ? `（用户：${key.username}）` : ''}`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await toggleKeyStatus(key.key)
        // 先更新本地状态
        const keyIndex = keys.value.findIndex(k => k.key === key.key)
        if (keyIndex === -1) return
        const targetKey = keys.value[keyIndex]
        if (!targetKey) return
        targetKey.is_active = !targetKey.is_active
        message.success(`密钥${action}成功`)
        // 延迟刷新以确保后端已处理
        setTimeout(() => {
          loadKeys()
        }, 500)
      } catch (error: any) {
        console.error('Failed to toggle key status:', error)
        message.error(error.message || `${action}密钥失败`)
        // 失败时重新加载数据
        loadKeys()
      }
    }
  })
}

async function handleDeleteKey(key: AdminKey) {
  dialog.error({
    title: '确认删除',
    content: `确定要删除密钥 ${key.masked_key} 吗？${key.username ? `（用户：${key.username}）` : ''}\n此操作不可恢复。`,
    positiveText: '删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await removeKey(key.key)
        // 先从本地列表中移除
        keys.value = keys.value.filter(k => k.key !== key.key)
        message.success('密钥删除成功')
        // 延迟刷新以确保后端已处理
        setTimeout(() => {
          loadKeys()
        }, 500)
      } catch (error: any) {
        console.error('Failed to delete key:', error)
        message.error(error.message || '删除密钥失败')
        // 失败时重新加载数据
        loadKeys()
      }
    }
  })
}

// 公告管理方法
async function loadAnnouncements() {
  announcementsLoading.value = true
  try {
    const response = await announcementApi.listAllAnnouncements()
    announcements.value = response.announcements || []
    if (response.total > 0) {
      message.success(`成功加载 ${response.total} 条公告`)
    }
  } catch (error: any) {
    console.error('Failed to load announcements:', error)
    
    // Provide user-friendly error messages
    let errorMessage = '加载公告列表失败'
    if (error.type === 'NETWORK_ERROR') {
      errorMessage = '网络连接失败，请检查网络后重试'
    } else if (error.type === 'UNAUTHORIZED') {
      errorMessage = '登录已过期，请重新登录'
    } else if (error.type === 'PERMISSION_ERROR') {
      errorMessage = '无权限访问，需要管理员权限'
    } else if (error.type === 'SERVER_ERROR') {
      errorMessage = '服务器错误，请稍后重试'
    } else if (error.message) {
      errorMessage = error.message
    }
    
    message.error(errorMessage)
  } finally {
    announcementsLoading.value = false
  }
}

async function handleCreateAnnouncement() {
  try {
    await announcementFormRef.value?.validate()
  } catch {
    message.warning('请填写完整的公告信息')
    return
  }

  announcementSubmitting.value = true
  try {
    await announcementApi.createAnnouncement({
      title: announcementFormData.value.title,
      content: announcementFormData.value.content
    })

    message.success('公告创建成功，所有用户将收到通知')
    showAnnouncementModal.value = false
    announcementFormData.value = {
      title: '',
      content: ''
    }
    await loadAnnouncements()
  } catch (error: any) {
    console.error('Failed to create announcement:', error)
    
    // Provide user-friendly error messages
    let errorMessage = '创建公告失败'
    if (error.type === 'NETWORK_ERROR') {
      errorMessage = '网络连接失败，请检查网络后重试'
    } else if (error.type === 'UNAUTHORIZED') {
      errorMessage = '登录已过期，请重新登录'
    } else if (error.type === 'PERMISSION_ERROR') {
      errorMessage = '无权限创建公告，需要管理员权限'
    } else if (error.type === 'BUSINESS_ERROR') {
      errorMessage = error.message || '公告内容不符合要求'
    } else if (error.type === 'SERVER_ERROR') {
      errorMessage = '服务器错误，请稍后重试'
    } else if (error.message) {
      errorMessage = error.message
    }
    
    message.error(errorMessage)
  } finally {
    announcementSubmitting.value = false
  }
}

async function handleDeleteAnnouncement(announcement: Announcement) {
  dialog.warning({
    title: '确认删除公告',
    content: `确定要删除公告"${announcement.title}"吗？\n\n此操作将：\n• 删除该公告\n• 删除所有用户的阅读记录\n• 此操作不可恢复`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await announcementApi.deleteAnnouncement(announcement.id)
        message.success('公告删除成功')
        await loadAnnouncements()
      } catch (error: any) {
        console.error('Failed to delete announcement:', error)
        
        // Provide user-friendly error messages
        let errorMessage = '删除公告失败'
        if (error.type === 'NETWORK_ERROR') {
          errorMessage = '网络连接失败，请检查网络后重试'
        } else if (error.type === 'UNAUTHORIZED') {
          errorMessage = '登录已过期，请重新登录'
        } else if (error.type === 'PERMISSION_ERROR') {
          errorMessage = '无权限删除公告，需要管理员权限'
        } else if (error.type === 'SERVER_ERROR') {
          errorMessage = '服务器错误，请稍后重试'
        } else if (error.message) {
          errorMessage = error.message
        }
        
        message.error(errorMessage)
      }
    }
  })
}

// 使用统计方法
function formatNumber(num: number): string {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

function getDateRangeParams() {
  if (!usageDateRange.value) return {}
  const [start, end] = usageDateRange.value
  return {
    start_date: new Date(start).toISOString().split('T')[0],
    end_date: new Date(end).toISOString().split('T')[0]
  }
}

async function loadUsageStats() {
  usageLoading.value = true
  try {
    const params = getDateRangeParams()
    console.log('Loading admin usage stats with params:', params)
    const result = await getAdminUsageStats(params)
    console.log('Admin usage stats result:', result)
    usageStats.value = result
  } catch (error: any) {
    console.error('Failed to load usage stats:', error)
    message.error(error.message || '加载使用统计失败')
  } finally {
    usageLoading.value = false
  }
}

async function loadUsageTrends() {
  trendsLoading.value = true
  try {
    const params = {
      days: trendView.value === 'daily' ? 30 : trendView.value === 'weekly' ? 90 : 365,
      view: trendView.value
    }
    console.log('Loading admin usage trends with params:', params)
    const result = await getUsageTrends(params)
    console.log('Admin usage trends result:', result)
    usageTrends.value = result
  } catch (error: any) {
    console.error('Failed to load usage trends:', error)
    message.error(error.message || '加载使用趋势失败')
  } finally {
    trendsLoading.value = false
  }
}

async function loadSessionUsage() {
  sessionUsageLoading.value = true
  try {
    const params = getDateRangeParams()
    console.log('Loading cursor session usage with params:', params)
    const result = await getCursorSessionUsage(params)
    console.log('Cursor session usage result:', result)
    sessionUsage.value = result
  } catch (error: any) {
    console.error('Failed to load session usage:', error)
    message.error(error.message || '加载 Session 使用统计失败')
  } finally {
    sessionUsageLoading.value = false
  }
}

function handleDateRangeChange() {
  loadUsageStats()
  loadSessionUsage()
}

async function handleExportCSV() {
  exportLoading.value = true
  try {
    const params = getDateRangeParams()
    const blob = await exportUsageData(params)
    downloadCSV(blob)
    message.success('导出成功')
  } catch (error: any) {
    console.error('Failed to export usage data:', error)
    message.error(error.message || '导出失败')
  } finally {
    exportLoading.value = false
  }
}

function formatTrendsForChart(trends: { date: string; requests: number; total_tokens: number }[]) {
  return trends.map(t => ({
    date: t.date,
    total_tokens: t.total_tokens,
    prompt_tokens: 0,
    completion_tokens: 0,
    request_count: t.requests
  }))
}

// 余额管理方法
async function loadUserBalances() {
  balancesLoading.value = true
  try {
    const offset = (balancePagination.value.page - 1) * balancePagination.value.pageSize
    const response = await getAllUserBalances({
      limit: balancePagination.value.pageSize,
      offset
    })
    userBalances.value = response.data.users || []
    balancePagination.value.itemCount = response.data.total || 0
    if (response.data.users?.length > 0) {
      message.success(`成功加载 ${response.data.total} 个用户余额`)
    }
  } catch (error: any) {
    console.error('Failed to load user balances:', error)
    message.error(error.message || '加载用户余额失败')
  } finally {
    balancesLoading.value = false
  }
}

function handleBalancePageChange(page: number) {
  balancePagination.value.page = page
  loadUserBalances()
}

function openAdjustBalanceModal(user: UserBalanceInfo) {
  selectedUserForBalance.value = user
  adjustBalanceFormData.value = {
    amount: 0,
    reason: ''
  }
  showAdjustBalanceModal.value = true
}

async function handleAdjustBalance() {
  try {
    await adjustBalanceFormRef.value?.validate()
  } catch {
    message.warning('请填写完整的调整信息')
    return
  }

  if (!selectedUserForBalance.value) {
    message.error('未选择用户')
    return
  }

  adjustBalanceSubmitting.value = true
  try {
    const response = await adjustUserBalance({
      user_id: selectedUserForBalance.value.user_id,
      amount: adjustBalanceFormData.value.amount,
      reason: adjustBalanceFormData.value.reason
    })

    const action = adjustBalanceFormData.value.amount > 0 ? '增加' : '扣除'
    message.success(`余额${action}成功，调整后余额: $${response.data.balance_after.toFixed(2)}`)
    showAdjustBalanceModal.value = false
    adjustBalanceFormData.value = {
      amount: 0,
      reason: ''
    }
    await loadUserBalances()
  } catch (error: any) {
    console.error('Failed to adjust balance:', error)
    
    let errorMessage = '调整余额失败'
    if (error.type === 'NETWORK_ERROR') {
      errorMessage = '网络连接失败，请检查网络后重试'
    } else if (error.type === 'UNAUTHORIZED') {
      errorMessage = '登录已过期，请重新登录'
    } else if (error.type === 'PERMISSION_ERROR') {
      errorMessage = '无权限调整余额，需要管理员权限'
    } else if (error.type === 'SERVER_ERROR') {
      errorMessage = '服务器错误，请稍后重试'
    } else if (error.message) {
      errorMessage = error.message
    }
    
    message.error(errorMessage)
  } finally {
    adjustBalanceSubmitting.value = false
  }
}

// 兑换记录管理方法
async function loadExchangeRecords() {
  exchangeRecordsLoading.value = true
  try {
    const offset = (exchangePagination.value.page - 1) * exchangePagination.value.pageSize
    const params: Record<string, any> = {
      limit: exchangePagination.value.pageSize,
      offset
    }
    
    // Add user_id filter if set
    if (exchangeFilterUserId.value) {
      params.user_id = exchangeFilterUserId.value
    }
    
    // Add date range filter if set
    if (exchangeDateRange.value) {
      const [start, end] = exchangeDateRange.value
      params.start_date = new Date(start).toISOString().split('T')[0]
      params.end_date = new Date(end).toISOString().split('T')[0]
    }
    
    const response = await getAdminExchangeRecords(params)
    exchangeRecords.value = response.data.records || []
    exchangePagination.value.itemCount = response.data.total || 0
    if (response.data.records?.length > 0) {
      message.success(`成功加载 ${response.data.total} 条兑换记录`)
    }
  } catch (error: any) {
    console.error('Failed to load exchange records:', error)
    message.error(error.message || '加载兑换记录失败')
  } finally {
    exchangeRecordsLoading.value = false
  }
}

async function loadExchangeStats() {
  exchangeStatsLoading.value = true
  try {
    const response = await getAdminExchangeStats()
    exchangeStats.value = response.data
  } catch (error: any) {
    console.error('Failed to load exchange stats:', error)
    message.error(error.message || '加载兑换统计失败')
  } finally {
    exchangeStatsLoading.value = false
  }
}

function handleExchangePageChange(page: number) {
  exchangePagination.value.page = page
  loadExchangeRecords()
}

function handleExchangeFilterChange() {
  exchangePagination.value.page = 1
  loadExchangeRecords()
}

onMounted(() => {
  loadSessions()
  loadUsers()
  loadKeys()
  loadAnnouncements()
  loadUserBalances()
  loadExchangeRecords()
  loadExchangeStats()
  loadUsageStats()
  loadUsageTrends()
  loadSessionUsage()
})
</script>

<style scoped>
.admin-panel {
  padding: 2rem;
  animation: fadeIn 0.6s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* Glassmorphism 卡片 - 主题感知 */
.glass-card {
  background: var(--bg-card);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--border-color);
  border-radius: 24px;
  box-shadow: var(--shadow-lg);
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

/* 页面标题 */
.page-header {
  padding: 2.5rem;
  margin-bottom: 2rem;
  text-align: center;
  animation: slideDown 0.8s ease-out;
}

@keyframes slideDown {
  from {
    opacity: 0;
    transform: translateY(-30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.page-header h1 {
  font-size: 2.5rem;
  font-weight: 700;
  margin: 0 0 0.5rem 0;
}

.gradient-text {
  color: var(--text-primary);
}

.subtitle {
  color: var(--text-secondary);
  font-size: 1.1rem;
  margin: 0;
}

/* 标签页样式 */
.admin-tabs {
  animation: fadeInUp 0.6s ease-out 0.2s both;
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

:deep(.n-tabs .n-tabs-nav) {
  background: var(--bg-secondary);
  backdrop-filter: blur(10px);
  -webkit-backdrop-filter: blur(10px);
  border-radius: 16px;
  padding: 0.5rem;
  margin-bottom: 1.5rem;
}

:deep(.n-tabs .n-tabs-tab) {
  color: var(--text-secondary);
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.3s ease;
  border-radius: 12px;
  padding: 0.75rem 1.5rem;
}

:deep(.n-tabs .n-tabs-tab:hover) {
  color: var(--text-primary);
  background: var(--bg-hover);
}

:deep(.n-tabs .n-tabs-tab--active) {
  color: var(--text-primary);
  background: var(--color-primary-light);
}

/* 标签页内容 */
.tab-content {
  padding: 2rem;
  animation: fadeIn 0.4s ease-out;
}

.action-bar {
  margin-bottom: 1.5rem;
}

.section-title {
  color: var(--text-primary);
  font-size: 1.5rem;
  font-weight: 700;
  margin: 0;
}

/* 按钮样式（增强对比度） */
.refresh-btn,
.add-btn {
  transition: all 0.3s ease;
}

.refresh-btn {
  background: var(--color-primary-light) !important;
  border: 1px solid var(--color-primary) !important;
  color: var(--color-primary) !important;
}

.refresh-btn:hover {
  transform: translateY(-2px);
  background: var(--color-primary) !important;
  color: var(--text-inverse) !important;
  box-shadow: var(--shadow-md);
}

.add-btn {
  background: var(--color-primary) !important;
  border: 1px solid var(--color-primary) !important;
  color: var(--text-inverse) !important;
  font-weight: 600 !important;
}

.add-btn:hover {
  transform: translateY(-2px);
  background: var(--color-primary-hover) !important;
  box-shadow: var(--shadow-lg);
}

/* 表格样式 - 主题感知 */
.modern-table {
  background: var(--bg-card);
  border-radius: 16px;
  overflow: hidden;
  border: 1px solid var(--border-color);
}

:deep(.modern-table .n-data-table-wrapper) {
  background: transparent;
}

:deep(.modern-table .n-data-table-th) {
  background: var(--bg-secondary) !important;
  color: var(--text-primary) !important;
  font-weight: 600;
  border-bottom: 1px solid var(--border-color) !important;
}

:deep(.modern-table .n-data-table-td) {
  color: var(--text-secondary) !important;
  border-bottom: 1px solid var(--border-color-light) !important;
  background: var(--bg-card) !important;
}

:deep(.modern-table .n-data-table-tr:hover .n-data-table-td) {
  background: var(--bg-hover) !important;
}

/* 表格内按钮样式 */
:deep(.modern-table .n-button) {
  font-weight: 600 !important;
}

:deep(.modern-table .n-button--primary-type) {
  background: var(--color-primary-light) !important;
  border-color: var(--color-primary) !important;
  color: var(--color-primary) !important;
}

:deep(.modern-table .n-button--error-type) {
  background: var(--color-error-light) !important;
  border-color: var(--color-error) !important;
  color: var(--color-error) !important;
}

:deep(.modern-table .n-button--success-type) {
  background: var(--color-success-light) !important;
  border-color: var(--color-success) !important;
  color: var(--color-success) !important;
}

:deep(.modern-table .n-button--warning-type) {
  background: var(--color-warning-light) !important;
  border-color: var(--color-warning) !important;
  color: var(--color-warning) !important;
}

:deep(.modern-table .n-button:hover) {
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

/* 表格内标签样式 */
:deep(.modern-table .n-tag) {
  font-weight: 600 !important;
  border: 1px solid currentColor !important;
}

:deep(.modern-table .n-tag--success-type) {
  background: var(--color-success-light) !important;
  color: var(--color-success) !important;
}

:deep(.modern-table .n-tag--error-type) {
  background: var(--color-error-light) !important;
  color: var(--color-error) !important;
}

:deep(.modern-table .n-tag--warning-type) {
  background: var(--color-warning-light) !important;
  color: var(--color-warning) !important;
}

:deep(.modern-table .n-tag--info-type) {
  background: var(--color-info-light) !important;
  color: var(--color-info) !important;
}

:deep(.modern-table .n-tag--default-type) {
  background: var(--bg-tertiary) !important;
  color: var(--text-muted) !important;
}

/* 模态框样式 */
:deep(.n-modal) {
  background: var(--bg-card);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid var(--border-color);
}

:deep(.n-dialog) {
  background: var(--bg-card) !important;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .admin-panel {
    padding: 1rem;
  }

  .page-header {
    padding: 1.5rem;
  }

  .page-header h1 {
    font-size: 1.8rem;
  }

  .tab-content {
    padding: 1rem;
  }

  .section-title {
    font-size: 1.2rem;
  }
}

@media (min-width: 1400px) {
  .admin-panel {
    padding: 2rem 4rem;
  }
}

/* 兑换记录统计样式 */
.exchange-stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
  max-width: 500px;
}

/* 使用统计样式 */
.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 1.5rem;
}

.stat-card {
  background: var(--color-primary-light);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  padding: 1.5rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.stat-icon {
  font-size: 2rem;
}

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 1.8rem;
  font-weight: 700;
  color: var(--text-primary);
}

.stat-value.positive {
  color: var(--color-success);
}

.stat-value.negative {
  color: var(--color-error);
}

.stat-label {
  color: var(--text-secondary);
  font-size: 0.9rem;
}

.chart-section {
  padding: 1.5rem;
}

.glass-card-inner {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 16px;
}

.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.chart-title {
  color: var(--text-primary);
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0;
}

.chart-container {
  min-height: 300px;
}

.two-column-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
  gap: 1.5rem;
}

.column-card {
  padding: 1.5rem;
}

.column-title {
  color: var(--text-primary);
  font-size: 1.2rem;
  font-weight: 600;
  margin: 0 0 1rem 0;
}

.inner-table {
  background: transparent;
}

:deep(.inner-table .n-data-table-wrapper) {
  background: transparent;
}

:deep(.inner-table .n-data-table-th) {
  background: var(--bg-secondary) !important;
  color: var(--text-primary) !important;
  font-weight: 600;
}

:deep(.inner-table .n-data-table-td) {
  color: var(--text-secondary) !important;
  background: transparent !important;
}

.session-section {
  padding: 1.5rem;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .two-column-grid {
    grid-template-columns: 1fr;
  }

  .stat-value {
    font-size: 1.4rem;
  }
}
</style>
