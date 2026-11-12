// #5440 (ADR 0222 §5): the rules of the deckbuilding screen.

import { describe, expect, it } from "vitest";

import type { DraftCard } from "./draft";
import {
  applySuggestion,
  basicsCount,
  buildProblem,
  buildSecondsLeft,
  curve,
  deckSize,
  initialBuild,
  landCount,
  landSuggestion,
  matchesSuggestion,
  MAX_BASICS,
  moveCard,
  sameDeck,
  setBasic,
  splitPool,
  toggleCard,
} from "./draftBuild";

function card(id: string, over: Partial<DraftCard> = {}): DraftCard {
  return {
    id,
    name: id,
    colors: ["G"],
    mana_cost: "{1}{G}",
    mana_value: 2,
    type_line: "Creature",
    ...over,
  };
}
const many = (n: number, prefix: string, over: Partial<DraftCard> = {}): DraftCard[] =>
  Array.from({ length: n }, (_, i) => card(`${prefix}${i}`, over));

describe("the deck being built", () => {
  const pool = [card("a"), card("b"), card("c")];

  it("opens empty, or on the deck the server holds, less what the pool does not have", () => {
    expect(initialBuild(pool, undefined)).toEqual({ main: [], basics: {} });
    expect(
      initialBuild(pool, {
        main: ["a", "zz", "a", "c"],
        basics: { Forest: 9, Wastes: 3, Island: 0 },
      }),
    ).toEqual({ main: ["a", "c"], basics: { Forest: 9 } });
  });

  it("moves a card to the deck and back, once", () => {
    let s = initialBuild(pool, undefined);
    s = toggleCard(s, "a");
    s = moveCard(s, "a", "deck");
    expect(s.main).toEqual(["a"]);
    s = moveCard(s, "b", "deck");
    expect(splitPool(pool, s)).toEqual({ deck: [pool[0], pool[1]], side: [pool[2]] });
    s = toggleCard(s, "a");
    s = moveCard(s, "a", "side");
    expect(s.main).toEqual(["b"]);
  });

  it("counts basics into the deck and holds a stepper to its range", () => {
    let s = initialBuild(pool, { main: ["a"] });
    s = setBasic(s, "Forest", 9);
    s = setBasic(s, "Island", 8);
    expect(basicsCount(s)).toBe(17);
    expect(deckSize(s)).toBe(18);
    expect(setBasic(s, "Forest", -3).basics).toEqual({ Island: 8 });
    expect(setBasic(s, "Forest", 1000).basics.Forest).toBe(MAX_BASICS);
    expect(setBasic(s, "Forest", Number.NaN).basics.Forest).toBeUndefined();
  });

  it("says how many cards short a deck is, and nothing once it is legal", () => {
    expect(buildProblem(39, 40)).toBe("Your deck has 39 cards. Add 1 more to reach 40.");
    expect(buildProblem(1, 40)).toBe("Your deck has 1 card. Add 39 more to reach 40.");
    expect(buildProblem(40, 40)).toBe("");
    expect(buildProblem(47, 40)).toBe("");
  });

  it("knows a deck from the one the server holds", () => {
    const s = setBasic(initialBuild(pool, { main: ["a", "b"] }), "Forest", 2);
    expect(sameDeck(s, { main: ["b", "a"], basics: { Forest: 2 } })).toBe(true);
    expect(sameDeck(s, { main: ["a", "b"], basics: { Forest: 3 } })).toBe(false);
    expect(sameDeck(s, { main: ["a", "c"], basics: { Forest: 2 } })).toBe(false);
    expect(sameDeck(s, undefined)).toBe(false);
  });
});

describe("the curve and the land count", () => {
  it("counts the cards that are not lands by mana value, 6 and more together", () => {
    const deck = [
      card("a", { mana_value: 0 }),
      card("b", { mana_value: 2 }),
      card("c", { mana_value: 2 }),
      card("d", { mana_value: 6 }),
      card("e", { mana_value: 9 }),
      card("l", { type_line: "Land", mana_value: 0, mana_cost: undefined }),
    ];
    expect(curve(deck).map((c) => `${c.label}:${c.count}`)).toEqual([
      "0:1",
      "1:0",
      "2:2",
      "3:0",
      "4:0",
      "5:0",
      "6+:2",
    ]);
  });

  it("counts the pool's own lands with the basics added", () => {
    const deck = [card("a"), card("l", { type_line: "Land — Gate", mana_cost: undefined })];
    expect(landCount(deck, { main: ["a", "l"], basics: { Forest: 16 } })).toBe(17);
  });
});

