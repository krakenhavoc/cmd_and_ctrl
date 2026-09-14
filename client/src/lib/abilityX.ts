import type { ActivatedAbilityView } from "./protocol";

// abilityX.ts — the arithmetic behind the X picker an activated
// ability opens (CR 602.2b), kept out of Board.svelte so it can be
// tested without a component harness.
//
// An ability's X is the SAME question a spell's X is, asked at the
// same point in the announcement, so it reuses XCostModal. What is
// different is the price and the floor, and both come off the wire
// rather than being re-derived from a cost string on this side:
//
//   x_slots  how many {X} tokens the cost carries. Treasure Vault's
//            "{X}{X}" is two, so the same mana buys half the X.
//   min_x    the printed floor. Helm of Obedience's "X can't be 0"
//            ships 1, and the server refuses anything under it.
//
// Re-parsing "{X}{X}" here would make the client a second parser of
// a syntax the server already owns, which is how the two drift.

// abilityDemandsX reports whether activating this ability needs the
// X prompt. The server derives the flag from the cost string, so
// there is exactly one place that reads the braces.
export function abilityDemandsX(ability: ActivatedAbilityView): boolean {
  return ability.demands_x === true;
}

// abilityMinX is the smallest value the player may announce. Absent
// means the ordinary floor of zero; a negative on the wire is
// nonsense and clamps to zero rather than being trusted.
export function abilityMinX(ability: ActivatedAbilityView): number {
  return Math.max(0, ability.min_x ?? 0);
}

// abilityXSlots is how many {X} tokens the cost carries. Anything
// missing or nonsensical reads as one, which is what every printed
// {X} but Treasure Vault's is — the safe direction, because it
// over-estimates the price rather than under-estimating it.
export function abilityXSlots(ability: ActivatedAbilityView): number {
  const n = ability.x_slots ?? 1;
  return n >= 1 ? n : 1;
}

// suggestedAbilityX is the picker's opening guess: the caller's
// estimate of spendable mana, divided by the number of {X} slots,
// and never below the printed floor.
//
// It is a suggestion, not a limit. The modal's live auto-tap preview
// is the real check, and the server's cost gate is the real answer —
// the guess exists so the common case is one keystroke of confirming
// rather than several of typing.
export function suggestedAbilityX(ability: ActivatedAbilityView, availableMana: number): number {
  const floor = abilityMinX(ability);
  const affordable = Math.floor(Math.max(0, availableMana) / abilityXSlots(ability));
  return Math.max(floor, affordable);
}
