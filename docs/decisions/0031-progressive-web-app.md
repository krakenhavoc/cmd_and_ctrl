# ADR 0031 — Progressive web app: installable shell, cached card art, nothing else offline

**Status:** Implemented · 2026-09-10 · Branch `feat/pwa`

## Context

The client had no PWA foundation at all: no manifest, no service
worker, no icons, not even a favicon. It is a Vite 6 + Svelte 5 SPA
served as static files by Caddy (`deploy/Caddyfile`), talking to the
Go server over one WebSocket plus a small JSON API.

That shape sets a hard limit on what "offline" can mean here. CMD &
CTRL is a **live multiplayer game whose authoritative state lives on
the server**. There is no local rules engine, no local game state
worth persisting, and no meaningful single-player mode. Anything
that markets itself as offline play would be a lie, and — worse —
anything that *serves a cached API response* would show a player a
table that does not exist.

So the win here is not offline play. It is four other things, in
descending order of how much they are actually worth:

1. **Card art caching.** `/cards/{scryfall_id}/image?size=…` is
   immutable content-addressed art. A four-player Commander board
   plus hands, graveyards and the zone browser can pull hundreds of
   distinct images. Today every one of those is a network round trip
   on every visit.
2. **Installability.** A home-screen launch into a standalone window,
   with the right icon and the right ink behind it.
3. **App-shell caching**, i.e. an instant cold start, and a real
   offline page instead of the browser's dinosaur.
4. **Mobile polish**: safe-area insets, `viewport-fit=cover`,
   `apple-touch-icon`, the iOS standalone metadata that Safari reads
   instead of the manifest.

## Decisions

### 1. Hand-rolled service worker, not `vite-plugin-pwa`

CI runs `npm ci`, which fails hard on a lockfile npm did not
produce, and the branch was prepared on a machine with no Node
toolchain — so adding `vite-plugin-pwa` meant hand-writing lockfile
entries for it *and* its Workbox dependency tree. Hand-editing a
lockfile is a worse risk than owning ~200 lines of service worker.

It is also a better fit. A generator's default is "precache the
build, runtime-cache the rest", i.e. opt-out caching. This app needs
the opposite: almost all of its traffic must never be cached, and
the exceptions are few enough to name individually. The hand-rolled
worker is deny-by-default and the deny list is written down (§3).

The build step is `client/vite-plugin-sw.ts`, ~120 lines: it builds
the precache list from the real bundle output plus a verified list of
files in `public/`, derives a build id from the contents of
everything precached, stamps both into `client/src/sw/service-worker.js`,
and emits the result as `/sw.js` at the dist root. It also fails the
build if `manifest.webmanifest` names an icon that does not exist —
a manifest pointing at a missing icon is silently un-installable.

### 2. Cache versioning is tied to the build, and the app shell can never outlive it

Two caches with deliberately different lifetimes:

| Cache | Name | Lifetime |
|---|---|---|
| App shell | `cmdctrl-shell-<build id>` | Deleted on activate whenever the build id changes |
| Card art | `cmdctrl-cards-v1` | Survives deploys; bounded by an LRU cap of 1200 entries |

The build id is a SHA-256 over the hashed asset filenames, the
contents of every precached static file, and the worker template
itself, truncated to 12 hex characters. It is deterministic: an
unchanged tree rebuilds to the same id and keeps the warm cache;
any precached byte changing invalidates it.

The shell must be build-scoped because the wire protocol is versioned
with the client (`PROTOCOL_VERSION` in `client/src/lib/protocol.ts`).
A stale shell against a newer server is exactly the failure mode
worth engineering against, and it is bounded on both ends: the cache
name changes with the build, and §4's update prompt reappears until
accepted.

The card cache is deliberately *not* build-scoped. A scryfall id
identifies one printing for ever, so dropping the art on every deploy
would re-download a whole battlefield for no reason.

### 3. Routing: deny by default, opt in per route class

