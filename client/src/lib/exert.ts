// exert.ts — the client's half of exert as it attacks (ADR 0130 §7,
// owner decision 1: every path asks).
//
// The choice is the player's (CR 508.1g, 701.43d), so no click path
// makes it for them. Whether a creature may be exerted is the server's
// answer (the digest's exert_on_attack, read through
// LegalActions.canExertOnAttack), never oracle text. Pure helpers, so
// the click flow is testable without rendering Game.svelte.

import type { LegalActions } from "./legalActions";

// AttackerSelection is the two-click attack flow's selected attacker.
// `exert` is absent for a creature that can't be exerted as it attacks,
// null while the player hasn't answered the dock's question, and true
// or false once they have.
export interface AttackerSelection {
  kind: "attacker";
  cardID: string;
  exert?: boolean | null;
}

// selectAttackerFor is the selection a click on `cardID` makes. A
// creature the server lists as exertable starts with the question
// open. One already declared is not listed (the enumerator drops a
// declared attacker), so a re-point asks nothing, and sends no exert:
// the server keeps an exert staged before the re-point, and refuses one
// chosen after the declaration (ruling: you exert as you declare it).
export function selectAttackerFor(cardID: string, gate: LegalActions): AttackerSelection {
  return gate.canExertOnAttack(cardID)
    ? { kind: "attacker", cardID, exert: null }
    : { kind: "attacker", cardID };
}

// exertUnanswered reports whether the selection still waits on the
// exert question. A seat click then commits nothing.
export function exertUnanswered(sel: AttackerSelection): boolean {
  return sel.exert === null;
}

// DeclareAttackerParams is declare_attacker's wire shape.
export type DeclareAttackerParams = {
  attacker: string;
  target: string;
  auto_tap: true;
  exert?: true;
};

// declareAttackerParams builds the single declaration for a seat or
// permanent click, or null while the exert question is open.
export function declareAttackerParams(
  sel: AttackerSelection,
  target: string,
): DeclareAttackerParams | null {
  if (exertUnanswered(sel)) return null;
  const p: DeclareAttackerParams = {
    attacker: sel.cardID,
    target,
    // ADR 0080 (#1063): the CR 508.1a attack tax may need lands
    // tapped. Inert at a table with no attack tax on it.
    auto_tap: true,
  };
  if (sel.exert === true) p.exert = true;
  return p;
}

// isCantExertError recognises the server's ErrCantExert refusal
// (`bad_request`, "this creature can't be exerted as it attacks").
export function isCantExertError(message: string | undefined): boolean {
  return !!message && message.includes("can't be exerted as it attacks");
}

// CANT_EXERT_SENTENCE is what the rejected toast says instead: what
// happened, that nothing was declared, and what to do next.
export const CANT_EXERT_SENTENCE =
  "That creature can't be exerted as it attacks right now, so nothing was declared. Attack without exerting it, or choose again.";
