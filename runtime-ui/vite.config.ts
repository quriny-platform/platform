import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

/**
 * Vite configuration for Quriny Runtime UI.
 *
 * - `proxy`: forwards /api, /health, /app-model, /app-graph, and /entities
 *   requests to the Go backend (default :8080) during development. This avoids
 *   CORS issues and keeps the frontend URL clean.
 *
 * - In production, a reverse proxy (nginx, Caddy) would handle this instead.
 */
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    proxy: {
      "/health": "http://localhost:8080",
      "/app-model": "http://localhost:8080",
      "/app-graph": "http://localhost:8080",
      "/entities": "http://localhost:8080",
    },
  },
});
