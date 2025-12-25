import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    proxy: {
      // API 请求代理
      '/api': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/v1': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/auth': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/profile': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/admin/keys': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/admin/cursor': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/admin/users': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/admin/announcements': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/admin/usage': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/admin/balance': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/admin/exchanges': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/announcements': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
      '/health': {
        target: 'http://localhost:8002',
        changeOrigin: true,
        secure: false,
      },
    },
  },
})
