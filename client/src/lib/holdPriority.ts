// #323 — "auto pass on cast unless I change a setting".
//
// Before this module the client stopped on EVERY non-empty stack:
// Game.svelte's auto-pass effect bailed out on `!stackEmpty(view)`.
// That is correct for an opponent's spell — you want the cursor when
// something is aimed at you — but it also meant casting your own
// spell handed you back priority and asked "Counter or Pass?" about
// the thing you had just decided to cast. One extra click on every
// single cast.
//
// The new rule is narrow on purpose:
//
//   auto-pass a non-empty stack ONLY when every item on it is
//   controlled by the viewer AND no trigger is still queuing AND
//   the viewer has not armed a hold.
//
// Anything an opponent put on the stack — a counterspell, a removal
// spell, their own trigger stacked under yours — leaves at least one
// item they control, so the cursor still stops. That keeps the
// "respond to what's aimed at me" window untouched.
//
// Holding priority after your own spell is legitimate and sometimes
// essential (respond to your own trigger, stack two effects in a
// chosen order, hold up a counter behind a bait spell), so the
// behaviour needs a hatch that is reachable BEFORE the auto-pass
// fires — by the time the snapshot lands there is no moment left to
// click in. `holdPriority` is that hatch: a session-scoped toggle
// wired to a visible button in the phase widget (always on screen)
// and in the stack card's header (on screen exactly while the stack
// is live). While it is on, the pre-#323 behaviour is restored
// verbatim — every stack stops.
//
// Scope note: this module decides *whether the client volunteers a
// pass*, never what counts as a legal response. The legality
// predicates live in priority.ts / timing.ts and are untouched.

import { writable, get, type Readable } from "svelte/store";
import type { GameView } from "./protocol";

const held = writable(false);

// holdPriority is the read-only view for reactive consumers
// (PhaseDisplay's button, StackOverlay's header, the auto-pass
// effect in Game.svelte).
//
// Session-scoped and deliberately NOT persisted: like the manual
// step pins in priorityStops.ts it is a now-intent, not a
// preference. The persistent version of the same question is
// `settings.gameplay.autoPassOwnStack`.
export const holdPriority: Readable<boolean> = { subscribe: held.subscribe };

// toggleHoldPriority flips the hatch and returns the new state so
// callers can flash a confirmation.
export function toggleHoldPriority(): boolean {
  let next = false;
  held.update((v) => {
    next = !v;
    return next;
  });
  return next;
}

// setHoldPriority pins the hatch to an explicit value.
export function setHoldPriority(on: boolean): void {
  held.set(on);
}

// isHoldingPriority is the non-reactive read, for callers outside a
// Svelte reactive scope.
export function isHoldingPriority(): boolean {
  return get(held);
}

// ownsEveryStackItem reports whether every item currently on the
// stack belongs to the viewer — the precondition for auto-passing a
// non-empty stack.
//
// Conservative by construction; it returns false for:
//   - an empty stack (nothing to pass on; the caller's own gates
//     handle that path)
//   - spectators and unknown viewers
//   - any item, spell or ability, controlled by someone else
//   - a pending-trigger queue that has not drained yet (CR 603.3b).
//     Those triggers are about to land on the stack and may belong
//     to anyone, so "all mine" isn't decided yet.
//
// Reads BOTH representations of the stack for the same reason
// timing.ts's stackEmpty does: spell items carry a CardView in the
// `stack` zone, ability items exist only in `stack_items`, and a
// snapshot mid-flight can populate one before the other.
export function ownsEveryStackItem(
  snap: GameView | null | undefined,
  viewerID: string | null | undefined,
): boolean {
  if (!snap || !viewerID) return false;
  if ((snap.pending_triggers?.length ?? 0) > 0) return false;

  const items = snap.stack_items ?? [];
  const cards = snap.stack?.cards ?? [];
  if (items.length === 0 && cards.length === 0) return false;

  for (const item of items) {
    if (item.controller !== viewerID) return false;
  }
  for (const card of cards) {
    if (card.controller !== viewerID) return false;
  }
  return true;
}

// _resetForTests is the vitest teardown hook. Not for prod use.
export function _resetForTests(): void {
  held.set(false);
}
