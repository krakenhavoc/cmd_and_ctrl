import { describe, expect, it } from "vitest";

import type { DraftCard } from "./draft";
import {
  cardRoles,
  colourStats,
  costPips,
  creatureStats,
  creatureTypes,
  curveStats,
  deckStats,
  namesOf,
  splitBasics,
  typeLineOf,
  typeLines,
  typeStats,
} from "./draftDeckStats";

let n = 0;
function c(name: string, over: Partial<DraftCard> = {}): DraftCard {
  return {
    id: `id${++n}`,
    name,
    colors: [],
    mana_value: 0,
    type_line: "Creature — Bear",
    ...over,
  };
}
const names = (cards: DraftCard[]): string[] => cards.map((x) => x.name);

describe("pips and the colours", () => {
  it("counts each coloured symbol of a cost, hybrid for both colours", () => {
    expect(costPips("{1}{W}{W}")).toEqual({ W: 2, U: 0, B: 0, R: 0, G: 0 });
    expect(costPips("{W/U}{2/G}{B/P}")).toEqual({ W: 1, U: 1, B: 1, R: 0, G: 1 });
    expect(costPips(undefined)).toEqual({ W: 0, U: 0, B: 0, R: 0, G: 0 });
  });

  it("tallies pips and cards for the colours of Main, most pips first", () => {
    const s = colourStats([
      c("A", { mana_cost: "{1}{B}{B}", colors: ["B"] }),
      c("B", { mana_cost: "{W}", colors: ["W"] }),
      c("C", { mana_cost: "{W}{B}", colors: ["W", "B"] }),
    ]);
    expect(s.rows).toEqual([
      { colour: "B", pips: 3, cards: 2 },
      { colour: "W", pips: 2, cards: 2 },
    ]);
    expect(s.totalPips).toBe(5);
  });

  it("splits 17 basics by the largest remainder", () => {
    // 3 : 2 of 17 is 10.2 and 6.8: 10 and 6 whole, the spare land to the bigger fraction.
    expect(splitBasics({ B: 3, W: 2 })).toEqual([
      ["B", 10],
      ["W", 7],
    ]);
    // Even pips, an odd land: the first colour in WUBRG order wins the tie.
    expect(splitBasics({ U: 1, R: 1 })).toEqual([
      ["U", 9],
      ["R", 8],
    ]);
    expect(splitBasics({ W: 1, U: 1, B: 1 }).reduce((t, [, k]) => t + k, 0)).toBe(17);
    expect(splitBasics({ G: 4 })).toEqual([["G", 17]]);
    expect(splitBasics({})).toEqual([]);
  });

  it("says the split in the order of the colours' pips", () => {
    const s = colourStats([
      c("A", { mana_cost: "{B}{B}{B}{B}{B}{B}{B}{B}{B}", colors: ["B"] }),
      c("B", { mana_cost: "{W}{W}{W}{W}{W}{W}{W}{W}", colors: ["W"] }),
    ]);
    expect(s.basics).toEqual([
      ["B", 9],
      ["W", 8],
    ]);
  });
});

describe("the curve", () => {
  it("leaves lands out and averages the mana value of the spells", () => {
    const s = curveStats([
      c("One", { mana_value: 1 }),
      c("Three", { mana_value: 3 }),
      c("Seven", { mana_value: 7 }),
      c("Forest", { type_line: "Basic Land — Forest", basic: true }),
    ]);
    expect(s.spells).toBe(3);
    expect(s.average).toBeCloseTo(11 / 3);
    expect(s.columns.map((x) => x.count)).toEqual([0, 1, 0, 1, 0, 0, 1]);
    expect(curveStats([]).average).toBe(0);
  });
});

