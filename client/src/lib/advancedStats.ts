// advancedStats.ts — the wire shapes and words of the advanced stats
// pages (ADR 0217 §3, §6 and §8; slice 6 of #4880): GET
// /me/stats/advanced, the share link (POST and DELETE
// /me/stats/share-link) and a shared view (GET /stats/shared/{token}).
// The words a chart says (its caption sentence, a rate with its count)
// live here, so they are unit-tested in node; the geometry is in
// statsCharts.ts.

import type { StatsGroup, StatsResult, StatsTally, TopCard } from "./myStats";
import { placeText } from "./myStats";
import { L } from "./labels";
import { authFetch, LobbyApiError } from "./session";

/** A count out of a total; `rate` is null below MinGames (3). Mirrors gamestats.Rate. */
export interface Rate {
  n: number;
  of: number;
  rate: number | null;
}

export interface KindCount {
  kind: string;
  games: number;
}

export interface PlaceRow {
  players: number;
  games: number;
  /** places[0] is the games finished 1st. */
  places: number[];
}

export interface TurnOrderRow {
  position: number;
  wins: Rate;
  avg_place: number | null;
}

export interface Spread {
  median: number | null;
  p25: number | null;
  p75: number | null;
}

export interface Bucket {
  value: number;
  games: number;
}

export interface OverTurn {
  own_turn: number;
  turns: number;
  attacks_per_turn: number | null;
  damage_per_turn: number | null;
}

export interface LandMiss {
  land: number;
  missed: number;
}

export interface ManaGame {
  screw: boolean;
  flood: boolean;
  colour_screw: boolean;
}

export interface CastCard {
  name: string;
  games: number;
  casts: number;
  scryfall_id?: string;
}

export interface AdvancedSummary {
  coverage: { games: number; v2: number; frames: number; stand_in: number };
  record: StatsTally;
  win_rate: Rate;
  places: PlaceRow[];
  win_kinds: KindCount[];
  combo_rules: KindCount[];
  out_kinds: KindCount[];
  turn_order: TurnOrderRow[];
  length: { turns: Spread; rounds: Spread; minutes: Spread; by_turns: Bucket[] };
  mulligans: { per_game: number | null; rows: { mulligans: string; wins: Rate }[] };
  commander: {
    casts_per_game: number | null;
    first_cast_turn: Spread;
    by_turn_5: Rate;
    later: Rate;
  };
  aggression: {
    own_turns: number;
    attacks_per_turn: number | null;
    damage_per_turn: number | null;
    attack_share: number | null;
    most_hit_share: number | null;
    evenness: number | null;
    spread_games: number;
    over_turns: OverTurn[];
  };
  mana: {
    land_drops_hit: number;
    land_drops_missed: number;
    missed_by_land: LandMiss[];
    screw: Rate;
    flood: Rate;
    colour_screw: Rate;
    trend: ManaGame[];
  };
  most_cast: CastCard[];
}

export interface AdvancedDeckRow {
  /** Absent for a deck with no saved id, and always in a shared view. */
  deck_id?: string;
  name: string;
  record: StatsTally;
  /** Newest first, at most 20. */
  last: StatsResult[];
}

export interface LibraryCard {
  name: string;
  scryfall_id?: string;
  decks: number;
}

export type Makeup = "all" | "people" | "with_bots" | "bots";
export type StatsRange = "all" | "90d" | "30d";

/** The body of GET /me/stats/advanced, and a shared view's overview. */
export interface AdvancedOverview extends AdvancedSummary {
  makeup: Makeup;
  range: StatsRange;
  most_played_deck: { deck_id?: string; name: string; games: number } | null;
  decks: AdvancedDeckRow[];
  colours: { colour: string; games: Rate }[];
  library: {
    card_index: boolean;
    decks: number;
    cards_in_most_decks: LibraryCard[];
    avg_mana_value: number | null;
    avg_mana_value_games: number;
  };
}

/** A shared view's basic record: ADR 0165's, without the recent games. */
export interface SharedBasic extends StatsTally {
  unfinished: number;
  by_deck: StatsGroup[];
  by_commander: StatsGroup[];
  top_cards: TopCard[];
}

/** The body of GET /stats/shared/{token}. */
export interface SharedStats {
  display_name: string;
  advanced: boolean;
  basic: SharedBasic;
  overview: AdvancedOverview | null;
}

/** The body of POST /me/stats/share-link. The token is shown this once. */
export interface ShareLink {
  token: string;
  path: string;
  created_at: number;
}

export const MAKEUPS: readonly Makeup[] = ["all", "people", "with_bots", "bots"];
export const RANGES: readonly StatsRange[] = ["all", "90d", "30d"];

export const MAKEUP_TEXT: Record<Makeup, string> = {
  all: L.statsMakeupAll,
  people: L.statsMakeupPeople,
  with_bots: L.statsMakeupWithBots,
  bots: L.statsMakeupBots,
};

