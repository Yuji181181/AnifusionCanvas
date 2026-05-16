import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          canvas: ['fabric'],
          query: ['@tanstack/react-query'],
          react: ['react', 'react-dom'],
          router: ['@tanstack/react-router'],
        },
      },
    },
  },
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 3000,
  },
})
