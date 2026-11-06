import { describe, expect, it } from "vitest";

import {
  aggressionCaption,
  coloursCaption,
  commanderCaption,
  compositionCaption,
  deckStatsHref,
  filterAnnouncement,
  filterQuery,
  kindText,
  landCurveCaption,
  landNumberText,
  missedCaption,
  NOT_ENOUGH,
  parseMakeup,
  parseRange,
  placesCaption,
  rateText,
  SHARE_REVEALS,
  shareURL,
  sortedKinds,
  spreadCaption,
  turnOrderCaption,
  versionText,
  winKindsCaption,
  type AdvancedSummary,
} from "./advancedStats";
import { L } from "./labels";

describe("advancedStats words", () => {
  it("shows a rate with its count, and none on fewer than 3 games", () => {
    expect(rateText({ n: 5, of: 13, rate: 5 / 13 })).toBe("38%, 5 of 13 games");
    expect(rateText({ n: 1, of: 2, rate: null })).toBe(NOT_ENOUGH);
    expect(rateText(null)).toBe(NOT_ENOUGH);
  });

  it("sorts win kinds by games and names them for a player", () => {
    const k = sortedKinds([
      { kind: "poison", games: 1 },
      { kind: "combat", games: 4 },
      { kind: "combo", games: 1 },
      { kind: "decking", games: 0 },
    ]);
    expect(k.map((x) => x.kind)).toEqual(["combat", "combo", "poison"]);
    expect(kindText("commander_damage")).toBe("Commander damage");
    expect(winKindsCaption(k)).toBe("Your most common win is combat damage: 4 of 6 wins.");
    expect(winKindsCaption([])).toMatch(/No wins/);
  });

  it("captions places at the most-played table size", () => {
    expect(placesCaption([{ players: 4, games: 16, places: [5, 4, 4, 3] }])).toBe(
      "At four-seat tables you finish 1st in 31% of 16 games.",
    );
    expect(placesCaption([{ players: 4, games: 2, places: [1, 1, 0, 0] }])).toContain(
      NOT_ENOUGH.toLowerCase(),
    );
  });

  it("names the best turn position only on enough games", () => {
    expect(
      turnOrderCaption([
        { position: 1, wins: { n: 1, of: 4, rate: 0.25 }, avg_place: 2 },
        { position: 3, wins: { n: 3, of: 4, rate: 0.75 }, avg_place: 1.5 },
      ]),
    ).toBe("You win most when you act 3rd: 75%, 3 of 4 games.");
    expect(
      turnOrderCaption([{ position: 1, wins: { n: 1, of: 2, rate: null }, avg_place: 2 }]),
    ).toContain(NOT_ENOUGH.toLowerCase());
  });

  const aggression: AdvancedSummary["aggression"] = {
    own_turns: 30,
    attacks_per_turn: 2.34,
    damage_per_turn: 4,
    attack_share: 0.6,
    most_hit_share: 0.52,
    evenness: 0.7,
    spread_games: 5,
    over_turns: [],
  };

  it("captions aggression and the spread, never naming an opponent", () => {
    expect(aggressionCaption(aggression)).toBe(
      "You attack with 2.3 creatures a turn on average, and 60% of the creatures that could attack did.",
    );
    expect(spreadCaption(aggression)).toBe(
      "In a typical game, 52% of your damage goes at one opponent.",
    );
    expect(spreadCaption({ ...aggression, spread_games: 2 })).toContain(NOT_ENOUGH.toLowerCase());
  });

  it("captions missed land drops by land number", () => {
    const mana = {
      land_drops_hit: 20,
      land_drops_missed: 4,
      missed_by_land: [
        { land: 3, missed: 1 },
        { land: 4, missed: 3 },
        { land: 6, missed: 0 },
      ],
      screw: { n: 0, of: 0, rate: null },
      flood: { n: 0, of: 0, rate: null },
      colour_screw: { n: 0, of: 0, rate: null },
      trend: [],
      land_curve: [],
    };
    expect(missedCaption(mana)).toBe("You miss your 4th land drop most often: 3 times.");
    expect(landNumberText(6)).toBe("6th or later");
    expect(missedCaption({ ...mana, land_drops_missed: 0, missed_by_land: [] })).toBe(
      "You hit every land drop: 20 of 20.",
    );
  });

  it("captions colours by their share of games", () => {
    expect(
      coloursCaption([
        { colour: "U", games: { n: 3, of: 6, rate: 0.5 } },
        { colour: "G", games: { n: 5, of: 6, rate: 5 / 6 } },
      ]),
    ).toBe("Green is in the most of your decks played: 83%, 5 of 6 games.");
  });

  it("reads and writes the filters", () => {
    expect(parseMakeup("with_bots")).toBe("with_bots");
    expect(parseMakeup("everyone")).toBe("all");
    expect(parseRange("30d")).toBe("30d");
    expect(parseRange(null)).toBe("all");
    expect(filterQuery("people", "90d")).toBe("?makeup=people&range=90d");
    expect(filterAnnouncement("bots", "30d")).toBe(
      `Showing ${L.statsMakeupBots.toLowerCase()}, ${L.statsRange30.toLowerCase()}.`,
    );
  });

  it("builds the share link and says what it reveals in one sentence", () => {
    expect(shareURL("https://cmd.labxp.io", "abc")).toBe("https://cmd.labxp.io/#/stats/shared/abc");
    expect(SHARE_REVEALS.split(". ").length).toBe(1);
    expect(SHARE_REVEALS).toMatch(/deck names, commanders/);
    expect(SHARE_REVEALS).toMatch(/never shows the other players/);
  });
});

