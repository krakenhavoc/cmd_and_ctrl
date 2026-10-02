// landTypes.ts — the chip a land wears while a resolved effect has
// changed its land types (ADR 0109 §1): Tidal Warrior's "becomes an
// Island until end of turn", Navigator's Compass's "in addition to its
// other types". The type line already shows what the land is; the chip
// says why, and for how long.

import type { LandTypeEffect } from "./protocol";

export interface LandTypeBadgeInput {
  land_type_effects?: LandTypeEffect[];
}

export interface LandTypeBadge {
  /** The short badge text: the newest effect's types, "+" for an add. */
  text: string;
  /** The tooltip: one line per effect, oldest first. */
  title: string;
}

/** "Island until end of turn — Tidal Warrior". */
export function landTypeEffectLine(e: LandTypeEffect): string {
  let line = joinTypes(e.types ?? []);
  if (e.in_addition) line += " in addition to its other types";
  if (e.until) line += ` ${e.until}`;
  if (e.source) line += ` — ${e.source}`;
  return line;
}

/**
 * The badge for a permanent's land-type effects, or null when it has
 * none. The newest effect is the one the text shows: a set replaces
 * whatever came before it in timestamp order (CR 613.7).
 */
export function landTypeBadge(card: LandTypeBadgeInput): LandTypeBadge | null {
  const effects = (card.land_type_effects ?? []).filter((e) => (e.types ?? []).length > 0);
  if (effects.length === 0) return null;
  const newest = effects[effects.length - 1];
  const types = newest.types.map((t) => t.toUpperCase()).join(" ");
  return {
    text: newest.in_addition ? `+${types}` : types,
    title: effects.map(landTypeEffectLine).join("\n"),
  };
}

/** "Island", "Mountain and Forest", "Mountain, Forest and Plains". */
function joinTypes(types: string[]): string {
  if (types.length <= 1) return types[0] ?? "";
  return `${types.slice(0, -1).join(", ")} and ${types[types.length - 1]}`;
}
