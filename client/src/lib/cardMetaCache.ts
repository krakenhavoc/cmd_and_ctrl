// Client-side cache for /cards/{id} metadata responses (Oracle text,
// type line, P/T, mana cost, etc.). Fetched lazily on first hover so
// the table doesn't pay a 100×-card round-trip cost up front.
//
// The browser's HTTP cache would serve the same purpose for repeated
// hovers (the server emits Cache-Control: public, max-age=3600), but
// piping every render through fetch() adds parse + Promise overhead
// per hover. The in-memory cache here returns synchronously after the
// first fetch resolves.

import { type Readable } from "svelte/store";

import { guardedWritable } from "./guardedStore";

// CardMeta is the wire shape of GET /cards/{id}. Mirrors the trimmed
// Scryfall projection in server/internal/cards/index.go — only the
// fields the hover panel reads are typed; the rest passes through
// untyped via the `extras` index signature (currently unused).
export interface CardMeta {
  id: string;
  name: string;
  type_line?: string;
  oracle_text?: string;
  mana_cost?: string;
  power?: string;
  toughness?: string;
  loyalty?: string;
  color_identity?: string[];
  card_faces?: Array<{
    name: string;
    type_line?: string;
    oracle_text?: string;
    mana_cost?: string;
  }>;
}

type CacheEntry =
  | { state: "loading"; promise: Promise<CardMeta | null> }
  | { state: "loaded"; meta: CardMeta }
  | { state: "error" };

const cache = new Map<string, CacheEntry>();

// Per-id reactive store. Components subscribe to one of these to get
// re-rendered when the fetch resolves. Created on demand and shared
// across subscribers so two cards hovered in quick succession both
// see the same store fill in once.
// #720: guarded, because these are the stores with a hand-written
// `.subscribe(...)` on them (HoverZoomOverlay.svelte). That callback
// runs inside svelte/store's shared drain loop, so a throw in it is
// the exact shape that freezes every other store in the app.
const stores = new Map<string, ReturnType<typeof guardedWritable<CardMeta | null>>>();

function storeFor(id: string) {
  let s = stores.get(id);
  if (!s) {
    s = guardedWritable<CardMeta | null>(null, `cardMeta:${id}`);
    stores.set(id, s);
  }
  return s;
}

// metaFor returns a Svelte readable store that resolves to the card's
// metadata once /cards/{id} responds, or stays null on error / unknown.
// Callers wire it up with `$metaStore` in their template; the store
// re-renders the consumer when the fetch lands.
//
// #740 — METAFOR MUST NOT WRITE TO THE STORE IT RETURNS, and neither
// may anything it calls. Its one caller reaches it from a `$derived`
// (`HoverZoomOverlay.svelte`), and a `set()` while a derived is being
// evaluated runs that store's subscribers inside the derived's
// reaction. HoverZoomOverlay's subscriber assigns to a `$state`, and
// Svelte answers a `$state` write from inside a derived by throwing
// `state_unsafe_mutation` (`svelte/internal/client/reactivity/sources.js`
// — not a dev-only check).
//
// The version before this comment took a `loaded` cache entry and
// re-`set()` the store with the meta it already held. That was not
// even a no-op write: svelte/store's `safe_not_equal` reports two
// objects as never equal, so it notified every subscriber. It only
// THREW when the hovered card changed to one with the same
// `scryfall_id` and the overlay's effect had therefore not torn down
// its subscription yet — a second copy of the same printing, or the
// same card hovered again after the first Card component unmounted
// without a `pointerleave` (a zone browser closing, a permanent
// leaving the battlefield). See hoverMeta.render.test.ts for the
// isolated case and boardFreeze.render.test.ts for the whole
// attack-declaration sequence #740 reported it from.
//
// Nothing needs the write: `storeFor` memoises per id and the fetch
// continuation below is what fills the store at the moment the cache
// entry becomes `loaded`, so a cached id's store already holds its
// meta. Every state of a known id is therefore the same answer —
// return the store and read it.
export function metaFor(scryfallID: string): Readable<CardMeta | null> {
  const existing = cache.get(scryfallID);
  if (existing) {
    return storeFor(scryfallID);
  }
  // First time — kick off the fetch.
  const promise = fetch(`/cards/${scryfallID}`, { credentials: "include" })
    .then(async (res) => {
      if (!res.ok) {
        cache.set(scryfallID, { state: "error" });
        return null;
      }
      const meta = (await res.json()) as CardMeta;
      cache.set(scryfallID, { state: "loaded", meta });
      storeFor(scryfallID).set(meta);
      return meta;
    })
    .catch(() => {
      cache.set(scryfallID, { state: "error" });
      return null;
    });
  cache.set(scryfallID, { state: "loading", promise });
  return storeFor(scryfallID);
}
