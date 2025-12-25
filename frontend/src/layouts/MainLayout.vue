<template>
  <div class="main-layout-wrapper">
    <!-- 简洁渐变背景 -->
    <div class="ai-tech-background"></div>

    <n-layout has-sider class="main-layout">
      <!-- 华丽玻璃侧边栏 -->
      <n-layout-sider
        bordered
        collapse-mode="width"
        :collapsed-width="72"
        :width="260"
        :native-scrollbar="false"
        class="glass-sider"
        :class="{ 'sider-collapsed': isCollapsed }"
        :collapsed="isCollapsed"
        @mouseenter="isCollapsed = false"
        @mouseleave="isCollapsed = true"
      >
        <!-- Logo区域 -->
        <div class="logo-container">
          <div class="logo-inner">
            <span class="logo-icon">
              <svg viewBox="0 0 40 40" class="logo-svg">
                <defs>
                  <linearGradient id="logoGrad" x1="0%" y1="0%" x2="100%" y2="100%">
                    <stop offset="0%" stop-color="#818cf8" />
                    <stop offset="50%" stop-color="#6366f1" />
                    <stop offset="100%" stop-color="#8b5cf6" />
                  </linearGradient>
                </defs>
                <circle cx="20" cy="20" r="18" fill="url(#logoGrad)" opacity="0.2"/>
                <circle cx="20" cy="20" r="14" fill="none" stroke="url(#logoGrad)" stroke-width="2"/>
                <path d="M14 20 L18 24 L26 16" stroke="#fff" stroke-width="2.5" fill="none" stroke-linecap="round" stroke-linejoin="round"/>
              </svg>
            </span>
            <span class="logo-text">
              <span class="logo-main">Curry2API</span>
              <span class="logo-sub">Control Panel</span>
            </span>
          </div>
        </div>

        <!-- 导航菜单 -->
        <div class="menu-wrapper">
          <n-menu
            v-model:value="activeKey"
            :collapsed-width="72"
            :collapsed-icon-size="24"
            :options="menuOptions"
            @update:value="handleMenuSelect"
            class="glass-menu"
          />
        </div>

        <!-- 侧边栏底部 -->
        <div class="sider-footer">
          <div class="version-badge">
            <span class="version-dot"></span>
            <span class="version-text">v2.0.0</span>
          </div>
        </div>
      </n-layout-sider>

      <n-layout>
        <!-- 华丽玻璃头部 -->
        <n-layout-header bordered class="glass-header">
          <div class="header-left">
            <div class="breadcrumb">
              <span class="breadcrumb-icon">
                <n-icon :component="currentPageIcon" size="18" />
              </span>
              <h3 class="page-title">{{ pageTitle }}</h3>
            </div>
          </div>

          <div class="header-center">
            <!-- 简化：移除装饰性元素 -->
          </div>

          <div class="header-right">
            <!-- 公告铃铛 -->
            <div class="header-action">
              <AnnouncementBell 
                ref="bellRef"
                @click="showAnnouncementModal = true" 
              />
            </div>

            <!-- 主题切换 -->
            <div class="header-action theme-toggle" @click="toggleTheme">
              <div class="toggle-track" :class="{ 'dark': isDark }">
                <div class="toggle-thumb">
                  <n-icon size="14">
                    <component :is="isDark ? MoonOutline : SunnyOutline" />
                  </n-icon>
                </div>
              </div>
            </div>

            <!-- 用户菜单 -->
            <n-dropdown :options="userMenuOptions" @select="handleUserMenuSelect" placement="bottom-end">
              <div class="user-profile">
                <div class="avatar">
                  <n-icon size="20"><PersonCircleOutline /></n-icon>
                </div>
                <div class="user-info">
                  <span class="user-name">{{ authStore.user?.username }}</span>
                  <span class="user-role">{{ authStore.isAdmin ? '管理员' : '用户' }}</span>
                </div>
                <n-icon size="16" class="dropdown-arrow"><ChevronDownOutline /></n-icon>
              </div>
            </n-dropdown>
          </div>
        </n-layout-header>

        <!-- 内容区域 -->
        <n-layout-content class="main-content">
          <div class="content-wrapper">
            <router-view v-slot="{ Component }">
              <transition name="page-fade" mode="out-in">
                <component :is="Component" />
              </transition>
            </router-view>
          </div>
        </n-layout-content>
      </n-layout>
    </n-layout>

    <!-- 公告弹窗 -->
    <AnnouncementModal 
      v-model:show="showAnnouncementModal"
      @read="handleAnnouncementRead"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NIcon } from 'naive-ui'
