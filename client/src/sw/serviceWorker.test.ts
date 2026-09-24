import { describe, it, expect } from "vitest";
import swSource from "./service-worker.js?raw";

// The service worker's routing table is the one piece of this PWA that can
// actually break the game: a single API response served from a cache would
// show a player a table that no longer exists. So these tests run the real
// shipped source rather than a copy of its rules.
//
// service-worker.js is a build template, not a module -- vite-plugin-sw.ts
// substitutes two placeholders and emits it as /sw.js. Here we do the same
// substitution and evaluate it against a stub worker global, then pull the
// routing helpers back off `self.__swInternals`.

const ORIGIN = "https://cmd.example.test";

const PRECACHE = [
  "/index.html",
  "/assets/index-a1b2c3.js",
  "/assets/index-d4e5f6.css",
  "/offline.html",
  "/manifest.webmanifest",
  "/icons/favicon.svg",
  "/icons/icon-192.png",
  "/card-back.jpg",
  "/card-back-small.jpg",
];

type Classify = (url: URL, method: string, mode: string) => string;

function stamp(): string {
  return swSource
    .replace('"__BUILD_ID__"', JSON.stringify("testbuild"))
    .replace('["__PRECACHE__"]', JSON.stringify(PRECACHE));
}

function loadWorker(): { classify: Classify; cacheable: (r: unknown) => boolean } {
  const source = stamp();

  const stub = {
    location: new URL(ORIGIN + "/"),
    addEventListener: () => {},
    skipWaiting: () => {},
    clients: { claim: () => {} },
  } as unknown as Record<string, unknown>;

  const run = new Function("self", "caches", "fetch", source + "\nreturn self.__swInternals;");
  return run(stub, undefined, undefined);
}

const { classify, cacheable } = loadWorker();

const kind = (path: string, method = "GET", mode = "no-cors") =>
  classify(new URL(ORIGIN + path), method, mode);

describe("service worker build template", () => {
  it("still carries both placeholders vite-plugin-sw.ts substitutes", () => {
    // If a rename ever breaks the substitution silently, the shipped worker
    // would precache the literal string "__PRECACHE__" and fail to install.
    expect(swSource).toContain('"__BUILD_ID__"');
    expect(swSource).toContain('["__PRECACHE__"]');
    const stamped = stamp();
    expect(stamped).not.toContain("__BUILD_ID__");
    expect(stamped).not.toContain("__PRECACHE__");
  });
});