export const RANGE_TEXT: Record<StatsRange, string> = {
  all: L.statsRangeAll,
  "90d": L.statsRange90,
  "30d": L.statsRange30,
};

/** The query string both reads take. */
export function filterQuery(makeup: Makeup, range: StatsRange): string {
  return `?makeup=${makeup}&range=${range}`;
}

export function parseMakeup(v: string | null | undefined): Makeup {
  return (MAKEUPS as readonly string[]).includes(v ?? "") ? (v as Makeup) : "all";
}

export function parseRange(v: string | null | undefined): StatsRange {
  return (RANGES as readonly string[]).includes(v ?? "") ? (v as StatsRange) : "all";
}

export async function fetchAdvancedOverview(
  makeup: Makeup,
  range: StatsRange,
): Promise<AdvancedOverview> {
  const res = await authFetch(`/me/stats/advanced${filterQuery(makeup, range)}`, {
    cache: "no-store",
  });
  return (await res.json()) as AdvancedOverview;
}

export async function createShareLink(): Promise<ShareLink> {
  const res = await authFetch("/me/stats/share-link", { method: "POST" });
  return (await res.json()) as ShareLink;
}

export async function revokeShareLink(): Promise<void> {
  await authFetch("/me/stats/share-link", { method: "DELETE" });
}

/**
 * fetchSharedStats reads a shared view. It needs no session, so it is a
 * plain fetch; an unknown or revoked link is a LobbyApiError(404).
 */
export async function fetchSharedStats(
  token: string,
  makeup: Makeup,
  range: StatsRange,
): Promise<SharedStats> {
  const res = await fetch(
    `/stats/shared/${encodeURIComponent(token)}${filterQuery(makeup, range)}`,
    { cache: "no-store", credentials: "omit" },
  );
  if (!res.ok) {
    let msg = "could not load these stats";
    try {
      const body = (await res.json()) as { error?: string };
      if (body.error) msg = body.error;
    } catch {
      /* not JSON */
    }
    throw new LobbyApiError(res.status, msg);
  }
  return (await res.json()) as SharedStats;
}

/** The link a person posts: this page's origin and the shared route. */
export function shareURL(origin: string, token: string): string {
  return `${origin}/#/stats/shared/${token}`;
}

// ---- Words ----

/** The line shown in place of a rate on too few games (ADR 0217 §3). */
export const NOT_ENOUGH = "Not enough games yet";

/** A whole percent: 0.384 -> "38%". */
export function pctText(v: number | null | undefined): string {
  return v == null ? "-" : `${Math.round(v * 100)}%`;
}

/** "38%, 5 of 13 games", or NOT_ENOUGH below three games. */
export function rateText(r: Rate | null | undefined, noun = "games"): string {
  if (!r || r.rate == null) return NOT_ENOUGH;
  return `${pctText(r.rate)}, ${r.n} of ${r.of} ${noun}`;
}

/** One decimal, or "-": 2.345 -> "2.3". */
export function oneDecimal(v: number | null | undefined): string {
  return v == null ? "-" : v.toFixed(1);
}

/** How a game was won, or how a seat left, as a player reads it. */
export const KIND_TEXT: Record<string, string> = {
  combo: "Combo",
  concession: "Concession",
  win_effect: "“You win” card",
  poison: "Poison",
  commander_damage: "Commander damage",
  combat: "Combat damage",
  noncombat: "Burn or drain",
  decking: "Decking",
};

export function kindText(kind: string): string {
  return KIND_TEXT[kind] ?? kind.replace(/_/g, " ");
}

/** Why a win counted as a combo (ADR 0217 §3's combo rule). */
export const COMBO_RULE_TEXT: Record<string, string> = {
  loop: "a loop",
  repeats: "one card used 10 or more times",
  actions: "30 or more spells and abilities",
  combats: "4 or more combats in a turn",
};

/** Sorted, most games first, then by the order the ADR lists them. */
export function sortedKinds(kinds: readonly KindCount[]): KindCount[] {
  const order = Object.keys(KIND_TEXT);
  return [...kinds]
    .filter((k) => k.games > 0)
    .sort((a, b) => b.games - a.games || order.indexOf(a.kind) - order.indexOf(b.kind));
}

/** "Most of your wins (5 of 9) came by combat damage." */
export function winKindsCaption(kinds: readonly KindCount[]): string {
  const s = sortedKinds(kinds);
  const total = s.reduce((t, k) => t + k.games, 0);
  if (total === 0) return "No wins with a known kind yet.";
  const top = s[0];
  return `Your most common win is ${kindText(top.kind).toLowerCase()}: ${top.games} of ${total} ${total === 1 ? "win" : "wins"}.`;
}

/** "You lose most often to combat damage: 7 of 12 games." */
export function outKindsCaption(kinds: readonly KindCount[]): string {
  const s = sortedKinds(kinds);
  const total = s.reduce((t, k) => t + k.games, 0);
  if (total === 0) return "No losses with a known kind yet.";
  const top = s[0];
  return `You go out most often to ${kindText(top.kind).toLowerCase()}: ${top.games} of ${total} ${total === 1 ? "game" : "games"}.`;
}

