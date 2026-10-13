import { describe, expect, it } from "vitest";
import {
  auraColors,
  auraGradient,
  deckArt,
  deckIdentity,
  deckNotes,
  deckRing,
  sortDecks,
} from "./deckGrid";
import type { MyDeckInfo } from "./myDecks";

const ZERO = { manual: 0, unreviewed: 0, caveats: 0, automated: 0, no_effect: 0 };

function deck(over: Partial<MyDeckInfo> = {}, printed = 98, size = 100): MyDeckInfo {
  return {
    id: over.name ?? "d",
    name: "Deck",
    commanders: [],
    card_count: size,
    updated_at: "2026-10-01T00:00:00Z",
    coverage: {
      counts: ZERO,
      unknown: 0,
      as_printed: 0,
      resolved: 0,
      copies: { ...ZERO, caveats: size - printed },
      unknown_copies: 0,
      as_printed_copies: printed,
      deck_size: size,
    },
    ...over,
  };
}

describe("deckRing", () => {
  it("is green at 96% and up, amber below", () => {
    expect(deckRing(deck({}, 96, 100))).toMatchObject({ n: 96, total: 100, tone: "good" });
    expect(deckRing(deck({}, 95, 100))?.tone).toBe("warn");
    expect(deckRing(deck({}, 100, 100))?.share).toBe(1);
  });

  it("is null without coverage", () => {
    expect(deckRing(deck({ coverage: undefined }))).toBeNull();
  });
});

describe("sortDecks", () => {
  const a = deck({ name: "beta" }, 90);
  const b = deck({ name: "Alpha" }, 99);
  const c = deck({ name: "gamma", coverage: undefined });
  it("keeps the server's order for Recent", () => {
    expect(sortDecks([a, b, c], "recent").map((d) => d.name)).toEqual(["beta", "Alpha", "gamma"]);
  });
  it("sorts by name, ignoring case", () => {
    expect(sortDecks([a, c, b], "name").map((d) => d.name)).toEqual(["Alpha", "beta", "gamma"]);
  });
  it("puts the best covered first and an ungraded deck last", () => {
    expect(sortDecks([c, a, b], "coverage").map((d) => d.name)).toEqual(["Alpha", "beta", "gamma"]);
  });
});

describe("the art and the aura", () => {
  it("draws only the commanders with a printing", () => {
    const d = deck({
      commander_cards: [
        { name: "A", scryfall_id: "id-a", color_identity: ["W"] },
        { name: "B", color_identity: null },
      ],
    });
    expect(deckArt(d)).toEqual([{ name: "A", scryfallId: "id-a" }]);
    expect(deckArt(deck())).toEqual([]);
  });

  it("tells unknown from colourless", () => {
    expect(deckIdentity(deck())).toBeNull();
    expect(
      deckIdentity(deck({ commander_cards: [{ name: "x", color_identity: null }] })),
    ).toBeNull();
    expect(deckIdentity(deck({ commander_cards: [{ name: "Karn", color_identity: [] }] }))).toEqual(
      [],
    );
    expect(
      deckIdentity(
        deck({
          commander_cards: [{ name: "x", color_identity: ["G", "W"] }],
          color_identity: ["G", "W"],
        }),
      ),
    ).toEqual(["W", "G"]);
  });

  it("maps each colour to its fixed token, colourless to steel, unknown to none", () => {
    expect(auraColors(null)).toEqual([]);
    expect(auraColors([])).toEqual(["var(--identity-c)"]);
    expect(auraColors(["U", "R"])).toEqual(["var(--identity-u)", "var(--identity-r)"]);
    expect(auraGradient(null)).toBe("");
    expect(auraGradient(["G"])).toBe(
      "linear-gradient(120deg, var(--identity-g), var(--identity-g))",
    );
  });
});

describe("deckNotes", () => {
  it("says what is simplified, or that every card plays as printed", () => {
    expect(deckNotes(deck({}, 98))).toBe("2 simplified");
    expect(deckNotes(deck({}, 100))).toBe("every card plays as printed");
    expect(deckNotes(deck({ coverage: undefined }))).toBe("");
  });
});
