// deckPicker.ts — the rows both deck pickers draw (YourDecksPicker and
// PrebuiltDeckPicker), and the filter over them.
//
// A row is one line: name, colour pips, a dim subtitle and a short
// coverage figure. Everything longer — a pre-built deck's summary, the
// full coverage sentence, the caveated cards — is said once, under the
// list, for the selected deck. A card per deck made the panel pages
// long once a library filled up (it holds up to 200); a row per deck
// plus a scroll cap keeps it one screen.
//
// Pure functions, like prebuiltDecks.ts and myDecks.ts, so what a row
// claims has a unit test.

import { deckSubtitle as librarySubtitle, type MyDeckInfo } from "./myDecks";
import {
  deckSubtitle as prebuiltSubtitle,
  orderedColors,
  type PrebuiltDeck,
} from "./prebuiltDecks";

/** How a row's coverage dot is coloured. `none` draws no figure at all. */
export type RowTone = "full" | "caveats" | "gap" | "none";

export interface PickerRow {
  id: string;
  name: string;
  /** WUBRG letters, in order. */
  colors: string[];
  /** The dim line after the name: archetype · commander, or commanders · N cards. */
  sub: string;
  /** "85/87 as printed", or "" when there is nothing to say. */
  cov: string;
  tone: RowTone;
  /** Lower-cased text the filter matches against. */
  haystack: string;
}

/** A list longer than this gets a filter box. */
export const FILTER_FROM = 8;

/**
 * prebuiltRow is a pre-built deck's row. The short figure keeps the
 * summariser's rule: "as printed" is only ever the `full` count, and a
 * deck with an unimplemented card says so instead.
 */
export function prebuiltRow(deck: PrebuiltDeck): PickerRow {
  const c = deck.coverage;
  let cov: string;
  let tone: RowTone;
  if (c.unregistered > 0) {
    cov = `${c.unregistered} not implemented`;
    tone = "gap";
  } else if (c.caveats === 0 && c.unreviewed === 0) {
    cov = `all ${c.cards} as printed`;
    tone = "full";
  } else {
    cov = `${c.full}/${c.cards} as printed`;
    tone = "caveats";
  }
  const sub = prebuiltSubtitle(deck);
  return {
    id: deck.id,
    name: deck.name,
    colors: orderedColors(deck.colors),
    sub,
    cov,
    tone,
    haystack: [deck.name, sub, deck.summary ?? ""].join(" ").toLowerCase(),
  };
}

/** libraryRow is a saved deck's row; no figure when the server sent no coverage. */
export function libraryRow(deck: MyDeckInfo): PickerRow {
  const c = deck.coverage;
  let cov = "";
  let tone: RowTone = "none";
  if (c && c.deck_size > 0) {
    cov = `${c.as_printed_copies}/${c.deck_size} as printed`;
    tone = c.as_printed_copies >= c.deck_size ? "full" : "caveats";
  }
  const sub = librarySubtitle(deck);
  return {
    id: deck.id,
    name: deck.name,
    colors: orderedColors(deck.color_identity),
    sub,
    cov,
    tone,
    haystack: [deck.name, sub].join(" ").toLowerCase(),
  };
}

/**
 * filterRows keeps the rows matching every word of the query, in any
 * field and any order: "nissa ramp" finds a deck named Ramp led by
 * Nissa. An empty query keeps everything.
 */
export function filterRows(rows: readonly PickerRow[], query: string): PickerRow[] {
  const words = query.toLowerCase().split(/\s+/).filter(Boolean);
  if (words.length === 0) return [...rows];
  return rows.filter((r) => words.every((w) => r.haystack.includes(w)));
}
