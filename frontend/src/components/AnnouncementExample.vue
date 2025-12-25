<!-- 
  Example usage of AnnouncementBell and AnnouncementModal components
  This file demonstrates how to integrate both components together
-->
<template>
  <div class="announcement-example">
    <!-- Bell icon that triggers the modal -->
    <AnnouncementBell 
      ref="bellRef"
      @click="showModal = true" 
    />
    
    <!-- Modal that displays announcements -->
    <AnnouncementModal 
      v-model:show="showModal"
      @read="handleAnnouncementRead"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AnnouncementBell from './AnnouncementBell.vue'
import AnnouncementModal from './AnnouncementModal.vue'

const showModal = ref(false)
const bellRef = ref<InstanceType<typeof AnnouncementBell>>()

// Handle when an announcement is marked as read
function handleAnnouncementRead(announcementId: number) {
  console.log('Announcement marked as read:', announcementId)
  
  // Refresh the unread count in the bell icon
  bellRef.value?.refreshCount()
}
</script>

<style scoped>
.announcement-example {
  display: inline-flex;
}
</style>
