// tableMoment.ts — when a hint may show at the table (ADR 0125 §3.5).
//
// A table hint shows only in a QUIET moment. Game.svelte computes the
// moment from what it already holds and publishes it here
// (publishTableMoment); the hint layer, mounted at the app shell, reads
// it. The predicate is pure, so every rule is pinned by
// tableMoment.test.ts:
//
//   1. No dock request is open for the viewer: mulligan, target,
//      payment, a choice, trigger order, attackers or blockers, the
//      opening-roll sheet or its confirm, a revealed-hand pick. Every
//      one of those registers with the dock (lib/dock.ts), so "any open
//      request" is the test, and an open dialog counts too.
//   2. The viewer is not mid-gesture: no targeting, no drag (a pointer
//      held down), no open card menu or ability popover.
//   3. The top of the stack, if there is one, is not another player's
//      item while the viewer holds priority. That is the moment they
//      decide whether to answer it, and ADR 0119's stack hold gives
//      them only a few seconds. The viewer's own item on top is quiet.
//   4. The table has been on screen for TABLE_SETTLE_MS, and no hint
//      has closed in the last TABLE_GAP_MS.
//
// And table hints are off while the tutorial coach is visible.
//
// A hint that loses its quiet moment steps aside at once and stays
// unseen; it comes back at the next quiet one. A hint is never on screen
// while the viewer has a decision to make.

import type { Readable } from "svelte/store";
import { guardedWritable } from "../guardedStore";
import type { GameView } from "../protocol";

/** How long the table is on screen before a hint may show. */
export const TABLE_SETTLE_MS = 5_000;
/** How long after a hint closes before the next may show. */
export const TABLE_GAP_MS = 20_000;

/** The table's moment, as Game.svelte sees it. Read-only to hints. */
export interface TableMoment {
  /** When the table came on screen (ms since the epoch). */
  since: number;
  /** A dock request is open for the viewer (rule 1). */
  dockRequest: boolean;
  /** A dialog or a sheet is open over the table (rule 1). */
  dialog: boolean;
  /** Targeting, a drag, an open card menu or ability popover (rule 2). */
  gesture: boolean;
  /** The viewer holds priority. */
  viewerHasPriority: boolean;
  /** Who controls the top item of the stack, or null when it is empty (rule 3). */
  stackTopController: string | null;
  /** The viewer's player id. */
  viewerID: string | null;
  /** The tutorial coach is on screen. */
  coachVisible: boolean;
}

/** Why a moment is not quiet; null when it is. */
export type NotQuiet =
  | "coach"
  | "dock-request"
  | "dialog"
  | "gesture"
  | "opponent-stack"
  | "settling"
  | "recent-hint";

/**
 * notQuietReason is the first rule a moment breaks, or null when it is
 * quiet. `lastClosedAt` is when the last hint closed (ms), or null.
 */
export function notQuietReason(
  m: TableMoment,
  now: number,
  lastClosedAt: number | null,
): NotQuiet | null {
  if (m.coachVisible) return "coach";
  if (m.dockRequest) return "dock-request";
  if (m.dialog) return "dialog";
  if (m.gesture) return "gesture";
  if (m.viewerHasPriority && m.stackTopController !== null && m.stackTopController !== m.viewerID) {
    return "opponent-stack";
  }
  if (now - m.since < TABLE_SETTLE_MS) return "settling";
  if (lastClosedAt !== null && now - lastClosedAt < TABLE_GAP_MS) return "recent-hint";
  return null;
}

/** isQuiet reports whether a table hint may be on screen now. */
export function isQuiet(m: TableMoment, now: number, lastClosedAt: number | null): boolean {
  return notQuietReason(m, now, lastClosedAt) === null;
}

/** stackTopController is who controls the top item of a view's stack, or null. */
export function stackTopController(view: GameView | null | undefined): string | null {
  const items = view?.stack_items ?? [];
  const top = items[items.length - 1];
  return top ? top.controller : null;
}

/** What the table publishes: the moment, and the view the hints may read. */
export interface TableState {
  moment: TableMoment;
  view: GameView | null;
  viewerID: string | null;
}

const store = guardedWritable<TableState | null>(null, "hintTable");

/** tableState is the mounted table's moment, or null when no table is on screen. */
export const tableState: Readable<TableState | null> = { subscribe: store.subscribe };

let token = 0;

/**
 * publishTableMoment registers a table: call it from the game route
 * while a seated viewer's table is on screen, and call `update` on every
 * change. `close` takes it down. Token-guarded like shortcutRuntime.ts,
 * because the keyed remount of one game route over another can run the
 * old one's cleanup after the new one registered.
 */
export function publishTableMoment(initial: TableState): {
  update(next: TableState): void;
  close(): void;
} {
  const mine = ++token;
  store.set(initial);
  return {
    update(next) {
      if (token === mine) store.set(next);
    },
    close() {
      if (token === mine) store.set(null);
    },
  };
}

/** _resetForTests is the vitest teardown hook. */
export function _resetTableMomentForTests(): void {
  token = 0;
  store.set(null);
}
