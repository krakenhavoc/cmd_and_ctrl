import { describe, expect, it } from "vitest";
import { filterRows, libraryRow, prebuiltRow } from "./deckPicker";
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

describe("prebuiltRow", () => {
  it("says 'all as printed' only for a deck that is", () => {
    const r = prebuiltRow(prebuilt({}));
    expect(r.cov).toBe("all 73 as printed");
    expect(r.tone).toBe("full");
  });

  it("gives a caveated deck its as-printed count, not a claim", () => {
    const r = prebuiltRow(prebuilt({ full: 72, caveats: 1 }));
    expect(r.cov).toBe("72/73 as printed");
    expect(r.tone).toBe("caveats");
    const u = prebuiltRow(prebuilt({ full: 70, unreviewed: 3 }));
    expect(u.tone).toBe("caveats");
  });

  it("says an unimplemented card out loud", () => {
    const r = prebuiltRow(prebuilt({ full: 70, unregistered: 3 }));
    expect(r.cov).toBe("3 not implemented");
    expect(r.tone).toBe("gap");
  });

  it("orders the pips and builds the subtitle", () => {
    const r = prebuiltRow(prebuilt({}, { colors: ["g", "U"] }));
    expect(r.colors).toEqual(["U", "G"]);
    expect(r.sub).toBe("aggro · Purphoros, God of the Forge");
  });
});

describe("libraryRow", () => {
  it("counts copies as printed, and is green only at the whole deck", () => {
    const r = libraryRow(saved({ coverage: LIB_COV, color_identity: ["G"] }));
    expect(r.cov).toBe("98/100 as printed");
    expect(r.tone).toBe("caveats");
    expect(r.colors).toEqual(["G"]);
    expect(r.sub).toBe("Nissa, Leyline Tamer · 100 cards");
    const full = libraryRow(saved({ coverage: { ...LIB_COV, as_printed_copies: 100 } }));
    expect(full.tone).toBe("full");
  });

  it("draws no figure without coverage", () => {
    const r = libraryRow(saved());
    expect(r.cov).toBe("");
    expect(r.tone).toBe("none");
    expect(r.colors).toEqual([]);
  });
});

describe("filterRows", () => {
  const rows = [
    libraryRow(saved()),
    libraryRow(saved({ id: "d2", name: "Sami Whammy", commanders: ["Sami, Wildcat Captain"] })),
    prebuiltRow(prebuilt({})),
  ];

  it("keeps everything for an empty query", () => {
    expect(filterRows(rows, "  ")).toHaveLength(3);
  });

  it("matches name, commander, archetype and summary, ignoring case", () => {
    expect(filterRows(rows, "WHAMMY").map((r) => r.id)).toEqual(["d2"]);
    expect(filterRows(rows, "leyline").map((r) => r.id)).toEqual(["d1"]);
    expect(filterRows(rows, "aggro").map((r) => r.id)).toEqual(["fury"]);
    expect(filterRows(rows, "goblins").map((r) => r.id)).toEqual(["fury"]);
  });

  it("needs every word, in any order", () => {
    expect(filterRows(rows, "tamer inverse").map((r) => r.id)).toEqual(["d1"]);
    expect(filterRows(rows, "tamer sami")).toEqual([]);
  });
});
