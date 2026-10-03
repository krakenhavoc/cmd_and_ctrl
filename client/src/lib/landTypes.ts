// landTypes.ts — the chip a land wears while a resolved effect has
// changed its land types (ADR 0109 §1): Tidal Warrior's "becomes an
// Island until end of turn", Navigator's Compass's "in addition to its
// other types", and (§2) Ultima's blight, which takes them all away.
// The type line already shows what the land is; the chip says why, and
// for how long.

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

/**
 * "Island until end of turn — Tidal Warrior", or for a land that lost
 * them all, 'No land types, no abilities, "{T}: Add {C}." for as long as
 * it has a blight counter on it — Ultima, Origin of Oblivion'.
 */
export function landTypeEffectLine(e: LandTypeEffect): string {
  let line = e.loses_all ? lostLine(e) : joinTypes(e.types ?? []);
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
  const effects = (card.land_type_effects ?? []).filter(
    (e) => e.loses_all || (e.types ?? []).length > 0,
  );
  if (effects.length === 0) return null;
  const newest = effects[effects.length - 1];
  const title = effects.map(landTypeEffectLine).join("\n");
  if (newest.loses_all) return { text: "NO LAND TYPES", title };
  const types = newest.types.map((t) => t.toUpperCase()).join(" ");
  return { text: newest.in_addition ? `+${types}` : types, title };
}

/**
 * 'No land types, no abilities, "{T}: Add {C}."': what ADR 0109 §2's
 * "loses all land types" took, and what the same effect gave.
 */
function lostLine(e: LandTypeEffect): string {
  const parts = ["No land types"];
  if (e.loses_abilities) parts.push("no abilities");
  for (const g of e.gains ?? []) parts.push(`"${g}"`);
  return parts.join(", ");
}

/** "Island", "Mountain and Forest", "Mountain, Forest and Plains". */
function joinTypes(types: string[]): string {
  if (types.length <= 1) return types[0] ?? "";
  return `${types.slice(0, -1).join(", ")} and ${types[types.length - 1]}`;
}
