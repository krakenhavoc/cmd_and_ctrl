import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import { serviceWorkerPlugin } from "./vite-plugin-sw";

// The Go server's port and this dev server's own. They default to
// `make server-dev`'s 8080 and Vite's 5173, so `make client-dev` is
// unchanged. The nightly E2E's Playwright shards set a distinct pair
// each, because the self-hosted runners share one host (#2661);
// tests-e2e/tests/env.ts reads the same two variables.
function portFromEnv(name: string, fallback: number): number {
  const raw = process.env[name];
  if (raw === undefined || raw === "") return fallback;
  const port = Number(raw);
  if (!Number.isInteger(port) || port <= 0 || port > 65535) {
    throw new Error(`${name}=${raw} is not a port`);
  }
  return port;
}

const clientPort = portFromEnv("CMDCTRL_DEV_CLIENT_PORT", 5173);
const goServer = `http://localhost:${portFromEnv("CMDCTRL_DEV_SERVER_PORT", 8080)}`;

export default defineConfig({
  // serviceWorkerPlugin is build-only: it stamps the precache list and a
  // content-derived build id into src/sw/service-worker.js and emits /sw.js.
  // Nothing registers a worker in dev (see src/lib/pwa.ts).
  plugins: [svelte(), serviceWorkerPlugin()],
  server: {
    port: clientPort,
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
        target: goServer,
        ws: true,
        changeOrigin: true,
        headers: { origin: goServer },
      },
      // Lobby + auth + cards routes — proxied to the Go server so
      // dev-mode clients hit the same URL space as production.
      "/admin": { target: goServer, changeOrigin: true },
      "/games": { target: goServer, changeOrigin: true },
      // The login page's bare-code join (ADR 0050). A sibling of
      // /games/{id}/join, but top-level because the code alone names
      // the table — so /games doesn't cover it and it needs its own
      // entry. This is the same allow-list gap that shipped it broken
      // behind Caddy: without it, POST /join gets Vite's index.html.
      "/join": { target: goServer, changeOrigin: true },
      "/cards": { target: goServer, changeOrigin: true },
      // The public card catalogue (JSON + its scoped image route).
      // Same silent failure as /config if it is missing here: Vite
      // answers /catalog with index.html, the JSON parse throws, and
      // the page renders "couldn't load" in dev only.
      "/catalog": { target: goServer, changeOrigin: true },
      // The public engine roadmap (ADR 0092). The SPA's own #/roadmap
      // is a hash route requested as `/`, so this never shadows it.
      "/roadmap": { target: goServer, changeOrigin: true },
      // The public deck coverage checker + deck requests (ADR 0095
      // §5, #1631). Same silent failure as /catalog above without
      // these: Vite answers with index.html, the JSON parse throws,
      // and #/deck-check renders "couldn't check that deck" in dev
      // only. The SPA's own #/deck-check hash route is unaffected,
      // same reasoning as /roadmap.
      "/deck-coverage": { target: goServer, changeOrigin: true },
      "/deck-requests": { target: goServer, changeOrigin: true },
      "/me": { target: goServer, changeOrigin: true },
      // POST /logout and POST /logout/everywhere (ADR 0051 decision
      // 6). Missing before S34 sub-PR 7: without it Vite answered the
      // POST itself, the server never cleared the cookie, and dev
      // logout only looked like it worked because the client drops its
      // stored copy regardless.
      "/logout": { target: goServer, changeOrigin: true },
      // Deployment identity + dev feature flags (ADR 0023). The client
      // fetches this at shell mount; without the proxy Vite answers
      // with index.html, the JSON parse fails, and env.ts falls back
      // to production — i.e. no dev features in `make client-dev`.
      "/config": { target: goServer, changeOrigin: true },
      // Dev-only tool routes (card search + spawn). Same failure mode
      // as /config but quieter: without this the card search gets
      // Vite's index.html, fails to parse, and searchDevCards's
      // swallow-everything error path renders "No matches" forever.
      "/dev": { target: goServer, changeOrigin: true },
      // S31: the Add-bot picker's options probe (tiers + curated
      // decks). Same failure mode again — without this the picker
      // parses index.html, finds no tiers, and offers nothing.
      "/bot": { target: goServer, changeOrigin: true },
      // The pre-built deck picker's catalog (GET /decks). Same
      // failure mode as /bot: without this the picker parses
      // Vite's index.html, finds no decks and renders nothing, in
      // dev only.
      "/decks": { target: goServer, changeOrigin: true },
      "/healthz": { target: goServer, changeOrigin: true },
      // S12.5: Discord OAuth round-trip. The redirect from Discord
      // lands on /auth/discord/callback; without this proxy the
      // dev browser hits Vite's index.html and the SPA never sees
      // the callback. Mirror the same shape used by the prod
      // reverse proxy (cmd.labxp.io → :8080).
      "/auth": { target: goServer, changeOrigin: true },
      // S12.5: Discord avatar cache, served by the Go side. Without
      // this proxy Vite returns its index.html and the browser
      // tries to render HTML as a PNG — onerror fires and the
      // PlayerHeader latches the failed-avatar state.
      "/avatars": { target: goServer, changeOrigin: true },
      // ADR 0128: a player's playmat image, served by the Go side. Same
      // reason as /avatars: without the proxy Vite answers with its
      // index.html and the board's background never loads.
      "/playmats": { target: goServer, changeOrigin: true },
    },
  },
});
