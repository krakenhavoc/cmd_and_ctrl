import { describe, expect, it } from "vitest";
import {
  botDeckTile,
  filterTiles,
  firstTab,
  libraryDetail,
  libraryTile,
  pickerTabs,
  prebuiltTile,
  sameName,
  tilePreview,
} from "./deckPicker";
import type { MyDeckInfo } from "./myDecks";
import type { DeckCoverage, PrebuiltDeck } from "./prebuiltDecks";

function prebuilt(cov: Partial<DeckCoverage>, extra: Partial<PrebuiltDeck> = {}): PrebuiltDeck {
  return {
    id: "fury",
    name: "Forge and Fury",
    archetype: "aggro",
    commander: "Purphoros, God of the Forge",
    summary: "Goblins and tokens.",
    colors: ["R"],
    card_count: 100,
    coverage: {
      cards: 73,
      full: 73,
      caveats: 0,
      unreviewed: 0,
      basics: 1,
      unregistered: 0,
      ...cov,
    },
    ...extra,
  };
}

function saved(extra: Partial<MyDeckInfo> = {}): MyDeckInfo {
  return {
    id: "d1",
    name: "Inverse Nissa",
    commanders: ["Nissa, Leyline Tamer"],
    card_count: 100,
    updated_at: "2026-10-01T00:00:00Z",
    ...extra,
  };
}

const LIB_COV = {
  counts: { manual: 0, unreviewed: 0, caveats: 2, automated: 50, no_effect: 46 },
  unknown: 0,
  as_printed: 96,
  resolved: 98,
  copies: { manual: 0, unreviewed: 0, caveats: 2, automated: 52, no_effect: 46 },
  unknown_copies: 0,
  as_printed_copies: 98,
  deck_size: 100,
};

describe("the line under a tile's name (ADR 0212 §3)", () => {
  it("is the commander when the deck is named for something else", () => {
    expect(libraryTile(saved()).sub).toBe("Nissa, Leyline Tamer");
    expect(prebuiltTile(prebuilt({})).sub).toBe("Purphoros, God of the Forge");
  });

  it("is the card count when the deck is named for its commander", () => {
    const t = libraryTile(saved({ name: "Krenko, Mob Boss", commanders: ["Krenko, Mob Boss"] }));
    expect(t.sub).toBe("100 cards");
    // Case and curly quotes do not make two names differ.
    expect(sameName("atraxa, praetors’ voice", "Atraxa, Praetors' Voice")).toBe(true);
    expect(sameName("", "")).toBe(false);
  });

  it("names a partner pair, and a pre-built deck named for its commander by archetype", () => {
    const pair = libraryTile(
      saved({ name: "Partner blitz", commanders: ["Tymna the Weaver", "Kraum, Ludevic's Opus"] }),
    );
    expect(pair.sub).toBe("Tymna the Weaver · Kraum, Ludevic's Opus");
    expect(pair.commander).toBe("Tymna the Weaver · Kraum, Ludevic's Opus");
    const named = prebuiltTile(prebuilt({}, { name: "Purphoros, God of the Forge" }));
    expect(named.sub).toBe("aggro · 100 cards");
  });

  it("is the count alone without a commander", () => {
    expect(libraryTile(saved({ commanders: [" "] })).sub).toBe("100 cards");
  });
});

describe("the gap flag: only a deck with cards that do not play as printed", () => {
  it("says nothing for a deck that plays as printed, or one with no coverage", () => {
    expect(prebuiltTile(prebuilt({})).flag).toBe("");
    expect(libraryTile(saved({ coverage: { ...LIB_COV, as_printed_copies: 100 } })).flag).toBe("");
    expect(libraryTile(saved()).flag).toBe("");
  });

  it("counts a saved deck's copies, and a pre-built deck's simplified and unreviewed cards", () => {
    const lib = libraryTile(saved({ coverage: LIB_COV }));
    expect(lib.flag).toBe("2 not as printed");
    expect(lib.flagTitle).toBe("2 cards in this deck do not play exactly as printed");
    expect(prebuiltTile(prebuilt({ full: 70, caveats: 2, unreviewed: 1 })).flag).toBe(
      "3 not as printed",
    );
  });

  it("says an unimplemented card out loud", () => {
    const t = prebuiltTile(prebuilt({ full: 70, caveats: 1, unregistered: 1 }));
    expect(t.flag).toBe("1 not implemented");
    expect(t.flagTitle).toBe("1 card in this deck has no implementation yet");
  });
});

