import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/Login.vue'),
      meta: { requiresGuest: true }
    },
    {
      path: '/',
      component: () => import('@/layouts/MainLayout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          redirect: '/dashboard'
        },
        {
          path: 'dashboard',
          name: 'Dashboard',
          component: () => import('@/views/Dashboard.vue')
        },
        {
          path: 'usage',
          name: 'Usage',
          component: () => import('@/views/UsageDashboard.vue')
        },
        {
          path: 'tokens',
          name: 'Tokens',
          component: () => import('@/views/TokenManagement.vue')
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/PersonalSettings.vue')
        },
        {
          path: 'referral',
          name: 'Referral',
          component: () => import('@/views/ReferralCenter.vue')
        },
        {
          path: 'models',
          name: 'Models',
          component: () => import('@/views/ModelMarketplace.vue')
        },
        {
          path: 'admin',
          name: 'Admin',
          component: () => import('@/views/AdminPanel.vue'),
          meta: { requiresAdmin: true }
        },
        {
          path: 'docs',
          name: 'Documentation',
          component: () => import('@/views/Documentation.vue')
        },
        // Game Center Routes - Requirements: 4.1, 5.1, 6.1, 7.1
        {
          path: 'games',
          name: 'GameCenter',
          component: () => import('@/views/GameCenter.vue')
        },
        {
          path: 'games/wheel',
          name: 'LuckyWheel',
          component: () => import('@/views/LuckyWheel.vue')
        },
        {
          path: 'games/coin',
          name: 'CoinFlip',
          component: () => import('@/views/CoinFlip.vue')
        },
        {
          path: 'games/number',
          name: 'NumberGuess',
          component: () => import('@/views/NumberGuess.vue')
        },
        // Balance Transfer Route - Requirements: 4.1, 4.2
        {
          path: 'transfer',
          name: 'BalanceTransfer',
          component: () => import('@/views/BalanceTransfer.vue')
        },
        // Chat Routes - Requirements: 5.1
        {
          path: 'chat',
          name: 'Chat',
          component: () => import('@/views/ChatPage.vue')
        },
        {
          path: 'chat/:id',
          name: 'ChatConversation',
          component: () => import('@/views/ChatPage.vue')
        }
      ]
    }
  ]
})

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()
  
  // 只在首次加载时获取用户信息
  if (!authStore.user && !authStore.loading && to.meta.requiresAuth) {
    const success = await authStore.fetchUser()
    
    // 如果获取失败且需要认证，跳转到登录页
    if (!success && to.meta.requiresAuth) {
      return next('/login')
    }
  }
  
  // 检查认证要求
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return next('/login')
  }
  
  // 已登录用户访问登录页，重定向到 dashboard
  if (to.meta.requiresGuest && authStore.isAuthenticated) {
    return next('/dashboard')
  }
  
  // 检查管理员权限
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    return next('/dashboard')
  }
  
  next()
})

export default router