import {
  GridOutline,
  KeyOutline,
  SettingsOutline,
  PersonCircleOutline,
  LogOutOutline,
  ShieldCheckmarkOutline,
  MoonOutline,
  SunnyOutline,
  StatsChartOutline,
  GiftOutline,
  AppsOutline,
  GameControllerOutline,
  SwapHorizontalOutline,
  ChatbubblesOutline,
  ChevronDownOutline
} from '@vicons/ionicons5'
import { useAuthStore } from '@/stores/auth'
import { useMessage } from 'naive-ui'
import { useDark } from '@vueuse/core'
import AnnouncementBell from '@/components/AnnouncementBell.vue'
import AnnouncementModal from '@/components/AnnouncementModal.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const message = useMessage()

const activeKey = ref<string>(route.name as string || 'Dashboard')
const showAnnouncementModal = ref(false)
const bellRef = ref<InstanceType<typeof AnnouncementBell> | null>(null)
const isCollapsed = ref(true)

// 主题切换 - 使用 useDark 与 App.vue 保持同步
const isDark = useDark()

function toggleTheme() {
  isDark.value = !isDark.value
  message.success(isDark.value ? '已切换到暗色模式' : '已切换到亮色模式')
}

const pageTitleMap: Record<string, { title: string; icon: any }> = {
  Dashboard: { title: '数据看板', icon: GridOutline },
  Usage: { title: '使用统计', icon: StatsChartOutline },
  Tokens: { title: 'API令牌', icon: KeyOutline },
  Referral: { title: '邀请中心', icon: GiftOutline },
  Models: { title: '模型广场', icon: AppsOutline },
  Settings: { title: '个人设置', icon: SettingsOutline },
  Admin: { title: '管理后台', icon: ShieldCheckmarkOutline },
  GameCenter: { title: '游戏中心', icon: GameControllerOutline },
  LuckyWheel: { title: '幸运转盘', icon: GameControllerOutline },
  CoinFlip: { title: '硬币翻转', icon: GameControllerOutline },
  NumberGuess: { title: '猜数字', icon: GameControllerOutline },
  BalanceTransfer: { title: '余额划转', icon: SwapHorizontalOutline },
  Chat: { title: 'AI 对话', icon: ChatbubblesOutline },
  ChatConversation: { title: 'AI 对话', icon: ChatbubblesOutline }
}

const pageTitle = computed(() => pageTitleMap[activeKey.value]?.title || '控制台')
const currentPageIcon = computed(() => pageTitleMap[activeKey.value]?.icon || GridOutline)

function renderIcon(icon: any) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions = computed(() => {
  const baseOptions = [
    { label: '数据看板', key: 'Dashboard', icon: renderIcon(GridOutline) },
    { label: 'AI 对话', key: 'Chat', icon: renderIcon(ChatbubblesOutline) },
    { label: '使用统计', key: 'Usage', icon: renderIcon(StatsChartOutline) },
    { label: 'API令牌', key: 'Tokens', icon: renderIcon(KeyOutline) },
    { label: '邀请中心', key: 'Referral', icon: renderIcon(GiftOutline) },
    { label: '模型广场', key: 'Models', icon: renderIcon(AppsOutline) },
    {
      label: '游戏中心',
      key: 'games',
      icon: renderIcon(GameControllerOutline),
      children: [
        { label: '游戏大厅', key: 'GameCenter' },
        { label: '幸运转盘', key: 'LuckyWheel' },
        { label: '硬币翻转', key: 'CoinFlip' },
        { label: '猜数字', key: 'NumberGuess' }
      ]
    },
    { label: '余额划转', key: 'BalanceTransfer', icon: renderIcon(SwapHorizontalOutline) },
    { label: '个人设置', key: 'Settings', icon: renderIcon(SettingsOutline) }
  ]

  if (authStore.isAdmin) {
    baseOptions.push({
      label: '管理后台',
      key: 'Admin',
      icon: renderIcon(ShieldCheckmarkOutline)
    })
  }

  return baseOptions
})

const userMenuOptions = [
  { label: '退出登录', key: 'logout', icon: renderIcon(LogOutOutline) }
]

function handleMenuSelect(key: string) {
  activeKey.value = key
  router.push({ name: key })
}

async function handleUserMenuSelect(key: string) {
  if (key === 'logout') {
    await authStore.logout()
    message.success('已退出登录')
    router.push('/login')
  }
}

