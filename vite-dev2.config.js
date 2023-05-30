import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  root: "./frontend",
  base: "./",
  server: {
    port: 7061,
    proxy: {
      "/ws": { target: "ws://localhost:7060", ws: true },
    },
  },
  plugins: [vue()],
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@import "./frontend/src/styles/main-styles.scss";`
      }
    }
  }
})
