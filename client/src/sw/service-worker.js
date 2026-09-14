// CMD & CTRL service worker.
//
// This file is a TEMPLATE. vite-plugin-sw.ts stamps the build id and the
// precache list into it at build time and emits the result as /sw.js at the
// dist root. It is never imported by the app bundle and never runs in `vite
// dev` -- lib/pwa.ts only registers it from a production build.
//
// DESIGN RULE, and the only one that really matters here: this worker is a
// DENY-BY-DEFAULT proxy. Anything not explicitly recognised below is left
// entirely alone -- no respondWith, no cache lookup, straight to the network.
// CMD & CTRL is a live multiplayer game whose authoritative state lives on the
// server; a single stale /games or /me response would show a player a table
// that does not exist. So caching is opt-in, per route class, and the API
// surface is additionally named in an explicit deny list that mirrors the
// @api matcher in deploy/Caddyfile.
//
// See docs/decisions/0031-progressive-web-app.md.

// BUILD_ID is substituted by vite-plugin-sw.ts with a hash over every
// precached file's contents plus this template. Same inputs, same id: a
// rebuild that changes nothing keeps the cache, and a rebuild that changes
// anything drops it. The shell cache name carries it, so an old app shell can
// never outlive its build -- which matters because the wire protocol
// (PROTOCOL_VERSION in lib/protocol.ts) is versioned with the client.
const BUILD_ID = "__BUILD_ID__";

// PRECACHE is replaced wholesale with the generated array of absolute paths.
// The literal below is a valid placeholder so this file stays lintable and
// unit-testable as-is.
const PRECACHE = ["__PRECACHE__"];

// Two caches, deliberately with different lifetimes.
//
// The shell cache is build-scoped: activate() deletes every shell cache whose
// name is not this build's.
//
// The card cache is NOT build-scoped. /cards/{scryfall_id}/image is immutable
// content-addressed art -- a given scryfall id is that card's printing for
// ever -- so throwing it away on every deploy would re-download a whole
// battlefield's worth of images for no reason. It is bounded by an LRU cap
// instead of by the build.
const SHELL_CACHE = "cmdctrl-shell-" + BUILD_ID;
const CARD_CACHE = "cmdctrl-cards-v1";

// A 4-player Commander board plus hands, graveyards and the zone browser can
// easily reference several hundred distinct printings across three sizes
// (small / normal / art_crop). 1200 entries covers a few games' worth of decks
// at roughly 40-80KB each, i.e. a few tens of MB -- comfortably inside a
// typical origin quota, and evicted oldest-first when exceeded.
const CARD_CACHE_LIMIT = 1200;

// Trimming walks every key in the cache, so do it every N writes rather than
// on each one; a mass board reveal writes dozens of images in a burst.
const CARD_TRIM_INTERVAL = 50;

const SELF_ORIGIN = self.location.origin;

// Immutable, content-addressed card art. Checked BEFORE the API deny list
// because it lives under /cards/, which the deny list otherwise covers whole.
// The query string (?size=small|normal|art_crop) is part of the cache key --
// the sizes are different images.
// /catalog/image/<id> is the public catalogue's own image route. Same
// bytes, same immutability, same cache -- it exists only because
// /cards/<id>/image is session-gated and the catalogue page is public
// (see server/internal/catalog/http.go). Tested before API_PATH below,
// which is what keeps "catalog" in that deny list from swallowing it.
const CARD_IMAGE_PATH = /^\/(?:cards\/[^/]+\/image|catalog\/image\/[^/]+)$/;

// The API deny list. Mirrors the @api path matcher in deploy/Caddyfile, which
// in turn mirrors the server's route table. Deliberately broader than that
// matcher (a bare /auth or /bugreport is denied too): over-denying costs a
// cache hit, under-denying corrupts game state.
//
// /ws is listed for completeness. Browsers do not route WebSocket handshakes
// through a service worker's fetch handler at all, so this worker cannot
// interfere with the live connection even by accident.
//
// /config and /dev are included ahead of PR #260, which adds them to the @api
// matcher: denying a route that does not exist yet costs nothing, and the
// alternative is a client that starts serving a stale /config from cache the
// day the route lands.
const API_PATH =
  /^\/(ws|healthz|me|logout|config|games|cards|admin|auth|avatars|bugreport|dev|bot|catalog|decks)(\/|$)/;

// Hashed build output. Vite content-hashes these filenames, so a given URL's
// bytes never change and cache-first is always correct.
const HASHED_ASSET_PATH = /^\/assets\//;

/**
 * classify decides how one request is handled. It is pure and takes
 * primitives so it can be unit-tested without a service worker environment
 * (see serviceWorker.test.ts).
 *
 * @param {URL} url      the parsed request URL
 * @param {string} method HTTP method
 * @param {string} mode   Request.mode ("navigate", "cors", "no-cors", ...)
 * @returns {"bypass"|"card-image"|"shell"|"asset"}
 */
function classify(url, method, mode) {
  // 1. Only GET is ever served from or written to a cache. POST /games,
  //    DELETE /games/x, the whole mutating surface: untouched.
  if (method !== "GET") return "bypass";

  // 2. Cross-origin is never cached. That includes the Google Fonts CDN:
  //    those responses are opaque, so their status cannot be inspected and
  //    their size is padded against the origin quota. app.css declares a
  //    metric-close fallback for all three families, so a blocked or offline
  //    font host degrades to system type rather than breaking.
  if (url.origin !== SELF_ORIGIN) return "bypass";

  // 3. Card art: immutable, and the single biggest repeat-visit win.
  if (CARD_IMAGE_PATH.test(url.pathname)) return "card-image";

  // 4. Everything else on the API surface: network only, always.
  if (API_PATH.test(url.pathname)) return "bypass";

  // 5. Navigations get the app shell.
  if (mode === "navigate") return "shell";

  // 6. Hashed build output, and the handful of unhashed files we precached.
  //    Anything else -- /sounds/*.mp3 among them -- falls through to the
  //    network and is left to the browser's own HTTP cache.
  if (HASHED_ASSET_PATH.test(url.pathname)) return "asset";
  if (PRECACHE.indexOf(url.pathname) !== -1) return "asset";

  return "bypass";
}

