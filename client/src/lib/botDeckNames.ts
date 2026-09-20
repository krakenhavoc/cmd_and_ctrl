// botDeckNames.ts — resolves a bot seat's curated deck ID
// (PlayerView.bot_deck) to a display name for the BOT chip tooltip
// (#688).
//
// PlayerView only carries the ID (e.g. "esper-control"); the deck's
// name lives in the curated-deck catalog behind GET /bot/options,
// already fetched by the lobby's Add-bot picker (fetchBotOptions,
// api.ts). This module resolves through that same endpoint instead of
// adding a new field to the wire (protocol.ts /
// server/internal/protocol/view.go / docs/protocol.md) — #688 allows
// either, and view.go is already near this repo's size ceiling.
//
// Client-only, cached for the session: one fetch, shared by every
// seat's chip. Modeled on env.ts's loadAppConfig/appConfig pair — a
// guarded store (#266/#720, guardedStore.ts) so a chip mounted before
// the fetch resolves re-renders once the names land, and a memoised
// in-flight promise so concurrent mounts share one request.
//
// The logic lives here rather than in PlayerIdentity.svelte for the
// same reason catalog.ts's filter and prebuiltDecks.ts's summariser
// do: this project has no jsdom, so a `.svelte` file cannot be unit
// tested, and "what does an unresolved deck ID render as" is exactly
// the kind of decision worth pinning down with a test.

import { type Readable } from "svelte/store";

import { guardedWritable } from "./guardedStore";
import { fetchBotOptions, type BotDeckInfo, type BotOptions } from "./api";

export type BotDeckNameMap = ReadonlyMap<string, BotDeckInfo>;

const EMPTY_MAP: BotDeckNameMap = new Map();

const store = guardedWritable<BotDeckNameMap>(EMPTY_MAP, "botDeckNames");

// botDeckNames is the live id -> curated-deck map. Read-only to
// consumers — only ensureBotDeckNamesLoaded (and, in tests,
// resetBotDeckNamesForTests) writes it.
export const botDeckNames: Readable<BotDeckNameMap> = { subscribe: store.subscribe };

/**
 * indexBotDecks turns GET /bot/options' deck list into an id -> deck
 * map. Pure, so "what does the map end up holding" is testable
 * without a network call.
 */
export function indexBotDecks(decks: readonly BotDeckInfo[] | undefined): Map<string, BotDeckInfo> {
  const map = new Map<string, BotDeckInfo>();
  for (const d of decks ?? []) {
    if (d.id) map.set(d.id, d);
  }
  return map;
}

/**
 * botDeckLabel is the display text for a bot seat's deck: the
 * curated deck's name when the catalog resolves the ID, else the raw
 * ID itself.
 *
 * That fallback is the point of #688's vitest coverage — an ID the
 * catalog doesn't recognize (a pasted custom decklist, a deck retired
 * since the seat was made, or a server whose catalog hasn't loaded
 * yet) must still render as something a player can read, never
 * blank.
 */
export function botDeckLabel(deckID: string | null | undefined, names: BotDeckNameMap): string {
  if (!deckID) return "";
  const name = names.get(deckID)?.name?.trim();
  return name || deckID;
}

let inflight: Promise<void> | null = null;

/**
 * ensureBotDeckNamesLoaded fetches the curated-deck catalog once per
 * session and fills the store. Safe to call from every mounted BOT
 * chip — only the first call does any work, and concurrent callers
 * share the same request (same posture as env.ts's loadAppConfig).
 *
 * `fetcher` defaults to the real GET /bot/options call and exists so
 * tests can inject a fake response without mocking a module.
 *
 * Failures are swallowed: the map is simply left as it was (empty, on
 * a first failed attempt). botDeckLabel's raw-ID fallback is always a
 * correct answer, so there is nothing for a caller to react to.
 */
export function ensureBotDeckNamesLoaded(
  fetcher: () => Promise<BotOptions> = fetchBotOptions,
): Promise<void> {
  if (!inflight) {
    inflight = fetcher()
      .then((opts) => {
        store.set(indexBotDecks(opts.decks));
      })
      .catch(() => {
        // Leave the store as-is; the raw-id fallback still applies.
      });
  }
  return inflight;
}

// resetBotDeckNamesForTests clears the memoised fetch and the store.
// Exported for unit tests only.
export function resetBotDeckNamesForTests(): void {
  inflight = null;
  store.set(EMPTY_MAP);
}
