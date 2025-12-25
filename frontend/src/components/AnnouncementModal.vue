<template>
  <n-modal
    v-model:show="isVisible"
    preset="card"
    title="系统公告"
    :style="{ width: '600px', maxWidth: '90vw' }"
    :segmented="{ content: true }"
    :closable="true"
    :mask-closable="true"
    @after-leave="handleAfterLeave"
  >
    <div class="announcement-modal-content">
      <!-- Loading state -->
      <div v-if="loading" class="loading-container">
        <n-spin size="medium" />
      </div>

      <!-- Empty state -->
      <n-empty
        v-else-if="announcements.length === 0"
        description="暂无公告"
        size="large"
      />

      <!-- Announcement list -->
      <div v-else class="announcement-list">
          <div
            v-for="announcement in announcements"
            :key="announcement.id"
            class="announcement-item"
            :class="{ 'unread': !announcement.is_read }"
            @click="handleAnnouncementClick(announcement)"
          >
            <div class="announcement-header">
              <h3 class="announcement-title">
                <n-badge v-if="!announcement.is_read" dot type="error" />
                {{ announcement.title }}
              </h3>
              <span class="announcement-date">
                {{ formatDate(announcement.created_at) }}
              </span>
            </div>
            <div class="announcement-content">
              {{ announcement.content }}
            </div>
          </div>
        </div>

      <!-- Load more button -->
      <div v-if="hasMore" class="load-more-container">
        <n-button
          :loading="loadingMore"
          @click="loadMore"
          text
          type="primary"
        >
          加载更多
        </n-button>
      </div>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { NModal, NSpin, NEmpty, NBadge, NButton, useMessage } from 'naive-ui'
import { announcementApi } from '@/api/announcement'
import type { Announcement } from '@/types'

// Props
const props = defineProps<{
  show: boolean
}>()

// Emits
const emit = defineEmits<{
  'update:show': [value: boolean]
  'read': [id: number]
}>()

// State
const message = useMessage()
const announcements = ref<Announcement[]>([])
const loading = ref(false)
const loadingMore = ref(false)
const currentOffset = ref(0)
const pageSize = 10
const total = ref(0)

// Computed
const isVisible = computed({
  get: () => props.show,
  set: (value) => emit('update:show', value)
})

const hasMore = computed(() => {
  return announcements.value.length < total.value
})

// Methods
async function fetchAnnouncements(append = false) {
  try {
    if (append) {
      loadingMore.value = true
    } else {
      loading.value = true
      currentOffset.value = 0
      announcements.value = []
    }

    const response = await announcementApi.listAnnouncements(
      pageSize,
      currentOffset.value
    )

    if (append) {
      announcements.value.push(...response.announcements)
    } else {
      announcements.value = response.announcements || []
    }

    total.value = response.total || 0
  } catch (error: any) {
    console.error('Failed to fetch announcements:', error)
    
    // Ensure announcements is empty array on error
    if (!append) {
      announcements.value = []
      total.value = 0
    }
    
    // Provide user-friendly error messages based on error type
    let errorMessage = '获取公告列表失败'
    if (error.type === 'NETWORK_ERROR') {
      errorMessage = '网络连接失败，请检查网络后重试'
    } else if (error.type === 'UNAUTHORIZED') {
      errorMessage = '登录已过期，请重新登录'
    } else if (error.type === 'SERVER_ERROR') {
      errorMessage = '服务器错误，请稍后重试'
    } else if (error.message) {
      errorMessage = error.message
    }
    
    message.error(errorMessage)
  } finally {
    loading.value = false
    loadingMore.value = false
  }
}