describe("a tile's pips, art and wash", () => {
  it("orders the pips, and washes a pre-built deck in its colours", () => {
    const t = prebuiltTile(prebuilt({}, { colors: ["g", "U"] }));
    expect(t.colors).toEqual(["U", "G"]);
    expect(t.identity).toEqual(["U", "G"]);
    expect(t.art).toEqual([]);
    expect(prebuiltTile(prebuilt({}, { colors: undefined })).identity).toBeNull();
  });

  it("draws a saved deck's commanders that have a printing, partners side by side", () => {
    const t = libraryTile(
      saved({
        commanders: ["Tymna the Weaver", "Kraum, Ludevic's Opus"],
        commander_cards: [
          { name: "Tymna the Weaver", scryfall_id: "t1", color_identity: ["W", "B"] },
          { name: "Kraum, Ludevic's Opus", scryfall_id: "k1", color_identity: ["U", "R"] },
        ],
        color_identity: ["W", "U", "B", "R"],
      }),
    );
    expect(t.art).toEqual([
      { name: "Tymna the Weaver", scryfallId: "t1" },
      { name: "Kraum, Ludevic's Opus", scryfallId: "k1" },
    ]);
    expect(t.colors).toEqual(["W", "U", "B", "R"]);
  });

  it("gives a bot deck the pre-built deck's tile, or a plain one without it", () => {
    const bot = { id: "fury", name: "Fury", commander: "Purphoros, God of the Forge" };
    expect(botDeckTile(bot, prebuilt({ caveats: 1, full: 72 })).flag).toBe("1 not as printed");
    expect(botDeckTile(bot, prebuilt({})).name).toBe("Fury");
    const plain = botDeckTile({ ...bot, colors: ["R"] });
    expect(plain.sub).toBe("Purphoros, God of the Forge");
    expect(plain.colors).toEqual(["R"]);
    expect(plain.flag).toBe("");
  });
});

describe("filterTiles", () => {
  const tiles = [
    libraryTile(saved()),
    libraryTile(saved({ id: "d2", name: "Sami Whammy", commanders: ["Sami, Wildcat Captain"] })),
    prebuiltTile(prebuilt({})),
  ];

  it("keeps everything for an empty query", () => {
    expect(filterTiles(tiles, "  ")).toHaveLength(3);
  });

  it("matches name, commander, archetype and summary, ignoring case", () => {
    expect(filterTiles(tiles, "WHAMMY").map((t) => t.id)).toEqual(["d2"]);
    expect(filterTiles(tiles, "leyline").map((t) => t.id)).toEqual(["d1"]);
    expect(filterTiles(tiles, "aggro").map((t) => t.id)).toEqual(["fury"]);
    expect(filterTiles(tiles, "goblins").map((t) => t.id)).toEqual(["fury"]);
  });

  it("needs every word, in any order", () => {
    expect(filterTiles(tiles, "tamer inverse").map((t) => t.id)).toEqual(["d1"]);
    expect(filterTiles(tiles, "tamer sami")).toEqual([]);
  });
});

describe("libraryDetail: the line under the grid for a saved deck", () => {
  it("says a deck plays as printed in words", () => {
    expect(libraryDetail(saved({ coverage: { ...LIB_COV, as_printed_copies: 100 } }))).toEqual([
      "Nissa, Leyline Tamer",
      "100 cards",
      "every card plays as printed",
    ]);
  });

  it("gives a deck with gaps its count and what the rest need", () => {
    expect(libraryDetail(saved({ coverage: LIB_COV }))).toEqual([
      "Nissa, Leyline Tamer",
      "100 cards",
      "98 of 100 cards play as printed",
      "2 simplified",
    ]);
  });

  it("leaves out the commander a deck is named for, and coverage the server did not send", () => {
    expect(
      libraryDetail(saved({ name: "Nissa, Leyline Tamer", commanders: ["Nissa, Leyline Tamer"] })),
    ).toEqual(["100 cards"]);
  });
});

describe("the tabs (owner decision 7)", () => {
  it("offers My decks only with saved decks, and Paste a list always", () => {
    expect(pickerTabs(3, 8)).toEqual(["library", "prebuilt", "paste"]);
    expect(pickerTabs(0, 8)).toEqual(["prebuilt", "paste"]);
    expect(pickerTabs(0, 0)).toEqual(["paste"]);
  });

  it("opens on My decks, unless the last deck was pre-built or there is no library", () => {
    const both = pickerTabs(3, 8);
    expect(firstTab(both, null)).toBe("library");
    expect(firstTab(both, { kind: "library", id: "x" })).toBe("library");
    expect(firstTab(both, { kind: "prebuilt", id: "x" })).toBe("prebuilt");
    expect(firstTab(pickerTabs(0, 8), null)).toBe("prebuilt");
    expect(firstTab(pickerTabs(3, 0), { kind: "prebuilt", id: "x" })).toBe("library");
    expect(firstTab(pickerTabs(0, 0), null)).toBe("paste");
  });
});

describe("tilePreview: your seat while you browse (§1)", () => {
  it("names the commander, then the deck when it is not the commander's", () => {
    const p = tilePreview(
      libraryTile(
        saved({ commander_cards: [{ name: "Nissa", scryfall_id: "n1", color_identity: ["G"] }] }),
      ),
    );
    expect(p.text).toEqual({
      what: "Nissa, Leyline Tamer",
      second: "Inverse Nissa",
      colors: [],
      ready: false,
    });
    expect(p.art).toEqual([{ name: "Nissa", scryfallId: "n1" }]);
    const same = tilePreview(
      libraryTile(saved({ name: "Krenko, Mob Boss", commanders: ["Krenko, Mob Boss"] })),
    );
    expect(same.text.second).toBe("");
  });

  it("says a pre-built deck as a bot's seat does", () => {
    const p = tilePreview(prebuiltTile(prebuilt({})));
    expect(p.text.what).toBe("Purphoros, God of the Forge");
    expect(p.text.second).toBe("Pre-built · Forge and Fury");
    expect(p.text.colors).toEqual(["R"]);
    expect(p.name).toBe("Forge and Fury");
  });
});
