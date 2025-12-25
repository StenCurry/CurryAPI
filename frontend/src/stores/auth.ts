import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types'
import { authApi } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const loading = ref(false)

  const isAuthenticated = computed(() => !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')

  async function fetchUser() {
    try {
      loading.value = true
      const data = await authApi.getCurrentUser()
      user.value = data.user
      return true
    } catch (error) {
      user.value = null
      return false
    } finally {
      loading.value = false
    }
  }

  function setUser(userData: User) {
    user.value = userData
  }

  function clearUser() {
    user.value = null
  }

  async function logout() {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      clearUser()
    }
  }

  return {
    user,
    loading,
    isAuthenticated,
    isAdmin,
    fetchUser,
    setUser,
    clearUser,
    logout
  }
})
