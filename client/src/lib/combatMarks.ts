// combatMarks.ts — the chip a creature wears when an effect has already
// decided it deals nothing in combat this turn (ADR 0187 §4; CR 510.1a):
// Carrion Rats' and Carrion Wurm's "assigns no combat damage this turn".

export interface CombatMarkInput {
  assigns_no_combat_damage?: boolean;
}

export interface CombatMarkBadge {
  /** The short badge text. */
  text: string;
  /** The tooltip: the rule in full. */
  title: string;
}

/** The badge for a creature that assigns no combat damage this turn, or null. */
export function noCombatDamageBadge(card: CombatMarkInput): CombatMarkBadge | null {
  if (card.assigns_no_combat_damage !== true) return null;
  return { text: "NO COMBAT DMG", title: "Assigns no combat damage this turn" };
}
