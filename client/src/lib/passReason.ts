// passReason.ts — say why an automatic pass passed, with a one-tap Undo
// (#3343, S72).
//
// Every pass_priority the client sends without a direct click goes
// through one helper in Game.svelte (`sendPass`), and every one of them
// carries a human reason. The reason is shown in a short-lived dock chip,
// "Passed for you: <reason>", and written into the client log line a bug
// report attaches, so neither the player nor we have to guess why a
// window went by.
//
// This module is the pure half: the wording (a mirror of the branches in
// autopassDecision.ts that end in a pass), the rule for when the chip's
// Undo still works, and the chip as a dock request. The component keeps
// the timers and the send.
//
// THE UNDO. The chip's Undo sends the existing `undo` action and spends
// the existing per-turn budget; there is no free undo here (the server
// has one for an automatic answer, `undo_auto_answer`, and none for a
// pass). It is offered only while the existing undo would take THIS pass
// back, which the client can read off four facts (passNoticeUndoable):
//
//   1. the server acknowledged the pass (an `ack` naming its frame), so
//      the pass is on the undo stack at all;
//   2. the snapshot on screen is that state or a later one;
//   3. this client has sent nothing since, so the pass is still the
//      viewer's most recent action; and
//   4. the seat's `undo_blocked` / budget say an undo is allowed now, and
//      its `undo_auto_answer` is empty (an automatic answer sitting on
//      top of the pass would be what Undo took back instead).
//
// Anything else and the chip shows the reason alone. That is
// conservative on purpose: a missing Undo costs a click on the dock's own
// Undo; a wrong one takes back something the player did not mean to.

import type { DockRequest } from "./dock";
import { L } from "./labels";
import type { AutopassGates, AutopassVerdict } from "./autopassDecision";
import { isBluff } from "./autopassDecision";
import type { PlayerView } from "./protocol";
import { stepLabel } from "./turn";

// ---- the wording ------------------------------------------------------------

/** What the chip says when the bluff timer is what passed. */
export const BLUFF_PASS_REASON = "bluff timer ended";

/** What the chip says when the stack hold ran out. */
export const HOLD_PASS_REASON = "stack hold ended, nothing to respond with";

/** The reason for a pass that is the second half of the player's own click. */
export const ATTACK_PASS_REASON = "attackers declared";
export const NO_BLOCKS_PASS_REASON = "no blocks declared";

/** The reason a direct click (Next, the key, End turn's pass) logs. */
export const MANUAL_PASS_REASON = "you passed";

/**
 * autopassReason says why autopassDecision(g) returned "pass", in the
 * order the decision reaches each branch. Only meaningful for a verdict
 * of "pass"; for anything else it still returns a sentence, never "".
 */
export function autopassReason(g: AutopassGates): string {
  // Rule 4: Skip to my turn passes everything.
  if (g.autopassToggle) return "skipping to your turn";
  if (!g.stackEmpty) {
    if (g.ownsEveryStackItem) return "everything on the stack is yours";
    return "nothing to respond with";
  }
  const step = stepLabel(g.step);
  const where = step ? ` in ${step}` : "";
  if (g.stepStop === true) return `nothing to do${where}`;
  if ((g.combatWindow || g.oppEndWindow) && !g.hasResponse) {
    return `nothing to respond with${where}`;
  }
  return step ? `no stop set for ${step}` : "no stop set here";
}

/**
 * timedPassReason says why a pass that waited (a timed bluff or a stack
 * hold) fired: `kind` is the timer that was armed, `verdict` what the
 * decision says now. A bluff verdict is a bluff timer ending whichever
 * timer was armed; a plain pass after a hold is the hold running out;
 * a plain pass after a bluff timer means the settings moved while it
 * waited, so the ordinary reason applies.
 */
export function timedPassReason(
  kind: "bluff" | "hold",
  verdict: AutopassVerdict,
  g: AutopassGates,
): string {
  if (isBluff(verdict)) return BLUFF_PASS_REASON;
  return kind === "hold" ? HOLD_PASS_REASON : autopassReason(g);
}

// ---- the notice -------------------------------------------------------------

/** How long the chip stays, unless another request takes the dock. */
export const PASS_NOTICE_MS = 6000;

/** What the chip is about: one automatic pass this client sent. */
export interface PassNotice {
  // The pass_priority frame's id (sendAction's return), the key the
  // server's ack and any error frame name it by.
  frame: string;
  // "nothing to respond with".
  reason: string;
  // client.actionsSent right after the pass left the socket. Any later
  // action by the viewer makes it differ.
  actionsSent: number;
}

/** The live facts passNoticeUndoable reads. */
export interface PassNoticeFacts {
  // The newest ack the server sent, or null.
  ack: { id: string; seq: number } | null;
  // The seq of the snapshot on screen.
  lastSeq: number;
  // client.actionsSent now.
  actionsSent: number;
  // The viewer's own seat in the snapshot on screen.
  seat: PlayerView | null | undefined;
  // undoBlockedReason(seat, isAdmin): null when the dock's Undo is
  // enabled, the reason phrase when it is not.
  undoBlocked: string | null;
}

/**
 * passNoticeUndoable is true while the chip's Undo would take this pass
 * back. See the file comment for the four conditions.
 */
export function passNoticeUndoable(n: PassNotice, f: PassNoticeFacts): boolean {
  if (!f.ack || f.ack.id !== n.frame) return false;
  if (f.lastSeq < f.ack.seq) return false;
  if (f.actionsSent !== n.actionsSent) return false;
  if ((f.seat?.undo_auto_answer ?? 0) !== 0) return false;
  return f.undoBlocked === null;
}

/** The chip's sentence. */
export function passNoticeText(n: PassNotice): string {
  return `Passed for you: ${n.reason}`;
}

/**
 * passNoticeRequest is the chip as a dock request: a status, not a
 * question (the autoAnswerNoticeRequest shape). It is a `step` request,
 * so it takes no keys, blocks nothing and leaves `next` where it is; any
 * other request outranks it. Undo is a row button only while `undoable`;
 * otherwise the chip is the reason alone.
 */
export function passNoticeRequest(
  n: PassNotice,
  undoable: boolean,
  onUndo: () => void,
): DockRequest {
  return {
    rank: "step",
    label: L.autoPassNotice,
    tag: "auto-pass",
    tone: "plain",
    question: passNoticeText(n),
    live: true,
    row: undoable
      ? [
          {
            id: "auto-pass-undo",
            label: "Undo",
            // Not "Undo…": the dock's own Undo ("Undo (3 left)") is the
            // one the e2e suite counts with /^Undo/.
            ariaLabel: L.undoAutoPass,
            title: "Take this pass back (uses one of your undos)",
            onPress: onUndo,
          },
        ]
      : undefined,
  };
}
