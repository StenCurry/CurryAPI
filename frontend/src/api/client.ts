import axios from 'axios'
import type { ApiResponse } from '@/types'

const client = axios.create({
  baseURL: import.meta.env.DEV 
    ? '' // 开发环境：使用相对路径，通过 Vite 代理
    : import.meta.env.VITE_API_BASE_URL || 'http://localhost:8002', // 生产环境：使用环境变量
  withCredentials: true, // 携带 Cookie
  headers: {
    'Content-Type': 'application/json',
    'Cache-Control': 'no-cache, no-store, must-revalidate',
    'Pragma': 'no-cache'
  }
})

// Request interceptor to add cache-busting timestamp
client.interceptors.request.use(
  (config) => {
    // Add timestamp to prevent caching for all requests
    config.params = {
      ...config.params,
      _t: Date.now()
    }
    
    // Add no-cache headers for all requests
    config.headers = {
      ...config.headers,
      'Cache-Control': 'no-cache, no-store, must-revalidate',
      'Pragma': 'no-cache',
      'Expires': '0'
    }
    
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for enhanced error handling
client.interceptors.response.use(
  (response) => response,
  (error) => {
    // Network error - backend service unavailable
    if (!error.response) {
      console.error('Network Error:', error)
      // Note: message notification should be handled at component level
      // to avoid circular dependency with router
      return Promise.reject({
        type: 'NETWORK_ERROR',
        message: '网络连接失败，请检查网络设置或后端服务是否可用',
        originalError: error
      })
    }

    const { response, config } = error
    const status = response.status

    // Handle different HTTP status codes
    switch (status) {
      case 401:
        // 所有 401 错误都不自动重定向，让调用方或路由守卫处理
        console.warn('Unauthorized access:', config.url)
        return Promise.reject({
          type: 'UNAUTHORIZED',
          message: '未登录或会话已过期',
          originalError: error
        })
      
      case 403:
        // Forbidden - no permission
        console.error('Forbidden: No permission to access this resource')
        return Promise.reject({
          type: 'PERMISSION_ERROR',
          message: '无权限访问此资源',
          originalError: error
        })
      
      case 400:
        // Bad Request - business logic error
        const errorMsg = response.data?.error?.message || '请求参数错误'
        console.error('Bad Request:', errorMsg)
        return Promise.reject({
          type: 'BUSINESS_ERROR',
          message: errorMsg,
          originalError: error
        })
      
      case 402:
        // Payment Required - insufficient balance
        // Requirements: 6.2 - Display balance warning
        console.warn('Insufficient balance:', config.url)
        return Promise.reject({
          type: 'INSUFFICIENT_BALANCE',
          message: response.data?.error?.message || '余额不足，请充值后再试',
          originalError: error
        })
      
      case 404:
        // Not Found
        console.warn('Resource not found:', config.url)
        return Promise.reject({
          type: 'NOT_FOUND',
          message: response.data?.error?.message || '请求的资源不存在',
          originalError: error
        })
      
      case 502:
        // Bad Gateway - AI service unavailable
        console.error('AI service unavailable:', config.url)
        return Promise.reject({
          type: 'SERVICE_UNAVAILABLE',
          message: response.data?.error?.message || 'AI 服务暂时不可用，请稍后重试',
          originalError: error
        })
      
      case 504:
        // Gateway Timeout - AI service timeout
        console.error('AI service timeout:', config.url)
        return Promise.reject({
          type: 'SERVICE_TIMEOUT',
          message: response.data?.error?.message || '请求超时，请稍后重试',
          originalError: error
        })
      
      case 500:
        // Internal Server Error
        console.error('Server Error:', response.data)
        return Promise.reject({
          type: 'SERVER_ERROR',
          message: '服务器错误，请稍后重试或联系管理员',
          originalError: error
        })
      
      default:
        // Other errors
        console.error(`HTTP Error ${status}:`, response.data)
        return Promise.reject({
          type: 'UNKNOWN_ERROR',
          message: response.data?.error?.message || '请求失败，请稍后重试',
          originalError: error
        })
    }

    return Promise.reject(error)
  }
)

export default client

// Helper function for API responses
export function handleApiResponse<T>(response: any): ApiResponse<T> {
  return response.data as ApiResponse<T>
}
