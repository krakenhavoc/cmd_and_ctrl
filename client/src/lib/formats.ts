// formats.ts: a table's format on the client (ADR 0222 §1, #2352).
//
// The server's format table (server/internal/game/format.go) is the
// authority: it sets the starting life and the seat cap and checks the
// deck. This file is only what the client needs to SAY about a format:
// its name, the seats to draw and the copy of the deck picker.
//
// Commander is the default. The server omits the format of a Commander
// table, deck and setup, so `undefined` and "" read as Commander, and a
// name this build does not know does too: an older client on a newer
// server still draws a table.

import { L } from "./labels";

/** A format's wire name. */
export type FormatID = "commander" | "standard" | "limited";

export interface FormatInfo {
  id: FormatID;
  /** The display name, from the labels registry. */
  name: string;
  /** One line under the name in the new-table form's picker. */
  blurb: string;
  /** Seats at a table of the format. */
  seats: number;
  /** Starting life, for the picker's blurb and the setup summary. */
  startingLife: number;
  /** Whether a deck of the format has a commander. */
  hasCommander: boolean;
  /** The deck picker's note on what a pasted list is checked against. */
  deckRules: string;
  /** The paste box's placeholder: a link, then the shape of a list. */
  pastePlaceholder: string;
}

const COMMANDER: FormatInfo = {
  id: "commander",
  name: L.formatCommander,
  blurb: "up to 4 players, 40 life, 100-card decks with a commander",
  seats: 4,
  startingLife: 40,
  hasCommander: true,
  deckRules:
    "Every deck is validated against Commander rules — 100-card singleton, color identity, format legality — whichever way it arrives.",
  pastePlaceholder:
    "https://archidekt.com/decks/123456\n\n— or paste the list —\n\nCommander:\n1 Atraxa, Praetors' Voice\n\nMainboard:\n1 Sol Ring\n...",
};

const STANDARD: FormatInfo = {
  id: "standard",
  name: L.formatStandard,
  blurb: "2 players, 20 life, 60-card decks, one game or best of three",
  seats: 2,
  startingLife: 20,
  hasCommander: false,
  deckRules:
    "Every deck is validated against Standard rules — at least 60 cards, at most 4 of a card other than basic lands, every card legal in Standard, a sideboard of up to 15 — whichever way it arrives. In a best of three you can swap sideboard cards in between games.",
  pastePlaceholder:
    "https://archidekt.com/decks/123456\n\n— or paste the list —\n\nDeck:\n4 Llanowar Elves\n4 Giant Growth\n22 Forest\n...\n\nSideboard:\n2 Naturalize\n...",
};

// Limited is a draft's format (ADR 0222 §6, #5440): a table of it is
// opened by a draft for one of its matches, with both decks already
// seated. It is never in the picker and no deck is ever pasted for it,
// so its deck copy is only ever read by a summary.
const LIMITED: FormatInfo = {
  id: "limited",
  name: L.formatLimited,
  blurb: "2 players, 20 life, 40-card decks built from a draft",
  seats: 2,
  startingLife: 20,
  hasCommander: false,
  deckRules:
    "A limited deck is built inside its draft: at least 40 cards from the cards you drafted, plus any number of basic lands.",
  pastePlaceholder: "",
};

/** Every format a table can be made in, in the order the picker lists them: Commander first. */
export const FORMATS: readonly FormatInfo[] = [COMMANDER, STANDARD];

/** Every format a table can have, the picker's and a draft's. */
const KNOWN: readonly FormatInfo[] = [...FORMATS, LIMITED];

/** The format a wire value names; Commander for none or an unknown one. */
export function formatOf(id: string | undefined | null): FormatInfo {
  return KNOWN.find((f) => f.id === id) ?? COMMANDER;
}

/** Whether a deck built for `deckFormat` may be seated at a `tableFormat` table. */
export function fitsFormat(
  deckFormat: string | undefined | null,
  tableFormat: string | undefined | null,
): boolean {
  return formatOf(deckFormat).id === formatOf(tableFormat).id;
}

/** The seats a table has: its own cap when the server names one, else its format's. */
export function tableSeats(g: { format?: string; max_seats?: number }): number {
  return g.max_seats && g.max_seats > 0 ? g.max_seats : formatOf(g.format).seats;
}
