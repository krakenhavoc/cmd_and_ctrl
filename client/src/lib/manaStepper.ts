// manaStepper.ts — the per-colour stepper behind a mana activation that
// adds two or more mana with a colour choice (ADR 0117 §4). Vivi
// Ornitier at power 3 is "{U|R}{U|R}{U|R}": instead of a button per
// answer (#1443's combinations, capped at 12) the player sets a count
// per colour, with a running "N of N", and confirms once.
//
// The server publishes the answer's shape as `color_options`: one list
// per picking slot, narrowed (CR 903.4f) and ordered identity-first
// (#843). The activation takes one colour per slot, in slot order, and
// checks each (`validateUpfrontManaColors`). So the client's whole job
// is to turn a count per colour into a colour per slot, and to refuse a
// count vector no slot assignment satisfies.
//
// Lists can differ (a fixed slot beside a wide one), so a count vector
// is valid only if each slot can be given a colour it offers: a
// bipartite matching between slots and colour units. Hall's condition
// decides it. For every non-empty set T of colours, the counts summed
// over T must not exceed the number of lists offering any colour in T.
// CR 106.1b names six types of mana, so there are at most 63 sets. The
// same test on a partial vector says whether it can still be completed
// (the unmatched slots take any colour they offer, which only adds), so
// it also decides whether "+" may be pressed.
//
// Everything here is pure, so it is tested without a DOM. The component
// is ManaSplitStepper, inside ManaSourcePicker.

import { manaSymbols } from "./manaSymbol";
import type { ManaAbilityView } from "./protocol";

/** How many of each colour the player has set, keyed by colour letter. */
export type SplitCounts = Record<string, number>;

/**
 * stepperSlots is the ability's `color_options` when the stepper is the
 * way to answer it, or null when it is not (ADR 0117 §4, "When"):
 *
 *   - two or more lists, every one non-empty, and at least one of them
 *     a real choice. All one-option lists is no choice at all, and the
 *     ability activates at once like a Forest;
 *   - one list keeps its buttons, `OneColorOfAmount` included;
 *   - a picking slot whose option carries a count (#742, "{W3|U3|…}")
 *     beside a second picking slot keeps one button per answer. No
 *     catalog card has that shape today.
 */
export function stepperSlots(a: ManaAbilityView): string[][] | null {
  const lists = a.color_options;
  if (!lists || lists.length < 2) return null;
  if (lists.some((l) => l.length === 0)) return null;
  if (!lists.some((l) => l.length >= 2)) return null;
  const counted = manaSymbols(a.produced ?? "").some(
    (slot) => slot.includes("|") && slot.split("|").some((o) => /\d/.test(o)),
  );
  if (counted) return null;
  return lists;
}

/**
 * splitColors is one entry per colour any list offers, in the order the
 * server sent them: the union of the lists by first appearance. That is
 * identity-first (#843), which docs/protocol.md asks clients to keep.
 */
export function splitColors(lists: readonly (readonly string[])[]): string[] {
  const out: string[] = [];
  for (const l of lists) for (const c of l) if (!out.includes(c)) out.push(c);
  return out;
}

/** The counts' sum. */
export function splitTotal(counts: SplitCounts): number {
  return Object.values(counts).reduce((n, k) => n + Math.max(0, k), 0);
}

/**
 * splitFeasible says whether `counts` can be met, or still completed,
 * by a slot assignment: Hall's condition over every non-empty set of
 * the offered colours. A count on a colour no list offers fails it.
 */
export function splitFeasible(lists: readonly (readonly string[])[], counts: SplitCounts): boolean {
  const colors = splitColors(lists);
  for (const [c, k] of Object.entries(counts)) {
    if (k < 0) return false;
    if (k > 0 && !colors.includes(c)) return false;
  }
  const masks = lists.map((l) => colors.reduce((m, c, i) => (l.includes(c) ? m | (1 << i) : m), 0));
  const sets = 1 << colors.length;
  for (let t = 1; t < sets; t++) {
    let want = 0;
    for (let i = 0; i < colors.length; i++) if (t & (1 << i)) want += counts[colors[i]] ?? 0;
    if (want === 0) continue;
    let offer = 0;
    for (const m of masks) if (m & t) offer++;
    if (want > offer) return false;
  }
  return true;
}

