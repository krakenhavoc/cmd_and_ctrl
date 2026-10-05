// damageBadge.ts — the damage badge a creature wears, and what it says
// when that damage can't kill it (#2257).
//
// A player blocked a 5/4 Solphim with a deathtouch creature, saw "4" on
// the damage badge and watched it live: it had an indestructible
// counter. The counter pip and the keyword badge were both on the card,
// but nothing tied them to the damage, so the board read as a rules bug.
// When the creature is indestructible the badge now carries the
// indestructible shield and a tooltip that says why the damage doesn't
// destroy it (CR 702.12b: neither lethal damage nor deathtouch does).
//
// Everything here is read off the snapshot. `toughness` on the wire is
// the current toughness (counters included) that the server's
// lethal-damage check uses, and `abilities` already holds a keyword
// granted by a counter, an anthem or an effect, not only a printed one.

export interface DamageBadgeInput {
  damage_marked?: number;
  toughness?: number;
  abilities?: string[];
  counters?: Record<string, number>;
}

export interface DamageBadge {
  /** The badge text: the damage marked. */
  text: string;
  /** The badge tooltip. */
  title: string;
  /** One short line for the card-detail panel. */
  summary: string;
  /** The damage is at least the creature's toughness. */
  lethal: boolean;
  /** The creature is indestructible, so the damage doesn't destroy it. */
  survives: boolean;
}

/**
 * The damage badge for a permanent, or null when it has no damage
 * marked.
 */
export function damageBadge(card: DamageBadgeInput): DamageBadge | null {
  const n = card.damage_marked ?? 0;
  if (n <= 0) return null;
  const toughness = card.toughness ?? 0;
  const lethal = toughness > 0 && n >= toughness;
  const survives = (card.abilities ?? []).includes("indestructible");
  const marked = `${n} damage marked`;
  if (!survives) {
    return { text: String(n), title: marked, summary: `${n} damage`, lethal, survives };
  }
  const from = (card.counters?.indestructible ?? 0) > 0 ? " (from an indestructible counter)" : "";
  const title = lethal
    ? `${marked}, enough to kill a creature with toughness ${toughness}, but it has indestructible${from}. ` +
      "Lethal damage and deathtouch don't destroy it. The damage wears off at end of turn."
    : `${marked}. It has indestructible${from}, so lethal damage and deathtouch don't destroy it.`;
  const summary = lethal
    ? `${n} damage: lethal, but indestructible`
    : `${n} damage: indestructible`;
  return { text: String(n), title, summary, lethal, survives };
}