describe("the deck page's words (ADR 0217 §4, §5 and §8)", () => {
  const rate = (n: number, of: number) => ({ n, of, rate: of >= 3 ? n / of : null });

  it("links a deck page, inside a shared view or not", () => {
    expect(deckStatsHref("d 1")).toBe("#/stats/deck/d%201");
    expect(deckStatsHref("d1", "tok_-1")).toBe("#/stats/shared/tok_-1/deck/d1");
  });

  it("says how often the 4th land drop was hit, or asks for games", () => {
    const pts = [1, 2, 3, 4, 5, 6, 7].map((n) => ({
      n,
      hit: rate(10 - n, 10),
      controlled: rate(10, 10),
    }));
    expect(landCurveCaption(pts)).toBe("You had hit land drop 4 by turn 4 in 60%, 6 of 10 games.");
    expect(landCurveCaption(pts.map((p) => ({ ...p, hit: rate(1, 2) })))).toBe(
      "Play 3 games with this deck to see its land curve.",
    );
    expect(landCurveCaption([])).toBe("Play 3 games with this deck to see its land curve.");
  });

  it("names a version by its list and dates", () => {
    const day = Date.UTC(2026, 10, 3, 12);
    expect(versionText({ hash: "a", current: true, first: day, last: day, games: 1 })).toMatch(
      /^Current list, .+ \(1 game\)$/,
    );
    const text = versionText({
      hash: "b",
      current: false,
      first: day,
      last: day + 9 * 864e5,
      games: 5,
    });
    expect(text).toMatch(/^Earlier list, .+ to .+ \(5 games\)$/);
  });

  it("captions the commander and the composition", () => {
    expect(
      commanderCaption({
        casts_per_game: 1.5,
        first_cast_turn: { median: 4, p25: 3, p75: 5 },
        by_turn_5: rate(3, 4),
        later: rate(0, 3),
      }),
    ).toBe("You cast your commander 1.5 times a game, first on turn 4 at the median.");
    expect(
      commanderCaption({
        casts_per_game: null,
        first_cast_turn: { median: null, p25: null, p75: null },
        by_turn_5: rate(0, 0),
        later: rate(0, 0),
      }),
    ).toBe("No commander casts recorded yet.");
    expect(
      compositionCaption([
        { label: "Creatures", value: 30, tick: 24 },
        { label: "Lands", value: 36, tick: 36.5 },
      ]),
    ).toBe(
      "Creatures stand out: 30 here against 24.0 in your other decks on average, more than usual.",
    );
    expect(compositionCaption([{ label: "Lands", value: 36 }])).toBe("How this deck is built.");
    expect(compositionCaption([{ label: "Lands", value: 36, tick: 35.5 }])).toBe(
      "This deck is built much like your other decks.",
    );
  });
});
