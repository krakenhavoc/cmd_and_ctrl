// How the back row's zones share its width (#3028, #4087).
//
// The back row holds the land strip and, since #4087, one zone for each
// other permanent type: planeswalkers, battles, artifacts, enchantments
// (and "other", for a card whose type line names none). A zone with no
// cards is not drawn, so a typical board has two or three.
//
// The land strip takes whatever the others leave (it opens and overlaps
// to fit). Every other zone is capped at a share of the row by how many
// piles it draws, so a long artifact row shrinks its own cards instead
// of crushing the lands, and a seat with three Sorins keeps them
// readable beside twenty tokens.

import type { BattlefieldBucket } from "./cardTypes";

/** The back row's zones other than the land strip, left to right. */
export type BackZone = Exclude<BattlefieldBucket, "creature" | "land">;

export const BACK_ZONES: readonly BackZone[] = [
  "planeswalker",
  "battle",
  "artifact",
  "enchantment",
  "other",
];

/**
 * What a zone's piles count for when the width is shared out. A
 * planeswalker or battle is never overlapped and is read off its
 * counter, so each of its piles claims the room of two; the rest claim
 * one.
 */
export const ZONE_WEIGHT: Record<BackZone, number> = {
  planeswalker: 2,
  battle: 2,
  artifact: 1,
  enchantment: 1,
  other: 1,
};

/** The label each zone's row carries (its aria-label and data-zone). */
export const ZONE_LABEL: Record<BackZone, string> = {
  planeswalker: "planeswalkers",
  battle: "battles",
  artifact: "artifacts",
  enchantment: "enchantments",
  other: "other",
};

/** A zone whose cards are never overlapped or folded into a pile. */
export function isSpreadZone(z: BackZone): boolean {
  return z === "planeswalker" || z === "battle";
}

/**
 * backRowShares gives each zone its max-width as a share of the back
 * row, 0..1. A zone with no piles, or one with nothing beside it to
 * share with, may take the whole row (1). The land strip is the zone
 * that gives way, so with no land piles the others split the row
 * between them.
 */
export function backRowShares(
  landPiles: number,
  piles: Partial<Record<BackZone, number>>,
): Record<BackZone, number> {
  const weighted = {} as Record<BackZone, number>;
  let total = Math.max(0, landPiles);
  for (const z of BACK_ZONES) {
    weighted[z] = Math.max(0, piles[z] ?? 0) * ZONE_WEIGHT[z];
    total += weighted[z];
  }
  const out = {} as Record<BackZone, number>;
  for (const z of BACK_ZONES) {
    const mine = weighted[z];
    out[z] = mine <= 0 || total - mine <= 0 ? 1 : mine / total;
  }
  return out;
}
