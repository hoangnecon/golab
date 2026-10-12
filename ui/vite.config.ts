import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      '/api': 'http://127.0.0.1:9120'
    }
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true
  }
})