`classify()` in `client/src/sw/service-worker.js` is the whole
policy. Anything it does not recognise gets no `respondWith` at all
— the browser performs the request exactly as if no worker were
installed.

| Route class | Strategy |
|---|---|
| Any non-GET method | **Network only** (no `respondWith`) |
| Any cross-origin request, Google Fonts included | **Network only** |
| Any request carrying a `Range` header | **Network only** |
| `/ws` | **Network only** |
| `/healthz`, `/me`, `/logout`, `/games`, `/games/*` | **Network only** |
| `/cards/`, `/cards/*` — except the row below | **Network only** |
| `/admin/*`, `/auth/*`, `/avatars/*`, `/bugreport`, `/bugreport/*` | **Network only** |
| `GET /cards/{id}/image?size=…` | **Cache-first**, card cache, LRU-capped |
| Navigations | **Cache-first** on the precached `/index.html`, falling back to `/offline.html` |
| `/assets/*` (content-hashed build output) | **Cache-first**, shell cache |
| Precached statics: `/offline.html`, `/manifest.webmanifest`, `/icons/favicon.svg`, `/icons/icon-192.png`, `/card-back.jpg`, `/card-back-small.jpg` | **Cache-first**, shell cache |
| Everything else, `/sounds/*.mp3` included | **Network only** |

The API deny list is a regex that mirrors the `@api` path matcher in
`deploy/Caddyfile`, and is deliberately *broader* than it: a bare
`/auth` or `/bugreport` is denied too. Over-denying costs a cache
hit; under-denying corrupts what a player sees.

`/ws` is listed for completeness and defence in depth. Browsers do
not route WebSocket handshakes through a worker's fetch handler at
all, so the worker cannot interfere with the live connection or with
the auto-reconnect in `client/src/lib/ws.ts` even by accident.

The card-image carve-out sits *before* the deny list because the art
lives under `/cards/`. It matches `^/cards/[^/]+/image$` only — the
query string is part of the cache key, since `?size=small`,
`?size=normal` and `?size=art_crop` are different images. `/cards/`
search and metadata endpoints stay network-only.

Responses are only stored if they are a clean `200`: not a redirect
(a cached redirect replays the target under the original URL), not a
`206`, not opaque. Range requests are skipped entirely, because a
cached `200` answered to a ranged request confuses media elements.

These rules are unit-tested against the **real shipped source** —
`client/src/sw/serviceWorker.test.ts` substitutes the build
placeholders exactly as the plugin does, evaluates the worker against
a stub global, and asserts every API route above resolves to
`bypass`.

### 4. The page is never reloaded on our own initiative

`skipWaiting()` is called from exactly one place: a message handler
that `client/src/lib/pwa.ts` triggers when the player clicks
**Reload** in the update prompt. A worker that skipped waiting on
install would reload the page under whoever happened to be holding
priority, mid-combat.

The flow:

1. A new worker installs in the background and enters `waiting`.
2. `pwa.ts` notices (`updatefound` → `statechange`, plus any worker
   already waiting at load) and sets the `updateReady` store.
3. `UpdatePrompt.svelte` shows a small non-modal toast at the bottom
   left: "New version available · Reload when you're not mid-turn",
   with **Reload** and a dismiss ×.
4. **Reload** posts `SKIP_WAITING`, and the page reloads on
   `controllerchange` — gated behind a flag that only the click sets,
   so an update activating for any other reason never reloads a tab.
5. **Dismiss** hides the toast until the next update check finds the
   waiting worker again.

Update checks run on registration, on tab focus (throttled to one
per 5 minutes, since focus changes constantly during play), and on a
30-minute timer. Registration uses `updateViaCache: "none"` so the
browser's HTTP cache cannot pin a client to an old `sw.js`.

The worker is **not registered in dev** — `vite dev` emits no worker,
and a caching layer in front of HMR is a debugging trap. Test it with
`npm run build && npm run preview`.

### 5. Google Fonts stay on the CDN, uncached

The two options were self-hosting the three families or caching the
CDN's opaque responses. Neither is free; we picked "neither", for
now:

