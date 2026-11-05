// advancedStatsFixture.ts — one advanced overview, as GET
// /me/stats/advanced answers it, for the render tests and the e2e spec
// (tests-e2e/tests/stats-pages-4880.spec.ts imports it). It imports only
// types, so the e2e package can import it too.

import type { AdvancedOverview, SharedStats } from "../advancedStats";

const rate = (n: number, of: number) => ({ n, of, rate: of >= 3 ? n / of : null });

export function overviewFixture(): AdvancedOverview {
  return {
    makeup: "all",
    range: "all",
    coverage: { games: 14, v2: 14, frames: 13, stand_in: 0 },
    record: { games: 14, wins: 5, losses: 9, draws: 0, avg_place: 2.21 },
    win_rate: rate(5, 14),
    places: [
      { players: 4, games: 12, places: [4, 3, 3, 2] },
      { players: 2, games: 2, places: [1, 1] },
    ],
    win_kinds: [
      { kind: "combat", games: 3 },
      { kind: "combo", games: 1 },
      { kind: "commander_damage", games: 1 },
    ],
    combo_rules: [{ kind: "repeats", games: 1 }],
    out_kinds: [
      { kind: "combat", games: 6 },
      { kind: "noncombat", games: 2 },
      { kind: "concession", games: 1 },
    ],
    turn_order: [
      { position: 1, wins: rate(2, 4), avg_place: 1.8 },
      { position: 2, wins: rate(1, 3), avg_place: 2.3 },
      { position: 3, wins: rate(1, 3), avg_place: 2.4 },
      { position: 4, wins: rate(0, 2), avg_place: 3 },
    ],
    length: {
      turns: { median: 9, p25: 8, p75: 11 },
      rounds: { median: 3, p25: 2, p75: 3 },
      minutes: { median: 52, p25: 41, p75: 70 },
      by_turns: [
        { value: 6, games: 1 },
        { value: 7, games: 1 },
        { value: 8, games: 3 },
        { value: 9, games: 4 },
        { value: 10, games: 2 },
        { value: 11, games: 2 },
        { value: 13, games: 1 },
      ],
    },
    mulligans: { per_game: 0.4, rows: [{ mulligans: "0", wins: rate(4, 9) }] },
    commander: {
      casts_per_game: 1.6,
      first_cast_turn: { median: 4, p25: 3, p75: 5 },
      by_turn_5: rate(4, 9),
      later: rate(1, 5),
    },
    aggression: {
      own_turns: 120,
      attacks_per_turn: 1.8,
      damage_per_turn: 4.2,
      attack_share: 0.58,
      most_hit_share: 0.52,
      evenness: 0.64,
      spread_games: 11,
      over_turns: Array.from({ length: 12 }, (_, i) => ({
        own_turn: i + 1,
        turns: 14 - i,
        attacks_per_turn: i < 1 ? null : Math.min(4, 0.3 * i + 0.2),
        damage_per_turn: i < 1 ? null : Math.min(9, 0.9 * i),
      })),
    },
    mana: {
      land_drops_hit: 96,
      land_drops_missed: 24,
      missed_by_land: [
        { land: 1, missed: 0 },
        { land: 2, missed: 1 },
        { land: 3, missed: 3 },
        { land: 4, missed: 6 },
        { land: 5, missed: 5 },
        { land: 6, missed: 9 },
      ],
      screw: rate(3, 14),
      flood: rate(2, 14),
      colour_screw: rate(4, 14),
      trend: Array.from({ length: 13 }, (_, i) => ({
        screw: i % 5 === 0,
        flood: i % 6 === 3,
        colour_screw: i % 3 === 1,
      })),
    },
    most_cast: [
      { name: "Sol Ring", games: 9, casts: 9 },
      { name: "Arcane Signet", games: 8, casts: 8 },
      { name: "Swords to Plowshares", games: 6, casts: 7 },
    ],
    most_played_deck: { name: "Atraxa Superfriends", games: 9 },
    decks: [
      {
        name: "Atraxa Superfriends",
        record: { games: 9, wins: 4, losses: 5, draws: 0, avg_place: 2 },
        last: ["win", "loss", "win", "loss", "loss", "win", "loss", "win", "loss"],
      },
      {
        name: "Krenko Goblins",
        record: { games: 5, wins: 1, losses: 4, draws: 0, avg_place: 2.6 },
        last: ["loss", "loss", "win", "loss", "loss"],
      },
    ],
    colours: [
      { colour: "W", games: rate(9, 14) },
      { colour: "U", games: rate(9, 14) },
      { colour: "B", games: rate(9, 14) },
      { colour: "R", games: rate(5, 14) },
      { colour: "G", games: rate(9, 14) },
    ],
    library: {
      card_index: true,
      decks: 2,
      cards_in_most_decks: [
        { name: "Sol Ring", decks: 2 },
        { name: "Arcane Signet", decks: 2 },
      ],
      avg_mana_value: 3.12,
      avg_mana_value_games: 14,
    },
  };
}

export function sharedFixture(): SharedStats {
  return {
    display_name: "Bea",
    advanced: true,
    basic: {
      games: 14,
      wins: 5,
      losses: 9,
      draws: 0,
      avg_place: 2.21,
      unfinished: 0,
      by_deck: [],
      by_commander: [],
      top_cards: [],
    },
    overview: overviewFixture(),
  };
}
