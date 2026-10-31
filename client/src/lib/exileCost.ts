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

import type { CardView, ExileCostZone, PlayerView } from "./protocol";

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
// already filtered them; nothing narrows them further here.
export function exileCostOptionCards(
  seat: PlayerView | undefined,
  ability: ExileCostFields,
): CardView[] {
  if (!seat) return [];
  const ids = new Set(ability.exile_cost_options ?? []);
  const pile = exileCostZone(ability) === "graveyard" ? seat.graveyard : seat.hand;
  return (pile?.cards ?? []).filter((c) => ids.has(c.instance_id));
}

// The modal's small print: where the cards come from, and that this is
// not a discard (nothing that watches discards sees it).
export function exileCostNote(ability: ExileCostFields): string {
  return exileCostZone(ability) === "graveyard"
    ? "exiled from your graveyard · not a discard"
    : "exiled from your hand · not a discard";
}

// Where the shortfall message says the cards have to be.
export function exileCostWhere(ability: ExileCostFields): string {
  return exileCostZone(ability) === "graveyard" ? "in your graveyard" : "in hand";
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
