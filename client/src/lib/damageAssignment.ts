// damageAssignment.ts — combat damage assignment (CR 510.1c-d) without
// the typing (#2956, ADR 0147).
//
// The server ships every damage_assignment prompt with its canonical
// split (`suggested`, legal.CanonicalDamageSplit: the answer the bots
// give), each blocker's lethal damage (`lethal`) and whether the
// attacker's damage covers lethal for every blocker (`covers_lethal`).
// This module never works out lethal damage itself: deathtouch, damage
// already marked and the order the split kills in are the server's.
//
//   - Auto-assign. With `gameplay.autoAssignCombatDamage` on (the
//     default) and `covers_lethal` set, there is nothing left to
//     choose: lethal to each blocker, the rest over with trample or onto
//     the last blocker without. DamageAutoAssign.svelte sends the
//     suggested split unasked and says so in the dock;
//     ChoicePromptModal holds its sheet back meanwhile.
//   - Otherwise the sheet opens pre-filled with the suggested split, and
//     the player ticks damage up and down on the blockers themselves
//     (DamageStepper on each blocking Card) or in the sheet. Both read
//     and write the one `boardDamageAssign` store, so the board and the
//     sheet always agree, and the sheet's "Deal damage" sends it.
//
// ResolveDamageAssignment stays the authority: a split it refuses is
// shown as the sheet's refusal, as before.

import { guardedWritable } from "./guardedStore";
import { L } from "./labels";
import type { DockRequest } from "./dock";
import type { DamageAssignmentView, PendingChoiceView } from "./protocol";

/** A split in progress: per-blocker amounts and the trample bucket. */
export interface DamageShares {
  amounts: Record<string, number>;
  trample: number;
}

/** The resolve_choice payload for a damage assignment. */
export interface DamageAnswer {
  assignments: { blocker_id: string; amount: number }[];
  trample_to_player: number;
}

/**
 * initialShares is what the prompt opens on: the server's suggested
 * split when it sent one, so trample's leftover is on the player and not
 * 0, else every blocker at 0.
 */
export function initialShares(frame: DamageAssignmentView): DamageShares {
  const amounts: Record<string, number> = {};
  for (const id of frame.blocker_card_ids) amounts[id] = 0;
  const s = frame.suggested;
  if (!s) return { amounts, trample: 0 };
  for (const a of s.assignments) {
    if (a.blocker_id in amounts) amounts[a.blocker_id] = Math.max(0, a.amount);
  }
  return { amounts, trample: frame.allow_trample ? Math.max(0, s.trample_to_player ?? 0) : 0 };
}

export function assignedTotal(s: DamageShares): number {
  let n = s.trample;
  for (const v of Object.values(s.amounts)) n += v;
  return n;
}

/** lethalOf is the server's lethal damage for a blocker, or null. */
export function lethalOf(frame: DamageAssignmentView, id: string): number | null {
  const i = frame.blocker_card_ids.indexOf(id);
  if (i < 0 || !frame.lethal || i >= frame.lethal.length) return null;
  return frame.lethal[i];
}

/**
 * stepBlocker ticks one blocker's damage by +1 or -1. Up takes from the
 * unassigned damage first, then from the trample bucket, so a split that
 * is already full can still be moved onto a blocker one click at a time.
 * Down returns the point to the unassigned damage. Returns the same
 * object when nothing can change.
 */
export function stepBlocker(
  s: DamageShares,
  id: string,
  delta: 1 | -1,
  power: number,
): DamageShares {
  if (!(id in s.amounts)) return s;
  const cur = s.amounts[id];
  if (delta < 0) {
    if (cur <= 0) return s;
    return { ...s, amounts: { ...s.amounts, [id]: cur - 1 } };
  }
  if (assignedTotal(s) < power) {
    return { ...s, amounts: { ...s.amounts, [id]: cur + 1 } };
  }
  if (s.trample > 0) {
    return { amounts: { ...s.amounts, [id]: cur + 1 }, trample: s.trample - 1 };
  }
  return s;
}

