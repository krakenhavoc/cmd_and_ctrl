// sacrificeCost.ts — #747: the client half of a sacrifice cost of N
// permanents ("Sacrifice two artifacts", "Sacrifice five Treasures").
//
// There is no count field on the wire. A sacrifice clause ships as
// `sacrifice_options` (a LegalTargetsView) on an activated ability, a
// mana ability, or a hand card's additional cost, and its `min` /
// `max` ARE the count — always equal, 1 for "Sacrifice a creature".
// The server ships the options in payment order (tokens first, then
// lower mana value, then the ability's own source, then board order),
// which is the order the legal-move enumerator takes its one payment
// from; "Choose for me" takes the first N of that list, so the button
// picks exactly what a bot would. See docs/protocol.md, "Sacrifice
// costs of N permanents".
//
// Pure functions only, so the picker and the menus share one
// definition of "how many" and "can this be paid" and vitest can pin
// both without mounting a component.

import type { CardView } from "./protocol";

// SacrificeOptionsShape is the part of a LegalTargetsView the helpers
// read. The menus' local cost shapes declare only cards / players, so
// min / max are optional here.
export interface SacrificeOptionsShape {
  cards?: string[];
  min?: number;
  max?: number;
  // #1213: the clause's count is the announced X ('Sacrifice X
  // Treasures'), so min / max say nothing and the number picked IS
  // the x_value the same message announces. See sacrificeRange.
  count_from_x?: boolean;
}

// sacrificeCount is how many permanents the clause sacrifices: the
// view's max, and never less than one — a clause that reached the
// client with no count (an older server, a hand-built fixture) is the
// "Sacrifice a creature" every cost was before #747.
export function sacrificeCount(opts: SacrificeOptionsShape | undefined): number {
  const n = opts?.max ?? opts?.min ?? 1;
  return n >= 1 ? n : 1;
}

// sacrificeShortfall is the reason a sacrifice cost can't be paid
// right now, or "" when it can. The one-permanent wording is unchanged
// from before #747; a count names what is missing: "needs three Foods
// (you have 2)".
export function sacrificeShortfall(opts: SacrificeOptionsShape | undefined, label: string): string {
  if (!opts) return "";
  const have = opts.cards?.length ?? 0;
  const need = sacrificeCount(opts);
  if (have >= need) return "";
  if (need === 1) return `nothing to sacrifice (${label})`;
  return `needs ${label} (you have ${have})`;
}

// orderSacrificeOptions resolves the server's option IDs against the
// cards the client holds, KEEPING THE SERVER'S ORDER — the picker lists
// them in payment order, so the "Choose for me" picks are the top of
// the list. IDs with no matching card are dropped.
export function orderSacrificeOptions(board: CardView[], ids: string[] | undefined): CardView[] {
  if (!ids || ids.length === 0) return [];
  const byID = new Map(board.map((c) => [c.instance_id, c]));
  const out: CardView[] = [];
  for (const id of ids) {
    const c = byID.get(id);
    if (c) out.push(c);
  }
  return out;
}

// toggleSacrificePick is one click in the picker. At a count of 1 a
// click replaces the pick, as the single-choice picker always did. At
// a count of N a click adds an unpicked permanent (unless N are
// already picked) or removes a picked one.
export function toggleSacrificePick(chosen: string[], id: string, count: number): string[] {
  if (count <= 1) return chosen.length === 1 && chosen[0] === id ? chosen : [id];
  if (chosen.includes(id)) return chosen.filter((c) => c !== id);
  if (chosen.length >= count) return chosen;
  return [...chosen, id];
}

// chooseForMeState is whether the picker shows the "Choose for me"
// button, and whether it can be pressed: shown only for a count of two
// or more (at one, a single click is already the whole choice), and
// disabled when fewer options than the count are on offer.
export function chooseForMeState(
  count: number,
  optionCount: number,
): { shown: boolean; disabled: boolean } {
  const shown = count > 1;
  return { shown, disabled: shown && optionCount < count };
}

// chooseSacrificeForMe fills the selection with the first N options in
// the order given — the server's payment order. It only SELECTS; the
// player still confirms, and can change the picks first.
export function chooseSacrificeForMe(options: string[], count: number): string[] {
  return options.slice(0, Math.max(0, count));
}