function handleAnnouncementRead(announcementId: number) {
  if (bellRef.value) {
    if (announcementId === -1) {
      bellRef.value.refreshCount()
    } else {
      bellRef.value.decrementCount()
      setTimeout(() => {
        if (bellRef.value) {
          bellRef.value.refreshCount()
        }
      }, 1000)
    }
  }
}

// 路由变化时更新activeKey
router.afterEach((to) => {
  if (to.name) {
    activeKey.value = to.name as string
  }
})
</script>

<style scoped>
/* ============================================
   布局容器
   ============================================ */
.main-layout-wrapper {
  position: relative;
  width: 100%;
  height: 100vh;
  overflow: hidden;
}

/* ============================================
   侧边栏样式 - 现代玻璃拟态
   ============================================ */
.glass-sider {
  background: var(--bg-card) !important;
  backdrop-filter: var(--backdrop-blur-lg);
  border-right: 1px solid var(--border-color);
  box-shadow: var(--shadow-sm);
  z-index: 50;
  transition: all var(--transition-normal);
}

/* Logo区域 */
.logo-container {
  height: 70px;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0 var(--spacing-md);
  border-bottom: 1px solid var(--border-color);
  background: rgba(255, 255, 255, 0.05);
}

.logo-inner {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;
  overflow: hidden;
}

.logo-icon {
  width: 40px;
  height: 40px;
  flex-shrink: 0;
  filter: drop-shadow(0 0 8px rgba(79, 70, 229, 0.3));
  transition: transform var(--transition-normal);
}

.logo-icon:hover {
  transform: scale(1.05) rotate(5deg);
}

.logo-svg {
  width: 100%;
  height: 100%;
}

.logo-text {
  display: flex;
  flex-direction: column;
  transition: opacity var(--transition-fast), width var(--transition-fast);
  white-space: nowrap;
}

.logo-main {
  font-size: 18px;
  font-weight: 700;
  line-height: 1.2;
  background: linear-gradient(135deg, var(--color-primary) 0%, #8b5cf6 100%);
  -webkit-background-clip: text;
  background-clip: text;
  -webkit-text-fill-color: transparent;
  letter-spacing: -0.5px;
}

.logo-sub {
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 500;
  letter-spacing: 0.5px;
  text-transform: uppercase;
}

/* 导航菜单优化 */
.menu-wrapper {
  padding: var(--spacing-md) var(--spacing-xs);
  flex: 1;
  overflow-y: auto;
}

/* 覆盖 Naive UI 菜单样式 */
:deep(.n-menu .n-menu-item-content) {
  padding-left: 16px !important;
  border-radius: var(--border-radius);
  transition: all var(--transition-fast);
  margin-bottom: 4px;
}

:deep(.n-menu .n-menu-item-content::before) {
  border-radius: var(--border-radius);
  left: 8px;
  right: 8px;
}

:deep(.n-menu .n-menu-item-content:hover) {
  background-color: var(--bg-hover) !important;
}

:deep(.n-menu .n-menu-item-content--selected) {
  background-color: var(--color-primary-light) !important;
}

:deep(.n-menu .n-menu-item-content--selected .n-menu-item-content-header) {
  color: var(--color-primary) !important;
  font-weight: 600;
}

:deep(.n-menu .n-menu-item-content--selected .n-icon) {
  color: var(--color-primary) !important;
}

:deep(.n-menu-item-content__icon) {
  margin-right: 12px !important;
}

/* 侧边栏底部 */
.sider-footer {
  padding: var(--spacing-md);
  border-top: 1px solid var(--border-color);
  background: rgba(255, 255, 255, 0.02);
}

.version-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--bg-secondary);
  border-radius: var(--border-radius-sm);
  border: 1px solid var(--border-color);
  transition: all var(--transition-fast);
}

.version-badge:hover {
  border-color: var(--border-color-hover);
  background: var(--bg-hover);
}

.version-dot {
  width: 6px;
  height: 6px;
  background: var(--color-success);
  border-radius: 50%;
  box-shadow: 0 0 8px var(--color-success);
}

.version-text {
  font-size: 12px;
  color: var(--text-secondary);
  font-weight: 600;
  font-family: monospace;
}

/* 折叠状态样式优化 */
.sider-collapsed .logo-inner {
  justify-content: center;
}

.sider-collapsed .logo-text {
  opacity: 0;
  width: 0;
  display: none;
}

.sider-collapsed .version-text {
  display: none;
}

