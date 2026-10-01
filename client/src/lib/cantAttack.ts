// cantAttack.ts — ADR 0106 §2 (#1794, owner decision 3): a creature
// that "can't attack its owner" (Xantcha, Sleeper Agent) says so on the
// card, naming the owner: "CAN'T ATTACK Alice".
//
// READ, NOT DERIVED. The chip is drawn from the server's
// `attack_target_restrictions` on the card view, with the player it
// names already resolved; this module only turns a seat id into a name.
// Which defenders a selected attacker may click is the legal_actions
// digest's answer (legalActions.ts `attackTargets`), which already
// leaves the owner out and greys their ring. Nothing here decides an
// attack.

import type { CardView, PlayerView } from "./protocol";

/** One card's chip: the name it shows and the tooltip that explains it. */
export interface CantAttackChip {
  /** The player's name, or names, after "CAN'T ATTACK". */
  label: string;
  /** "can't attack Alice or planeswalkers Alice controls (Xantcha, Sleeper Agent)". */
  title: string;
}

/**
 * cantAttackChip builds one card's chip from its restrictions, or null
 * when it has none. Two restrictions naming the same player (a printed
 * one and a granted one) make one name, not two.
 */
export function cantAttackChip(
  card: Pick<CardView, "attack_target_restrictions">,
  seats: readonly PlayerView[] | undefined,
): CantAttackChip | null {
  const rows = card.attack_target_restrictions ?? [];
  if (rows.length === 0) return null;
  const names = new Map((seats ?? []).map((s) => [s.id, s.name]));
  const byPlayer = new Map<string, { walkers: boolean; sources: string[] }>();
  for (const r of rows) {
    if (!r.player) continue;
    const e = byPlayer.get(r.player) ?? { walkers: false, sources: [] };
    e.walkers = e.walkers || !!r.planeswalkers;
    if (r.source && !e.sources.includes(r.source)) e.sources.push(r.source);
    byPlayer.set(r.player, e);
  }
  if (byPlayer.size === 0) return null;
  const labels: string[] = [];
  const titles: string[] = [];
  for (const [id, e] of byPlayer) {
    const name = names.get(id) ?? "another player";
    labels.push(name);
    let t = `can't attack ${name}`;
    if (e.walkers) t += ` or planeswalkers ${name} controls`;
    if (e.sources.length > 0) t += ` (${e.sources.join(", ")})`;
    titles.push(t);
  }
  return { label: labels.join(", "), title: titles.join("; ") };
}

/** The chip for every battlefield card that has one, keyed by instance ID. */
export function cantAttackByCard(
  cards: readonly CardView[] | undefined,
  seats: readonly PlayerView[] | undefined,
): Record<string, CantAttackChip> {
  const out: Record<string, CantAttackChip> = {};
  for (const c of cards ?? []) {
    const chip = cantAttackChip(c, seats);
    if (chip) out[c.instance_id] = chip;
  }
  return out;
}
