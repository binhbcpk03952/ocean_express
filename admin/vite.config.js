import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    allowedHosts: true,
    // Cố định cổng 5173 (đã nằm trong CORS allow-list của backend).
    // strictPort: nếu 5173 bận thì báo lỗi thay vì nhảy sang 5174 rồi bị CORS chặn.
    port: 5173,
    strictPort: true,
    watch: {
      usePolling: true,
    },
  },
})
