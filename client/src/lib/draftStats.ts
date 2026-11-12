// draftStats.ts — what a pool or a deck is made of (#5702): creatures
// against everything else, a count for each card type, the mana curve
// and the colours. Pure functions over the viewer's own cards.
//
// A card's types are read from the front face of its type line, before
// the dash: "Artifact Creature — Golem" is an artifact and a creature,
// and so counts under each type once, and once under creatures in the
// creature/non-creature split. Subtypes are not read ("Creature — Elf
// Druid" has no type "Druid"). Lands are counted as a type but kept out
// of the curve and of the colours, which are about what is cast.

import { isLand, type DraftCard } from "./draft";
import { curve, type CurveColumn } from "./draftBuild";

export const CARD_TYPES = [
  "creature",
  "instant",
  "sorcery",
  "artifact",
  "enchantment",
  "planeswalker",
  "land",
] as const;
export type CardType = (typeof CARD_TYPES)[number];

/** The words of the type line's front face that come before its dash. */
function frontTypes(card: Pick<DraftCard, "type_line">): string {
  const front = (card.type_line ?? "").split("//")[0];
  return front.split("—")[0].split(/\s-\s/)[0];
}

/** Every type of the card, in CARD_TYPES order, each once. */
export function cardTypes(card: Pick<DraftCard, "type_line" | "basic">): CardType[] {
  const words = new Set(
    frontTypes(card)
      .split(/\s+/)
      .map((w) => w.toLowerCase()),
  );
  return CARD_TYPES.filter((t) => (t === "land" ? isLand(card) : words.has(t)));
}

export const STAT_COLOURS = ["W", "U", "B", "R", "G"] as const;
export type StatColour = (typeof STAT_COLOURS)[number];

export interface ColourTally {
  /** Cards of each colour; a card of two colours is in both counts. */
  byColour: Record<StatColour, number>;
  /** Cards of two or more colours. */
  multi: number;
  /** Cards with no colour that are not lands. */
  colourless: number;
}

export interface PoolStats {
  total: number;
  creatures: number;
  nonCreatures: number;
  /** Cards of each type; a card with two types is in two counts. */
  types: Record<CardType, number>;
  /** Mana value 0 to 6+, lands left out. */
  curve: CurveColumn[];
  colours: ColourTally;
}

/** The tallies of a list of cards. */
export function tally(cards: readonly DraftCard[]): PoolStats {
  const types = Object.fromEntries(CARD_TYPES.map((t) => [t, 0])) as Record<CardType, number>;
  const byColour = Object.fromEntries(STAT_COLOURS.map((c) => [c, 0])) as Record<
    StatColour,
    number
  >;
  let creatures = 0;
  let multi = 0;
  let colourless = 0;
  for (const card of cards) {
    const own = cardTypes(card);
    for (const t of own) types[t]++;
    if (own.includes("creature")) creatures++;
    if (own.includes("land")) continue;
    const colours = STAT_COLOURS.filter((c) => (card.colors ?? []).includes(c));
    if (colours.length === 0) {
      colourless++;
      continue;
    }
    for (const c of colours) byColour[c]++;
    if (colours.length > 1) multi++;
  }
  return {
    total: cards.length,
    creatures,
    nonCreatures: cards.length - creatures,
    types,
    curve: curve(cards),
    colours: { byColour, multi, colourless },
  };
}

/** The curve as a sentence for a screen reader: "0 at 0, 1 at 1, ... 2 at 6 or more". */
export function curveSummary(columns: readonly CurveColumn[]): string {
  return columns
    .map((c, i) => `${c.count} at ${i === columns.length - 1 ? `${c.value} or more` : c.value}`)
    .join(", ");
}
