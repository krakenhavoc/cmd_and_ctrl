import { describe, expect, it } from "vitest";
import {
  bestOfSetting,
  clockText,
  deckProblem,
  gameOfText,
  lastGameText,
  mainOf,
  matchOf,
  matchOverText,
  moveCard,
  poolLines,
  poolOf,
  sameDeck,
  scoreLine,
  secondsLeft,
  sideboardingOf,
} from "./match";
import type { CardView, MatchPoolView, MatchView, SideboardingView } from "./protocol";

const seats = [
  { seat: 0, name: "Alice" },
  { seat: 1, name: "Bob" },
];

function card(id: string, name: string): CardView {
  return { instance_id: id, name, owner: "bob", controller: "bob" } as CardView;
}

const pool: MatchPoolView = {
  seat: 1,
  player_id: "bob",
  deck: [card("m1", "Shock"), card("m2", "Shock"), card("m3", "Mountain")],
  sideboard: [card("s1", "Abrade"), card("s2", "Abrade")],
  registered: ["m1", "m2", "m3"],
};

const sb: SideboardingView = {
  deadline: "2026-11-11T00:03:00Z",
  seconds_left: 180,
  ready: [],
  chooser_seat: 1,
  deck_min: 60,
  sideboard_max: 15,
  pools: [pool],
};

const match: MatchView = {
  best_of: 3,
  game: 1,
  wins: [1, 0],
  games: [{ winner_seat: 0, starting_seat: 0, turns: 7 }],
  sideboarding: sb,
};

describe("a table's match", () => {
  it("has nothing to show at a Commander table or a best-of-one", () => {
    expect(matchOf({})).toBeNull();
    expect(matchOf(null)).toBeNull();
    expect(matchOf({ match: { best_of: 1, game: 1, wins: [0, 0] } })).toBeNull();
    expect(sideboardingOf({})).toBeNull();
    expect(matchOf({ match })).toBe(match);
    expect(sideboardingOf({ match })).toBe(sb);
  });

  it("says the score, the game and the last result", () => {
    expect(scoreLine(match, seats)).toBe("Alice 1–0 Bob");
    expect(scoreLine(match, [...seats].reverse())).toBe("Alice 1–0 Bob");
    expect(gameOfText(match)).toBe("Game 1 of 3");
    expect(lastGameText(match, seats)).toBe("Alice won game 1.");
    expect(lastGameText({ ...match, games: [{ starting_seat: 0, turns: 3 }] }, seats)).toBe(
      "Game 1 was a draw.",
    );
    expect(lastGameText({ ...match, games: [] }, seats)).toBe("");
  });

  it("words a finished match", () => {
    const won: MatchView = { best_of: 3, game: 3, wins: [2, 1], over: true, winner_seat: 0 };
    expect(matchOverText(won, seats)).toBe("wins the match 2–1.");
    expect(matchOverText({ ...won, wins: [1, 0], conceded_seat: 1 }, seats)).toBe(
      "wins the match. Bob conceded it at 1–0.",
    );
    expect(matchOverText({ best_of: 3, game: 3, wins: [1, 1], over: true }, seats)).toBe(
      "The match is a draw, 1–1.",
    );
  });

  it("gives a pool to its owner alone", () => {
    expect(poolOf(sb, "bob")).toBe(pool);
    expect(poolOf(sb, "alice")).toBeNull();
    expect(poolOf(sb, null)).toBeNull();
    expect(poolOf({ ...sb, pools: undefined }, "bob")).toBeNull();
  });
});

describe("the sideboarding screen's arithmetic", () => {
  it("lists each side by name with its copies", () => {
    const main = new Set(mainOf(pool));
    expect(poolLines(pool, main, true).map((l) => `${l.ids.length} ${l.name}`)).toEqual([
      "1 Mountain",
      "2 Shock",
    ]);
    expect(poolLines(pool, main, false).map((l) => `${l.ids.length} ${l.name}`)).toEqual([
      "2 Abrade",
    ]);
    expect(poolLines(pool, main, true).every((l) => l.moved === 0)).toBe(true);
  });

  it("moves one card across and marks what changed", () => {
    let main = mainOf(pool);
    main = moveCard(main, "m1");
    main = moveCard(main, "s1");
    expect(sameDeck(main, ["m2", "m3", "s1"])).toBe(true);
    expect(sameDeck(main, mainOf(pool))).toBe(false);
    const set = new Set(main);
    const deck = poolLines(pool, set, true);
    const side = poolLines(pool, set, false);
    expect(deck.map((l) => `${l.ids.length} ${l.name} +${l.moved}`)).toEqual([
      "1 Abrade +1",
      "1 Mountain +0",
      "1 Shock +0",
    ]);
    expect(side.map((l) => `${l.ids.length} ${l.name} +${l.moved}`)).toEqual([
      "1 Abrade +0",
      "1 Shock +1",
    ]);
    // And back again is the registered deck.
    main = moveCard(moveCard(main, "s1"), "m1");
    expect(sameDeck(main, pool.registered)).toBe(true);
  });

  it("says what the deck check would refuse", () => {
    expect(deckProblem(60, 15, sb)).toBe("");
    expect(deckProblem(61, 14, sb)).toBe("");
    expect(deckProblem(59, 16, sb)).toBe("Your deck has 59 cards. Bring in 1 more to reach 60.");
    expect(deckProblem(60, 16, sb)).toContain("it holds 15");
  });

  it("counts down from the server's clock, not this device's", () => {
    expect(secondsLeft(sb, 1_000, 1_000)).toBe(180);
    expect(secondsLeft(sb, 1_000, 31_000)).toBe(150);
    expect(secondsLeft(sb, 1_000, 999_000)).toBe(0);
    expect(clockText(125)).toBe("2:05");
    expect(clockText(0)).toBe("0:00");
    expect(clockText(-4)).toBe("0:00");
  });

  it("reads a table's match length", () => {
    expect(bestOfSetting(3)).toBe(3);
    expect(bestOfSetting(undefined)).toBe(1);
    expect(bestOfSetting(1)).toBe(1);
  });
});
