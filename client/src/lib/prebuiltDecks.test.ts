import { describe, expect, it } from "vitest";
import {
  deckSubtitle,
  imperfectCardLine,
  orderedColors,
  summariseCoverage,
  type DeckCoverage,
  type PrebuiltDeck,
} from "./prebuiltDecks";

function cov(partial: Partial<DeckCoverage>): DeckCoverage {
  return {
    cards: 0,
    full: 0,
    caveats: 0,
    unreviewed: 0,
    basics: 0,
    unregistered: 0,
    ...partial,
  };
}

describe("summariseCoverage", () => {
  it("only claims 'exactly as printed' for a deck that is", () => {
    const s = summariseCoverage(cov({ cards: 88, full: 88, basics: 2 }));
    expect(s.tone).toBe("full");
    expect(s.headline).toBe("All 88 nonbasic cards play exactly as printed");
  });

  // The failure this whole feature has to avoid: a deck with twenty-two
  // simplified cards must not be advertised as fully implemented. The
  // honest claim — every card IS implemented — is still worth making,
  // and is the one that is true.
  it("never says a caveated deck plays as printed", () => {
    const s = summariseCoverage(
      cov({ cards: 88, full: 55, caveats: 22, unreviewed: 11, basics: 2 }),
    );
    expect(s.tone).toBe("caveats");
    expect(s.headline).toContain("Every card is implemented");
    expect(s.headline).toContain("55 of 88");
    expect(s.headline).not.toMatch(/^All /);
    expect(s.detail).toContain("22 simplified");
    expect(s.detail).toContain("11 not yet reviewed");
  });

  it("names only the buckets that have cards in them", () => {
    const caveatsOnly = summariseCoverage(cov({ cards: 90, full: 86, caveats: 4, basics: 1 }));
    expect(caveatsOnly.detail).toContain("4 simplified");
    expect(caveatsOnly.detail).not.toContain("not yet reviewed");

    const unreviewedOnly = summariseCoverage(
      cov({ cards: 90, full: 86, unreviewed: 4, basics: 1 }),
    );
    expect(unreviewedOnly.detail).toContain("4 not yet reviewed");
    expect(unreviewedOnly.detail).not.toContain("simplified");
  });

  // Unreachable in practice (a server-side build test fails first), and
  // rendered rather than assumed away: if it ever happens, the picker
  // says so instead of quietly counting the card as fine.
  it("leads with the gap when a card has no implementation at all", () => {
    const s = summariseCoverage(cov({ cards: 88, full: 87, unregistered: 1 }));
    expect(s.tone).toBe("gap");
    expect(s.headline).toContain("1 of 88");
    expect(s.headline).toContain("no implementation");
  });

  it("mentions basics only when the deck has some", () => {
    expect(summariseCoverage(cov({ cards: 88, full: 88, basics: 2 })).detail).toContain(
      "Basic lands",
    );
    expect(summariseCoverage(cov({ cards: 90, full: 90, basics: 0 })).detail).toBe("");
  });
});

describe("imperfectCardLine", () => {
  it("joins the declared caveats", () => {
    expect(imperfectCardLine({ name: "Farseek", caveats: ["Only basic lands are found."] })).toBe(
      "Only basic lands are found.",
    );
  });

  // An unreviewed card has no caveat text, and an empty bullet reads
  // like a bug rather than like "nobody has checked this one".
  it("says something for an unreviewed card", () => {
    const line = imperfectCardLine({ name: "Thought Vessel", unreviewed: true });
    expect(line).toContain("Not yet reviewed");
  });

  it("ignores blank caveat strings", () => {
    expect(imperfectCardLine({ name: "X", caveats: ["", "  "], unreviewed: true })).toContain(
      "Not yet reviewed",
    );
  });
});

describe("orderedColors", () => {
  it("puts identities in WUBRG order and drops duplicates", () => {
    expect(orderedColors(["R", "U", "u"])).toEqual(["U", "R"]);
  });
  it("treats colourless as no pips", () => {
    expect(orderedColors([])).toEqual([]);
    expect(orderedColors(undefined)).toEqual([]);
  });
});

describe("deckSubtitle", () => {
  const base: PrebuiltDeck = {
    id: "izzet-aggro",
    name: "Raid and Ransack",
    card_count: 100,
    coverage: cov({}),
  };

  it("joins archetype and commander", () => {
    expect(
      deckSubtitle({ ...base, archetype: "aggro", commander: "Mary Read and Anne Bonny" }),
    ).toBe("aggro · Mary Read and Anne Bonny");
  });

  // Both halves are optional on the wire; a stray separator is the
  // giveaway that nobody tried it with one missing.
  it("renders no stray separator when a half is missing", () => {
    expect(deckSubtitle({ ...base, archetype: "aggro" })).toBe("aggro");
    expect(deckSubtitle({ ...base, commander: "Tatyova" })).toBe("Tatyova");
    expect(deckSubtitle(base)).toBe("");
  });
});
