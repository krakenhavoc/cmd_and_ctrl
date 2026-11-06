// advancedStatsFixture.ts — one advanced overview, as GET
// /me/stats/advanced answers it, for the render tests and the e2e spec
// (tests-e2e/tests/stats-pages-4880.spec.ts imports it). It imports only
// types, so the e2e package can import it too.

import type { AdvancedDeck, AdvancedOverview, SharedStats } from "../advancedStats";

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
      land_curve: Array.from({ length: 7 }, (_, i) => ({
        n: i + 1,
        hit: rate(Math.max(3, 14 - i * 2), 14),
        controlled: rate(Math.max(3, 14 - i), 14),
      })),
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
        deck_id: "11111111-2222-4333-8444-555555555555",
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

/** One deck's page, as GET /me/stats/decks/{id} answers it. */
export function deckFixture(): AdvancedDeck {
  const o = overviewFixture();
  const roles = [
    "removal",
    "board_wipe",
    "counterspell",
    "fog",
    "ramp",
    "mana_rock",
    "mana_dork",
    "draw",
    "tutor",
    "protection",
    "recursion",
    "graveyard_hate",
    "burn",
    "lifegain",
    "token_maker",
    "sacrifice_outlet",
    "wincon",
  ].map((role, i) => ({ role, cards: i % 4 === 0 ? 6 : i % 4 === 1 ? 2 : 0 }));
  return {
    makeup: "all",
    range: "all",
    coverage: o.coverage,
    record: { games: 9, wins: 4, losses: 5, draws: 0, avg_place: 2 },
    win_rate: rate(4, 9),
    places: o.places,
    win_kinds: o.win_kinds,
    combo_rules: o.combo_rules,
    out_kinds: o.out_kinds,
    turn_order: o.turn_order,
    length: o.length,
    mulligans: o.mulligans,
    commander: o.commander,
    aggression: o.aggression,
    mana: o.mana,
    most_cast: o.most_cast,
    deck: {
      id: "11111111-2222-4333-8444-555555555555",
      name: "Atraxa Superfriends",
      updated_at: Date.UTC(2026, 10, 1),
      commanders: ["Atraxa, Praetors' Voice"],
      commander_cards: [
        {
          name: "Atraxa, Praetors' Voice",
          scryfall_id: "22222222-2222-4222-8222-222222222222",
          color_identity: ["W", "U", "B", "G"],
        },
      ],
      roles_at: Date.UTC(2026, 10, 4),
      facts: {
        hash: "1f0c9a7d2b3e4f50",
        cards: 100,
        unresolved: 2,
        types: [
          { type: "land", cards: 37 },
          { type: "creature", cards: 21 },
          { type: "artifact", cards: 12 },
          { type: "enchantment", cards: 6 },
          { type: "instant", cards: 10 },
          { type: "sorcery", cards: 8 },
          { type: "planeswalker", cards: 4 },
          { type: "battle", cards: 0 },
        ],
        roles_available: true,
        roles,
        mana_curve: [2, 9, 14, 16, 10, 6, 4, 2],
        nonland: 63,
        avg_mana_value: 3.2,
        colour_share: [
          { colour: "W", symbols: 18.5, share: 0.27 },
          { colour: "U", symbols: 17, share: 0.25 },
          { colour: "B", symbols: 12.5, share: 0.18 },
          { colour: "R", symbols: 0, share: 0 },
          { colour: "G", symbols: 20, share: 0.3 },
        ],
        colour_identity: ["W", "U", "B", "G"],
      },
    },
    version: "current",
    no_games_since_edit: false,
    versions: [
      {
        hash: "1f0c9a7d2b3e4f50",
        current: true,
        first: Date.UTC(2026, 10, 2),
        last: Date.UTC(2026, 10, 4),
        games: 6,
      },
      {
        hash: "0f0f0f0f0f0f0f0f",
        current: false,
        first: Date.UTC(2026, 9, 10),
        last: Date.UTC(2026, 9, 20),
        games: 3,
      },
    ],
    dead_cards: [
      {
        name: "Craterhoof Behemoth",
        dead: rate(4, 5),
        could_afford: false,
        scryfall_id: "33333333-3333-4333-8333-333333333333",
      },
      { name: "Swords to Plowshares", dead: rate(3, 6), could_afford: true },
    ],
    other_decks: {
      decks: 3,
      types: [
        { type: "land", average: 36.3 },
        { type: "creature", average: 28 },
      ],
      roles: roles.map((r) => ({ role: r.role, average: r.cards > 0 ? 5 : 0 })),
    },
  };
}