describe("card types", () => {
  it("counts a card with two types under both, and lists the rest as other spells", () => {
    const s = typeStats([
      c("Golem", { type_line: "Artifact Creature — Golem" }),
      c("Bolt", { type_line: "Instant" }),
      c("Aura", { type_line: "Enchantment — Aura" }),
      c("Tower", { type_line: "Land" }),
    ]);
    const of = (t: string): string[] => names(s.rows.find((r) => r.type === t)!.cards);
    expect(of("creature")).toEqual(["Golem"]);
    expect(of("artifact")).toEqual(["Golem"]);
    expect(of("instant")).toEqual(["Bolt"]);
    expect(of("land")).toEqual(["Tower"]);
    expect(names(s.otherSpells)).toEqual(["Bolt", "Aura"]);
    expect(s.creatures).toBe(1);
  });

  it("reads a type line's front face and its subtypes", () => {
    expect(typeLineOf({ type_line: "Artifact Creature — Golem" })).toBe("Artifact Creature");
    expect(typeLineOf({ type_line: "Enchantment — Aura" })).toBe("Enchantment (Aura)");
    expect(typeLineOf({ type_line: "Artifact — Equipment" })).toBe("Artifact (Equipment)");
    expect(typeLineOf({ type_line: "Instant" })).toBe("Instant");
    expect(typeLineOf({ type_line: "Creature — Elf // Sorcery" })).toBe("Creature");
  });

  it("lists the type lines most common first, with their cards", () => {
    const lines = typeLines([
      c("A", { type_line: "Creature — Elf" }),
      c("B", { type_line: "Creature — Elf Druid" }),
      c("C", { type_line: "Enchantment — Aura" }),
      c("D", { type_line: "Artifact Creature — Golem" }),
    ]);
    expect(lines.map((r) => [r.key, r.cards.length])).toEqual([
      ["Creature", 2],
      ["Artifact Creature", 1],
      ["Enchantment (Aura)", 1],
    ]);
  });
});

describe("creatures", () => {
  const deck = [
    c("Small", { power: "2", toughness: "1" }),
    c("Medium", { power: "3", toughness: "3" }),
    c("Big", { power: "6", toughness: "5", keywords: ["Trample"] }),
    c("Wild", { power: "*", toughness: "*" }),
    c("Bird", { power: "1", toughness: "1", keywords: ["Flying"] }),
    c("Ghost", { power: "2", toughness: "2", oracle_text: "Ghost can't be blocked." }),
    c("Wall", { power: "0", toughness: "4", oracle_text: "Wall can't block creatures." }),
    c("Bolt", { type_line: "Instant" }),
  ];

  it("buckets by power and averages only the printed numbers", () => {
    const s = creatureStats(deck);
    expect(names(s.creatures)).toHaveLength(7);
    expect(names(s.small)).toEqual(["Small", "Bird", "Ghost", "Wall"]);
    expect(names(s.medium)).toEqual(["Medium"]);
    expect(names(s.big)).toEqual(["Big"]);
    // (2+3+6+1+2+0) / 6 and (1+3+5+1+2+4) / 6: Wild's * is left out.
    expect(s.avgPower).toBeCloseTo(14 / 6);
    expect(s.avgToughness).toBeCloseTo(16 / 6);
  });

  it("calls flying, trample, menace and unblockable evasive, and not 'can't block'", () => {
    expect(names(creatureStats(deck).evasive)).toEqual(["Big", "Bird", "Ghost"]);
  });

  it("has no average without a sized creature", () => {
    expect(creatureStats([c("Bolt", { type_line: "Instant" })]).avgPower).toBeNull();
  });

  it("lists the eight most common creature types", () => {
    const many = [
      ...["Elf", "Elf", "Elf", "Goblin", "Goblin", "Wizard"].map((t) =>
        c(`${t} guy`, { type_line: `Creature — ${t}` }),
      ),
      ...["A", "B", "C", "D", "E", "F"].map((t) => c(t, { type_line: `Creature — ${t}` })),
    ];
    const rows = creatureTypes(many);
    expect(rows).toHaveLength(8);
    expect(rows.slice(0, 2).map((r) => [r.key, r.cards.length])).toEqual([
      ["Elf", 3],
      ["Goblin", 2],
    ]);
    // A two-type creature counts under each of its types.
    const both = creatureTypes([c("Druid", { type_line: "Creature — Elf Druid" })]);
    expect(both.map((r) => r.key)).toEqual(["Druid", "Elf"]);
  });
});

