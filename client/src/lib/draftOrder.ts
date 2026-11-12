// draftOrder.ts — the order a pack is shown in (#5701).
//
// Mythic rares, rares, uncommons, commons, then basic lands. Inside one
// rarity: white, blue, black, red, green, then cards of several colours,
// then colourless ones; inside one colour, by name. The key is the
// card's own printed attributes and its draft id, never its place in the
// pack, so a card keeps its place as the others leave (pick to pick) and
// the same pack always comes out in the same order.
//
// The same order serves pick-one, pick-two and cube drafts: a cube has
// no rarity slots, so each card sorts by the rarity it is printed with.

import { type DraftCard } from "./draft";

/** Mythic first. A rarity not named here (special, bonus) sits after common. */
const RARITY_RANK: Record<string, number> = {
  mythic: 0,
  rare: 1,
  uncommon: 2,
  common: 3,
};
const OTHER_RARITY = 4;
const BASIC_LAND = 5;

const COLOUR_RANK: Record<string, number> = { W: 0, U: 1, B: 2, R: 3, G: 4 };
const MULTICOLOUR = 5;
const COLOURLESS = 6;

/** A basic land: the flag, or a front face that says "Basic ... Land". */
export function isBasicLand(card: Pick<DraftCard, "type_line" | "basic">): boolean {
  if (card.basic) return true;
  const front = (card.type_line ?? "").split("//")[0];
  return /\bBasic\b/.test(front) && /\bLand\b/.test(front);
}

/** Where a card's rarity puts it: 0 mythic, 1 rare, 2 uncommon, 3 common, 4 other, 5 basic land. */
export function rarityRank(card: Pick<DraftCard, "rarity" | "type_line" | "basic">): number {
  if (isBasicLand(card)) return BASIC_LAND;
  return RARITY_RANK[(card.rarity ?? "").toLowerCase()] ?? OTHER_RARITY;
}

/** Where a card's colours put it: 0 to 4 for W U B R G, 5 for several colours, 6 for none. */
export function colourRank(card: Pick<DraftCard, "colors">): number {
  const colours = [...new Set((card.colors ?? []).filter((c) => c in COLOUR_RANK))];
  if (colours.length === 0) return COLOURLESS;
  if (colours.length > 1) return MULTICOLOUR;
  return COLOUR_RANK[colours[0]];
}

const byName = (a: DraftCard, b: DraftCard): number =>
  a.name.localeCompare(b.name, "en") || (a.id < b.id ? -1 : a.id > b.id ? 1 : 0);

/** The pack in display order. Returns a new array; the input is left as the server sent it. */
export function sortPack(pack: readonly DraftCard[]): DraftCard[] {
  return [...pack].sort(
    (a, b) => rarityRank(a) - rarityRank(b) || colourRank(a) - colourRank(b) || byName(a, b),
  );
}