describe("the land-count suggestion", () => {
  it("is 17 lands for 23 other cards, split by the deck's colours", () => {
    const deck = [...many(14, "g"), ...many(9, "r", { colors: ["R"], mana_cost: "{2}{R}" })];
    const s = landSuggestion(deck, 40);
    expect(s.lands).toBe(17);
    expect(s.add).toBe(17);
    expect(s.basics).toEqual({ Mountain: 7, Forest: 10 });
  });

  it("asks for enough lands to reach the minimum deck", () => {
    expect(landSuggestion(many(20, "g"), 40).lands).toBe(20);
    expect(landSuggestion(many(26, "g"), 40).lands).toBe(19);
  });

  it("takes off the lands the pool already gave", () => {
    const deck = [...many(23, "g"), ...many(2, "l", { type_line: "Land", mana_cost: undefined })];
    const s = landSuggestion(deck, 40);
    expect(s.lands).toBe(17);
    expect(s.add).toBe(15);
    expect(s.basics).toEqual({ Forest: 15 });
  });

  it("gives every colour the deck plays at least one land, and counts a hybrid for both", () => {
    const deck = [...many(22, "g"), card("u", { colors: ["U"], mana_cost: "{U}" })];
    const s = landSuggestion(deck, 40);
    expect(s.basics.Island).toBeGreaterThanOrEqual(1);
    expect(s.basics.Island + s.basics.Forest).toBe(17);
    const hybrid = landSuggestion(
      many(23, "h", { colors: ["W", "B"], mana_cost: "{W/B}{W/B}" }),
      40,
    );
    expect(hybrid.basics).toEqual({ Plains: 9, Swamp: 8 });
  });

  it("has no split for a deck with no colour, and nothing to say for an empty one", () => {
    const grey = landSuggestion(many(23, "a", { colors: [], mana_cost: "{3}" }), 40);
    expect(grey.lands).toBe(17);
    expect(grey.basics).toEqual({});
    expect(landSuggestion([], 40)).toEqual({ lands: 0, basics: {}, add: 0 });
  });

  it("can be applied, and then matches", () => {
    const s = landSuggestion(many(23, "g"), 40);
    let state = initialBuild([], undefined);
    expect(matchesSuggestion(state, s)).toBe(false);
    state = applySuggestion(state, s);
    expect(state.basics).toEqual({ Forest: 17 });
    expect(matchesSuggestion(state, s)).toBe(true);
  });
});

describe("the build timer", () => {
  const build = {
    deadline: "2026-11-11T19:15:00Z",
    now: "2026-11-11T19:00:00Z",
    seconds: 900,
    deck_min: 40,
    submitted: false,
  };
  it("counts from the server's clock, less the time passed here", () => {
    expect(buildSecondsLeft(build, 0)).toBe(900);
    expect(buildSecondsLeft(build, 30_500)).toBe(870);
    expect(buildSecondsLeft(build, 2_000_000)).toBe(0);
    expect(buildSecondsLeft(undefined, 0)).toBeNull();
    expect(buildSecondsLeft({ ...build, deadline: "soon" }, 0)).toBeNull();
  });

  it("has nothing to count when the build is untimed, with or without a stray deadline", () => {
    const untimed = { now: build.now, seconds: 0, untimed: true, deck_min: 40, submitted: false };
    expect(buildSecondsLeft(untimed, 0)).toBeNull();
    expect(buildSecondsLeft({ ...untimed, deadline: build.deadline }, 0)).toBeNull();
    // An older server's frame has a deadline and no flag, and counts as ever.
    expect(buildSecondsLeft({ ...build, untimed: undefined }, 0)).toBe(900);
  });
});
