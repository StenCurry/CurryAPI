import client from './client'

export interface User {
  id: number
  username: string
  email: string
  role: string
  created_at: string
  last_login?: string
  is_active: boolean
}

export const usersApi = {
  // 列出所有用户
  async listUsers() {
    const response = await client.get('/admin/users')
    return response.data
  },

  // 获取用户信息
  async getUser(id: number) {
    const response = await client.get(`/admin/users/${id}`)
    return response.data
  },

  // 更新用户角色
  async updateUserRole(id: number, role: string) {
    const response = await client.put(`/admin/users/${id}/role`, { role })
    return response.data
  },

  // 启用/禁用用户
  async toggleUserStatus(id: number) {
    const response = await client.put(`/admin/users/${id}/status`)
    return response.data
  },

  // 删除用户
  async deleteUser(id: number) {
    const response = await client.delete(`/admin/users/${id}`)
    return response.data
  }
}
