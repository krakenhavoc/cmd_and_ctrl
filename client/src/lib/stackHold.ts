// ADR 0119 §2 — the stack hold.
//
// Smart auto-pass passes an opponent's spell the viewer cannot answer
// in about one round trip, so the item is often on screen for less
// than a second. The hold keeps an AUTOMATIC pass waiting until the
// top stack item has been on this client's screen for
// the table's pace (ADR 0143 §2.6), when that item is controlled by someone other
// than the viewer. The rules let a player take their time before
// passing (CR 117.3d), and the item resolves only once every seat has
// passed (CR 117.4), so one seat's hold is the table's.
//
// It never holds the viewer's explicit `next`, the viewer's own top
// item (CR 117.3c: the caster gets priority first, and
// autoPassOwnStack passes it at once), or an empty stack.
//
// autopassDecision stays the pure precedence and does not know about
// time. This module answers how much longer a `pass` verdict should
// wait, from the stack, the viewer, the first-seen times, the hold
// and now. Game.svelte owns the first-seen map and the timer, the same
// machinery as the timed bluff (bluff.ts), and the action dock shows
// the countdown through stackHoldStatus below.

import type { Readable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import type { GameView, TableSettingsView } from "./protocol";

/** The longest hold any pace sets (Slow). The considering chip waits past it. */
export const STACK_HOLD_MAX_MS = 3000;

/** Normal pace's hold, about 2 s (ADR 0119 owner answer 2a). */
export const DEFAULT_STACK_HOLD_MS = 2000;

/** A table's pace, the host's setting (`settings.bot_pace` on the wire). */
export type TablePace = TableSettingsView["bot_pace"];

/**
 * ADR 0143 §2.6: the hold is the table's, set by its pace, for people
 * and bots alike. These are aiseat's botPacePresets StackHold values
 * (server/internal/aiseat/runner.go), and stackHold.test.ts reads that
 * file so the two tables cannot drift apart.
 */
export const STACK_HOLD_BY_PACE: Readonly<Record<TablePace, number>> = Object.freeze({
  fast: 0,
  normal: 2000,
  slow: 3000,
});

/**
 * stackHoldMsForPace is the hold for a table's pace. An unknown pace
 * (an older server, a frame with no settings) is Normal's, the default
 * pace.
 */
export function stackHoldMsForPace(pace: unknown): number {
  return typeof pace === "string" && pace in STACK_HOLD_BY_PACE
    ? STACK_HOLD_BY_PACE[pace as TablePace]
    : DEFAULT_STACK_HOLD_MS;
}

/**
 * clampStackHoldMs reads a hold the way bluff.ts reads its
 * bounds: clamped where it is used, so a hand-edited blob can neither
 * stall the table nor go negative. A value that is not a number at all
 * falls back to the default.
 */
export function clampStackHoldMs(n: unknown): number {
  if (typeof n !== "number" || !Number.isFinite(n)) return DEFAULT_STACK_HOLD_MS;
  return Math.min(STACK_HOLD_MAX_MS, Math.max(0, Math.round(n)));
}

/** The top of the stack: the item that resolves next (CR 608.1). */
export interface StackTop {
  id: string;
  controller: string;
}

// topStackItem reads both representations of the stack, for the
// reason timing.ts's stackEmpty and holdPriority.ts's
// ownsEveryStackItem do: ability items exist only in `stack_items`,
// and a snapshot mid-flight can populate one before the other. The
// wire sends `stack_items` bottom to top, so the top is the last one.
export function topStackItem(view: GameView | null | undefined): StackTop | null {
  const items = view?.stack_items ?? [];
  if (items.length > 0) {
    const top = items[items.length - 1];
    return { id: top.id, controller: top.controller };
  }
  const cards = view?.stack?.cards ?? [];
  if (cards.length > 0) {
    const top = cards[cards.length - 1];
    return { id: top.instance_id, controller: top.controller ?? "" };
  }
  return null;
}

/** When this client first saw each item now on the stack, by item ID. */
export type FirstSeen = ReadonlyMap<string, number>;

/**
 * noteStackSeen returns the first-seen map for this frame: every item
 * still on the stack keeps the time it was first seen, a new one is
 * stamped `now`, and an item that has left is dropped. Pure; the
 * caller keeps the result.
 */
export function noteStackSeen(
  prev: FirstSeen,
  view: GameView | null | undefined,
  now: number,
): Map<string, number> {
  const next = new Map<string, number>();
  const note = (id: string): void => {
    if (next.has(id)) return;
    next.set(id, prev.get(id) ?? now);
  };
  for (const it of view?.stack_items ?? []) note(it.id);
  for (const c of view?.stack?.cards ?? []) note(c.instance_id);
  return next;
}

export interface StackHoldInput {
  view: GameView | null | undefined;
  viewerID: string | null | undefined;
  firstSeen: FirstSeen;
  holdMs: unknown;
  now: number;
}

/**
 * stackHoldRemainingMs is how much longer an automatic pass waits: 0
 * when there is nothing to hold for (an empty stack, the viewer's own
 * top item, a spectator, the setting at 0), otherwise the rest of the
 * hold measured from when this client first saw the top item. A top
 * item missing from the map counts as seen now.
 */
export function stackHoldRemainingMs(i: StackHoldInput): number {
  const hold = clampStackHoldMs(i.holdMs);
  if (hold <= 0 || !i.viewerID) return 0;
  const top = topStackItem(i.view);
  if (!top || top.controller === i.viewerID) return 0;
  const seen = i.firstSeen.get(top.id) ?? i.now;
  return Math.max(0, seen + hold - i.now);
}

/**
 * combinedPassDelayMs is the wait when a timed bluff and the hold both
 * apply: the longer of the two, not their sum. Both end in the same
 * pass, and the hold is the floor under it.
 */
export function combinedPassDelayMs(bluffMs: number, holdMs: number): number {
  return Math.max(0, bluffMs, holdMs);
}

// StackHoldStatus is what a running hold looks like to the action dock:
// the wall-clock time the automatic pass goes at.
export type StackHoldStatus = { passesAt: number } | null;

const status = guardedWritable<StackHoldStatus>(null, "stackHoldStatus");
export const stackHoldStatus: Readable<StackHoldStatus> = { subscribe: status.subscribe };

export function setStackHoldStatus(s: StackHoldStatus): void {
  status.set(s);
}

/**
 * stackHoldStatusText renders the dock's line: "auto-pass in 1.4 s".
 * Tenths, because the hold is short and the line is there so the table
 * does not look stalled; `next` still passes at once.
 */
export function stackHoldStatusText(s: StackHoldStatus, now: number): string {
  if (!s) return "";
  const secs = Math.max(0, (s.passesAt - now) / 1000);
  return `auto-pass in ${secs.toFixed(1)} s`;
}

// _resetForTests is the vitest teardown hook. Not for prod use.
export function _resetForTests(): void {
  status.set(null);
}
