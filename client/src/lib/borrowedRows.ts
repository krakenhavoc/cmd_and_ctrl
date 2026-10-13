// borrowedRows.ts — ADR 0155 owner decisions 2 and 3 on the client: how a
// permanent that "has all activated abilities of …" other objects shows
// them.
//
// The server publishes one row per lent ability, each marked with the
// lender it was read off (`borrowed_from`) and with a `borrow:` ref. This
// file does the presenting and nothing else:
//
//   - identical rows are shown ONCE (collapseBorrowed): two Llanowar Elves
//     in two graveyards are two instances of the same ability (CR 113.2c),
//     and Necrotic Ooze in a creature-heavy graveyard would otherwise list
//     twenty. Display only. The engine keeps both rows, and the row that
//     is shown carries the first one's `index` and `ref`, so that is what
//     a click sends;
//   - the rows that remain are grouped under "From <lender>" (splitBorrowed);
//   - a lender the catalog does not automate is a disabled
//     "<lender>: not automated" line (unautomatedLines).

import type { CardView, GrantedByView } from "./protocol";
import { borrowedRowKey, isBorrowedRow } from "./abilityRef";

/** The slice of an ability row this file reads. */
export interface BorrowableRow {
  ref?: string;
  label?: string;
  borrowed_from?: GrantedByView;
}

/**
 * Drops each borrowed row that repeats an earlier row of the list, own or
 * lent, keeping the first. Two rows repeat when they are the same catalog
 * row lent twice (the same `borrow:` ref up to its instance count), or when
 * they read the same: the server writes a lent label with the lender's name
 * read as the borrower's (CR 201.5b), so one ability lent by two different
 * creatures, or one the borrower already has, has one label. Returns `rows`
 * itself when nothing is borrowed or nothing repeats, so a list with no
 * borrow is never rebuilt.
 */
export function collapseBorrowed<T extends BorrowableRow>(rows: T[]): T[] {
  if (!rows.some(isBorrowedRow)) return rows;
  const labels = new Set<string>();
  const keys = new Set<string>();
  const out: T[] = [];
  for (const row of rows) {
    const key = borrowedRowKey(row.ref);
    const label = row.label ?? "";
    if (isBorrowedRow(row) && ((key && keys.has(key)) || (label && labels.has(label)))) continue;
    if (key) keys.add(key);
    if (label) labels.add(label);
    out.push(row);
  }
  return out.length === rows.length ? rows : out;
}

/** One lender's rows. */
export interface BorrowGroup<T> {
  /** The lender's object id, or "" when the server did not name it. */
  lenderID: string;
  /** "From Llanowar Elves", or "From another object" for an unnamed lender. */
  heading: string;
  rows: T[];
}

/** A list split into the permanent's own (and granted) rows and its lent ones. */
export interface SplitRows<T> {
  own: T[];
  groups: BorrowGroup<T>[];
}

/** Heading of the group a lender's rows sit under. */
export function fromHeading(lender: GrantedByView | undefined): string {
  return lender?.name ? `From ${lender.name}` : "From another object";
}

/**
 * Splits a list into own rows and one group per lender, in the order the
 * lenders first appear. Pass the collapsed list: a row that was dropped as
 * a duplicate then stays under the first lender that has it.
 */
export function splitBorrowed<T extends BorrowableRow>(rows: readonly T[]): SplitRows<T> {
  const own: T[] = [];
  const groups: BorrowGroup<T>[] = [];
  const byLender = new Map<string, BorrowGroup<T>>();
  for (const row of rows) {
    if (!isBorrowedRow(row)) {
      own.push(row);
      continue;
    }
    const key = row.borrowed_from?.id ?? "";
    let group = byLender.get(key);
    if (!group) {
      group = { lenderID: key, heading: fromHeading(row.borrowed_from), rows: [] };
      byLender.set(key, group);
      groups.push(group);
    }
    group.rows.push(row);
  }
  return { own, groups };
}

/** One disabled "<lender>: not automated" line of the menu. */
export interface UnautomatedLine {
  id: string;
  label: string;
  hint: string;
}

/**
 * The permanent's "<lender>: not automated" lines (owner decision 2), one
 * per distinct lender name: the engine lends nothing from an object the
 * catalog has no entry for, and says so rather than staying silent.
 */
export function unautomatedLines(card: CardView): UnautomatedLine[] {
  const seen = new Set<string>();
  const out: UnautomatedLine[] = [];
  for (const lender of card.unautomated_lenders ?? []) {
    const name = lender.name ?? "";
    if (!name || seen.has(name)) continue;
    seen.add(name);
    out.push({
      id: `unautomated-${lender.id}`,
      label: `${name}: not automated`,
      hint: `${name}'s abilities are not automated, so they are not lent. Resolve them by hand.`,
    });
  }
  return out;
}
