import client from './client'
import type { LoginRequest, RegisterRequest, SendCodeRequest, User } from '@/types'

export const authApi = {
  // 登录
  async login(data: LoginRequest) {
    const response = await client.post('/auth/login', data)
    return response.data
  },

  // 注册
  async register(data: RegisterRequest) {
    const response = await client.post('/auth/register', data)
    return response.data
  },

  // 发送验证码
  async sendCode(data: SendCodeRequest) {
    const response = await client.post('/auth/send-code', data)
    return response.data
  },

  // 获取当前用户信息
  async getCurrentUser(): Promise<{ user: User }> {
    const response = await client.get('/auth/me')
    return response.data
  },

  // 登出
  async logout() {
    const response = await client.post('/auth/logout')
    return response.data
  }
}
