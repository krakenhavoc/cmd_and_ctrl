// attackSelect.ts — #4326: select several attackers, then one target.
//
// At a table where a creature has more than one legal attack target, a
// click on it selects it and a click on a player, planeswalker or battle
// stages it there (#3266, lib/attackStage.ts). Sending five creatures at
// one opponent took ten clicks. Now a click on a second such creature
// joins the selection, so several are selected at once, and one target
// click stages every one of them that may attack it. Those that may not
// stay selected, to be pointed somewhere else.
//
// What joins: a creature not yet declared whose selection asks nothing.
// A creature with an exert question (ADR 0130 §7) is selected alone,
// since the dock asks that question about one creature, and a declared
// attacker being re-pointed is declare_attacker's, one at a time, as
// before. Clicking either starts a new selection, as every click did.
//
// The selection is Game.svelte's: `combatSelection` is the creature
// clicked last (the dock names it and the board lights its targets) and
// `more` the others, oldest first. All pure, apart from the store the
// board reads to draw the others as selected.

import type { AttackerSelection } from "./exert";
import { guardedWritable } from "./guardedStore";
import { stageAt, type StagedAttack } from "./attackStage";

/**
 * attackPicked is the selected attackers other than the one the board
 * already draws from `selectedCombatCardID`. Game.svelte publishes it;
 * BattlefieldRow reads it to draw them as selected.
 */
export const attackPicked = guardedWritable<ReadonlySet<string>>(new Set(), "attackPicked");

/**
 * joinsSelection says whether a click making `next` adds it to the
 * selection held in `current` rather than replacing it.
 */
export function joinsSelection(
  current: AttackerSelection | null,
  next: AttackerSelection,
  declared: (cardID: string) => boolean,
): boolean {
  if (!current || current.cardID === next.cardID) return false;
  if (current.exert !== undefined || next.exert !== undefined) return false;
  // #5658: so is one with an enlist question.
  if (current.enlist !== undefined || next.enlist !== undefined) return false;
  return !declared(current.cardID) && !declared(next.cardID);
}

/**
 * dropFromSelection takes `cardID` out of a selection of `primary` and
 * `more`. When it was the primary, the last of the others takes its
 * place, so the selection only empties when nothing is left in it.
 */
export function dropFromSelection(
  primary: AttackerSelection,
  more: readonly string[],
  cardID: string,
): { primary: AttackerSelection | null; more: string[] } {
  if (primary.cardID !== cardID) {
    return { primary, more: more.filter((id) => id !== cardID) };
  }
  if (more.length === 0) return { primary: null, more: [] };
  const last = more[more.length - 1]!;
  return { primary: { kind: "attacker", cardID: last }, more: more.slice(0, -1) };
}

/**
 * stageSelectionAt stages every selected creature that may attack
 * `target` there, in the order they were selected, and returns what is
 * left selected. `ids` is the selection oldest first, the primary last.
 * Null while the primary's exert question is open (only a lone
 * selection can have one): nothing is staged until it is answered.
 */
export function stageSelectionAt(
  stage: readonly StagedAttack[],
  primary: AttackerSelection,
  more: readonly string[],
  target: string,
  mayAttack: (cardID: string, target: string) => boolean,
): { stage: StagedAttack[]; left: string[] } | null {
  let next: StagedAttack[] = [...stage];
  const left: string[] = [];
  for (const id of [...more, primary.cardID]) {
    if (!mayAttack(id, target)) {
      left.push(id);
      continue;
    }
    const sel: AttackerSelection =
      id === primary.cardID ? primary : { kind: "attacker", cardID: id };
    const staged = stageAt(next, sel, target);
    if (!staged) return null;
    next = staged;
  }
  return { stage: next, left };
}

/**
 * selectionName is how the dock names the selection: the creature
 * clicked last, and how many others are with it.
 */
export function selectionName(name: string, others: number): string {
  return others > 0 ? `${name} and ${others} more` : name;
}
