// S13.6 — intelligent priority auto-pass.
//
// Answers one question: "does the viewer have anything to do in this
// priority window?" The autoPassPriority effect uses it to skip
// *stopped* steps where the viewer would otherwise have to click pass
// with nothing to do.
//
// S31 sub-PR 2 turned this from a derivation into a lookup. It used
// to walk the viewer's hand, command zone and battlefield running the
// S13.3 timing predicates over every card, which meant it inherited
// every bug in those predicates plus one of its own: it could not see
// mana, so "you have a response" meant "you hold a card that is legal
// at this speed", affordable or not. The server now enumerates the
// seat's legal moves and ships them as `legal_moves`, so the question
// is `legal_moves.some(m => m.kind !== "pass")` — the engine's own
// answer, mana and targets and all.
//
// The server is still authoritative for every action; this module
// exists purely to decide whether the client should auto-pass
// priority on the viewer's behalf when `smartAutoPass` is on.

import { hasNonPassMove, hasPriority } from "./timing";
import type { GameView } from "./protocol";

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
//
// Kept as its own signal even though `legal_moves` now carries
// `block` moves too: this one has to be readable on a frame where the
// defender holds no priority, and the enumerator's answer is derived
// from the same server-side eligibility test anyway (#328), so the
// two cannot disagree.
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

// hasDeclaredAttackers reports whether the viewer is the active
// player, standing in their own declare-attackers step, with at least
// one attack already declared.
//
// #599. It is the attacking half of owesBlockDecision above, and it
// exists for the same reason: an auto-pass that closes a window the
// player cannot get back. The enumerator offers one attack move per
// ELIGIBLE creature (legal/combat.go, game.AttackerEligible), and
// declaring an attacker taps it, so a wide declaration removes every
// non-pass move the seat had. Smart auto-pass then reads "nothing to
// do", yields, and the cursor leaves the step — taking the #318
// cluster, and the Undo button inside it, with it.
//
// That undo is not a nicety. attackAll.ts sends ONE bulk
// declare_attackers precisely so a single undo restores declarations
// and tap state together; nothing else in the client can put a
// declared attacker back. Auto-passing out of the step the instant
// the declaration lands makes the feature's own escape hatch
// unreachable, which is how it shipped and why the e2e guard for it
// has failed every nightly since 2026-09-12.
//
// Deliberately NOT gated on the seat's undo budget. The budget is a
// server number the viewer may not be able to read on the frame that
// matters, an admin bypasses it entirely, and "let me look at the
// attack I just declared" is worth the window on its own.
export function hasDeclaredAttackers(
  snap: GameView | null | undefined,
  viewerID: string | null,
): boolean {
  if (!snap || !viewerID) return false;
  if (snap.turn?.step !== "declare_attackers") return false;
  const active = snap.turn?.active_seat ?? -1;
  if (active < 0 || snap.seats?.[active]?.id !== viewerID) return false;
  return (snap.battlefield?.cards ?? []).some(
    (c) => c.controller === viewerID && !!c.attacking_target,
  );
}

// hasAnyLegalResponse reports whether the viewer could fire *any*
// action against the current snapshot. Used by the smart-skip
// auto-pass to decide whether to pass through a step the viewer has
// pinned in their stops grid.
//
// Returns false for spectators and for viewers who don't hold
// priority. An owed declare-blockers decision counts as "you have
// something to do here" even though it isn't a response (#328).
//
// `snapSeq` is vestigial: the answer is one scan of a field the
// server already computed, so the per-seq memo the card walk needed
// is gone. The parameter stays so call sites don't churn.
export function hasAnyLegalResponse(
  snap: GameView | null | undefined,
  viewerID: string | null,
  snapSeq = -1,
): boolean {
  void snapSeq;
  if (!snap || !viewerID) return false;

  // #328 first: it's the one entry here that isn't a priority-gated
  // action, and the one whose absence silently costs the player the
  // game.
  if (owesBlockDecision(snap, viewerID)) return true;

  if (!hasPriority(snap, viewerID)) return false;

  // #599: a declaration the viewer just made is something to look at,
  // even though the engine has no further move to offer them.
  if (hasDeclaredAttackers(snap, viewerID)) return true;

  // No move list on a frame where the viewer holds priority means a
  // server older than S31, or a field we dropped; err toward
  // stopping. A false positive costs one click, a false negative eats
  // a window the player was entitled to — the asymmetry ADR 0009 §3
  // calls for.
  return hasNonPassMove(snap) ?? true;
}

// autopassSuspended reports whether the server has told this table to
// stop passing AUTOMATICALLY (#628, CR 726).
//
// The engine raises `loop_notice` when one triggered ability has
// resolved 25 times in a turn with nobody casting, activating,
// answering a prompt or declaring a creature in between — a trigger
// loop, which with every seat on autopass is a tight
// pass → resolve → broadcast → pass spin that no one at the table can
// get out of except by finding the toggle mid-flight.
//
// Deliberately NOT a "can the viewer act" question, so it sits with
// the mulligan / game-over guards rather than with the stops grid:
// the suspension is table-wide, it outranks every gate including the
// autopass toggle itself, and it holds until a player makes a real
// decision. Pressing "next" by hand still passes — stopping the game
// is the server's job and it declines to.
export function autopassSuspended(snap: GameView | null | undefined): boolean {
  return !!snap?.loop_notice;
}

// loopNoticeText renders the banner line: "Mirror Engine — create a
// Spark has resolved 25 times this turn." The label already reads
// "<card> — <ability>" by catalog convention, so there is nothing to
// look up on the board. Empty string when nothing is suspected.
export function loopNoticeText(snap: GameView | null | undefined): string {
  const n = snap?.loop_notice;
  if (!n) return "";
  const label = n.label || "An ability";
  return `${label} has resolved ${n.count} times this turn.`;
}

// Testing hook: retained as a no-op. The memo cache it used to clear
// went away with the card walk, but vitest teardowns still call it
// and a missing export is a worse failure than a no-op.
export function _resetCacheForTests(): void {}
