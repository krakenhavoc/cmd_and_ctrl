// dragCast.ts — #1508. Drag a card out of your hand onto the table to
// cast it, with its mana auto-tapped.
//
// The gesture is a small state machine, kept here as pure functions so
// the thresholds and the release decision read as a specification and
// are tested without a DOM. Hand.svelte owns the pointer listeners, the
// ghost and the drop zone; everything it DECIDES comes from this file.
//
//   idle ──pointerdown on a hand card──▶ pressed
//   pressed ──moved ≥ DRAG_ACTIVATION_PX──▶ dragging
//   pressed ──pointerup──▶ idle, outcome "click" (the ordinary click
//            handler fires as it always did; a drag never started)
//   dragging ──pointerup──▶ idle, outcome "cast" or "snapBack"
//   any ──Escape / pointercancel──▶ idle, outcome "snapBack" if a drag
//            had started
//
// What a drop DOES is not decided here either: a "cast" outcome hands
// the card to Board.svelte's handlePlayCard — the one cast entry point
// (AGENTS.md §7) — with the drag flag set, so the same prompt chain a
// click runs (face, costs, X, modes, targets) runs after the drop. The
// only difference is the mana posture: applyCastChoices stamps
// `strict: true, auto_tap: true` on a dragged cast, whatever the
// viewer's strictMana setting says (owner decision 3 on the issue).

import type { CardView } from "./protocol";
import type { AutoTapPreview } from "./api";
import type { Legality } from "./timing";
import { isLand } from "./cardTypes";
import { needsFacePicker } from "./faces";
import { alternativeCostsOf, printedCostClaimable, tapCostOf } from "./targeting";
import { guardedWritable } from "./guardedStore";

// DRAG_ACTIVATION_PX is how far the pointer has to travel from where
// it went down before a press becomes a drag. Below it the gesture is
// an ordinary click, and the existing click handler casts exactly as
// it did before drag existed.
export const DRAG_ACTIVATION_PX = 8;

// CAST_ZONE_MARGIN_PX is the margin above the hand's top edge the card
// must be released past to count as "on the table" (owner decision 1).
// Released short of it, the card snaps back with no message: that is a
// change of mind, not a refusal.
export const CAST_ZONE_MARGIN_PX = 60;

// SNAP_REASON_MS is how long a refusal's reason stays on screen after
// a red card snaps back.
export const SNAP_REASON_MS = 2200;

// LAND_DRAG_REASON is what a dragged land says (owner decision 5: lands
// are not drag-played for now). The land still enters the drag — so the
// player sees why nothing happened rather than a card that will not
// move — and always snaps back.
export const LAND_DRAG_REASON = "Play lands by clicking";

export type DragPhase = "idle" | "pressed" | "dragging";

export interface DragState {
  phase: DragPhase;
  cardID: string | null;
  // Where the pointer went down, and where it is now (client px).
  startX: number;
  startY: number;
  x: number;
  y: number;
  // The hand's top edge (client px), measured at pointerdown. The cast
  // zone is everything more than CAST_ZONE_MARGIN_PX above it.
  handTop: number;
  // True while dragging with the pointer past the line.
  inCastZone: boolean;
}

export const IDLE: DragState = {
  phase: "idle",
  cardID: null,
  startX: 0,
  startY: 0,
  x: 0,
  y: 0,
  handTop: 0,
  inCastZone: false,
};

// Verdict is the live answer to "if I let go here, would it cast?" —
// the gold / red glow while dragging, and the snap-back reason.
export interface Verdict {
  castable: boolean;
  reason?: string;
}

export type DragOutcome =
  | { kind: "none" }
  // A press that never became a drag. The native click does the work.
  | { kind: "click" }
  // Released over the table while castable: hand the card to the cast
  // chain with the drag flag.
  | { kind: "cast"; cardID: string }
  // Released short of the line (no reason: the player changed their
  // mind), released while red (the reason), or cancelled.
  | { kind: "snapBack"; reason?: string };

export type DragInput =
  | { type: "down"; cardID: string; x: number; y: number; handTop: number }
  | { type: "move"; x: number; y: number }
  | { type: "up"; x: number; y: number; verdict: Verdict }
  | { type: "cancel" };

// inCastZone: is the pointer far enough above the hand to count as
// "on the table"? Strictly past the margin, so a release exactly on
// the line snaps back.
export function inCastZone(y: number, handTop: number, margin = CAST_ZONE_MARGIN_PX): boolean {
  return y < handTop - margin;
}

// pastActivation: has a press moved far enough to become a drag?
export function pastActivation(
  startX: number,
  startY: number,
  x: number,
  y: number,
  threshold = DRAG_ACTIVATION_PX,
): boolean {
  const dx = x - startX;
  const dy = y - startY;
  return dx * dx + dy * dy >= threshold * threshold;
}