describe("what the cards do", () => {
  const roles = (over: Partial<DraftCard>): string[] => cardRoles(c("X", over));

  it("reads removal from the rules text", () => {
    expect(
      roles({ type_line: "Instant", oracle_text: "Destroy target nonblack creature." }),
    ).toContain("removal");
    expect(
      roles({ type_line: "Instant", oracle_text: "Lightning Bolt deals 3 damage to any target." }),
    ).toContain("removal");
    expect(
      roles({ type_line: "Sorcery", oracle_text: "Exile target creature or planeswalker." }),
    ).toContain("removal");
    expect(
      roles({
        type_line: "Enchantment — Aura",
        oracle_text: "Enchanted creature can't attack or block.",
      }),
    ).toContain("removal");
    // Not removal: a creature that destroys something else, a pump spell.
    expect(roles({ oracle_text: "Destroy target artifact." })).not.toContain("removal");
    expect(
      roles({ type_line: "Instant", oracle_text: "Target creature gets +3/+3 until end of turn." }),
    ).not.toContain("removal");
  });

  it("reads fliers from the keyword, and only for creatures", () => {
    expect(roles({ keywords: ["Flying"] })).toContain("fliers");
    expect(roles({ type_line: "Enchantment", keywords: ["Flying"] })).not.toContain("fliers");
    expect(roles({ oracle_text: "Flying" })).not.toContain("fliers");
  });

  it("reads draw, tricks, counters, bounce, tokens, ramp, life gain, discard and answers", () => {
    expect(roles({ type_line: "Sorcery", oracle_text: "Draw two cards." })).toContain("draw");
    expect(
      roles({ type_line: "Instant", oracle_text: "Target creature gets +3/+3 until end of turn." }),
    ).toContain("tricks");
    expect(
      roles({ type_line: "Sorcery", oracle_text: "Target creature gets +3/+3 until end of turn." }),
    ).not.toContain("tricks");
    expect(roles({ type_line: "Instant", oracle_text: "Counter target spell." })).toContain(
      "counters",
    );
    expect(
      roles({
        type_line: "Instant",
        oracle_text: "Return target creature to its owner's hand.",
      }),
    ).toContain("bounce");
    expect(
      roles({
        type_line: "Instant",
        oracle_text: "Return target creature card from your graveyard to your hand.",
      }),
    ).not.toContain("bounce");
    expect(
      roles({ type_line: "Instant", oracle_text: "Create two 1/1 white Soldier creature tokens." }),
    ).toContain("tokens");
    expect(roles({ type_line: "Artifact", oracle_text: "{T}: Add {C}." })).toContain("ramp");
    expect(roles({ type_line: "Land", oracle_text: "{T}: Add {G}." })).toEqual([]);
    expect(roles({ keywords: ["Lifelink"] })).toContain("lifegain");
    expect(roles({ type_line: "Instant", oracle_text: "You gain 3 life." })).toContain("lifegain");
    expect(
      roles({ type_line: "Sorcery", oracle_text: "Target player discards two cards." }),
    ).toContain("discard");
    expect(roles({ oracle_text: "{1}, Discard a card: Draw a card." })).not.toContain("discard");
    expect(
      roles({ type_line: "Instant", oracle_text: "Destroy target artifact or enchantment." }),
    ).toContain("answers");
  });
});

describe("the whole overlay", () => {
  it("fills every section, lists names with a repeat as ×2, and keeps zero rows", () => {
    const main = [
      c("Angel", {
        type_line: "Creature — Angel",
        keywords: ["Flying"],
        colors: ["W"],
        mana_value: 5,
        power: "4",
        toughness: "4",
      }),
      c("Doom Blade", {
        type_line: "Instant",
        colors: ["B"],
        mana_value: 2,
        oracle_text: "Destroy target nonblack creature.",
      }),
      c("Doom Blade", {
        type_line: "Instant",
        colors: ["B"],
        mana_value: 2,
        oracle_text: "Destroy target nonblack creature.",
      }),
      c("Helix", { type_line: "Instant", colors: ["R", "W"], mana_value: 2 }),
      c("Rock", { type_line: "Artifact", mana_value: 2 }),
    ];
    const s = deckStats(main);
    expect(s.total).toBe(5);
    expect(names(s.instant)).toEqual(["Doom Blade", "Doom Blade", "Helix"]);
    expect(names(s.sorcery)).toEqual(["Angel", "Rock"]);
    expect(names(s.multicolour)).toEqual(["Helix"]);
    expect(names(s.colourless)).toEqual(["Rock"]);
    expect(s.roles.find((r) => r.role === "removal")!.cards).toHaveLength(2);
    expect(s.roles.find((r) => r.role === "counters")!.cards).toHaveLength(0);
    expect(namesOf(s.instant)).toEqual(["Doom Blade ×2", "Helix"]);
    expect(s.roles).toHaveLength(11);
  });
});
