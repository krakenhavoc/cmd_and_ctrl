import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { serviceWorkerPlugin } from "./vite-plugin-sw";

export default defineConfig({
  // serviceWorkerPlugin is build-only: it stamps the precache list and a
  // content-derived build id into src/sw/service-worker.js and emits /sw.js.
  // Nothing registers a worker in dev (see src/lib/pwa.ts).
  plugins: [svelte(), serviceWorkerPlugin()],
  server: {
    port: 5173,
    strictPort: true,
    proxy: {
      // http:// target + ws:true is the canonical Vite form; http-proxy
      // handles the WebSocket upgrade on top of the HTTP base. A
      // ws:// target confuses HTTP probes hitting /ws in a browser.
      //
      // `changeOrigin: true` rewrites the Host header; it does NOT
      // rewrite Origin. The Go hub's checkOrigin compares Origin
      // against Host and rejects cross-origin upgrades (see
      // server/internal/ws/hub.go:checkOrigin). Without this override
      // the browser's "http://localhost:5173" Origin header survives
      // the proxy and the upgrade 403s. Overriding Origin here makes
      // the dev flow read as same-origin on the Go side without
      // having to set CMDCTRL_ALLOWED_ORIGINS for every developer.
      "/ws": {
        target: "http://localhost:8080",
        ws: true,
        changeOrigin: true,
        headers: { origin: "http://localhost:8080" },
      },
      // Lobby + auth + cards routes — proxied to the Go server so
      // dev-mode clients hit the same URL space as production.
      "/admin": { target: "http://localhost:8080", changeOrigin: true },
      "/games": { target: "http://localhost:8080", changeOrigin: true },
      // The login page's bare-code join (ADR 0050). A sibling of
      // /games/{id}/join, but top-level because the code alone names
      // the table — so /games doesn't cover it and it needs its own
      // entry. This is the same allow-list gap that shipped it broken
      // behind Caddy: without it, POST /join gets Vite's index.html.
      "/join": { target: "http://localhost:8080", changeOrigin: true },
      "/cards": { target: "http://localhost:8080", changeOrigin: true },
      // The public card catalogue (JSON + its scoped image route).
      // Same silent failure as /config if it is missing here: Vite
      // answers /catalog with index.html, the JSON parse throws, and
      // the page renders "couldn't load" in dev only.
      "/catalog": { target: "http://localhost:8080", changeOrigin: true },
      // The public engine roadmap (ADR 0092). The SPA's own #/roadmap
      // is a hash route requested as `/`, so this never shadows it.
      "/roadmap": { target: "http://localhost:8080", changeOrigin: true },
      // The public deck coverage checker + deck requests (ADR 0095
      // §5, #1631). Same silent failure as /catalog above without
      // these: Vite answers with index.html, the JSON parse throws,
      // and #/deck-check renders "couldn't check that deck" in dev
      // only. The SPA's own #/deck-check hash route is unaffected,
      // same reasoning as /roadmap.
      "/deck-coverage": { target: "http://localhost:8080", changeOrigin: true },
      "/deck-requests": { target: "http://localhost:8080", changeOrigin: true },
      "/me": { target: "http://localhost:8080", changeOrigin: true },
      // POST /logout and POST /logout/everywhere (ADR 0051 decision
      // 6). Missing before S34 sub-PR 7: without it Vite answered the
      // POST itself, the server never cleared the cookie, and dev
      // logout only looked like it worked because the client drops its
      // stored copy regardless.
      "/logout": { target: "http://localhost:8080", changeOrigin: true },
      // Deployment identity + dev feature flags (ADR 0023). The client
      // fetches this at shell mount; without the proxy Vite answers
      // with index.html, the JSON parse fails, and env.ts falls back
      // to production — i.e. no dev features in `make client-dev`.
      "/config": { target: "http://localhost:8080", changeOrigin: true },
      // Dev-only tool routes (card search + spawn). Same failure mode
      // as /config but quieter: without this the card search gets
      // Vite's index.html, fails to parse, and searchDevCards's
      // swallow-everything error path renders "No matches" forever.
      "/dev": { target: "http://localhost:8080", changeOrigin: true },
      // S31: the Add-bot picker's options probe (tiers + curated
      // decks). Same failure mode again — without this the picker
      // parses index.html, finds no tiers, and offers nothing.
      "/bot": { target: "http://localhost:8080", changeOrigin: true },
      // The pre-built deck picker's catalog (GET /decks). Same
      // failure mode as /bot: without this the picker parses
      // Vite's index.html, finds no decks and renders nothing, in
      // dev only.
      "/decks": { target: "http://localhost:8080", changeOrigin: true },
      "/healthz": { target: "http://localhost:8080", changeOrigin: true },
      // S12.5: Discord OAuth round-trip. The redirect from Discord
      // lands on /auth/discord/callback; without this proxy the
      // dev browser hits Vite's index.html and the SPA never sees
      // the callback. Mirror the same shape used by the prod
      // reverse proxy (cmd.labxp.io → :8080).
      "/auth": { target: "http://localhost:8080", changeOrigin: true },
      // S12.5: Discord avatar cache, served by the Go side. Without
      // this proxy Vite returns its index.html and the browser
      // tries to render HTML as a PNG — onerror fires and the
      // PlayerHeader latches the failed-avatar state.
      "/avatars": { target: "http://localhost:8080", changeOrigin: true },
    },
  },
});
