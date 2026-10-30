// deckPicker.ts — the deck picker's tiles and tabs (ADR 0212 §3), and the
// filter over them.
//
// The picker is a grid of commander art behind three tabs: My decks,
// Pre-built and Paste a list. A tile says the deck's name once; the line
// under it is the commander only when the name is not the commander's,
// otherwise the card count (and a pre-built deck's archetype). A tile
// carries a flag only when the deck has cards that do not play as
// printed; a deck that plays as printed says nothing. Everything longer
// (the coverage sentence, the source, a pre-built deck's summary and its
// caveats) is said once, for the selected deck, under the grid.
//
// Pure functions, like prebuiltDecks.ts and myDecks.ts, so what a tile
// claims has a unit test.

import type { BotDeckInfo } from "./api";
import { deckArt, deckIdentity, type DeckArt } from "./deckGrid";
import type { LastDeck } from "./lastDeck";
import { coverageDetail, coverageLine, type MyDeckInfo } from "./myDecks";
import { orderedColors, type PrebuiltDeck } from "./prebuiltDecks";
import type { SeatTileText } from "./tablesPage";

export type TileKind = "library" | "prebuilt";

export interface DeckTile {
  id: string;
  kind: TileKind;
  name: string;
  /** The line under the name: the commander(s) when not the name, else the card count. */
  sub: string;
  /** The commander, or a partner pair joined with " · "; "" when unknown. */
  commander: string;
  /** WUBRG letters, in order, for the pips. */
  colors: string[];
  /** The identity the art slot washes in: WUBRG, [] for colourless, null when unknown. */
  identity: string[] | null;
  /** The commanders with a printing to draw, partners side by side. */
  art: DeckArt[];
  /** "2 not as printed", "1 not implemented", or "" for a deck that plays as printed. */
  flag: string;
  /** What the flag means, for its tooltip. */
  flagTitle: string;
  /** Lower-cased text the filter matches against. */
  haystack: string;
}

/** A list longer than this gets a filter box. */
export const FILTER_FROM = 8;

/** How a partner pair, or a commander with a Background, is named on one line. */
export const COMMANDER_JOIN = " · ";

const plural = (n: number, word: string): string => `${n} ${word}${n === 1 ? "" : "s"}`;

