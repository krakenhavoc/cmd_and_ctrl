import { describe, expect, it } from "vitest";

import { boughtList, opponentLabel, type GameCards } from "./gameCards";

const G: GameCards = {
  ready: true,
  opponents: [
    {
      commanders: ["Atraxa"],
      cards: [
        { name: "Sol Ring", casts: 3 },
        { name: "Arcane Signet", casts: 1 },
      ],
    },
    { commanders: ["Tymna", "Atraxa"], cards: [{ name: "sol ring", casts: 2 }] },
  ],
};

describe("gameCards", () => {
  it("lists each distinct card once, commanders first", () => {
    expect(boughtList(G).map((l) => l.name)).toEqual([
      "Atraxa",
      "Tymna",
      "Sol Ring",
      "Arcane Signet",
    ]);
    expect(boughtList(G).every((l) => l.count === 1)).toBe(true);
  });

  it("is empty for a game with no opponents' cards", () => {
    expect(boughtList({ ready: false, opponents: [] })).toEqual([]);
  });

  it("tells opponents apart without naming them", () => {
    expect(opponentLabel(0)).toBe("Opponent 1");
    expect(opponentLabel(2)).toBe("Opponent 3");
  });
});
