// manaAbilityCost.ts — the one predicate both mana-ability click paths
// ask: does this activation need an answer before it can be sent?
//
// Two copies used to exist, one in Board's context-menu path and one in
// PlayerPanel's click path, and they had drifted: the panel's asked
// about the sacrifice picker and the counter cost but not about the
// DISCARD component #1213 added, so clicking Skirge Familiar on the
// board sent `activate_mana_ability` with no `discard_ids` and the
// server refused it. #1283 added a fourth component (Cadaverous
// Bloom's "Exile a card from your hand"), which is the moment to make
// it one function.

import { counterCostNeedsPrompt } from "./counterCost";
import type { ManaAbilityView } from "./protocol";

export function manaAbilityNeedsPrompt(ability: ManaAbilityView): boolean {
  return (
    !!ability.sacrifice_options ||
    !!ability.tap_others_options ||
    !!ability.discard_cost_n ||
    !!ability.exile_cost_n ||
    // Collect evidence N (Cryptex) asks which graveyard cards.
    !!ability.exile_cost_any_number ||
    // #1600: Food Chain's "Exile a creature you control".
    !!ability.exile_permanent_options ||
    // #5397: an either/or cost (forage) asks which way to pay.
    (ability.cost_branches?.length ?? 0) > 0 ||
    counterCostNeedsPrompt(ability)
  );
}

// The optional payment fragment carried by activate_mana_ability.
// Kept beside the prompt predicate so both click paths use and test
// the exact wire name (#758).
export function manaTapPayment(tapIDs: string[]): { tap_ids?: string[] } {
  return tapIDs.length > 0 ? { tap_ids: [...tapIDs] } : {};
}

// #1600: the permanents Food Chain's "Exile a creature you control"
// names, on their own wire field — `exile_ids` names cards in a hand or
// a graveyard. Absent when there are none, so every other mana ability's
// payload is unchanged.
export function manaExilePermanentPayment(ids: string[]): { exile_permanent_ids?: string[] } {
  return ids.length > 0 ? { exile_permanent_ids: [...ids] } : {};
}

// #5397: the branch of an either/or cost (forage) the activation pays,
// on activate_mana_ability's `cost_branch`. Absent when the row has no
// branches, so every other mana ability's payload is unchanged.
export function manaCostBranchPayment(k: number | undefined): { cost_branch?: number } {
  return k === undefined ? {} : { cost_branch: k };
}
