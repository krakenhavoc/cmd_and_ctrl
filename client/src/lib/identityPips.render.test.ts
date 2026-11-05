// @vitest-environment jsdom
//
// #4847: a deck's colour identity is drawn as its mana symbols, {W} {U}
// {B} {R} {G} in WUBRG order, the same ManaSymbol a cost and the mana
// pool draw, not as letters on plain discs. A deck known to be
// colourless shows {C}; one whose identity is unknown shows nothing. The
// row is one image with one name. It holds in the deck picker's grid
// (and the Add bot grid, which is the same component) and on a seat tile.

import { afterEach, beforeEach, describe, expect, it } from "vitest";

import DeckGrid from "./components/DeckGrid.svelte";
import IdentityPips from "./components/IdentityPips.svelte";
import SeatTile from "./components/SeatTile.svelte";
import { libraryTile, prebuiltTile } from "./deckPicker";
import type { MyDeckInfo } from "./myDecks";
import { seatTileText } from "./tablesPage";
import { cleanup, render } from "./test/render.svelte";

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
});
afterEach(cleanup);

/** The identity row inside `root`, or null when none is drawn. */
const pipsIn = (root: Element) =>
  root.querySelector<HTMLElement>('[role="img"][aria-label^="colour identity"]');

/** The mana symbols the row draws, by symbol. */
const symbolsIn = (root: Element) =>
  [...root.querySelectorAll("svg.mana-symbol")].map((s) => s.getAttribute("data-symbol"));

function pips(props: Record<string, unknown>) {
  return render(IdentityPips as never, props as never).container;
}

describe("IdentityPips", () => {
  it("draws each colour as its mana symbol, in WUBRG order, under one name", () => {
    const c = pips({ colors: ["g", "U", "W"] });
    const row = pipsIn(c)!;
    expect(row.getAttribute("aria-label")).toBe("colour identity: white, blue, green");
    expect(symbolsIn(row)).toEqual(["W", "U", "G"]);
    // The symbols are decorative inside the one image, and no letter is
    // left as text beside them.
    for (const s of row.querySelectorAll("svg.mana-symbol")) {
      expect(s.getAttribute("aria-hidden")).toBe("true");
    }
    expect(row.textContent?.trim()).toBe("");
  });

  it("draws {C} for a deck known to be colourless", () => {
    const row = pipsIn(pips({ colors: [], colorless: true }))!;
    expect(row.getAttribute("aria-label")).toBe("colour identity: colourless");
    expect(symbolsIn(row)).toEqual(["C"]);
  });

  it("draws nothing when the identity is unknown", () => {
    expect(pipsIn(pips({ colors: [] }))).toBeNull();
    expect(pipsIn(pips({ colors: undefined }))).toBeNull();
  });

  it("lets the colours win over a stray colourless flag", () => {
    expect(symbolsIn(pips({ colors: ["R"], colorless: true }))).toEqual(["R"]);
  });
});

describe("the deck picker's tiles", () => {
  const saved = (id: string, extra: Partial<MyDeckInfo>): MyDeckInfo => ({
    id,
    name: id,
    commanders: [],
    card_count: 100,
    updated_at: "2026-10-01T00:00:00Z",
    ...extra,
  });

  it("draws mana symbols, {C} for a colourless commander, nothing for an unknown one", () => {
    const tiles = [
      saved("esper", {
        commanders: ["Raffine, Scheming Seer"],
        commander_cards: [{ name: "Raffine, Scheming Seer", color_identity: ["W", "U", "B"] }],
        color_identity: ["W", "U", "B"],
      }),
      saved("eldrazi", {
        commanders: ["Kozilek, the Great Distortion"],
        commander_cards: [{ name: "Kozilek, the Great Distortion", color_identity: [] }],
      }),
      saved("pasted", {}),
    ].map(libraryTile);
    const { container } = render(
      DeckGrid as never,
      { tiles, name: "your-deck", label: "your decks", selected: "" } as never,
    );
    const tile = (id: string) => container.querySelector(`.dtile[data-id="${id}"]`)!;
    expect(symbolsIn(tile("esper"))).toEqual(["W", "U", "B"]);
    expect(pipsIn(tile("esper"))?.getAttribute("aria-label")).toBe(
      "colour identity: white, blue, black",
    );
    expect(symbolsIn(tile("eldrazi"))).toEqual(["C"]);
    expect(pipsIn(tile("pasted"))).toBeNull();
  });

  it("draws a pre-built deck's colours as symbols too", () => {
    const tiles = [
      prebuiltTile({
        id: "fury",
        name: "Forge and Fury",
        commander: "Purphoros, God of the Forge",
        colors: ["R"],
        card_count: 100,
        coverage: { cards: 73, full: 73, caveats: 0, unreviewed: 0, basics: 1, unregistered: 0 },
      }),
    ];
    const { container } = render(
      DeckGrid as never,
      { tiles, name: "prebuilt-deck", label: "pre-built decks", selected: "fury" } as never,
    );
    expect(symbolsIn(container.querySelector(".dtile")!)).toEqual(["R"]);
  });
});

describe("a seat tile", () => {
  function seatTile(p: Record<string, unknown>) {
    const seat = { player_id: "a1", name: "Alice", seat: 0, deck_uploaded: true, ...p };
    return render(
      SeatTile as never,
      { seat: 0, p: seat, text: seatTileText(seat as never, undefined) } as never,
    ).container.querySelector("li.tile")!;
  }

  it("draws a person's identity as mana symbols", () => {
    const li = seatTile({
      deck_name: "Blood and thunder",
      commander_cards: [
        { name: "Tymna the Weaver", color_identity: ["W", "B"] },
        { name: "Kraum, Ludevic's Opus", color_identity: ["U", "R"] },
      ],
      color_identity: ["R", "W", "B", "U"],
    });
    expect(symbolsIn(li)).toEqual(["W", "U", "B", "R"]);
  });

  it("draws {C} for a colourless commander, and nothing for a seat without commanders", () => {
    const kozilek = seatTile({
      deck_name: "Eldrazi",
      commander_cards: [{ name: "Kozilek", color_identity: [] }],
    });
    expect(symbolsIn(kozilek)).toEqual(["C"]);
    cleanup();
    expect(pipsIn(seatTile({ deck_name: "Kenrith's court" }))).toBeNull();
  });
});