/** The place row's caption: "At four-seat tables you finish 1st in 31% of 16 games." */
export function placesCaption(rows: readonly PlaceRow[]): string {
  const row = [...rows].sort((a, b) => b.games - a.games)[0];
  if (!row || row.games === 0) return "No finishing places yet.";
  const first = row.places[0] ?? 0;
  const size =
    { 2: "two", 3: "three", 4: "four", 5: "five", 6: "six" }[row.players] ?? `${row.players}`;
  if (row.games < 3)
    return `At ${size}-seat tables you have ${row.games} ${row.games === 1 ? "game" : "games"}: ${NOT_ENOUGH.toLowerCase()}.`;
  return `At ${size}-seat tables you finish 1st in ${pctText(first / row.games)} of ${row.games} games.`;
}

/** The first-turn effect's caption, naming the best position. */
export function turnOrderCaption(rows: readonly TurnOrderRow[]): string {
  const rated = rows.filter((r) => r.wins.rate != null);
  if (rated.length === 0)
    return `Turn order needs 3 four-seat games at a position: ${NOT_ENOUGH.toLowerCase()}.`;
  const best = [...rated].sort((a, b) => (b.wins.rate ?? 0) - (a.wins.rate ?? 0))[0];
  return `You win most when you act ${placeText(best.position)}: ${rateText(best.wins)}.`;
}

/** The aggression chart's caption. */
export function aggressionCaption(a: AdvancedSummary["aggression"]): string {
  if (a.attacks_per_turn == null) return "No turns with attacks recorded yet.";
  const share =
    a.attack_share == null
      ? ""
      : `, and ${pctText(a.attack_share)} of the creatures that could attack did`;
  return `You attack with ${oneDecimal(a.attacks_per_turn)} creatures a turn on average${share}.`;
}

/** The spread's caption: "52% of your damage goes at one opponent." */
export function spreadCaption(a: AdvancedSummary["aggression"]): string {
  if (a.most_hit_share == null || a.spread_games < 3) {
    return `How you spread damage needs 3 games at three or more seats: ${NOT_ENOUGH.toLowerCase()}.`;
  }
  return `In a typical game, ${pctText(a.most_hit_share)} of your damage goes at one opponent.`;
}

/** Missed drops' caption, naming the most-missed land number. */
export function missedCaption(m: AdvancedSummary["mana"]): string {
  const total = m.land_drops_hit + m.land_drops_missed;
  if (total === 0) return "No land drops recorded yet.";
  const worst = [...m.missed_by_land].sort((a, b) => b.missed - a.missed || a.land - b.land)[0];
  if (!worst || worst.missed === 0) return `You hit every land drop: ${total} of ${total}.`;
  return `You miss your ${landNumberText(worst.land)} land drop most often: ${worst.missed} ${worst.missed === 1 ? "time" : "times"}.`;
}

/** 3 -> "3rd", 6 (the top bucket) -> "6th or later". */
export function landNumberText(land: number, top = 6): string {
  return land >= top ? `${placeText(top)} or later` : placeText(land);
}

/** The colours chart's caption. */
export function coloursCaption(rows: readonly { colour: string; games: Rate }[]): string {
  const rated = rows.filter((r) => r.games.rate != null && r.games.n > 0);
  if (rated.length === 0)
    return `Colours need 3 games whose commanders we know: ${NOT_ENOUGH.toLowerCase()}.`;
  const top = [...rated].sort((a, b) => (b.games.rate ?? 0) - (a.games.rate ?? 0))[0];
  return `${COLOUR_TEXT[top.colour] ?? top.colour} is in the most of your decks played: ${rateText(top.games)}.`;
}

export const COLOUR_TEXT: Record<string, string> = {
  W: "White",
  U: "Blue",
  B: "Black",
  R: "Red",
  G: "Green",
};

/** Game length's caption. */
export function lengthCaption(l: AdvancedSummary["length"]): string {
  if (l.turns.median == null) return "No game lengths yet.";
  const mins = l.minutes.median == null ? "" : ` and about ${Math.round(l.minutes.median)} minutes`;
  return `Your games last ${Math.round(l.turns.median)} turns at the median${mins}.`;
}

/** "Median 9 turns", for the tile. */
export function medianLengthText(l: AdvancedSummary["length"]): string {
  return l.turns.median == null ? "-" : `${Math.round(l.turns.median)} turns`;
}

/** What the share dialog says the link reveals (ADR 0217 §6), one sentence. */
export const SHARE_REVEALS =
  "Anyone with the link can see your stats under your display name, including your deck names, commanders and the cards you cast, which shows much of your decklists; it never shows the other players’ names.";

/** The announcement after a filter change: once, through the polite region. */
export function filterAnnouncement(makeup: Makeup, range: StatsRange): string {
  return `Showing ${MAKEUP_TEXT[makeup].toLowerCase()}, ${RANGE_TEXT[range].toLowerCase()}.`;
}