// step advances the gesture by one input and says what, if anything,
// the component should do about it. It never throws and never mutates
// its argument.
export function step(
  state: DragState,
  input: DragInput,
): { state: DragState; outcome: DragOutcome } {
  const none: DragOutcome = { kind: "none" };
  switch (input.type) {
    case "down":
      // A second pointer (another finger) while one gesture is live is
      // ignored rather than restarting it.
      if (state.phase !== "idle") return { state, outcome: none };
      return {
        state: {
          phase: "pressed",
          cardID: input.cardID,
          startX: input.x,
          startY: input.y,
          x: input.x,
          y: input.y,
          handTop: input.handTop,
          inCastZone: false,
        },
        outcome: none,
      };
    case "move": {
      if (state.phase === "idle") return { state, outcome: none };
      const moved = { ...state, x: input.x, y: input.y };
      if (state.phase === "pressed") {
        if (!pastActivation(state.startX, state.startY, input.x, input.y)) {
          return { state: moved, outcome: none };
        }
        moved.phase = "dragging";
      }
      moved.inCastZone = inCastZone(input.y, state.handTop);
      return { state: moved, outcome: none };
    }
    case "up": {
      if (state.phase === "idle") return { state, outcome: none };
      if (state.phase === "pressed") return { state: IDLE, outcome: { kind: "click" } };
      return { state: IDLE, outcome: release(state, input.y, input.verdict) };
    }
    case "cancel":
      if (state.phase === "dragging") return { state: IDLE, outcome: { kind: "snapBack" } };
      return { state: IDLE, outcome: none };
  }
}

// release is the drop decision: cast only when the card is released
// past the line AND the live verdict is castable.
export function release(state: DragState, y: number, verdict: Verdict): DragOutcome {
  if (!state.cardID || !inCastZone(y, state.handTop)) return { kind: "snapBack" };
  if (!verdict.castable) return { kind: "snapBack", reason: verdict.reason };
  return { kind: "cast", cardID: state.cardID };
}

// previewDecidesMana: may the auto-tap preview's "can't pay" answer
// turn this card red? The preview prices the PRINTED cost with no
// announce-time choices, so its "no" is only the cast's "no" when
// nothing announced later can lower the price or change what is being
// cast:
//
//   - a modal DFC has not picked a face yet (the other half may be the
//     cheap one, or a land);
//   - an alternative cost (evoke, Force of Will's pitch) may be the one
//     the player can afford, and a card whose printed cost is not
//     claimable from here is only ever cast for one;
//   - convoke / waterbend taps creatures the auto-tapper does not;
//   - a Phyrexian symbol may be paid with life.
//
// In every one of those cases the drag stays gold and the server — the
// only real answer — refuses the strict cast if the board can't pay.
// A red card that would have cast is the worse lie.
export function previewDecidesMana(card: CardView): boolean {
  if (needsFacePicker(card)) return false;
  if (alternativeCostsOf(card).length > 0) return false;
  if (!printedCostClaimable(card)) return false;
  if (tapCostOf(card)) return false;
  if ((card.phyrexian_symbols ?? 0) > 0) return false;
  return true;
}

// cantPayReason words a preview that could not plan the payment.
export function cantPayReason(preview: AutoTapPreview): string {
  const missing = (preview.missing ?? []).join("");
  return missing ? `Can't pay ${missing}` : "Can't pay this cost";
}

// dragVerdict is the live gold / red answer. The timing and legality
// half is the same Legality the hand's click path greys cards with
// (canCastFromHand — the server's legal-move list); the mana half is
// the auto-tap preview's answer, when it has arrived and when it is
// decisive for this card. A preview that failed or is still in flight
// says nothing, and the drag stays gold.
export function dragVerdict(
  card: CardView,
  legality: Legality,
  preview: AutoTapPreview | null | undefined,
): Verdict {
  if (isLand(card)) return { castable: false, reason: LAND_DRAG_REASON };
  if (!legality.legal) return { castable: false, reason: legality.reason ?? "Can't cast now" };
  if (preview && !preview.ok && previewDecidesMana(card)) {
    return { castable: false, reason: cantPayReason(preview) };
  }
  return { castable: true };
}

// wantsPreview: is an auto-tap preview worth fetching for this drag?
// Not for a land (it never casts) and not for a card the legality check
// already refused (it will snap back whatever the mana says).
export function wantsPreview(card: CardView, legality: Legality): boolean {
  return !isLand(card) && legality.legal;
}

// previewSourceIDs lists the sources the auto-tapper would use, for the
// board highlight. `sources` (#1285) is the described form and names
// hand cards too (a Spirit Guide); `plan` is the older bare list.
export function previewSourceIDs(preview: AutoTapPreview | null | undefined): string[] {
  if (!preview || !preview.ok) return [];
  if (preview.sources && preview.sources.length > 0) {
    return preview.sources.map((s) => s.card_id).filter((id) => id !== "");
  }
  return (preview.plan ?? []).filter((id) => id !== "");
}

// autoTapHighlight is the set of permanents (and hand cards) the
// in-flight drag's auto-tap preview would spend. Card.svelte rings the
// ones it renders. Written only by the drag: set when the preview
// lands, cleared on drop, snap-back or cancel.
export const autoTapHighlight = guardedWritable<ReadonlySet<string>>(
  new Set<string>(),
  "autoTapHighlight",
);

export function clearAutoTapHighlight(): void {
  autoTapHighlight.set(new Set<string>());
}
