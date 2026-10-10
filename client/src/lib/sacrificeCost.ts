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

import type { CardView, SacrificeGroupView, TargetSharesView } from "./protocol";

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
  // #2526: the clause's set rule ("Sacrifice a Swamp and a Forest"):
  // the picks must fill every group one-to-one. See fillsEachOf.
  each_of?: SacrificeGroupView[];
  // ADR 0137's amendment: a craft clause's "two that share a card type".
  // See fitsShares.
  shares?: TargetSharesView;
  // #2097: the clause takes every permanent in `cards` ("sacrifice all
  // creatures you control"). See sacrificesAll.
  all?: boolean;
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
  if (have >= need && canFillEachOf(opts.each_of)) return "";
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

// exilePermanentOptions resolves an exile-a-permanent cost's options
// (`exile_permanent_options`). Since craft (ADR 0137, CR 702.167b) an
// option may also be a card in the activator's own graveyard, so the
// pool is the battlefield plus that graveyard, still in the server's
// order.
export function exilePermanentOptions(
  battlefield: CardView[],
  graveyard: CardView[] | undefined,
  ids: string[] | undefined,
): CardView[] {
  return orderSacrificeOptions([...battlefield, ...(graveyard ?? [])], ids);
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
// removes on a second click. ADR 0100: when zero is a legal count
// (`min` 0 — "sacrifice any number of creatures" with one creature on
// the board), a second click on the one pick clears it, so the player
// can go back to paying nothing.
export function toggleSacrificePickInRange(
  chosen: string[],
  id: string,
  ceiling: number,
  min = 1,
): string[] {
  if (min <= 0 && chosen.length === 1 && chosen[0] === id) return [];
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
  if (have >= need && canFillEachOf(opts.each_of)) return "";
  if (need === 1) return `nothing to sacrifice (${label})`;
  return `needs ${label} (you have ${have})`;
}

// --- ADR 0100 §3: a variable count on a CAST ----------------------------
//
// A spell's additional cost can print two more counts, and in both of
// them zero is a legal payment:
//
//   "sacrifice any number of creatures"   min 0, max 0 (explicitly sent)
//   "sacrifice X lands"                   count_from_x; X may be 0
//
// The fixed vocabulary above reads a 0 / 0 view as the one-permanent
// clause a bounds-less fixture was before #747, and the activated
// abilities keep that reading. These read the additional cost's view,
// where the server always sends its real bounds, so 0 / 0 is the open
// count ADR 0100 §4 says it is.

// castSacrificeRange is the bounds a cast's sacrifice picker enforces.
// Any other clause reads exactly as sacrificeRange reads it.
export function castSacrificeRange(opts: SacrificeOptionsShape | undefined): SacrificeRange {
  if (opts?.count_from_x) return { min: 0, max: 0 };
  if (opts?.min === 0 && opts?.max === 0) return { min: 0, max: 0 };
  return sacrificeRange(opts);
}

// castSacrificeFloor is how many permanents the cast must be able to
// sacrifice to be castable at all — 0 for the two variable counts.
export function castSacrificeFloor(opts: SacrificeOptionsShape | undefined): number {
  // #2097: "sacrifice all" is paid by whatever is there, nothing included.
  if (opts?.all) return 0;
  return castSacrificeRange(opts).min;
}

// --- #2097: "sacrifice all creatures you control" --------------------------
//
// The server ships `all: true` with min = max = the number of permanents
// listed: the cost takes every one of them and the caster chooses none.
// The picker becomes a confirmation, and the cast sends `sacrifice_ids`
// empty so the server takes the set as it is when the spell is cast (a
// creature that arrived while the sheet was open goes too).

// sacrificesAll reports a clause that takes everything it lists.
export function sacrificesAll(opts: SacrificeOptionsShape | undefined): boolean {
  return opts?.all === true;
}

// sacrificeAllWarning is the sheet's warning line: how many permanents
// the cast takes, said plainly.
export function sacrificeAllWarning(count: number): string {
  if (count === 1) return "This sacrifices the one permanent below. You don't choose.";
  return `This sacrifices all ${count} permanents below. You don't choose.`;
}

// --- #2526: a clause with a SET RULE ---------------------------------------
//
// "Sacrifice a Swamp and a Forest" ships min 2 / max 2 over the union of
// both kinds in `cards`, plus `each_of`: one group per printed part, each
// listing the candidates that could fill it. The picks must fill every
// group with a DIFFERENT permanent — two Swamps are not a Swamp and a
// Forest, and a Swamp Forest fills one part but not both. This is the
// server's own matching (game.SacrificeSetSatisfiedForEffect), so the
// confirm button never opens on a pick the server would refuse.
//
// Absent or empty groups mean no set rule, and every function here then
// answers "yes" / passes its input through, so a plain clause reads
// exactly as it did.

// assignGroups finds one pick per group, no pick used twice, trying
// `order` in turn. Returns the picks in group order, or null.
function assignGroups(groups: SacrificeGroupView[], order: string[]): string[] | null {
  const out: string[] = [];
  const used = new Set<string>();
  const walk = (i: number): boolean => {
    if (i === groups.length) return true;
    const fits = new Set(groups[i].cards ?? []);
    for (const id of order) {
      if (used.has(id) || !fits.has(id)) continue;
      used.add(id);
      out.push(id);
      if (walk(i + 1)) return true;
      out.pop();
      used.delete(id);
    }
    return false;
  };
  return walk(0) ? out : null;
}

// canFillEachOf is whether the permanents on offer can fill every part
// at all — the "can this cost be paid" test for a greyed row.
export function canFillEachOf(groups: SacrificeGroupView[] | undefined): boolean {
  if (!groups || groups.length === 0) return true;
  const all = [...new Set(groups.flatMap((g) => g.cards ?? []))];
  return assignGroups(groups, all) !== null;
}

// fillsEachOf is the confirm gate for a set rule: the picks fill every
// group, one permanent each. No groups means no rule.
export function fillsEachOf(chosen: string[], groups: SacrificeGroupView[] | undefined): boolean {
  if (!groups || groups.length === 0) return true;
  if (chosen.length !== groups.length) return false;
  return assignGroups(groups, chosen) !== null;
}

// chooseSacrificeSetForMe is "Choose for me" under a set rule: a set
// that fills every group, taking a permanent that fits only some groups
// before one that fits them all (so a dual land stays on the board while
// basics will do), and otherwise keeping the server's payment order.
// Empty when the board cannot pay.
export function chooseSacrificeSetForMe(
  options: string[],
  groups: SacrificeGroupView[] | undefined,
): string[] {
  if (!groups || groups.length === 0) return [];
  const fitCount = (id: string) => groups.filter((g) => (g.cards ?? []).includes(id)).length;
  const order = options
    .map((id, i) => ({ id, i, n: fitCount(id) }))
    .sort((a, b) => a.n - b.n || a.i - b.i)
    .map((x) => x.id);
  return assignGroups(groups, order) ?? [];
}

// --- ADR 0137's amendment: "two that share a card type" -------------------
//
// A craft clause (Eye of Ojer Taq) ships `shares`: each candidate's keys
// (its card types), and the picks must all have one key in common. A
// candidate may have several (an artifact creature is both), so this is
// not `same`, which keys each card once. The server lists only keys enough
// candidates have to pay the count, so a candidate with no listed key can
// be part of no payment.
//
// Absent `shares` means no rule, and every function here then answers
// "yes" / passes its input through.

// sharedKeys is the keys every pick has, in the first pick's order.
export function sharedKeys(chosen: string[], shares: TargetSharesView | undefined): string[] {
  if (!shares || chosen.length === 0) return [];
  const keys = shares.keys ?? {};
  let out = keys[chosen[0]] ?? [];
  for (const id of chosen.slice(1)) {
    const k = new Set(keys[id] ?? []);
    out = out.filter((key) => k.has(key));
  }
  return out;
}

// fitsShares is the confirm gate: the picks share a key. No rule means
// no gate.
export function fitsShares(chosen: string[], shares: TargetSharesView | undefined): boolean {
  if (!shares) return true;
  return chosen.length > 0 && sharedKeys(chosen, shares).length > 0;
}

// sharesAllows is whether `id` may be added to `chosen`: it keeps a key
// in common with every pick so far. The greying test.
export function sharesAllows(
  chosen: string[],
  id: string,
  shares: TargetSharesView | undefined,
): boolean {
  if (!shares || chosen.includes(id)) return true;
  return sharedKeys([...chosen, id], shares).length > 0;
}

// canFillShares is whether `need` of the candidates share a key at all —
// the "can this cost be paid" test for a greyed row.
export function canFillShares(shares: TargetSharesView | undefined, need: number): boolean {
  if (!shares) return true;
  const counts = new Map<string, number>();
  for (const ks of Object.values(shares.keys ?? {})) {
    for (const k of ks) counts.set(k, (counts.get(k) ?? 0) + 1);
  }
  return [...counts.values()].some((n) => n >= need);
}

// chooseSharedForMe is "Choose for me" under the rule: the first `need`
// options, in the server's payment order, that have the earliest key
// enough of them share. Empty when none does.
export function chooseSharedForMe(
  options: string[],
  need: number,
  shares: TargetSharesView | undefined,
): string[] {
  if (!shares) return chooseSacrificeForMe(options, need);
  const keys = shares.keys ?? {};
  const order: string[] = [];
  for (const id of options) {
    for (const k of keys[id] ?? []) if (!order.includes(k)) order.push(k);
  }
  for (const k of order) {
    const fit = options.filter((id) => (keys[id] ?? []).includes(k));
    if (fit.length >= need) return fit.slice(0, need);
  }
  return [];
}

// exilePermanentAutoPick is the exile picker's skip (#1600): the ids to
// send without asking when the board offers exactly what the clause
// takes — a fixed count with no more on offer, or an open count's floor
// with nothing beyond it. Null when the player has to choose. A set rule
// is never skipped unless the options are exactly the floor, because a
// wider pool has to be chosen from.
export function exilePermanentAutoPick(opts: SacrificeOptionsShape | undefined): string[] | null {
  if (!opts) return [];
  const options = opts.cards ?? [];
  const range = sacrificeRange(opts);
  const hasRule = (opts.each_of?.length ?? 0) > 0 || !!opts.shares;
  if (hasRule && options.length !== range.min) return null;
  const ceiling = range.max === 0 ? range.min : range.max;
  return options.length <= ceiling ? options : null;
}
