// #5701: the order a pack is shown in.

import { describe, expect, it } from "vitest";

import type { DraftCard } from "./draft";
import { colourRank, isBasicLand, rarityRank, sortPack } from "./draftOrder";

function card(id: string, over: Partial<DraftCard> = {}): DraftCard {
  return { id, name: id, rarity: "common", colors: [], type_line: "Creature — Elf", ...over };
}
const names = (cards: DraftCard[]): string[] => cards.map((c) => c.name);

describe("the rank of a rarity", () => {
  it("runs mythic, rare, uncommon, common, other, basic land", () => {
    const ranks = [
      rarityRank(card("m", { rarity: "mythic" })),
      rarityRank(card("r", { rarity: "rare" })),
      rarityRank(card("u", { rarity: "uncommon" })),
      rarityRank(card("c", { rarity: "common" })),
      rarityRank(card("s", { rarity: "special" })),
      rarityRank(card("b", { rarity: "common", basic: true })),
    ];
    expect(ranks).toEqual([0, 1, 2, 3, 4, 5]);
  });

  it("puts a card with no rarity after the commons", () => {
    expect(rarityRank(card("x", { rarity: undefined }))).toBe(4);
  });

  it("knows a basic land by its flag or by its type line, but not a land that is not basic", () => {
    expect(isBasicLand(card("a", { basic: true, type_line: undefined }))).toBe(true);
    expect(isBasicLand(card("b", { type_line: "Basic Land — Forest" }))).toBe(true);
    expect(isBasicLand(card("c", { type_line: "Basic Snow Land — Island" }))).toBe(true);
    expect(isBasicLand(card("d", { type_line: "Land" }))).toBe(false);
    expect(isBasicLand(card("e", { type_line: "Land — Forest Plains" }))).toBe(false);
  });

  it("ranks a basic land last even when its rarity says common", () => {
    expect(rarityRank(card("f", { type_line: "Basic Land — Forest", rarity: "common" }))).toBe(5);
  });
});

describe("the rank of a colour", () => {
  it("runs WUBRG, then several colours, then none", () => {
    const rank = (colors: string[]): number => colourRank({ colors });
    expect(["W", "U", "B", "R", "G"].map((c) => rank([c]))).toEqual([0, 1, 2, 3, 4]);
    expect(rank(["W", "U"])).toBe(5);
    expect(rank(["G", "W", "B"])).toBe(5);
    expect(rank([])).toBe(6);
    expect(colourRank({ colors: undefined })).toBe(6);
  });
});

describe("sorting a pack", () => {
  it("shows mythic, rare, uncommon, common, then basic land", () => {
    const pack = [
      card("1", { name: "Forest", type_line: "Basic Land — Forest", basic: true }),
      card("2", { name: "Common", rarity: "common" }),
      card("3", { name: "Rare", rarity: "rare" }),
      card("4", { name: "Uncommon", rarity: "uncommon" }),
      card("5", { name: "Mythic", rarity: "mythic" }),
    ];
    expect(names(sortPack(pack))).toEqual(["Mythic", "Rare", "Uncommon", "Common", "Forest"]);
  });

  it("orders one rarity by colour, then name", () => {
    const pack = [
      card("1", { name: "Zed", rarity: "rare", colors: [] }),
      card("2", { name: "Bolt", rarity: "rare", colors: ["R"] }),
      card("3", { name: "Gold B", rarity: "rare", colors: ["U", "R"] }),
      card("4", { name: "Gold A", rarity: "rare", colors: ["W", "B"] }),
      card("5", { name: "Elf", rarity: "rare", colors: ["G"] }),
      card("6", { name: "Angel", rarity: "rare", colors: ["W"] }),
      card("7", { name: "Aven", rarity: "rare", colors: ["W"] }),
      card("8", { name: "Ink", rarity: "rare", colors: ["B"] }),
      card("9", { name: "Sea", rarity: "rare", colors: ["U"] }),
    ];
    expect(names(sortPack(pack))).toEqual([
      "Angel",
      "Aven",
      "Sea",
      "Ink",
      "Bolt",
      "Elf",
      "Gold A",
      "Gold B",
      "Zed",
    ]);
  });

  it("keeps a rarity above the colour: a common white card is below a rare green one", () => {
    const pack = [
      card("1", { name: "White common", colors: ["W"] }),
      card("2", { name: "Green rare", rarity: "rare", colors: ["G"] }),
    ];
    expect(names(sortPack(pack))).toEqual(["Green rare", "White common"]);
  });

  it("is stable as cards leave: the cards that stay keep their order", () => {
    const pack = Array.from({ length: 15 }, (_, i) =>
      card(`c${i}`, {
        name: `Card ${String.fromCharCode(97 + ((i * 7) % 15))}`,
        rarity: ["common", "uncommon", "rare", "mythic"][i % 4],
        colors: [["W", "U", "B", "R", "G"][i % 5]],
      }),
    );
    const first = sortPack(pack);
    let rest = pack;
    for (const gone of ["c3", "c9", "c0", "c14"]) {
      rest = rest.filter((c) => c.id !== gone);
      const stays = first.filter((c) => rest.some((r) => r.id === c.id));
      expect(sortPack(rest).map((c) => c.id)).toEqual(stays.map((c) => c.id));
    }
  });

  it("does not depend on the order the server sent", () => {
    const pack = [
      card("1", { name: "A", rarity: "rare", colors: ["W"] }),
      card("2", { name: "B", rarity: "uncommon", colors: ["U"] }),
      card("3", { name: "C", rarity: "common", colors: ["B"] }),
      card("4", { name: "D", rarity: "mythic", colors: [] }),
    ];
    expect(names(sortPack([...pack].reverse()))).toEqual(names(sortPack(pack)));
  });

  it("breaks a tie of rarity, colour and name by id, so two copies keep their order", () => {
    const pack = [card("c9", { name: "Same" }), card("c10", { name: "Same" })];
    expect(sortPack(pack).map((c) => c.id)).toEqual(["c10", "c9"]);
    expect(sortPack([...pack].reverse()).map((c) => c.id)).toEqual(["c10", "c9"]);
  });

  it("orders a cube pack by each card's printed rarity, in the same way", () => {
    const pack = [
      card("1", { name: "Chump", rarity: "common", colors: ["G"] }),
      card("2", { name: "Bomb", rarity: "mythic", colors: ["R"] }),
      card("3", { name: "Value", rarity: "rare", colors: ["W"] }),
    ];
    expect(names(sortPack(pack))).toEqual(["Bomb", "Value", "Chump"]);
  });

  it("does not change the list it is given", () => {
    const pack = [card("1", { rarity: "common" }), card("2", { rarity: "mythic" })];
    const before = pack.map((c) => c.id);
    sortPack(pack);
    expect(pack.map((c) => c.id)).toEqual(before);
  });
});
