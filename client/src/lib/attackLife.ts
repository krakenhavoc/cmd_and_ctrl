// attackLife.ts — paying a Phyrexian attack tax with life (#4198).
//
// Norn's Annex: "Creatures can't attack you or planeswalkers you control
// unless their controller pays {W/P} for each of those creatures." Each
// {W/P} is paid with {W} or 2 life (CR 107.4f), and which is part of the
// cost to attack (CR 508.1h), so it is announced with the declaration as
// `phyrexian_life` on declare_attacker / declare_attackers — the field a
// cast and an activation already send, through the same server helper.
//
// The sums here are the cast chain's (phyrexianLife.ts): the ceiling is
// the symbols the tax prints, capped by CR 119.4. Only the count of
// symbols differs. It is read off the view (`attack_targets[].
// tax_phyrexian`, per attacking creature) and multiplied by the attackers
// aimed at each target; no client code parses a mana string.
//
// A printed Phyrexian symbol always asks, as a cast does: mana-only
// attackers get a question they answer with "0". A symbol payable with
// life only through a grant (K'rrik) does not ask up front; the picker
// still offers it, and a refused declaration opens the picker.

import { clampPhyrexianLife, maxPhyrexianLife } from "./phyrexianLife";
import type { GameView } from "./protocol";

/** The Phyrexian symbols of a declaration's tax, and how many are granted. */
export interface AttackLifeSymbols {
  total: number;
  granted: number;
}

/**
 * attackLifeSymbols counts the Phyrexian symbols in the tax of a
 * declaration whose attackers are aimed at `targets` (one entry per
 * attacker; a target repeats once per creature sent at it). A target the
 * view lists no symbols for adds none.
 */
export function attackLifeSymbols(
  view: GameView | null | undefined,
  targets: readonly string[],
): AttackLifeSymbols {
  const rows = view?.turn?.attack_targets ?? [];
  let total = 0;
  let granted = 0;
  for (const id of targets) {
    const row = rows.find((t) => t.id === id);
    total += row?.tax_phyrexian ?? 0;
    granted += row?.tax_phyrexian_granted ?? 0;
  }
  return { total, granted: Math.min(granted, total) };
}

/** attackLifeCeiling is the most symbols the viewer may pay with life. */
export function attackLifeCeiling(symbols: number, life: number | undefined): number {
  return maxPhyrexianLife(symbols, life);
}

/**
 * attackLifeAsks says whether sending this declaration should first ask
 * mana or life: the tax prints a Phyrexian symbol and the viewer's life
 * total could pay for at least one. Symbols that are payable with life
 * only through a grant do not ask.
 */
export function attackLifeAsks(
  view: GameView | null | undefined,
  targets: readonly string[],
  life: number | undefined,
): boolean {
  const { total, granted } = attackLifeSymbols(view, targets);
  return total - granted > 0 && attackLifeCeiling(total, life) > 0;
}

/**
 * clampAttackLife holds a claim inside what the declaration can take, so
 * a stale count (an attacker unchecked, a life total that fell) is
 * never sent.
 */
export function clampAttackLife(n: number, symbols: number, life: number | undefined): number {
  return clampPhyrexianLife(n, attackLifeCeiling(symbols, life));
}

/**
 * withAttackLife adds `phyrexian_life` to a declaration's params when
 * the claim is above zero. Omitted at zero, the way every other
 * caller of the field leaves it, so a declaration at a table with no
 * Phyrexian tax is byte for byte what it was.
 */
export function withAttackLife<T extends object>(
  params: T,
  n: number,
): T & { phyrexian_life?: number } {
  return n > 0 ? { ...params, phyrexian_life: n } : params;
}

/** What AttackLifeModal asks about. */
export interface AttackLifePrompt {
  // Phyrexian symbols in the whole declaration's tax.
  symbols: number;
  // The tax one creature pays ("{W/P}"), and how many attack.
  each: string;
  attackers: number;
  // The confirm button's verb ("Attack with 3").
  verb: string;
}

/**
 * attackLifePrompt builds the question for a declaration whose attackers
 * are aimed at `targets`, one entry each. `each` is the first taxed
 * target's per-creature price; a table with two different taxed targets
 * shows the first, and the symbol count still covers all of them.
 */
export function attackLifePrompt(
  view: GameView | null | undefined,
  targets: readonly string[],
  verb: string,
): AttackLifePrompt {
  const rows = view?.turn?.attack_targets ?? [];
  const first = rows.find((t) => targets.includes(t.id) && (t.tax_phyrexian ?? 0) > 0);
  return {
    symbols: attackLifeSymbols(view, targets).total,
    each: first?.tax ?? "",
    attackers: targets.length,
    verb,
  };
}

/** attackLifeSentence is the stepper's one-line summary. */
export function attackLifeSentence(n: number, symbols: number): string {
  if (n <= 0) return symbols === 1 ? "paid with mana" : "all paid with mana";
  if (n >= symbols) return symbols === 1 ? "paid with life" : "all paid with life";
  return `${n} with life, ${symbols - n} with mana`;
}
