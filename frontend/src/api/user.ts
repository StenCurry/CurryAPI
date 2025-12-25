import apiClient from './client'

export const getProfile = () => apiClient.get('/auth/me')

export const updateUsername = (username: string) =>
  apiClient.put('/profile/username', { username })

export const updatePassword = (oldPassword: string, newPassword: string) =>
  apiClient.put('/profile/password', { old_password: oldPassword, new_password: newPassword })
