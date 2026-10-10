// boardChoicePick.ts — a pending choice's permanents, picked on the
// board as well as in its sheet (#2880, ADR 0117 click to act, ADR 0111
// the action dock).
//
// "Sacrifice a creature" (Fleshbag Marauder) used to be answered only in
// the sheet: a grid of small copies of the permanents, with the dock's
// "Sacrifice" waiting on a pick there. A click on the creature itself did
// nothing. Now the permanents the choice offers are highlighted where
// they sit, a click on one picks it (a second click puts it back), and
// the sheet and the board share ONE selection: a pick in either shows in
// both, and the sheet's "n / m selected" and its confirm in the dock read
// it. The answer is the sheet's own {choice_id, card_ids}, sent by the
// sheet's own confirm, so the server sees no difference.
//
// Who owns what:
//   - ChoicePromptModal owns the open prompt. While its card grid is up
//     and some of the options are permanents on the battlefield, it
//     publishes them here (`eligible`) with its bounds and its
//     selection, and it copies a selection changed here back into its
//     own.
//   - The board reads this store: Card draws the highlight and the pick,
//     Board's click intercept toggles a pick (and swallows a click on
//     any other permanent while the choice is open), and a seat that
//     holds an eligible permanent is drawn in full.
//
// This is not #2394's board pick (boardAnsweredChoice.ts). That one takes
// a choice OFF the sheet entirely ("untap up to five lands" is answered
// on the board, with "Show as list" as the way back). This one keeps the
// sheet: a sacrifice is a decision a player wants to read before making,
// and the sheet says what is being asked and why. The two never overlap:
// a choice #2394 answers on the board has no sheet to publish this.
//
// Pure functions plus one store, so the shared selection and the dock's
// enable rule are pinned by boardChoicePick.test.ts without a component.

import { guardedWritable } from "./guardedStore";
import type { GameView, PendingChoiceView } from "./protocol";

export interface BoardChoicePick {
  // The pending choice this selection answers.
  choiceID: string;
  // The options that are permanents on the battlefield: what the board
  // highlights and lets the player click. Every other permanent is not
  // clickable while the choice is open.
  eligible: ReadonlySet<string>;
  // The one selection, shared with the sheet. It may also hold an option
  // that is not a permanent (proliferate's players): the board never
  // toggles one, but it counts toward `max`.
  selected: ReadonlySet<string>;
  // The sheet's bounds: the confirm is enabled between them.
  min: number;
  max: number;
}

export const boardChoicePick = guardedWritable<BoardChoicePick | null>(null, "boardChoicePick");

// The choice kinds whose sheet is the card grid and whose options are,
// or may be, permanents on the battlefield. Every one is answered with
// {card_ids}, from the grid's one `selected` set.
//
// Left out on purpose: the board-answered kinds (pick_target,
// legend_rule, retarget, choose_protector, and choose_cards /
// untap_choice while #2394 answers them on the board — boardAnsweredChoice.ts);
// pay_unless's sacrifice and tap payments (their own pick lists, beside a
// "Pay" that is not one stray click away); damage_assignment and
// divide_shield (a number per permanent, not a pick); and every kind whose
// options are cards in a hand, a library or a revealed set.
const BOARD_PICK_KINDS = new Set<string>([
  // "Each player sacrifices a creature" (Fleshbag Marauder, Grave Pact).
  "sacrifice_choice",
  // "Sacrifice a Forest so it enters" (Heart of Yavimaya).
  "entry_sacrifice",
  // "Choose among the permanents somebody else controls" (Tragic Arrogance).
  "their_permanents",
  // "Choose N of your own permanents" (Scapeshift).
  "own_permanents",
  // "Choose your Ring-bearer".
  "ring_bearer",
  // Proliferate: the permanents are on the board; the players stay chips.
  "proliferate",
  // "Enter as a copy of a creature on the battlefield" (Clone).
  "copy_target",
  // "A source of your choice": the ones on the battlefield.
  "choose_source",
  // #2394's two card-set picks, once the player asked for the list: the
  // sheet is open, and the board still works beside it.
  "choose_cards",
  "untap_choice",
]);

export function isBoardPickKind(kind: string): boolean {
  return BOARD_PICK_KINDS.has(kind);
}

// boardPickEligible is the set of a choice's options that are permanents
// on the battlefield, or null when the choice is not one the board can
// answer (another kind, or none of its options on the battlefield).
export function boardPickEligible(
  choice: PendingChoiceView | null | undefined,
  view: Pick<GameView, "battlefield">,
): Set<string> | null {
  if (!choice || !isBoardPickKind(choice.kind)) return null;
  const onBattlefield = new Set((view.battlefield?.cards ?? []).map((c) => c.instance_id));
  const ids = (choice.options ?? [])
    .map((o) => o.instance_id)
    .filter((id) => onBattlefield.has(id));
  return ids.length > 0 ? new Set(ids) : null;
}

// selectionLegal is the confirm's rule, the sheet's and the dock's: the
// selection is between the choice's floor and its ceiling.
export function selectionLegal(size: number, min: number, max: number): boolean {
  return size >= min && size <= max;
}

export function canConfirmBoardPick(s: BoardChoicePick | null): boolean {
  return s !== null && selectionLegal(s.selected.size, s.min, s.max);
}

// isBoardPickable reports whether a click on this permanent picks it.
export function isBoardPickable(s: BoardChoicePick | null, id: string): boolean {
  return s !== null && s.eligible.has(id);
}

export function isBoardPicked(s: BoardChoicePick | null, id: string): boolean {
  return s !== null && s.selected.has(id);
}

// toggleBoardPick is a board click, by the sheet grid's own rule: a
// picked permanent is put back, an unpicked one is picked while there is
// room, and at the ceiling a click on another does nothing until one is
// put back. A permanent the choice does not offer changes nothing.
// Returns the same object when nothing changed.
export function toggleBoardPick(s: BoardChoicePick, id: string): BoardChoicePick {
  if (!s.eligible.has(id)) return s;
  const next = new Set(s.selected);
  if (next.has(id)) {
    next.delete(id);
  } else if (next.size < s.max) {
    next.add(id);
  } else {
    return s;
  }
  return { ...s, selected: next };
}

// sameSelection compares two selections as sets.
export function sameSelection(a: ReadonlySet<string>, b: ReadonlySet<string>): boolean {
  if (a.size !== b.size) return false;
  for (const id of a) if (!b.has(id)) return false;
  return true;
}

// publishBoardPick is ChoicePromptModal's half: the open choice's board
// pick, or null when there is none. It writes only when something
// changed, so the modal's selection and this store can follow each other
// without a loop.
export function publishBoardPick(next: BoardChoicePick | null): void {
  boardChoicePick.update((cur) => {
    if (next === null) return cur === null ? cur : null;
    if (
      cur !== null &&
      cur.choiceID === next.choiceID &&
      cur.min === next.min &&
      cur.max === next.max &&
      sameSelection(cur.eligible, next.eligible) &&
      sameSelection(cur.selected, next.selected)
    ) {
      return cur;
    }
    return next;
  });
}

// pickOnBoard is the board's half: a click on a permanent while a choice
// is open. True when the click belongs to the choice — it toggled an
// eligible permanent, or it landed on one the choice does not offer and
// is swallowed, so a stray click cannot tap a land or open a menu while
// the game waits on the answer. False when no choice is open.
export function pickOnBoard(id: string): boolean {
  let handled = false;
  boardChoicePick.update((cur) => {
    if (cur === null) return cur;
    handled = true;
    return toggleBoardPick(cur, id);
  });
  return handled;
}