async function handleAnnouncementClick(announcement: Announcement) {
  // If already read, do nothing
  if (announcement.is_read) {
    return
  }

  // Optimistic update: mark as read immediately in UI
  const wasRead = announcement.is_read
  announcement.is_read = true
  
  // Emit read event immediately for optimistic UI update
  emit('read', announcement.id)

  try {
    // Make API call in background
    await announcementApi.markAsRead(announcement.id)
  } catch (error: any) {
    console.error('Failed to mark as read:', error)
    
    // Rollback optimistic update on error
    announcement.is_read = wasRead
    
    // Provide user-friendly error messages
    let errorMessage = '标记已读失败'
    if (error.type === 'NETWORK_ERROR') {
      errorMessage = '网络连接失败，无法标记已读'
    } else if (error.type === 'UNAUTHORIZED') {
      errorMessage = '登录已过期，请重新登录'
    } else if (error.type === 'SERVER_ERROR') {
      errorMessage = '服务器错误，请稍后重试'
    } else if (error.message) {
      errorMessage = error.message
    }
    
    message.error(errorMessage)
    
    // Emit read event again to revert the count
    // This is a negative emit to indicate rollback
    // Parent should handle this by refreshing the actual count
    emit('read', -1)
  }
}

function loadMore() {
  currentOffset.value += pageSize
  fetchAnnouncements(true)
}

function formatDate(dateString: string): string {
  const date = new Date(dateString)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  
  // Less than 1 minute
  if (diff < 60000) {
    return '刚刚'
  }
  
  // Less than 1 hour
  if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return `${minutes}分钟前`
  }
  
  // Less than 1 day
  if (diff < 86400000) {
    const hours = Math.floor(diff / 3600000)
    return `${hours}小时前`
  }
  
  // Less than 7 days
  if (diff < 604800000) {
    const days = Math.floor(diff / 86400000)
    return `${days}天前`
  }
  
  // Format as date
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  
  return `${year}-${month}-${day} ${hours}:${minutes}`
}

function handleAfterLeave() {
  // Reset state when modal is closed
  currentOffset.value = 0
}

// Watch for modal visibility changes
watch(() => props.show, (newValue) => {
  if (newValue) {
    fetchAnnouncements()
  }
})

// Expose methods for parent components
defineExpose({
  refresh: () => fetchAnnouncements()
})
</script>

<style scoped>
.announcement-modal-content {
  min-height: 200px;
  max-height: 60vh;
  overflow-y: auto;
  padding: 4px;
}

.loading-container {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 200px;
  padding: 40px 0;
}

.announcement-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

/* Staggered animation for list items */
.announcement-item {
  padding: 16px;
  border-radius: 8px;
  background-color: #f8f9fa;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid transparent;
  animation: item-slide-in 0.4s ease-out backwards;
}

/* Stagger animation delay for each item */
.announcement-item:nth-child(1) { animation-delay: 0.05s; }
.announcement-item:nth-child(2) { animation-delay: 0.1s; }
.announcement-item:nth-child(3) { animation-delay: 0.15s; }
.announcement-item:nth-child(4) { animation-delay: 0.2s; }
.announcement-item:nth-child(5) { animation-delay: 0.25s; }
.announcement-item:nth-child(n+6) { animation-delay: 0.3s; }

