// exileCost.ts — #1297: the pieces of the "Exile N cards from your …"
// cost picker that do not depend on Svelte.
//
// #1283 built the picker for one owner and one pile: Cadaverous Bloom's
// mana ability, "Exile a card from your hand". #1297 gave the component
// a second owner (a CR 602 activated ability) and a second pile (the
// graveyard — Grim Lavamancer's "Exile two cards from your graveyard"),
// so the ids the server offers have to be resolved out of whichever
// zone `exile_cost_zone` names, and the modal has to say which one.
// Both owners call these, so the two cannot drift apart.

import type { CardView, ExileCostZone, PlayerView, ZoneView } from "./protocol";
import { manaValueOf } from "./targetPrices";

interface ExileCostFields {
  exile_cost_options?: string[];
  exile_cost_zone?: ExileCostZone;
}

// The pile a clause reads. An older server that sent no zone only ever
// meant the hand (#1283 had no other).
export function exileCostZone(ability: ExileCostFields): ExileCostZone {
  return ability.exile_cost_zone ?? "hand";
}

// The cards the server offered for this clause, resolved out of the
// seat's own hand or graveyard, in that zone's order. The server has
// already filtered them; nothing narrows them further here. #5457: a
// processing clause ("put a card an opponent owns from exile into that
// player's graveyard") reads the shared exile zone instead. #5655:
// "exile a spell you control" reads the shared stack.
export function exileCostOptionCards(
  seat: PlayerView | undefined,
  ability: ExileCostFields,
  exile?: ZoneView,
  stack?: ZoneView,
): CardView[] {
  const ids = new Set(ability.exile_cost_options ?? []);
  const zone = exileCostZone(ability);
  if (zone === "exile") return (exile?.cards ?? []).filter((c) => ids.has(c.instance_id));
  if (zone === "stack") return (stack?.cards ?? []).filter((c) => ids.has(c.instance_id));
  if (!seat) return [];
  const pile = zone === "graveyard" ? seat.graveyard : seat.hand;
  return (pile?.cards ?? []).filter((c) => ids.has(c.instance_id));
}

// The modal's small print: where the cards come from, and that this is
// not a discard (nothing that watches discards sees it).
export function exileCostNote(ability: ExileCostFields): string {
  switch (exileCostZone(ability)) {
    case "exile":
      return "an opponent's card in exile · it goes to its owner's graveyard";
    case "stack":
      return "a spell you control on the stack · it is exiled without resolving";
    case "graveyard":
      return "exiled from your graveyard · not a discard";
    default:
      return "exiled from your hand · not a discard";
  }
}

// Where the shortfall message says the cards have to be.
export function exileCostWhere(ability: ExileCostFields): string {
  switch (exileCostZone(ability)) {
    case "exile":
      return "in exile";
    case "stack":
      return "on the stack";
    case "graveyard":
      return "in your graveyard";
    default:
      return "in hand";
  }
}

// The picker's button: every pile but exile is exiled from; a card
// already in exile is put into its owner's graveyard (#5457).
export function exileCostVerb(ability: ExileCostFields): string {
  return exileCostZone(ability) === "exile" ? "Put into graveyard" : "Exile";
}

// The mana symbols of one colour in a printed mana cost: {B}, a hybrid
// {B/G}, a Phyrexian {B/P} and a {2/B} each count once, which is how the
// engine counts them (game.Card.ColorSymbols, CR 700.5).
export function colorSymbolCount(manaCost: string | undefined, color: string): number {
  let n = 0;
  for (const m of (manaCost ?? "").matchAll(/\{([^}]*)\}/g)) {
    if (m[1].toUpperCase().split("/").includes(color.toUpperCase())) n++;
  }
  return n;
}

// "Exile any number of … with N or more <colour> mana symbols" (Baron
// Helmut Zemo's boast): the running total of the picked cards.
export function exileCostSymbolTotal(cards: CardView[], ids: string[], color: string): number {
  const picked = new Set(ids);
  return cards
    .filter((c) => picked.has(c.instance_id))
    .reduce((sum, c) => sum + colorSymbolCount(c.mana_cost, color), 0);
}

// Collect evidence N (CR 701.59a, #5193): the running total mana value of
// the picked cards, each read off its printed cost as the engine reads it
// in the graveyard.
export function exileCostManaValueTotal(cards: CardView[], ids: string[]): number {
  const picked = new Set(ids);
  return cards
    .filter((c) => picked.has(c.instance_id))
    .reduce((sum, c) => sum + manaValueOf(c.mana_cost ?? ""), 0);
}
