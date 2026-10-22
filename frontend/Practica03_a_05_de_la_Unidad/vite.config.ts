import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'
import { VitePWA } from 'vite-plugin-pwa'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react(),
    tailwindcss(),
    VitePWA({
      registerType: 'autoUpdate',
      includeAssets: ['favicon.svg', 'apple-touch-icon.png'],
      manifest: {
        name: 'MultiCatálogo - Tienda Cosmética MLM',
        short_name: 'MultiCatálogo',
        description: 'Tienda cosmética con red de ventas multinivel',
        theme_color: '#4f46e5',
        background_color: '#ffffff',
        display: 'standalone',
        orientation: 'portrait',
        scope: '/',
        start_url: '/',
        icons: [
          {
            src: '/pwa-72x72.svg',
            sizes: '72x72',
            type: 'image/svg+xml',
          },
          {
            src: '/pwa-96x96.svg',
            sizes: '96x96',
            type: 'image/svg+xml',
          },
          {
            src: '/pwa-128x128.svg',
            sizes: '128x128',
            type: 'image/svg+xml',
          },
          {
            src: '/pwa-144x144.svg',
            sizes: '144x144',
            type: 'image/svg+xml',
          },
          {
            src: '/pwa-152x152.svg',
            sizes: '152x152',
            type: 'image/svg+xml',
          },
          {
            src: '/pwa-192x192.svg',
            sizes: '192x192',
            type: 'image/svg+xml',
          },
          {
            src: '/pwa-384x384.svg',
            sizes: '384x384',
            type: 'image/svg+xml',
          },
          {
            src: '/pwa-512x512.svg',
            sizes: '512x512',
            type: 'image/svg+xml',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg}'],
        runtimeCaching: [
          {
            urlPattern: /\.(?:png|jpg|jpeg|svg|gif|webp)$/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'image-cache',
              expiration: {
                maxEntries: 60,
                maxAgeSeconds: 86400, // 24 horas
              },
            },
          },
          {
            urlPattern: /\.(?:js|css)$/i,
            handler: 'CacheFirst',
            options: {
              cacheName: 'static-cache',
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 86400, // 24 horas
              },
            },
          },
        ],
      },
    }),
  ],
  server: {
    host: true,
  },
})
