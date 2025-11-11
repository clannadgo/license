import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    headers: {
      'Content-Security-Policy': "default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline' blob:; style-src 'self' 'unsafe-inline' blob:; img-src 'self' data: blob:; font-src 'self' data: blob:; connect-src 'self' ws: wss: http: https:; worker-src 'self' blob:; child-src 'self' blob:; object-src 'none'; frame-src 'none';"
    }
  },
  preview: {
    headers: {
      'Content-Security-Policy': "default-src 'self'; script-src 'self' 'unsafe-eval' 'unsafe-inline' blob:; style-src 'self' 'unsafe-inline' blob:; img-src 'self' data: blob:; font-src 'self' data: blob:; connect-src 'self' ws: wss: http: https:; worker-src 'self' blob:; child-src 'self' blob:; object-src 'none'; frame-src 'none';"
    }
  }
})
