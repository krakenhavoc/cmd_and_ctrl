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
    // #6125: an {X} in the row's cost asks for X.
    !!ability.demands_x ||
    // #6085: Rosheen's "Reveal any number of …" asks which cards to show,
    // when the hand holds any.
    (ability.reveal_cost_options?.length ?? 0) > 0 ||
    // #1600: Food Chain's "Exile a creature you control".
    !!ability.exile_permanent_options ||
    // #5397: an either/or cost (forage) asks which way to pay.
    (ability.cost_branches?.length ?? 0) > 0 ||
    // #6101: "Choose a player" (Spectral Searchlight) asks which player.
    !!ability.chosen_player ||
    counterCostNeedsPrompt(ability)
  );
}

// #5447: whether a mana ability's tap-others cost needs the picker, or
// has exactly one payment the client can send on its own. A fixed count
// (`min` equal to `max`, Springleaf Drum) asks only when there are more
// options than the count. "Tap X untapped tokens you control: Add X
// mana" (Hazel of the Rootbloom) ships `min` 0 with `max` every option:
// how many to tap is the player's choice, so it always asks when there
// is anything to tap.
export function manaTapNeedsPicker(
  opts: { cards?: string[]; min?: number; max?: number } | undefined,
): boolean {
  if (!opts) return false;
  const options = opts.cards ?? [];
  const need = opts.max ?? opts.min ?? 1;
  const min = opts.min ?? need;
  return options.length > need || (min < need && options.length > 0);
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

// #6101: the player a "Choose a player. That player adds …" row names,
// on activate_mana_ability's `chosen_player`. Absent when none was asked,
// so every other mana ability's payload is unchanged.
export function manaChosenPlayerPayment(id: string | undefined): { chosen_player?: string } {
  return id ? { chosen_player: id } : {};
}

// #6101: a colour named up front rides the activation only when the
// activator chose themselves: anyone else chooses their own colour.
export function manaColorsForChosenPlayer(
  colors: string[],
  chosen: string | undefined,
  activator: string | undefined,
): string[] {
  return chosen && chosen !== activator ? [] : colors;
}

// #6101: the players a "Choose a player" row offers: everyone still in
// the game, the activator first (as "You"), then the others in seat
// order.
export interface ChosenPlayerOption {
  id: string;
  name: string;
  you: boolean;
}

export function chosenPlayerOptions(
  seats: readonly { id: string; name: string; eliminated?: boolean }[],
  activator: string | undefined,
): ChosenPlayerOption[] {
  const live = seats.filter((s) => !s.eliminated);
  const me = live.filter((s) => s.id === activator);
  const others = live.filter((s) => s.id !== activator);
  return [
    ...me.map((s) => ({ id: s.id, name: s.name, you: true })),
    ...others.map((s) => ({ id: s.id, name: s.name, you: false })),
  ];
}

// #5397: the branch of an either/or cost (forage) the activation pays,
// on activate_mana_ability's `cost_branch`. Absent when the row has no
// branches, so every other mana ability's payload is unchanged.
export function manaCostBranchPayment(k: number | undefined): { cost_branch?: number } {
  return k === undefined ? {} : { cost_branch: k };
}

// manaXPayment is the `x_value` a mana activation sends (#6125): for a
// "Sacrifice X …" clause the number of permanents sacrificed, for an {X}
// in the cost the X the player named, and nothing for a row with no X.
export function manaXPayment(
  ability: ManaAbilityView,
  named: number | undefined,
  sacrificeIDs: string[] | undefined,
): { x_value?: number } {
  if (!ability.demands_x) return {};
  if (ability.sacrifice_options?.count_from_x) return { x_value: sacrificeIDs?.length ?? 0 };
  return named !== undefined && named > 0 ? { x_value: named } : {};
}
