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
      registerType: 'prompt',
      injectRegister: false,
      includeAssets: ['favicon.svg', 'apple-touch-icon.png'],
      manifest: {
        name: 'MultiCatálogo - Tienda Cosmética MLM',
        short_name: 'MultiCat',
        description: 'Tienda cosmética con red de ventas multinivel',
        lang: 'es-EC',
        theme_color: '#4f46e5',
        background_color: '#ffffff',
        display: 'standalone',
        scope: '/',
        start_url: '/',
        icons: [
          {
            src: '/pwa-192x192.png',
            sizes: '192x192',
            type: 'image/png',
            purpose: 'any',
          },
          {
            src: '/pwa-512x512.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'any',
          },
          {
            src: '/pwa-512x512-maskable.png',
            sizes: '512x512',
            type: 'image/png',
            purpose: 'maskable',
          },
        ],
      },
      workbox: {
        globPatterns: ['**/*.{js,css,html,ico,png,svg,webmanifest,woff,woff2,ttf,eot}'],
        navigateFallback: '/index.html',
        navigateFallbackDenylist: [/^\/api\//],
        runtimeCaching: [
          {
            urlPattern: ({ request, url }) => request.method === 'GET'
              && request.destination === 'image'
              && (/\.(?:png|jpg|jpeg|svg|gif|webp|avif)$/i.test(url.pathname)
                || url.hostname === 'picsum.photos'
                || url.hostname.endsWith('.picsum.photos')),
            handler: 'CacheFirst',
            options: {
              cacheName: 'image-cache',
              cacheableResponse: { statuses: [0, 200] },
              expiration: {
                maxEntries: 60,
                maxAgeSeconds: 2592000, // 30 días
              },
            },
          },
          {
            urlPattern: ({ request, url }) => request.method === 'GET'
              && !request.headers.has('Authorization')
              && /^\/api\/productos(?:\/\d+)?$/.test(url.pathname),
            handler: 'NetworkFirst',
            options: {
              cacheName: 'public-product-catalog',
              networkTimeoutSeconds: 3,
              cacheableResponse: { statuses: [200] },
              expiration: {
                maxEntries: 50,
                maxAgeSeconds: 86400,
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