/** stepTrample ticks the trample bucket; up only from unassigned damage. */
export function stepTrample(s: DamageShares, delta: 1 | -1, power: number): DamageShares {
  if (delta < 0) return s.trample > 0 ? { ...s, trample: s.trample - 1 } : s;
  return assignedTotal(s) < power ? { ...s, trample: s.trample + 1 } : s;
}

/** setBlocker and setTrample are the sheet's number fields. */
export function setBlocker(s: DamageShares, id: string, n: number): DamageShares {
  if (!(id in s.amounts)) return s;
  return { ...s, amounts: { ...s.amounts, [id]: Math.max(0, Math.floor(n) || 0) } };
}

export function setTrample(s: DamageShares, n: number): DamageShares {
  return { ...s, trample: Math.max(0, Math.floor(n) || 0) };
}

/**
 * trampleShort is the blockers still short of lethal while damage is
 * going over (CR 702.19b): the server refuses that split, so the sheet
 * says why instead of offering a confirm that fails. Empty when nothing
 * tramples over, or the server sent no lethal figures.
 */
export function trampleShort(frame: DamageAssignmentView, s: DamageShares): string[] {
  if (s.trample <= 0) return [];
  return frame.blocker_card_ids.filter((id) => {
    const lethal = lethalOf(frame, id);
    return lethal !== null && (s.amounts[id] ?? 0) < lethal;
  });
}

export function canDealDamage(frame: DamageAssignmentView, s: DamageShares): boolean {
  return assignedTotal(s) === frame.attacker_power && trampleShort(frame, s).length === 0;
}

/** toAnswer is the resolve_choice payload, in the frame's blocker order. */
export function toAnswer(frame: DamageAssignmentView, s: DamageShares): DamageAnswer {
  return {
    assignments: frame.blocker_card_ids.map((id) => ({
      blocker_id: id,
      amount: s.amounts[id] ?? 0,
    })),
    trample_to_player: s.trample,
  };
}

// ---- Auto-assign --------------------------------------------------------

/**
 * autoAssignAnswer is the answer to send for the viewer unasked, or null
 * when the prompt must be asked: the setting is off, the prompt is not
 * the viewer's, or the damage does not cover lethal for every blocker
 * (so which of them die is a real choice).
 */
export function autoAssignAnswer(
  choice: PendingChoiceView | null | undefined,
  viewerID: string | null,
  enabled: boolean,
): DamageAnswer | null {
  if (!enabled || !choice || !viewerID) return null;
  if (choice.kind !== "damage_assignment" || choice.chooser !== viewerID) return null;
  const frame = choice.damage_assignment;
  if (!frame || !frame.covers_lethal || !frame.suggested) return null;
  const s = initialShares(frame);
  if (assignedTotal(s) !== frame.attacker_power) return null;
  return toAnswer(frame, s);
}

/**
 * One auto-assigned prompt: when it was sent, and whether the sheet may
 * take it back. `expired` is set a few seconds after sending: if the
 * prompt is still (or again, after an undo) pending then, the sheet asks.
 */
export interface AutoAssignRecord {
  sentAt: number;
  expired: boolean;
}

export const autoAssignRecords = guardedWritable<Record<string, AutoAssignRecord>>(
  {},
  "autoAssignRecords",
);

/** How long the sheet waits on a sent auto-assignment before it asks. */
export const AUTO_ASSIGN_GRACE_MS = 4000;

/**
 * autoAssignHolds is ChoicePromptModal's rule: true while the sheet
 * should stay closed for this prompt because the auto-assignment is
 * about to be sent or is on its way. A prompt whose record expired (the
 * answer was refused, or an undo brought the prompt back) is asked.
 */