// canConfirmSacrifice is the confirm gate: exactly N distinct picks.
export function canConfirmSacrifice(chosen: string[], count: number): boolean {
  return chosen.length === count && new Set(chosen).size === count;
}

// keepAvailablePicks drops picks that are no longer among the options:
// a chosen Treasure destroyed in response while the picker is open
// would otherwise stay in the selection unseen, unable to be unticked,
// holding a slot that disables every other row, and a confirm would
// send an ID the server refuses. Returns `chosen` itself when nothing
// was dropped, so a reactive caller can compare by reference.
export function keepAvailablePicks(chosen: string[], optionIDs: string[]): string[] {
  const available = new Set(optionIDs);
  const kept = chosen.filter((id) => available.has(id));
  return kept.length === chosen.length ? chosen : kept;
}

// --- #1213: a count the ACTIVATOR announces --------------------------
//
// Two printed clauses no longer ship min == max:
//
//   "Sacrifice one or more artifacts"  min 1, max 0 (no ceiling)
//   "Sacrifice X Treasures"            count_from_x, the count IS the X
//
// Everything above stays the fixed-count vocabulary, byte for byte,
// because every other cost site still uses it. What follows is the
// RANGE vocabulary, and a fixed clause passes through it unchanged:
// its bounds are N..N and every answer these give is the answer the
// functions above give.

// SacrificeRange is how many permanents one payment may name. `max`
// of 0 means "no printed ceiling" — the board is the only bound — and
// a caller that needs a concrete number substitutes the option count.
export interface SacrificeRange {
  min: number;
  max: number;
}

// sacrificeRange reads the bounds off the view. A clause whose count
// is the announced X (`count_from_x`) has no printed bounds at all, so
// it reads as "at least one, as many as you control": the client sends
// the number picked AS the x_value, which is the only reading that
// cannot disagree with the server.
//
// A view with no bounds at all is the one-permanent clause every cost
// was before #747, which is what sacrificeCount already assumes.
export function sacrificeRange(opts: SacrificeOptionsShape | undefined): SacrificeRange {
  if (opts?.count_from_x) return { min: 1, max: 0 };
  const min = opts?.min ?? 0;
  const max = opts?.max ?? 0;
  if (min <= 0 && max <= 0) return { min: 1, max: 1 };
  if (min <= 0) return { min: max, max };
  return { min, max };
}

// sacrificeCeiling is the range's `max` resolved against what is
// actually on offer: an unbounded clause is capped by the board.
export function sacrificeCeiling(range: SacrificeRange, optionCount: number): number {
  return range.max > 0 ? range.max : optionCount;
}

// canConfirmSacrificeRange is the confirm gate for a range: at least
// `min` distinct picks, and no more than the ceiling.
export function canConfirmSacrificeRange(
  chosen: string[],
  range: SacrificeRange,
  optionCount: number,
): boolean {
  if (new Set(chosen).size !== chosen.length) return false;
  return chosen.length >= range.min && chosen.length <= sacrificeCeiling(range, optionCount);
}

// toggleSacrificePickInRange is one click in a ranged picker. A
// single-pick clause (ceiling 1) replaces the pick, exactly as
// toggleSacrificePick does; anything wider adds until the ceiling and
// removes on a second click.
export function toggleSacrificePickInRange(
  chosen: string[],
  id: string,
  ceiling: number,
): string[] {
  return toggleSacrificePick(chosen, id, ceiling);
}

// sacrificeRangeShortfall is the reason a ranged cost cannot be paid
// right now, or "" when it can — the greyed-row text, one clause over
// from sacrificeShortfall. It reads the FLOOR, because a clause with
// no ceiling is payable the moment one permanent can pay it.
export function sacrificeRangeShortfall(
  opts: SacrificeOptionsShape | undefined,
  label: string,
): string {
  if (!opts) return "";
  const have = opts.cards?.length ?? 0;
  const need = sacrificeRange(opts).min;
  if (have >= need) return "";
  if (need === 1) return `nothing to sacrifice (${label})`;
  return `needs ${label} (you have ${have})`;
}
