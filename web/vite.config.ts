import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// Vite build/dev config. Vitest reads `vitest.config.ts` separately to avoid a
// Vitest↔Vite type-version mismatch when both configs live in one file.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    // Proxy /api/* to the Go backend during local dev so the frontend can
    // call a same-origin URL and skip CORS concerns entirely.
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
});
