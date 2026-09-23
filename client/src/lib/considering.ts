// "Considering a response…" chip (#1307).
//
// Smart autopass (ADR 0009, autopassDecision.ts) now passes in one
// round trip — milliseconds — whenever a seat has no response, and
// only pauses for a click when it does. That makes the LENGTH of a
// pause a tell: an opponent who watches the clock can read "held
// priority for a while" as "they have something" and a snap pass as
// "they don't", which is exactly the information a bluff (ADR 0009's
// amendment, bluff.ts) exists to deny them.
//
// The fix is a chip every seat sees the same way regardless of WHY a
// human is holding — a real hold, a timed bluff, a manual bluff, or
// someone away from the keyboard. That only works if the chip is
// derived from PUBLIC state and elapsed time alone: nothing here may
// read (or imply) whether the holder actually has a response, only
// whether the window they're holding in is the kind where holding
// would be meaningful at all.
//
// isResponseWindowFor and responseWindowKey are pure and read only
// the fields every viewer already receives — Board.svelte owns the
// timer and the 800ms threshold (see its "considering" section).

import type { GameView } from "./protocol";

// stackDepth counts everything on the stack. Reads both
// representations for the reason timing.ts's stackEmpty and
// holdPriority.ts's ownsEveryStackItem do: `stack_items` carries
// abilities that have no card of their own, and a snapshot mid-flight
// can populate one representation before the other.
function stackDepth(view: GameView | null | undefined): number {
  if (!view) return 0;
  return Math.max(view.stack_items?.length ?? 0, view.stack?.cards?.length ?? 0);
}

// hasBlockingChoice reports whether any pending choice is open. A
// table waiting on somebody's answer to a prompt is a different kind
// of pause than a priority hold — showing the "considering" chip
// through one would blur two different waits into one.
function hasBlockingChoice(view: GameView | null | undefined): boolean {
  return (view?.pending_choices?.length ?? 0) > 0;
}

// isResponseWindowFor reports whether `seat` — a seat index, matching
// turn.active_seat / turn.priority_holder — currently holds priority
// in a window where holding it would mean something: the stack is
// non-empty (there's something to respond to), or priority has moved
// off the active player with an empty stack (the classic "hold
// priority into their end step" pattern). False whenever the table is
// waiting on a pending choice or an owed block decision instead —
// neither is a response window, however long either one takes.
export function isResponseWindowFor(view: GameView | null | undefined, seat: number): boolean {
  if (!view?.turn) return false;
  if (view.turn.priority_holder !== seat) return false;
  const stackNonEmpty = stackDepth(view) > 0;
  const holderNotActive = view.turn.priority_holder !== view.turn.active_seat;
  if (!stackNonEmpty && !holderNotActive) return false;
  if (hasBlockingChoice(view)) return false;
  if ((view.turn.block_decision_seats?.length ?? 0) > 0) return false;
  return true;
}

// responseWindowKey identifies "the same priority window" across
// frames — turn number, step, who holds priority, and the stack
// depth. Board.svelte restarts its timer whenever this changes, so a
// seat holding through several of its own decisions in one window
// doesn't get a fresh timer for each of them, and a genuinely new
// window (a new turn, a new step, priority moving, something new
// landing on the stack) always does.
export function responseWindowKey(view: GameView | null | undefined): string {
  if (!view?.turn) return "";
  return [view.turn.number, view.turn.step, view.turn.priority_holder, stackDepth(view)].join("|");
}
