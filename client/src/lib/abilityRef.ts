// abilityRef.ts — ADR 0093 Decision 5 on the client.
//
// Every ability row the server publishes carries a stable `ref`
// ("own:0", "land:G", "grant:cryptolith-rite/any-color:0:0"). The
// activate verbs take it back beside `ability_index`, and the server
// refuses a ref that no longer names the row at that index — a grant
// that appeared or vanished between this snapshot and the click moves
// rows under their indexes, and firing whatever now sits at the index
// would be the #544 bug in the player's hands.
//
// The ref is read off the SAME card snapshot the click was made on,
// which is the point: it is the client's statement of which row it
// meant. Absent (an older server, or a row the card no longer lists)
// sends nothing, which the server accepts exactly as before.

import type { ActivatedAbilityView, CardView, ManaAbilityView } from "./protocol";

function refIn(
  rows: readonly { index: number; ref?: string }[] | undefined,
  index: number,
): {
  ref?: string;
} {
  const ref = rows?.find((r) => r.index === index)?.ref;
  return ref ? { ref } : {};
}

/** The `ref` fragment of an activate_ability payload for `index`. */
export function activatedAbilityRef(card: CardView, index: number): { ref?: string } {
  return refIn(card.activated_abilities ?? card.zone_abilities, index);
}

/** The `ref` fragment of an activate_mana_ability payload for `index`. */
export function manaAbilityRef(card: CardView, index: number): { ref?: string } {
  return refIn(card.mana_abilities ?? card.zone_mana_abilities, index);
}

/** Whether any of a permanent's ability rows was granted by another permanent. */
export function hasGrantedAbility(card: CardView): boolean {
  const granted = (rows?: readonly (ActivatedAbilityView | ManaAbilityView)[]) =>
    (rows ?? []).some((r) => !!r.granted_by);
  return granted(card.mana_abilities) || granted(card.activated_abilities);
}

/** Whether a permanent has a GRANTED activated (non-mana) ability row. */
export function hasGrantedActivatedAbility(card: CardView): boolean {
  return (card.activated_abilities ?? []).some((r) => !!r.granted_by);
}

/** "from Cryptolith Rite", or "" for the permanent's own row. */
export function grantedFromLabel(row: { granted_by?: { name?: string } }): string {
  if (!row.granted_by) return "";
  return row.granted_by.name ? `from ${row.granted_by.name}` : "granted";
}

/**
 * The stale-ref refusal, recognised by the server's message (ADR 0093:
 * "that ability is no longer at that position — the board changed").
 */
export function isStaleAbilityRefError(message: string | undefined): boolean {
  return !!message && message.includes("no longer at that position");
}
