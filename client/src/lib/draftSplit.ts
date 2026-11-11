// draftSplit.ts — the picks split into Main and Sideboard (#5699).
//
// The split is this browser's own state, keyed by the draft's id. The
// server keeps nothing about it until a deck is submitted (ADR 0149:
// nothing new is stored about a person). What is held is the Sideboard,
// as a list of card ids; everything else in the picks is Main, so a card
// that has just been picked lands in Main without anyone having to say
// so. It is read again on a reload or a reconnect, is the starting state
// of the deckbuilding screen (components/draft/DraftBuild.svelte), and is
// dropped once the draft is finished (pruneSplits).
//
// Every storage access is wrapped: storage can be off or full, and the
// page must work without it.

import type { DraftCard } from "./draft";

export type Zone = "main" | "side";

const KEY_PREFIX = "cmdctrl.draftSplit.";
/** A split nobody has touched for this long is dropped, whatever became of its draft. */
const MAX_AGE_MS = 14 * 24 * 60 * 60 * 1000;

const key = (id: string): string => KEY_PREFIX + id;

/** The cards in Main and in the Sideboard; an id in `side` that is not a card is ignored. */
export function splitCards(
  cards: readonly DraftCard[],
  side: readonly string[],
): { main: DraftCard[]; side: DraftCard[] } {
  const set = new Set(side);
  const main: DraftCard[] = [];
  const out: DraftCard[] = [];
  for (const c of cards) (set.has(c.id) ? out : main).push(c);
  return { main, side: out };
}

/** The Sideboard after a card is moved to `to`. The same list when nothing changes. */
export function moveTo(side: readonly string[], id: string, to: Zone): string[] {
  const has = side.includes(id);
  if (to === "side") return has ? (side as string[]) : [...side, id];
  return has ? side.filter((x) => x !== id) : (side as string[]);
}

/** `side` less every id that is not one of the cards. */
export function pruneSide(cards: readonly DraftCard[], side: readonly string[]): string[] {
  const ids = new Set(cards.map((c) => c.id));
  return side.filter((id, i, all) => ids.has(id) && all.indexOf(id) === i);
}

/** The Sideboard this browser remembers for a draft; empty when there is none. */
export function loadSplit(draftID: string, now: number = Date.now()): string[] {
  try {
    const raw = localStorage.getItem(key(draftID));
    if (!raw) return [];
    const v = JSON.parse(raw) as { side?: unknown; at?: unknown } | null;
    if (!v || !Array.isArray(v.side)) return [];
    if (typeof v.at === "number" && now - v.at > MAX_AGE_MS) {
      localStorage.removeItem(key(draftID));
      return [];
    }
    return v.side.filter((x): x is string => typeof x === "string");
  } catch {
    return [];
  }
}

/** Remembers the Sideboard. An empty one is not worth a record. */
export function saveSplit(
  draftID: string,
  side: readonly string[],
  now: number = Date.now(),
): void {
  try {
    if (side.length === 0) localStorage.removeItem(key(draftID));
    else localStorage.setItem(key(draftID), JSON.stringify({ side, at: now }));
  } catch {
    // storage is off or full: the split lasts as long as this page does
  }
}

/** Forgets a draft's split: the draft is finished, or gone. */
export function forgetSplit(draftID: string): void {
  try {
    localStorage.removeItem(key(draftID));
  } catch {
    // nothing to forget
  }
}

/**
 * Drops every split whose draft this browser no longer holds a seat in
 * (`live` are the ids it does), and every one gone stale.
 */
export function pruneSplits(live: readonly string[], now: number = Date.now()): void {
  try {
    const keep = new Set(live);
    const drop: string[] = [];
    for (let i = 0; i < localStorage.length; i++) {
      const k = localStorage.key(i);
      if (!k?.startsWith(KEY_PREFIX)) continue;
      const id = k.slice(KEY_PREFIX.length);
      let stale = !keep.has(id);
      if (!stale) {
        try {
          const v = JSON.parse(localStorage.getItem(k) ?? "null") as { at?: unknown } | null;
          stale = !v || (typeof v.at === "number" && now - v.at > MAX_AGE_MS);
        } catch {
          stale = true;
        }
      }
      if (stale) drop.push(k);
    }
    for (const k of drop) localStorage.removeItem(k);
  } catch {
    // nothing to prune
  }
}
