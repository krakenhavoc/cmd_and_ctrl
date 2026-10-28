// Tints a signature leaves on the survivors of its batch until the turn
// ends (ADR 0209 §4, #4416). Toxic Deluge gives every creature -X/-X until
// end of turn, so the creatures it did not kill keep a faint dark-green
// stain for as long as the effect lasts.
//
// Presentation only: the client remembers which permanents were marked,
// in which game and on which turn, and forgets them as soon as the turn
// number (or the game) changes. Nothing goes on the wire or into the
// snapshot. A reconnect mid-turn may forget a tint; that is the whole
// cost. TableEffects.svelte marks and syncs; FxTint.svelte draws it on a
// permanent's tile (Card.svelte).

import { get, type Readable } from "svelte/store";
import { guardedDerived, guardedWritable } from "../guardedStore";

/** The kinds of tint a scene can leave. */
export type TintKind = "deluge";

export interface TintMark {
  kind: TintKind;
  game: string;
  turn: number;
}

/** A tint book: instance ID to its mark. */
export type TintBook = ReadonlyMap<string, TintMark>;

/** markTints adds `ids` to the book under this game and turn (a later mark of one card wins). */
export function markTints(
  book: TintBook,
  ids: Iterable<string>,
  kind: TintKind,
  game: string,
  turn: number,
): TintBook {
  const next = new Map(book);
  for (const id of ids) next.set(id, { kind, game, turn });
  return next;
}

/**
 * syncTints drops every mark made in another game or on another turn. It
 * returns the same book when nothing changed, so a store can skip the set.
 */
export function syncTints(book: TintBook, game: string, turn: number): TintBook {
  let stale = false;
  for (const m of book.values()) {
    if (m.game !== game || m.turn !== turn) {
      stale = true;
      break;
    }
  }
  if (!stale) return book;
  const next = new Map<string, TintMark>();
  for (const [id, m] of book) if (m.game === game && m.turn === turn) next.set(id, m);
  return next;
}

const book = guardedWritable<TintBook>(new Map(), "tintBook");

/** The app's tints, by instance ID: what each permanent's tile draws. */
export const tableTints: Readable<ReadonlyMap<string, TintKind>> = guardedDerived(
  book,
  (b) => {
    const out = new Map<string, TintKind>();
    for (const [id, m] of b) out.set(id, m.kind);
    return out;
  },
  "tableTints",
  new Map(),
);

/** tintSurvivors marks the permanents a scene's batch left standing. */
export function tintSurvivors(ids: readonly string[], kind: TintKind, game: string, turn: number) {
  if (ids.length === 0) return;
  book.update((b) => markTints(syncTints(b, game, turn), ids, kind, game, turn));
}

/** syncTableTints forgets the tints of another game or an earlier turn. */
export function syncTableTints(game: string, turn: number): void {
  // Every frame calls this: it only notifies the tiles when a tint goes.
  const now = get(book);
  const next = syncTints(now, game, turn);
  if (next !== now) book.set(next);
}

/** clearTableTints forgets every tint (tests, leaving a game). */
export function clearTableTints(): void {
  book.set(new Map());
}
