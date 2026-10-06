// abilityChips.ts — #2219: the art tile's ability chips.
//
// The server lists a card's non-keyword triggered, static and activated
// abilities in `ability_rows`, each with a label it wrote (game.AbilityRowsOf).
// The tile draws one chip per kind present, in a fixed order, with the
// number of rows of that kind; hovering or focusing the chip lists the
// labels. Nothing here reads oracle text or decides what counts: an
// absent or empty list is no chips.

import type { AbilityRowKind, AbilityRowView } from "./protocol";

export interface AbilityChip {
  kind: AbilityRowKind;
  count: number;
  labels: string[];
}

/** The chips' order on the tile: ⚡ triggered, ◆ static, ↻ activated. */
export const ABILITY_CHIP_ORDER: readonly AbilityRowKind[] = ["triggered", "static", "activated"];

/** The list card's heading for each kind. */
export const ABILITY_CHIP_TITLE: Record<AbilityRowKind, string> = {
  triggered: "Triggered abilities",
  static: "Static abilities",
  activated: "Activated abilities",
};

/**
 * abilityChips groups a card's rows by kind, in chip order, dropping a
 * kind with no rows and any row whose kind this client does not know.
 */
export function abilityChips(rows: readonly AbilityRowView[] | null | undefined): AbilityChip[] {
  if (!rows || rows.length === 0) return [];
  const out: AbilityChip[] = [];
  for (const kind of ABILITY_CHIP_ORDER) {
    const labels = rows.filter((r) => r.kind === kind).map((r) => r.label);
    if (labels.length > 0) out.push({ kind, count: labels.length, labels });
  }
  return out;
}

/** abilityChipDescription is what a screen reader hears after the chip's name. */
export function abilityChipDescription(chip: AbilityChip): string {
  return chip.labels.join("; ");
}