describe("service worker routing", () => {
  // The list below mirrors the @api matcher in deploy/Caddyfile. If a route
  // is added there and not here, this test is the reminder.
  const apiPaths = [
    "/ws",
    "/healthz",
    "/me",
    "/logout",
    // Sign out everywhere (ADR 0051 decision 6): Caddy needs /logout/*
    // as its own entry, the worker's prefix match already covers it.
    "/logout/everywhere",
    "/admin/users/7f3c/revoke-sessions",
    // /config and /dev land with PR #260; denying them early is free.
    "/config",
    "/dev",
    "/dev/seed",
    "/games",
    "/games/7f3c",
    "/games/7f3c/seats",
    // The login page's bare-code join (ADR 0050). Top-level, so /games
    // doesn't cover it. Listed so this keeps mirroring the @api matcher;
    // the navigation test below is the one that actually guards the entry.
    "/join",
    "/cards/",
    "/cards/search?q=sol+ring",
    "/catalog",
    // The public engine roadmap (ADR 0092). Same reasoning as /decks:
    // a cached copy would show last deploy's roadmap.
    "/roadmap",
    "/admin/games",
    "/auth/discord/callback",
    "/avatars/123.png",
    "/bugreport",
    "/bugreport/upload",
    // S31 bot seats: the Add-bot picker's options probe.
    "/bot/options",
    "/games/7f3c/seats/bot",
    // The pre-built deck picker's catalog. A cached answer here would
    // show yesterday's coverage numbers after a deploy taught the
    // server better, which is the one thing that route must not do.
    "/decks",
  ];

  it.each(apiPaths)("never caches the API route %s", (path) => {
    expect(kind(path)).toBe("bypass");
  });

  it("never caches an API route requested as a navigation either", () => {
    expect(kind("/auth/discord/callback", "GET", "navigate")).toBe("bypass");
  });

  // /join is POST-only, so the method check keeps the real request uncached
  // on its own, and a plain GET falls through to "bypass" whether or not the
  // path is in API_PATH. A navigation is the one shape where the entry decides
  // the answer: without it the worker hands back the app shell. That makes
  // this the case that actually guards the entry — the list above would pass
  // with it missing.
  it("sends a navigation to /join to the network, not the app shell", () => {
    expect(kind("/join", "GET", "navigate")).toBe("bypass");
  });

  it("caches card art, which is immutable per scryfall id", () => {
    expect(kind("/cards/4f1b/image?size=normal")).toBe("card-image");
    expect(kind("/cards/4f1b/image?size=small")).toBe("card-image");
    expect(kind("/cards/4f1b/image?size=art_crop")).toBe("card-image");
  });

  it("does not mistake other /cards/ shapes for card art", () => {
    expect(kind("/cards/4f1b")).toBe("bypass");
    expect(kind("/cards/4f1b/image/extra")).toBe("bypass");
    expect(kind("/cards/4f1b/meta")).toBe("bypass");
  });

  // The public catalogue serves the same immutable bytes from its own
  // path, because /cards/<id>/image needs a session and the catalogue
  // page does not. It must land in the card-image cache and not be
  // swallowed by "catalog" in the API deny list.
  it("caches catalogue art the same way as card art", () => {
    expect(kind("/catalog/image/4f1b?size=normal")).toBe("card-image");
    expect(kind("/catalog/image/4f1b?size=small&face=1")).toBe("card-image");
  });

  it("does not mistake the catalogue JSON for card art", () => {
    expect(kind("/catalog")).toBe("bypass");
    expect(kind("/catalog/image")).toBe("bypass");
    expect(kind("/catalog/image/4f1b/extra")).toBe("bypass");
  });

  it("bypasses every non-GET method, card art included", () => {
    for (const method of ["POST", "PUT", "PATCH", "DELETE", "HEAD"]) {
      expect(kind("/cards/4f1b/image?size=normal", method)).toBe("bypass");
      expect(kind("/", method, "navigate")).toBe("bypass");
    }
  });

  it("bypasses cross-origin requests, including the font CDN", () => {
    const fonts = new URL("https://fonts.googleapis.com/css2?family=Instrument+Sans");
    const files = new URL("https://fonts.gstatic.com/s/instrumentsans/v1/font.woff2");
    expect(classify(fonts, "GET", "cors")).toBe("bypass");
    expect(classify(files, "GET", "cors")).toBe("bypass");
  });

  it("serves navigations from the app shell", () => {
    expect(kind("/", "GET", "navigate")).toBe("shell");
    expect(kind("/index.html", "GET", "navigate")).toBe("shell");
  });

  it("serves hashed build output and precached statics from the cache", () => {
    expect(kind("/assets/index-a1b2c3.js")).toBe("asset");
    expect(kind("/assets/some-chunk-999.css")).toBe("asset");
    expect(kind("/card-back.jpg")).toBe("asset");
    expect(kind("/icons/icon-192.png")).toBe("asset");
    expect(kind("/manifest.webmanifest")).toBe("asset");
  });

  it("leaves audio and anything else unrecognised to the network", () => {
    expect(kind("/sounds/tap-1.mp3")).toBe("bypass");
    expect(kind("/sounds/ambient_bronze_thunder.mp3")).toBe("bypass");
    expect(kind("/icons/maskable-512.png")).toBe("bypass");
    expect(kind("/robots.txt")).toBe("bypass");
  });
});

describe("service worker response filter", () => {
  it("stores only clean same-origin 200s", () => {
    expect(cacheable({ status: 200, redirected: false, type: "basic" })).toBe(true);
    expect(cacheable({ status: 206, redirected: false, type: "basic" })).toBe(false);
    expect(cacheable({ status: 404, redirected: false, type: "basic" })).toBe(false);
    expect(cacheable({ status: 200, redirected: true, type: "basic" })).toBe(false);
    expect(cacheable({ status: 200, redirected: false, type: "opaque" })).toBe(false);
    expect(cacheable(null)).toBe(false);
  });
});