@keyframes item-slide-in {
  from {
    opacity: 0;
    transform: translateX(-20px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

/* Enhanced hover effect with smooth transitions */
.announcement-item:hover {
  background-color: #e9ecef;
  border-color: #667eea;
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.2);
}

.announcement-item:active {
  transform: translateY(0);
  transition: all 0.1s ease;
}

/* Unread announcement styling with gradient */
.announcement-item.unread {
  background: linear-gradient(135deg, #e8f4fd 0%, #f0f8ff 100%);
  border-color: #91caff;
  position: relative;
}

.announcement-item.unread::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 4px;
  background: linear-gradient(180deg, #667eea 0%, #91caff 100%);
  border-radius: 8px 0 0 8px;
  animation: unread-pulse 2s ease-in-out infinite;
}

@keyframes unread-pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.6;
  }
}

.announcement-item.unread:hover {
  background: linear-gradient(135deg, #d6ebff 0%, #e8f4fd 100%);
  border-color: #667eea;
}

.announcement-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
  gap: 12px;
}

.announcement-title {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #2c3e50;
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  transition: color 0.2s ease;
}

.announcement-item:hover .announcement-title {
  color: #667eea;
}

.announcement-date {
  font-size: 12px;
  color: #8c8c8c;
  white-space: nowrap;
  flex-shrink: 0;
  transition: color 0.2s ease;
}

.announcement-item:hover .announcement-date {
  color: #667eea;
}

.announcement-content {
  font-size: 14px;
  color: #5c5c5c;
  line-height: 1.6;
  word-break: break-word;
  white-space: pre-wrap;
  transition: color 0.2s ease;
}

.announcement-item:hover .announcement-content {
  color: #2c3e50;
}

.load-more-container {
  display: flex;
  justify-content: center;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e8e8e8;
  animation: fade-in 0.3s ease-in;
}

@keyframes fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* Enhanced scrollbar styling */
.announcement-modal-content::-webkit-scrollbar {
  width: 8px;
}

.announcement-modal-content::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 4px;
  margin: 4px 0;
}

.announcement-modal-content::-webkit-scrollbar-thumb {
  background: linear-gradient(180deg, #c1c1c1 0%, #a8a8a8 100%);
  border-radius: 4px;
  transition: background 0.2s ease;
}

.announcement-modal-content::-webkit-scrollbar-thumb:hover {
  background: linear-gradient(180deg, #a8a8a8 0%, #8c8c8c 100%);
}

/* Smooth modal entrance animation */
:deep(.n-card) {
  animation: modal-fade-in 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes modal-fade-in {
  0% {
    opacity: 0;
    transform: scale(0.9) translateY(-30px);
  }
  100% {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

/* Modal backdrop animation */
:deep(.n-modal-mask) {
  animation: backdrop-fade-in 0.3s ease-out;
}

@keyframes backdrop-fade-in {
  from {
    opacity: 0;
  }
  to {
    opacity: 1;
  }
}

/* Empty state styling */
:deep(.n-empty) {
  padding: 40px 0;
  animation: fade-in 0.5s ease-in;
}

/* Mobile responsive adjustments */
@media (max-width: 768px) {
  .announcement-modal-content {
    max-height: 70vh;
  }
  
  .announcement-item {
    padding: 12px;
  }
  
  .announcement-title {
    font-size: 15px;
  }
  
  .announcement-content {
    font-size: 13px;
  }
  
  .announcement-date {
    font-size: 11px;
  }
  
  /* Reduce hover effects on mobile */
  .announcement-item:hover {
    transform: none;
    box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
  }
}

/* Extra small screens */
@media (max-width: 480px) {
  .announcement-modal-content {
    max-height: 75vh;
  }
  
  .announcement-item {
    padding: 10px;
    border-radius: 6px;
  }
  
  .announcement-header {
    flex-direction: column;
    gap: 4px;
  }
  
  .announcement-date {
    align-self: flex-start;
  }
  
  .announcement-title {
    font-size: 14px;
  }
  
  .announcement-content {
    font-size: 12px;
    line-height: 1.5;
  }
  
  .announcement-modal-content::-webkit-scrollbar {
    width: 4px;
  }
}

/* Tablet landscape */
@media (min-width: 769px) and (max-width: 1024px) {
  .announcement-modal-content {
    max-height: 65vh;
  }
}

/* Loading state animation */
:deep(.n-spin-container) {
  transition: opacity 0.3s ease;
}

/* Badge dot animation */
:deep(.n-badge .n-badge-sup) {
  animation: badge-dot-pulse 1.5s ease-in-out infinite;
}

@keyframes badge-dot-pulse {
  0%, 100% {
    transform: scale(1);
  }
  50% {
    transform: scale(1.2);
  }
}
</style>
