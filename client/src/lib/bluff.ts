// #1307 — bluffing.
//
// Smart autopass passes in the time it takes one frame to round-trip
// and holds only when the viewer has an answer, so the length of a
// pause is a tell: anyone watching the priority pills knows who has
// something. A bluff removes it by pausing when the viewer has
// nothing, either for a random while (timed) or until they click
// next (manual).
//
// The decision that a window is bluffable lives in
// autopassDecision.ts. This module holds the delay and two session
// stores: whether bluffing is armed this game, and what the phase
// widget should say while a bluff is running. Game.svelte owns the
// timer.

import { type Readable } from "svelte/store";
import { guardedWritable } from "./guardedStore";

// Clamp for the delay bounds. The floor keeps a bluff longer than an
// automatic pass (which is the whole point); the ceiling keeps a
// hand-edited setting from stalling the table.
export const BLUFF_MIN_MS = 500;
export const BLUFF_MAX_MS = 15000;

function clampMs(n: number): number {
  if (!Number.isFinite(n)) return BLUFF_MIN_MS;
  return Math.min(BLUFF_MAX_MS, Math.max(BLUFF_MIN_MS, n));
}

// bluffDelayMs picks a delay uniformly between the two bounds, both
// clamped to [BLUFF_MIN_MS, BLUFF_MAX_MS]. Bounds given the wrong way
// round are swapped rather than refused, because the two sliders can
// cross while a player drags them. `rand` is injectable for tests.
export function bluffDelayMs(min: number, max: number, rand: () => number = Math.random): number {
  let lo = clampMs(min);
  let hi = clampMs(max);
  if (lo > hi) [lo, hi] = [hi, lo];
  return Math.round(lo + rand() * (hi - lo));
}

// bluffArmed is the in-game switch. Session-scoped like hold and the
// manual pins: it defaults at game mount from the settings (armed if
// either bluff is on) and the phase widget's bluff button flips it
// for the rest of the game.
const armed = guardedWritable(false, "bluffArmed");
export const bluffArmed: Readable<boolean> = { subscribe: armed.subscribe };

export function setBluffArmed(on: boolean): void {
  armed.set(on);
}

export function toggleBluffArmed(): boolean {
  let next = false;
  armed.update((v) => {
    next = !v;
    return next;
  });
  return next;
}

// initBluffArmed is the game-mount default.
export function initBluffArmed(g: { bluffCounterspell: boolean; bluffInstant: boolean }): void {
  armed.set(g.bluffCounterspell || g.bluffInstant);
}

// BluffStatus is what a running bluff looks like to the phase widget:
// timed with the wall-clock time it will pass at, or manual.
export type BluffStatus = { manual: false; passesAt: number } | { manual: true } | null;

const status = guardedWritable<BluffStatus>(null, "bluffStatus");
export const bluffStatus: Readable<BluffStatus> = { subscribe: status.subscribe };

export function setBluffStatus(s: BluffStatus): void {
  status.set(s);
}

// bluffStatusText renders the widget's line: "bluffing — passes in 3s"
// or "bluffing — click next". Coarse on purpose; it is a reminder to
// the viewer, not a timer anyone else sees.
export function bluffStatusText(s: BluffStatus, now: number): string {
  if (!s) return "";
  if (s.manual) return "bluffing — click next";
  const secs = Math.max(0, Math.ceil((s.passesAt - now) / 1000));
  return `bluffing — passes in ${secs}s`;
}

// _resetForTests is the vitest teardown hook. Not for prod use.
export function _resetForTests(): void {
  armed.set(false);
  status.set(null);
}
