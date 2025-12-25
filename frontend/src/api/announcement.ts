import client from './client'
import type { Announcement } from '@/types'

export interface CreateAnnouncementRequest {
  title: string
  content: string
}

export interface AnnouncementListResponse {
  total: number
  announcements: Announcement[]
}

export interface UnreadCountResponse {
  count: number
}

export interface MessageResponse {
  message: string
}

// Retry configuration
const MAX_RETRIES = 3
const RETRY_DELAY = 1000 // 1 second

// Helper function to retry failed requests
async function retryRequest<T>(
  fn: () => Promise<T>,
  retries = MAX_RETRIES,
  delay = RETRY_DELAY
): Promise<T> {
  try {
    return await fn()
  } catch (error: any) {
    // Don't retry on client errors (4xx) except 408 (timeout) and 429 (rate limit)
    if (error.originalError?.response?.status) {
      const status = error.originalError.response.status
      if (status >= 400 && status < 500 && status !== 408 && status !== 429) {
        throw error
      }
    }

    // Retry on network errors or server errors
    if (retries > 0 && (error.type === 'NETWORK_ERROR' || error.type === 'SERVER_ERROR')) {
      await new Promise(resolve => setTimeout(resolve, delay))
      return retryRequest(fn, retries - 1, delay * 1.5) // Exponential backoff
    }

    throw error
  }
}

// 管理员接口
export const announcementApi = {
  // 创建公告
  async createAnnouncement(data: CreateAnnouncementRequest): Promise<Announcement> {
    return retryRequest(async () => {
      const response = await client.post('/admin/announcements', data)
      return response.data
    })
  },

  // 获取所有公告列表（管理员）
  async listAllAnnouncements(): Promise<AnnouncementListResponse> {
    return retryRequest(async () => {
      const response = await client.get('/admin/announcements')
      return response.data
    })
  },

  // 删除公告
  async deleteAnnouncement(id: number): Promise<MessageResponse> {
    return retryRequest(async () => {
      const response = await client.delete(`/admin/announcements/${id}`)
      return response.data
    })
  },

  // 获取公告列表（用户）
  async listAnnouncements(limit = 10, offset = 0): Promise<AnnouncementListResponse> {
    return retryRequest(async () => {
      const response = await client.get(`/announcements?limit=${limit}&offset=${offset}`)
      return response.data
    })
  },

  // 获取未读公告数量
  async getUnreadCount(): Promise<UnreadCountResponse> {
    return retryRequest(async () => {
      const response = await client.get('/announcements/unread-count')
      return response.data
    })
  },

  // 标记公告为已读
  async markAsRead(id: number): Promise<MessageResponse> {
    return retryRequest(async () => {
      const response = await client.post(`/announcements/${id}/read`)
      return response.data
    })
  }
}