export function autoAssignHolds(
  choice: PendingChoiceView,
  viewerID: string | null,
  enabled: boolean,
  records: Record<string, AutoAssignRecord>,
): boolean {
  if (autoAssignAnswer(choice, viewerID, enabled) === null) return false;
  const rec = records[choice.id];
  return !rec || !rec.expired;
}

/**
 * autoAssignText is the dock notice's line: what went where.
 * "Vivi Ornitier: 4 to Solphim, 3 to Professional Face-Breaker, 2 to
 * Corsair Captain, 35 to Bob (trample)".
 */
export function autoAssignText(
  attacker: string,
  answer: DamageAnswer,
  name: (id: string) => string,
  defender: string,
): string {
  const parts = answer.assignments
    .filter((a) => a.amount > 0)
    .map((a) => `${a.amount} to ${name(a.blocker_id)}`);
  if (answer.trample_to_player > 0) {
    parts.push(`${answer.trample_to_player} to ${defender} (trample)`);
  }
  return `${attacker}: ${parts.join(", ")}`;
}

/** How long the notice stays, unless another request takes the dock. */
export const AUTO_ASSIGN_NOTICE_MS = 6000;

/**
 * autoAssignNoticeRequest is the notice as a dock request: a status,
 * like ADR 0127's automatic-answer notice. Undo takes the assignment
 * back and asks; Always ask turns the setting off.
 */
export function autoAssignNoticeRequest(
  text: string,
  undoable: boolean,
  onUndo: () => void,
  onAlwaysAsk: () => void,
): DockRequest {
  return {
    rank: "step",
    label: L.damageAutoAssigned,
    tag: "damage",
    tone: "plain",
    question: text,
    detail: "assigned for you",
    live: true,
    row: [
      {
        id: "auto-assign-undo",
        label: L.undoAutomaticAnswer,
        disabled: !undoable,
        title: undoable
          ? "Take this assignment back and assign it by hand"
          : "Someone has acted since",
        onPress: onUndo,
      },
      {
        id: "auto-assign-always-ask",
        label: L.alwaysAskCombatDamage,
        title: "Turn off Auto-assign combat damage; this assignment stands",
        onPress: onAlwaysAsk,
      },
    ],
  };
}

// ---- The board ----------------------------------------------------------

/**
 * BoardDamageAssign is an open damage-assignment prompt as the board
 * draws it: steppers on each blocker, the running total on the attacker.
 * ChoicePromptModal publishes it while its sheet is up and copies a
 * change made on the board back into its own shares.
 */
export interface BoardDamageAssign {
  choiceID: string;
  attackerID: string;
  power: number;
  allowTrample: boolean;
  blockerDivides: boolean;
  lethal: Record<string, number>;
  shares: DamageShares;
}

export const boardDamageAssign = guardedWritable<BoardDamageAssign | null>(
  null,
  "boardDamageAssign",
);

export function publishBoardDamage(next: BoardDamageAssign | null): void {
  boardDamageAssign.set(next);
}

/** stepOnBoard is a stepper press on a blocker's card. */
export function stepOnBoard(id: string, delta: 1 | -1): void {
  boardDamageAssign.update((s) => {
    if (!s) return s;
    const shares = stepBlocker(s.shares, id, delta, s.power);
    return shares === s.shares ? s : { ...s, shares };
  });
}

/** stepTrampleOnBoard is a stepper press on the attacker's trample line. */
export function stepTrampleOnBoard(delta: 1 | -1): void {
  boardDamageAssign.update((s) => {
    if (!s || !s.allowTrample) return s;
    const shares = stepTrample(s.shares, delta, s.power);
    return shares === s.shares ? s : { ...s, shares };
  });
}

export function sameShares(a: DamageShares, b: DamageShares): boolean {
  if (a.trample !== b.trample) return false;
  const ka = Object.keys(a.amounts);
  if (ka.length !== Object.keys(b.amounts).length) return false;
  return ka.every((k) => a.amounts[k] === b.amounts[k]);
}
