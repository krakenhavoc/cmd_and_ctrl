// banding.ts — attacking in a band (CR 702.22, ADR 0224, #5723).
//
// The server owns the rule: a band is the entries of one
// declare_attackers that share a `band` number, and the declaration is
// refused whole (illegal_attack, reason "band") when a band has two
// creatures without banding or splits across defenders. The client's
// part is the minimal affordance ADR 0224 §5 names: the "Choose
// attackers" picker's per-row Band toggle, which bands the toggled rows
// at that defender, and this module's check, which says why a band the
// player has toggled would be refused before they send it. All pure.

import type { CardView } from "./protocol";

/** The engine's ability token for banding (game.KeywordBanding). */
export const BANDING = "banding";

/** hasBanding reports whether a card has banding right now. */
export function hasBanding(card: Pick<CardView, "abilities"> | null | undefined): boolean {
  return !!card?.abilities?.some((a) => a.toLowerCase() === BANDING);
}

/**
 * bandProblem says why the creatures toggled into one band would be
 * refused (CR 702.22c), or null when the band is legal or is not a band
 * at all (fewer than two members). The server checks it again.
 */
export function bandProblem(
  members: readonly Pick<CardView, "abilities" | "name">[],
): string | null {
  if (members.length < 2) return null;
  const without = members.filter((c) => !hasBanding(c));
  if (without.length <= 1) return null;
  const names = without.map((c) => c.name || "an unknown creature");
  return `A band can have only one creature without banding, and ${names[0]} and ${names[1]} both lack it.`;
}

/**
 * withBand puts the declaration's entries named in `ids` in band 1: the
 * picker forms one band at its defender. Fewer than two named entries is
 * no band, and the entries are returned as they were.
 */
export function withBand<T extends { attacker: string; band?: number }>(
  entries: readonly T[],
  ids: readonly string[],
): T[] {
  const members = entries.filter((e) => ids.includes(e.attacker));
  if (members.length < 2) return entries.map((e) => ({ ...e }));
  return entries.map((e) => (ids.includes(e.attacker) ? { ...e, band: 1 } : { ...e }));
}

/**
 * bandTitle is the board's tooltip for an attacker in a band: which
 * band, and what that means for the defender.
 */
export function bandTitle(band: number): string {
  return `attacking in band ${band} — blocking one creature in a band blocks all of them`;
}