- Opaque responses cannot be inspected (their status is unknowable,
  so a cached 404 looks like a cached font), are padded against the
  origin's storage quota, and Google's stylesheet is UA-dependent and
  mutable — caching it can pin a browser to the wrong `woff2` variant.
- Self-hosting is the better long-term answer but adds ~10 font files
  to the repo and touches `app.css` and `index.html`, which other
  work is actively in. It is the top follow-up (§7).

The consequence is honest and small: offline, or with the font host
blocked, type falls back to the metric-close stacks `app.css` already
declares. The app looks plainer; nothing breaks.

### 6. Caddy serves the shell with the right MIME type and no HTTP caching

`deploy/Caddyfile` gains three `header` blocks and **no `try_files`**.

No rewrite is needed because the client routes on `location.hash`
(`client/src/lib/router.ts`), so every deep link — invite links
included — is requested from the origin as `/`. Client deep links do
**not** 404 today, and adding a catch-all rewrite would start
answering unknown paths with `index.html`: the exact silent failure
the `@api` comment in that file already warns about.

- `/`, `/index.html`, `/sw.js`, `/offline.html` → `Cache-Control:
  no-cache`. The deploy rsyncs with `--delete`, so a browser holding
  an old `index.html` in its HTTP cache is pinned to a build whose
  hashed assets no longer exist on the origin.
- `/manifest.webmanifest` → `Content-Type:
  application/manifest+json` plus `no-cache`. Go's mime table does
  not know `.webmanifest` on every host, and `file_server` would fall
  back to sniffing it.
- `/assets/*` → `public, max-age=31536000, immutable`. Vite
  content-hashes those filenames, so they are immutable by
  construction.

Caddy's `header` handler runs before `file_server`, which only sets a
`Content-Type` when one is not already present.

### 7. Icons are generated from SVG, and the PNGs are committed

`client/public/icons/icon.svg` and `maskable.svg` are the source of
truth; the mark is the login badge from `Login.svelte` (a gold
rounded square rotated 45° with a centre pip) on `--bg` `#0b0a09` in
`--gold` `#d9b45c`. The maskable variant is full-bleed and keeps the
mark inside the 40%-radius safe circle.

The PNGs beside them are build products, committed because CI has no
image toolchain and the manifest names them. Regenerate with
`client/tools/gen-icons.sh`, which prefers `rsvg-convert`, then
ImageMagick, then macOS `qlmanage` (which is what produced the
committed set).

## What is deliberately NOT offline

- **Playing.** No local rules engine, no local state. `/offline.html`
  says so in as many words rather than letting the player think the
  app is broken.
- **The lobby, your seat, your decks, the card database, avatars, bug
  reports.** Every one of those is a network-only API route (§3).
- **Reconnecting.** Handled by `ws.ts`'s existing auto-reconnect; the
  worker is not involved and must not be.
- **Audio.** `public/sounds/` is 17MB, most of it one ambient track,
  and audio without a live game is pointless. Network-only, left to
  the browser's HTTP cache.

## Follow-ups (deliberately not in this change)

1. **Self-host the three Google Fonts families** (§5) and drop the
   cross-origin dependency entirely.
2. **Surface the card cache in Settings** — an entry count and a
   "clear cached card art" button, via a worker message.
3. **Warm the card cache from a game snapshot**: on entering a game,
   the client already knows every scryfall id in the visible zones;
   it could prefetch them into the cache during the mulligan rather
   than at first render.
4. **Manifest `screenshots`** for a richer Chromium install dialog,
   and a `shortcuts` entry pointing at `#/lobby`.
5. **An install prompt** (`beforeinstallprompt`), which needs a home
   for the button in the lobby chrome.
6. **Periodic Background Sync / push** for "it's your turn" while the
   tab is backgrounded — the genuinely valuable PWA capability this
   app has not touched, and a design problem of its own.
7. **`display_override: ["window-controls-overlay"]`** for desktop
   installs, once the app chrome has somewhere to put a titlebar.