// Case, curly quotes and runs of spaces do not make two names differ.
const norm = (s: string): string =>
  s.toLowerCase().replace(/[‘’`]/g, "'").replace(/\s+/g, " ").trim();

/** sameName: a deck named for its commander ("Krenko, Mob Boss"). */
export function sameName(a: string, b: string): boolean {
  return norm(a) !== "" && norm(a) === norm(b);
}

function flagged(n: number, what: string, title: string): Pick<DeckTile, "flag" | "flagTitle"> {
  return n > 0 ? { flag: `${n} ${what}`, flagTitle: title } : { flag: "", flagTitle: "" };
}

/**
 * libraryTile is a saved deck's tile. Its flag counts copies that do not
 * play as printed (the strict reading of ADR 0110 §6, as the Decks page's
 * ring does), and is absent when the server sent no coverage.
 */
export function libraryTile(deck: MyDeckInfo): DeckTile {
  const commander = deck.commanders.filter((c) => c.trim() !== "").join(COMMANDER_JOIN);
  const c = deck.coverage;
  const gap = c && c.deck_size > 0 ? Math.max(0, c.deck_size - c.as_printed_copies) : 0;
  return {
    id: deck.id,
    kind: "library",
    name: deck.name,
    sub: commander && !sameName(deck.name, commander) ? commander : plural(deck.card_count, "card"),
    commander,
    colors: orderedColors(deck.color_identity),
    identity: deckIdentity(deck),
    art: deckArt(deck),
    ...flagged(
      gap,
      "not as printed",
      `${plural(gap, "card")} in this deck ${gap === 1 ? "does" : "do"} not play exactly as printed`,
    ),
    haystack: [deck.name, commander].join(" ").toLowerCase(),
  };
}

/**
 * prebuiltTile is a pre-built deck's tile. Its art is the commander's
 * crop (ADR 0212 phase B1), the same as a saved deck's; with no printing
 * to draw (an older server, or a commander with no image) the art slot
 * washes in the deck's colours. A card with no implementation at all is
 * said as such; otherwise the flag counts the cards that are simplified
 * or not yet reviewed, which is everything the summariser does not call
 * "as printed".
 */
export function prebuiltTile(deck: PrebuiltDeck): DeckTile {
  const commander = (deck.commander ?? "").trim();
  const c = deck.coverage;
  const colors = orderedColors(deck.colors);
  const short = c.caveats + c.unreviewed;
  return {
    id: deck.id,
    kind: "prebuilt",
    name: deck.name,
    sub:
      commander && !sameName(deck.name, commander)
        ? commander
        : [deck.archetype, plural(deck.card_count, "card")].filter(Boolean).join(" · "),
    commander,
    colors,
    identity: deck.colors ? colors : null,
    art: deckArt(deck),
    ...(c.unregistered > 0
      ? flagged(
          c.unregistered,
          "not implemented",
          `${plural(c.unregistered, "card")} in this deck ${c.unregistered === 1 ? "has" : "have"} no implementation yet`,
        )
      : flagged(
          short,
          "not as printed",
          `${plural(short, "card")} in this deck ${short === 1 ? "is" : "are"} simplified or not yet reviewed`,
        )),
    haystack: [deck.name, commander, deck.archetype ?? "", deck.summary ?? ""]
      .join(" ")
      .toLowerCase(),
  };
}

/**
 * botDeckTile is a bot deck's tile in the Add bot picker. The bot decks
 * are the pre-built decks, so GET /decks' entry, when it is in, gives
 * the coverage; a bot deck the list does not carry still gets a tile.
 */
export function botDeckTile(deck: BotDeckInfo, prebuilt?: PrebuiltDeck): DeckTile {
  if (prebuilt) return { ...prebuiltTile(prebuilt), id: deck.id, name: deck.name };
  const commander = (deck.commander ?? "").trim();
  const colors = orderedColors(deck.colors);
  return {
    id: deck.id,
    kind: "prebuilt",
    name: deck.name,
    sub: commander && !sameName(deck.name, commander) ? commander : "Pre-built",
    commander,
    colors,
    identity: deck.colors ? colors : null,
    art: deckArt(deck),
    flag: "",
    flagTitle: "",
    haystack: [deck.name, commander, deck.description ?? ""].join(" ").toLowerCase(),
  };
}

/**
 * filterTiles keeps the tiles matching every word of the query, in any
 * field and any order: "nissa ramp" finds a deck named Ramp led by
 * Nissa. An empty query keeps everything.
 */
export function filterTiles(tiles: readonly DeckTile[], query: string): DeckTile[] {
  const words = query.toLowerCase().split(/\s+/).filter(Boolean);
  if (words.length === 0) return [...tiles];
  return tiles.filter((t) => words.every((w) => t.haystack.includes(w)));
}

/**
 * libraryDetail is the detail line under the grid for a saved deck,
 * after its name and before its source: the commander when the name is
 * not it, the card count, and the coverage sentence ("every card plays
 * as printed", or how many do and what the rest need).
 */
export function libraryDetail(deck: MyDeckInfo): string[] {
  const t = libraryTile(deck);
  const parts: string[] = [];
  if (t.commander && !sameName(deck.name, t.commander)) parts.push(t.commander);
  parts.push(plural(deck.card_count, "card"));
  if (deck.coverage) {
    if (t.flag) parts.push(coverageLine(deck), coverageDetail(deck));
    else parts.push("every card plays as printed");
  }
  return parts.filter(Boolean);
}

// --- Tabs (ADR 0212 §3, owner decision 7) -----------------------------

export type PickerTab = "library" | "prebuilt" | "paste";

/**
 * pickerTabs: My decks for a signed-in person with saved decks, then
 * Pre-built when the server offers any, then Paste a list, which is
 * always there. A guest, or an empty library, gets no My decks tab.
 */
export function pickerTabs(libraryCount: number, prebuiltCount: number): PickerTab[] {
  const tabs: PickerTab[] = [];
  if (libraryCount > 0) tabs.push("library");
  if (prebuiltCount > 0) tabs.push("prebuilt");
  tabs.push("paste");
  return tabs;
}

/**
 * firstTab is the tab the picker opens on: My decks, unless the last deck
 * this person seated was a pre-built one, or there is no library. With
 * neither list it is Paste a list.
 */
export function firstTab(tabs: readonly PickerTab[], last: LastDeck | null): PickerTab {
  if (tabs.includes("prebuilt") && (last?.kind === "prebuilt" || !tabs.includes("library"))) {
    return "prebuilt";
  }
  return tabs[0];
}

// --- What the picker tells its caller --------------------------------

/** The commit the picker offers: the lobby card's footer primary, or its own button. */
export interface PickerCommit {
  label: string;
  disabled: boolean;
  /** What it does, or why it cannot (DESIGN.md §6). */
  title: string;
  busy: boolean;
  /** Nothing to send yet: Paste a list with an empty box. */
  empty?: boolean;
}

/** The selected deck, drawn in your seat's tile before anything is sent (§1). */
export interface DeckPreview {
  /** The deck's name, to tell it from the deck already in the seat. */
  name: string;
  text: SeatTileText;
  art: DeckArt[];
}

export interface PickerState {
  /** null while there is nothing to commit (no deck in the open tab). */
  commit: PickerCommit | null;
  preview: DeckPreview | null;
}

/**
 * tilePreview is what your seat's tile says while a tile is selected: the
 * commander, then the deck's name when it is not the commander's (a
 * pre-built deck says "Pre-built · <deck>", as a bot's seat does), the
 * pips and the art. It is not ready: nothing has been sent.
 */
export function tilePreview(t: DeckTile): DeckPreview {
  const second =
    t.kind === "prebuilt"
      ? `Pre-built · ${t.name}`
      : t.commander && !sameName(t.name, t.commander)
        ? t.name
        : "";
  return {
    name: t.name,
    text: { what: t.commander || t.name, second, colors: t.colors, ready: false },
    art: t.art,
  };
}
