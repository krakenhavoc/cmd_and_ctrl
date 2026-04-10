import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";

export default defineConfig({
  plugins: [svelte()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      // http:// target + ws:true is the canonical Vite form; http-proxy
      // handles the WebSocket upgrade on top of the HTTP base. A
      // ws:// target confuses HTTP probes hitting /ws in a browser.
      "/ws": {
        target: "http://localhost:8080",
        ws: true,
        changeOrigin: true,
      },
    },
  },
});