// cacheable rejects anything we should not be storing: redirects (the cached
// copy would replay the redirect target under the original URL), partial
// content from Range requests, opaque cross-origin responses, and errors.
function cacheable(response) {
  return Boolean(
    response && response.status === 200 && !response.redirected && response.type !== "opaque",
  );
}

let writesSinceTrim = 0;

async function trimCardCache() {
  const cache = await caches.open(CARD_CACHE);
  const keys = await cache.keys();
  // Cache.keys() returns insertion order, so the head is the oldest entry.
  // Trim past the limit so this does not run again on the very next write.
  const excess = keys.length - CARD_CACHE_LIMIT;
  if (excess <= 0) return;
  const doomed = keys.slice(0, excess + Math.floor(CARD_CACHE_LIMIT * 0.1));
  await Promise.all(doomed.map((request) => cache.delete(request)));
}

async function putCardImage(request, response) {
  const cache = await caches.open(CARD_CACHE);
  try {
    await cache.put(request, response);
  } catch {
    // Quota exhausted (or the storage layer said no). Drop what we have and
    // carry on -- a cache write failing must never fail the image.
    await trimCardCache();
    return;
  }
  writesSinceTrim += 1;
  if (writesSinceTrim >= CARD_TRIM_INTERVAL) {
    writesSinceTrim = 0;
    await trimCardCache();
  }
}

// Cache-first. On a full battlefield this is the difference between hundreds
// of network round-trips and none.
//
// The cache write is handed to waitUntil rather than awaited: the image
// resolves as soon as the network does, and the worker is kept alive long
// enough to finish storing it.
async function cardImage(event) {
  const request = event.request;
  const cached = await caches.match(request, { cacheName: CARD_CACHE });
  if (cached) return cached;
  const response = await fetch(request);
  if (cacheable(response)) {
    event.waitUntil(putCardImage(request, response.clone()));
  }
  return response;
}

// Cache-first on the precached shell, which is what makes a cold start
// instant. The shell is only ever replaced by a new service worker
// activating, and the player decides when that happens (see lib/pwa.ts).
async function shell(event) {
  const cache = await caches.open(SHELL_CACHE);
  const cached = await cache.match("/index.html");
  if (cached) return cached;
  try {
    return await fetch(event.request);
  } catch {
    const offline = await cache.match("/offline.html");
    if (offline) return offline;
    throw new Error("offline and no cached shell");
  }
}

// Cache-first for content-hashed assets and precached static files.
async function asset(event) {
  const request = event.request;
  const cache = await caches.open(SHELL_CACHE);
  const cached = await cache.match(request);
  if (cached) return cached;
  const response = await fetch(request);
  if (cacheable(response)) {
    event.waitUntil(cache.put(request, response.clone()).catch(() => {}));
  }
  return response;
}

async function precacheShell() {
  const cache = await caches.open(SHELL_CACHE);
  // cache: "reload" so a precache never picks up the very HTTP-cached copy
  // this deploy is replacing. addAll is atomic: if any entry 404s the install
  // fails and the previous worker stays in charge, which is the correct
  // outcome -- better no update at all than a half-cached shell.
  await cache.addAll(PRECACHE.map((path) => new Request(path, { cache: "reload" })));
}

self.addEventListener("install", (event) => {
  // No skipWaiting(). A new worker sits in `waiting` until the player accepts
  // the update prompt; reloading a game mid-match on our own initiative would
  // be destructive.
  event.waitUntil(precacheShell());
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    (async () => {
      const names = await caches.keys();
      await Promise.all(
        names
          .filter((name) => name.startsWith("cmdctrl-shell-") && name !== SHELL_CACHE)
          .map((name) => caches.delete(name)),
      );
      // Take over any already-open tab. On a first install this is what lets
      // the current page benefit immediately; on an update it only runs after
      // the player has accepted, because nothing else calls skipWaiting().
      await self.clients.claim();
    })(),
  );
});

self.addEventListener("fetch", (event) => {
  const request = event.request;
  let url;
  try {
    url = new URL(request.url);
  } catch {
    return;
  }
  // Range requests (audio scrubbing) would produce a 206 we must not store,
  // and a cached 200 answered to a ranged request confuses media elements.
  if (request.headers.has("range")) return;

  switch (classify(url, request.method, request.mode)) {
    case "card-image":
      event.respondWith(cardImage(event));
      return;
    case "shell":
      event.respondWith(shell(event));
      return;
    case "asset":
      event.respondWith(asset(event));
      return;
    default:
      // Deny by default: no respondWith at all, so the browser performs the
      // request exactly as if no service worker were installed.
      return;
  }
});

self.addEventListener("message", (event) => {
  // The ONLY path to skipWaiting, and it is driven by the player clicking
  // "reload" in the update prompt. lib/pwa.ts sends this.
  if (event.data && event.data.type === "SKIP_WAITING") {
    self.skipWaiting();
  }
});

// Test seam. The routing rules are the safety-critical part of this file, and
// they are exercised directly by serviceWorker.test.ts, which evaluates this
// source against a stub global. Costs one property on the worker global.
self.__swInternals = { classify, cacheable, BUILD_ID, SHELL_CACHE, CARD_CACHE, PRECACHE };