.sider-collapsed .version-badge {
  padding: 8px;
  background: transparent;
  border-color: transparent;
}

/* ============================================
   头部样式 - 玻璃拟态与交互
   ============================================ */
.glass-header {
  height: 70px;
  padding: 0 var(--spacing-xl);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: rgba(255, 255, 255, 0.8);
  backdrop-filter: var(--backdrop-blur-lg);
  border-bottom: 1px solid var(--border-color);
  position: sticky;
  top: 0;
  z-index: 40;
  transition: all var(--transition-normal);
}

.dark-theme .glass-header {
  background: rgba(15, 23, 42, 0.8);
}

.header-left {
  display: flex;
  align-items: center;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 12px;
  background: var(--bg-secondary);
  border-radius: var(--border-radius-full);
  border: 1px solid var(--border-color);
  transition: all var(--transition-fast);
}

.breadcrumb:hover {
  border-color: var(--border-color-hover);
  box-shadow: var(--shadow-xs);
  transform: translateY(-1px);
}

.breadcrumb-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--bg-primary);
  border-radius: 50%;
  color: var(--color-primary);
  box-shadow: var(--shadow-sm);
}

.page-title {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  padding-right: 4px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.header-action {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  transition: all var(--transition-fast);
  cursor: pointer;
}

.header-action:hover {
  background: var(--bg-hover);
}

/* 主题切换开关优化 */
.theme-toggle {
  width: auto;
  height: auto;
  border-radius: 0;
  background: transparent;
}

.toggle-track {
  width: 52px;
  height: 28px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 14px;
  position: relative;
  transition: all var(--transition-normal);
  cursor: pointer;
  box-shadow: inset 0 2px 4px rgba(0, 0, 0, 0.05);
}

.toggle-track:hover {
  border-color: var(--border-color-hover);
}

.toggle-track.dark {
  background: var(--bg-tertiary);
  border-color: var(--color-primary);
}

.toggle-thumb {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 20px;
  height: 20px;
  background: #ffffff;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all var(--transition-normal);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  color: #f59e0b;
}

.toggle-track.dark .toggle-thumb {
  left: 27px;
  background: var(--color-primary);
  color: #ffffff;
  box-shadow: 0 0 8px var(--color-primary);
}

/* 用户菜单优化 */
.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 12px 4px 4px;
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--border-radius-full);
  cursor: pointer;
  transition: all var(--transition-fast);
}

.user-profile:hover {
  background: var(--bg-primary);
  border-color: var(--color-primary);
  box-shadow: var(--shadow-sm);
  transform: translateY(-1px);
}

.avatar {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, var(--color-primary) 0%, #8b5cf6 100%);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #ffffff;
  box-shadow: var(--shadow-sm);
  font-weight: 600;
}

.user-info {
  display: flex;
  flex-direction: column;
  line-height: 1.2;
}

.user-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.user-role {
  font-size: 11px;
  color: var(--text-muted);
}

.dropdown-arrow {
  color: var(--text-muted);
  transition: transform var(--transition-fast);
}

.user-profile:hover .dropdown-arrow {
  transform: rotate(180deg);
  color: var(--color-primary);
}

/* ============================================
   内容区域 - 增加层次感
   ============================================ */
.main-content {
  background: transparent !important;
  height: calc(100vh - 70px);
  overflow-y: auto;
  overflow-x: hidden;
  position: relative;
}

.content-wrapper {
  padding: var(--spacing-xl);
  max-width: 1600px;
  margin: 0 auto;
  min-height: 100%;
}

/* 页面切换动画 - 更加平滑 */
.page-fade-enter-active,
.page-fade-leave-active {
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
}

.page-fade-enter-from {
  opacity: 0;
  transform: translateY(10px);
  filter: blur(4px);
}

.page-fade-leave-to {
  opacity: 0;
  transform: translateY(-10px);
  filter: blur(4px);
}

/* ============================================
   响应式设计
   ============================================ */
@media (max-width: 768px) {
  .glass-header {
    padding: 0 var(--spacing-md);
    height: 60px;
  }

  .page-title {
    display: none;
  }
  
  .breadcrumb {
    padding: 6px;
    border-radius: 50%;
  }

  .header-center {
    display: none;
  }

  .user-info {
    display: none;
  }
  
  .user-profile {
    padding: 4px;
    border-radius: 50%;
  }
  
  .dropdown-arrow {
    display: none;
  }

  .content-wrapper {
    padding: var(--spacing-md);
  }
}
</style>
