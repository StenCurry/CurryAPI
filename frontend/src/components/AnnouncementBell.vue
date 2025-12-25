<template>
  <div class="announcement-bell">
    <n-badge :value="unreadCount" :max="99" :show="unreadCount > 0">
      <n-button
        text
        :class="{ 'bell-shake': shouldShake }"
        @click="handleClick"
        title="公告通知"
      >
        <template #icon>
          <n-icon size="20">
            <NotificationsOutline />
          </n-icon>
        </template>
      </n-button>
    </n-badge>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { NIcon } from 'naive-ui'
import { NotificationsOutline } from '@vicons/ionicons5'
import { announcementApi } from '@/api/announcement'

// Props
const emit = defineEmits<{
  click: []
}>()

// State - Reactive unread count
const unreadCount = ref(0)
const shouldShake = ref(false)
const isPageVisible = ref(!document.hidden)
let pollingInterval: number | null = null
let previousCount = 0

// Fetch unread count
async function fetchUnreadCount() {
  try {
    const response = await announcementApi.getUnreadCount()
    const newCount = response.count
    
    // Trigger shake animation if count increased
    if (newCount > previousCount && previousCount > 0) {
      triggerShake()
    }
    
    previousCount = newCount
    unreadCount.value = newCount
  } catch (error: any) {
    console.error('Failed to fetch unread count:', error)
    // Silently fail for polling - don't show error messages for background updates
    // Only log network errors, don't disrupt user experience
    if (error.type === 'NETWORK_ERROR') {
      console.warn('Network error while fetching unread count, will retry on next poll')
    }
  }
}

// Optimistic update: decrease count immediately
function decrementCount() {
  if (unreadCount.value > 0) {
    unreadCount.value--
    previousCount = unreadCount.value
  }
}

// Trigger shake animation
function triggerShake() {
  shouldShake.value = true
  setTimeout(() => {
    shouldShake.value = false
  }, 1000)
}

// Handle click event
function handleClick() {
  emit('click')
}

// Start polling
function startPolling() {
  // Don't start if already polling
  if (pollingInterval !== null) {
    return
  }
  
  // Initial fetch
  fetchUnreadCount()
  
  // Poll every 30 seconds
  pollingInterval = window.setInterval(() => {
    fetchUnreadCount()
  }, 30000)
}

// Stop polling
function stopPolling() {
  if (pollingInterval !== null) {
    clearInterval(pollingInterval)
    pollingInterval = null
  }
}

// Handle visibility change - stop polling when page is not visible
function handleVisibilityChange() {
  isPageVisible.value = !document.hidden
  
  if (document.hidden) {
    // Page is hidden, stop polling to save resources
    stopPolling()
  } else {
    // Page is visible again, resume polling and fetch immediately
    startPolling()
  }
}

// Watch page visibility for reactive updates
watch(isPageVisible, (visible) => {
  if (visible) {
    // Fetch immediately when page becomes visible
    fetchUnreadCount()
  }
})

// Lifecycle hooks
onMounted(() => {
  startPolling()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onUnmounted(() => {
  stopPolling()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})

// Expose methods for parent components
defineExpose({
  refreshCount: fetchUnreadCount,
  decrementCount: decrementCount
})
</script>

<style scoped>
.announcement-bell {
  display: inline-flex;
  align-items: center;
}

/* Shake animation for new announcements */
@keyframes bell-shake {
  0%, 100% {
    transform: rotate(0deg);
  }
  10%, 30%, 50%, 70%, 90% {
    transform: rotate(-10deg);
  }
  20%, 40%, 60%, 80% {
    transform: rotate(10deg);
  }
}

.bell-shake {
  animation: bell-shake 0.8s ease-in-out;
}

/* Enhanced hover effect with smooth transitions */
.announcement-bell :deep(.n-button) {
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border-radius: 50%;
  position: relative;
}

.announcement-bell :deep(.n-button:hover) {
  color: #667eea;
  transform: scale(1.15);
  background-color: rgba(102, 126, 234, 0.08);
}

.announcement-bell :deep(.n-button:active) {
  transform: scale(1.05);
  transition: all 0.1s ease;
}

/* Badge styling with fade-in animation */
.announcement-bell :deep(.n-badge-sup) {
  animation: badge-fade-in 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
  font-weight: 600;
  box-shadow: 0 2px 8px rgba(245, 34, 45, 0.3);
}

@keyframes badge-fade-in {
  0% {
    opacity: 0;
    transform: scale(0.3) translateY(-10px);
  }
  50% {
    transform: scale(1.1);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* Pulse animation for badge when count changes */
.announcement-bell :deep(.n-badge-sup) {
  animation: badge-fade-in 0.4s cubic-bezier(0.34, 1.56, 0.64, 1),
             badge-pulse 2s ease-in-out 0.4s infinite;
}

@keyframes badge-pulse {
  0%, 100% {
    box-shadow: 0 2px 8px rgba(245, 34, 45, 0.3);
  }
  50% {
    box-shadow: 0 2px 12px rgba(245, 34, 45, 0.5);
  }
}

/* Mobile responsive adjustments */
@media (max-width: 768px) {
  .announcement-bell :deep(.n-button) {
    padding: 8px;
  }
  
  .announcement-bell :deep(.n-icon) {
    font-size: 18px;
  }
  
  .announcement-bell :deep(.n-badge-sup) {
    font-size: 11px;
    min-width: 16px;
    height: 16px;
    line-height: 16px;
    padding: 0 4px;
  }
}

/* Extra small screens */
@media (max-width: 480px) {
  .announcement-bell :deep(.n-button:hover) {
    transform: scale(1.1);
  }
}
</style>
