// S13.6 — intelligent priority auto-pass.
//
// Folds the S13.3 timing predicates (canCastFromHand /
// canActivateAbility) into a single "does the viewer have any legal
// response right now?" question. The autoPassPriority effect uses
// this to skip *stopped* steps where the viewer would otherwise have
// to click pass with nothing to do.
//
// Scope:
//   - Checks cards in the viewer's hand + command zone (cast paths)
//     and cards the viewer controls on the battlefield (activated
//     abilities via canActivateAbility, which is the generic
//     instant-speed gate).
//   - Does NOT check mana affordability — that needs the server's
//     AutoTapForCost preview endpoint, which we deliberately keep
//     off the priority hot path. The predicate stays conservative:
//     false positives (claims "you could do something" when you
//     actually can't afford it) are safe — they just mean we stop
//     at a step where the player gets to decide. False negatives
//     (claims "nothing to do" when something IS castable) would
//     eat the player's priority window, so we err on the true side.
//   - Does NOT consider triggered-ability responses — those aren't
//     priority-gated actions from the viewer's perspective, and S19
//     will auto-fire them anyway.
//
// The server is still authoritative for every action; this module
// exists purely to decide whether the client should auto-pass
// priority on the viewer's behalf when `smartAutoPass` is on.

import { canActivateAbility, canCastFromHand, hasPriority } from "./timing";
import type { GameView } from "./protocol";

// Cache the result per (snap.seq, viewerID). Computing
// hasAnyLegalResponse can scan dozens of cards across hand +
// battlefield + command; a single snapshot tick can trigger
// multiple callers (the autoPassPriority effect + any UI surface
// that wants to badge "you have a response available") and we
// don't want each to re-walk the same state.
const cache = new Map<string, boolean>();
let cacheSeq = -1;

function cacheKey(seq: number, viewerID: string): string {
  if (seq !== cacheSeq) {
    cache.clear();
    cacheSeq = seq;
  }
  return viewerID;
}

// owesBlockDecision reports whether the viewer is facing a
// declare-blockers decision they have not been given the chance to
// make: they're under attack and hold at least one creature that
// could legally block one of the attackers.
//
// #328. Blocking is a TURN-BASED ACTION (CR 509.1), not a response to
// anything, so no amount of "does this player have a legal response?"
// reasoning can see it — which is why every auto-pass gate in this
// module sailed straight through the one window a defending player
// cannot afford to lose. The reporter's replay has them passing
// declare_blockers 2ms after the snapshot arrived and dropping from
// 26 to 18 life against three unblocked attackers, with an untapped
// creature on the table the whole time.
//
// The answer is the server's: block legality is a rules question
// (CR 509.1a untapped, CR 509.1b evasion) and re-deriving it in
// TypeScript is how client and engine drift apart. The engine ships
// `turn.block_decision_seats` and this predicate just asks whether
// the viewer's seat is in it.
//
// Deliberately independent of who holds priority. The active player
// gets priority first on entering the step, so the defender's seat is
// listed before priority ever reaches them; the auto-pass effect
// needs the guard to already be true when it does.
export function owesBlockDecision(
  snap: GameView | null | undefined,
  viewerID: string | null,
): boolean {
  if (!snap || !viewerID) return false;
  const seats = snap.turn?.block_decision_seats;
  if (!seats || seats.length === 0) return false;
  const idx = snap.seats?.findIndex((s) => s.id === viewerID) ?? -1;
  if (idx < 0) return false;
  return seats.includes(idx);
}

// hasAnyLegalResponse reports whether the viewer could fire *any*
// priority-gated action against the current snapshot. Used by the
// smart-skip auto-pass to decide whether to pass through a step the
// viewer has pinned in their stops grid.
//
// Returns false for spectators, for viewers who don't hold priority,
// and for any snap where the timing helpers reject every card.
//
// #328: an owed declare-blockers decision also counts as "you have
// something to do here", even though it isn't a response. Folding it
// in here rather than only at the auto-pass call site keeps every
// consumer — the skip decision and any UI badge — agreeing about
// whether the window is live.
export function hasAnyLegalResponse(
  snap: GameView | null | undefined,
  viewerID: string | null,
  snapSeq = -1,
): boolean {
  if (!snap || !viewerID) return false;
  if (!hasPriority(snap, viewerID)) return false;

  const key = cacheKey(snapSeq, viewerID);
  const cached = cache.get(key);
  if (cached !== undefined) return cached;

  const answer = compute(snap, viewerID);
  cache.set(key, answer);
  return answer;
}

function compute(snap: GameView, viewerID: string): boolean {
  const seat = snap.seats?.find((s) => s.id === viewerID);
  if (!seat) return false;

  // #328: an owed block declaration first — it's the one entry here
  // that isn't a priority-gated action, and the one whose absence
  // silently costs the player the game.
  if (owesBlockDecision(snap, viewerID)) return true;

  // Hand: any castable card?
  for (const card of seat.hand?.cards ?? []) {
    if (canCastFromHand(card, snap, viewerID).legal) return true;
  }

  // Command zone: same predicate shape — a commander you could cast
  // counts as a legal response on your own main phase.
  for (const card of seat.command?.cards ?? []) {
    if (canCastFromHand(card, snap, viewerID).legal) return true;
  }

  // Battlefield: any card the viewer controls that could fire an
  // instant-speed activation. canActivateAbility is intentionally
  // permissive (the engine doesn't know per-card ability lists) —
  // any battlefield card you control during a priority window is
  // treated as "could activate something." Conservative but matches
  // S13.3's existing UX: the grey/un-grey dance on the battlefield
  // already signals when activation is legal.
  for (const card of snap.battlefield?.cards ?? []) {
    if (card.controller !== viewerID) continue;
    if (canActivateAbility(card, snap, viewerID).legal) return true;
  }

  return false;
}

// Testing hook: clears the memo cache. Intended for vitest teardown
// only — production calls are invalidated by snap.seq changes.
export function _resetCacheForTests(): void {
  cache.clear();
  cacheSeq = -1;
}
