// draftSets.ts — which sets can be drafted: types, fetch, and the pure
// sorting the #/draft/sets page runs over the list (ADR 0222 §4, #5437).
//
// Mirrors server/internal/draftpacks/draftable.go (Response / SetStatus).
// Hand-maintained, like catalog.ts and roadmap.ts: update both sides in
// lockstep.
//
// The body is set names and card names, public card data like the
// catalogue: the route and this page need no session.

import { L } from "./labels";

export const DRAFT_SETS_HASH = "#/draft/sets";
export const DRAFT_SETS_TITLE = "Sets you can draft";

/**
 * What every surface that shows a pack must say (ADR 0222 §4): a pack is
 * built to slot rates by rarity, not taken from a real print sheet, and
 * it follows the booster the set was drafted with when it was sold.
 */
export const PACKS_ARE_APPROXIMATIONS =
  "Packs are approximations of retail packs. Each set uses the booster it was drafted with: a 14-card play booster for sets released from February 2024 on, and a 15-card draft booster for older sets. Cards come in slots by rarity, with one printing of each card. Foils, showcase frames and tokens are left out, the exact mix of a real pack is not reproduced, and set and collector boosters are not used.";

/**
 * The kind of booster a set's packs are modelled on: the draftable
 * booster of the set's era. Mirrors draftpacks.Kind.
 */
export type BoosterKind = "play" | "draft";

/** The name of a booster kind, or "" for one this client does not know. */
export function boosterName(kind: string | undefined): string {
  switch (kind) {
    case "play":
      return L.draftPlayBooster;
    case "draft":
      return L.draftDraftBooster;
    default:
      return "";
  }
}

export interface DraftSet {
  /** Scryfall's lowercase set code. */
  code: string;
  name: string;
  /** "YYYY-MM-DD"; absent when the data has no date. */
  released?: string;
  /** Which booster the set's packs are modelled on. */
  booster: BoosterKind;
  /** How many different cards the set's packs can contain. */
  cards: number;
  /** True only when no card blocks the set. */
  draftable: boolean;
  /** Names of the cards that are not fully built yet, sorted. */
  blockers: string[];
}

export interface DraftSetsResponse {
  /** Every set a pack can be made for, newest first. */
  sets: DraftSet[];
  /** How many of them are draftable. */
  draftable: number;
}

/**
 * fetchDraftSets loads the list.
 *
 * Plain `fetch`, not `authFetch`: the route is public, and authFetch
 * reads any 401 as an expired session and clears it.
 */
export async function fetchDraftSets(signal?: AbortSignal): Promise<DraftSetsResponse> {
  const res = await fetch("/draft/sets", {
    headers: { Accept: "application/json" },
    credentials: "same-origin",
    signal,
  });
  if (!res.ok) {
    let message = `${res.status} ${res.statusText}`;
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // not JSON — keep the status line
    }
    throw new Error(message);
  }
  const body = (await res.json()) as Partial<DraftSetsResponse>;
  const sets = Array.isArray(body.sets) ? body.sets : [];
  return {
    sets: sets.map((s) => ({ ...s, blockers: Array.isArray(s.blockers) ? s.blockers : [] })),
    draftable: sets.filter((s) => s.draftable).length,
  };
}

/** The sets that can be drafted, in the server's order (newest first). */
export function readySets(sets: DraftSet[]): DraftSet[] {
  return sets.filter((s) => s.draftable);
}

/**
 * The sets that cannot be drafted yet, the closest to ready first: fewest
 * blocking cards, then the server's order.
 */
export function blockedSets(sets: DraftSet[]): DraftSet[] {
  return sets
    .map((s, i) => ({ s, i }))
    .filter(({ s }) => !s.draftable)
    .sort((a, b) => a.s.blockers.length - b.s.blockers.length || a.i - b.i)
    .map(({ s }) => s);
}

/** "1 card to go" / "12 cards to go". */
export function cardsToGo(n: number): string {
  return `${n} ${n === 1 ? "card" : "cards"} to go`;
}

/** A set's release year, or "" when it has no date. */
export function releaseYear(released: string | undefined): string {
  const m = /^(\d{4})-\d{2}-\d{2}$/.exec(released ?? "");
  return m ? m[1] : "";
}
