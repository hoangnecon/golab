import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

const DEV_PORT = 5173;
const DAEMON = 'http://127.0.0.1:9120';
/** Origins of this dev server itself; the daemon only accepts its own origin. */
const DEV_ORIGINS = new Set([`http://127.0.0.1:${DEV_PORT}`, `http://localhost:${DEV_PORT}`]);

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) },
  },
  server: {
    port: DEV_PORT,
    proxy: {
      '/api': {
        target: DAEMON,
        // The string shorthand implies this; the object form does not.
        changeOrigin: true,
        configure: (proxy) => {
          // The proxy rewrites Host to the daemon but keeps the browser's Origin,
          // which the daemon's same-origin guard rejects. Rewrite it only for
          // this dev server's own pages: a foreign Origin must still be refused.
          proxy.on('proxyReq', (proxyReq, req) => {
            const origin = req.headers.origin;
            if (origin && DEV_ORIGINS.has(origin)) proxyReq.setHeader('Origin', DAEMON);
          });
        },
      },
    },
  },
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
});
