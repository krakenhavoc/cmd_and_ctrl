// enlist.ts — the client's half of enlist (CR 702.154, #5658).
//
// "As this creature attacks, you may tap up to one untapped creature you
// control that you didn't choose to attack with" is an optional cost to
// attack (CR 702.154b, 508.1g), so, like exert (lib/exert.ts), no click
// path chooses it for the player. Which creatures an attacker may enlist
// is the server's answer (the digest's enlist_on_attack, read through
// LegalActions.enlistOnAttack), never oracle text. Pure helpers.

import type { CardView } from "./protocol";
import type { AttackerSelection } from "./exert";
import type { LegalActions } from "./legalActions";

// EnlistStage is the part of a staged attack this file reads.
export interface EnlistStage {
  attacker: string;
  enlist?: string;
}

// ENLIST_NOTE is the reminder the question carries, in the card's own
// reminder-text words.
export const ENLIST_NOTE =
  "The enlisted creature is tapped, and its power is added to the attacker's until end of turn.";

// enlistTaken is the creatures a new attacker can't enlist because of
// what is already staged: the staged attackers and the creatures they
// enlist. One creature is tapped once and can't also attack.
export function enlistTaken(stage: readonly EnlistStage[]): Set<string> {
  const out = new Set<string>();
  for (const s of stage) {
    out.add(s.attacker);
    if (s.enlist) out.add(s.enlist);
  }
  return out;
}

// enlistCandidates is who `cardID` may enlist: the server's list, less
// the creatures the stage has taken.
export function enlistCandidates(
  cardID: string,
  gate: LegalActions,
  taken: ReadonlySet<string> = new Set(),
): string[] {
  return gate.enlistOnAttack(cardID).filter((id) => id !== cardID && !taken.has(id));
}

// withEnlistQuestion opens the enlist question on a fresh selection when
// the creature has someone to enlist. A selection with the exert question
// keeps only that one: no card asks both.
export function withEnlistQuestion(
  sel: AttackerSelection,
  gate: LegalActions,
  taken: ReadonlySet<string> = new Set(),
): AttackerSelection {
  if (sel.exert !== undefined) return sel;
  return enlistCandidates(sel.cardID, gate, taken).length > 0 ? { ...sel, enlist: null } : sel;
}

// enlistUnanswered reports whether the selection still waits on the
// enlist question. A seat click then commits nothing.
export function enlistUnanswered(sel: AttackerSelection): boolean {
  return sel.enlist === null;
}

// EnlistOption is one creature the dock or the picker offers to enlist.
export interface EnlistOption {
  id: string;
  name: string;
  // The creature's power as the board shows it, when it has one.
  power?: number;
}

// enlistOptions names the candidates from the board, in the server's
// order. Creatures that read the same (name and power) are offered once:
// which of five identical tokens is tapped is not a choice.
export function enlistOptions(
  ids: readonly string[],
  cards: readonly CardView[] | undefined,
): EnlistOption[] {
  const out: EnlistOption[] = [];
  const seen = new Set<string>();
  for (const id of ids) {
    const c = cards?.find((x) => x.instance_id === id);
    const name = c?.name || "unknown creature";
    const power = typeof c?.power === "number" ? c.power : undefined;
    const key = `${name}\u0000${power ?? ""}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push(power === undefined ? { id, name } : { id, name, power });
  }
  return out;
}

// enlistOptionLabel is "Elk (+3)": the creature and what it adds.
export function enlistOptionLabel(o: EnlistOption): string {
  return o.power === undefined ? o.name : `${o.name} (+${Math.max(o.power, 0)})`;
}

// isCantEnlistError recognises the server's ErrCantEnlist and
// ErrEnlistedCantAttack refusals (`bad_request`).
export function isCantEnlistError(message: string | undefined): boolean {
  return (
    !!message &&
    (message.includes("can't be enlisted by this attacker") ||
      message.includes("has been enlisted, so it can't attack"))
  );
}

// CANT_ENLIST_SENTENCE is what the rejected toast says instead: what
// happened, that nothing was declared, and what to do next.
export const CANT_ENLIST_SENTENCE =
  "That creature can't be enlisted right now, or an enlisted creature was also sent to attack, so nothing was declared. Choose your attackers again.";