/** Whether `counts` fills every slot and some assignment meets it. */
export function splitComplete(lists: readonly (readonly string[])[], counts: SplitCounts): boolean {
  return splitTotal(counts) === lists.length && splitFeasible(lists, counts);
}

/** `counts` with `color` moved by `delta`, never below zero. */
export function splitAdjust(counts: SplitCounts, color: string, delta: number): SplitCounts {
  return { ...counts, [color]: Math.max(0, (counts[color] ?? 0) + delta) };
}

/**
 * fixedShare is how many slots offer `color` and nothing else: the part
 * of its count that cannot be stepped away (Command Tower narrowed to
 * one colour, CR 903.4f).
 */
export function fixedShare(lists: readonly (readonly string[])[], color: string): number {
  return lists.filter((l) => l.length === 1 && l[0] === color).length;
}

/** "+" is enabled while the total is short and one more stays feasible. */
export function canAdd(
  lists: readonly (readonly string[])[],
  counts: SplitCounts,
  color: string,
): boolean {
  if (splitTotal(counts) >= lists.length) return false;
  return splitFeasible(lists, splitAdjust(counts, color, 1));
}

/** "−" is disabled at 0 and on a fixed slot's share. */
export function canRemove(
  lists: readonly (readonly string[])[],
  counts: SplitCounts,
  color: string,
): boolean {
  return (counts[color] ?? 0) > fixedShare(lists, color);
}

/**
 * splitAssignment is the colour for each slot, in slot order, that
 * meets `counts` exactly: the `colors[]` the activation carries. A
 * maximum matching by augmenting paths (each colour a node whose
 * capacity is its count). Null when the counts do not fill every slot
 * or no assignment meets them.
 */
export function splitAssignment(
  lists: readonly (readonly string[])[],
  counts: SplitCounts,
): string[] | null {
  if (splitTotal(counts) !== lists.length) return null;
  const assign: (string | undefined)[] = new Array(lists.length).fill(undefined);
  const holders = (c: string): number[] => assign.flatMap((a, i) => (a === c ? [i] : []));

  function place(slot: number, seen: Set<string>): boolean {
    for (const c of lists[slot]) {
      if (seen.has(c)) continue;
      seen.add(c);
      const held = holders(c);
      if (held.length < (counts[c] ?? 0)) {
        assign[slot] = c;
        return true;
      }
      for (const other of held) {
        if (place(other, seen)) {
          assign[slot] = c;
          return true;
        }
      }
    }
    return false;
  }

  for (let slot = 0; slot < lists.length; slot++) {
    if (!place(slot, new Set())) return null;
  }
  return assign as string[];
}

/** The counts of an assignment, one per colour it names. */
function countsOf(colors: readonly string[]): SplitCounts {
  const out: SplitCounts = {};
  for (const c of colors) out[c] = (out[c] ?? 0) + 1;
  return out;
}

/**
 * splitStart is where the stepper opens (ADR 0117 §4, "Start state"):
 * the split last confirmed for this card in this game, if it still
 * fits the current lists; otherwise all of it on the first colour every
 * list offers; otherwise each slot on its own first option. It always
 * fills the total, so the same split again is one press.
 */
export function splitStart(
  lists: readonly (readonly string[])[],
  remembered?: SplitCounts | null,
): SplitCounts {
  if (remembered && splitComplete(lists, remembered)) {
    const out: SplitCounts = {};
    for (const [c, k] of Object.entries(remembered)) if (k > 0) out[c] = k;
    return out;
  }
  const common = splitColors(lists).find((c) => lists.every((l) => l.includes(c)));
  if (common) return { [common]: lists.length };
  return countsOf(lists.map((l) => l[0]));
}

// ---- the split last confirmed, per card, in client memory only ----

const remembered = new Map<string, SplitCounts>();

/** The memory key: one game, one card, one ability. */
export function splitMemoryKey(gameID: string, cardID: string, abilityIndex: number): string {
  return `${gameID}\u0000${cardID}\u0000${abilityIndex}`;
}

export function rememberSplit(key: string, counts: SplitCounts): void {
  remembered.set(key, { ...counts });
}

export function rememberedSplit(key: string): SplitCounts | null {
  const c = remembered.get(key);
  return c ? { ...c } : null;
}

/** Tests only: forget every remembered split. */
export function forgetSplits(): void {
  remembered.clear();
}
