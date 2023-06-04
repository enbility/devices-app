import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  root: "./frontend",
  publicDir: "public",
  base: "./",
  build: {
    outDir: "../dist/",
    emptyOutDir: true,
  },
  server: {
    port: 7051,
    proxy: {
      "/ws": { target: "ws://localhost:7050", ws: true },
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
