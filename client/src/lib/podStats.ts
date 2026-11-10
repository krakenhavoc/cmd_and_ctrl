// podStats.ts — a pod's stats and leaderboard (ADR 0220 §7, slice 5 of
// #4880 = ADR 0217 slice 7): the wire types for GET /pods/{id}/stats and
// GET /pods/{id}/leaderboard, the calls, and the sentences each figure
// says.
//
// Mirrors server/internal/lobby/pod_stats.go. Hand-maintained, like
// pods.ts; update both sides in lockstep. A member who is not sharing
// has no numbers; a member hidden by a block is simply not in the lists.

import { authFetch } from "./session";
import {
  kindText,
  oneDecimal,
  pctText,
  type KindCount,
  type StatsRange,
  type TurnOrderRow,
} from "./advancedStats";

export { RANGES, RANGE_TEXT, parseRange, type StatsRange } from "./advancedStats";

/** The leaderboard's minimum, ADR 0217 owner answer 7. The server sends its own. */
export const POD_MIN_GAMES = 5;

/** A record: games, wins, losses, draws and the mean place. */
export interface PodTally {
  games: number;
  wins: number;
  losses: number;
  draws: number;
  avg_place: number | null;
}

export interface PodStatsMember {
  user_id: string;
  display_name: string;
  avatar_url?: string;
  owner?: boolean;
  you?: boolean;
  sharing: boolean;
  /** The member's record in the pod's games; absent while not sharing. */
  pod_games?: PodTally;
  /** The member's record over every game in the range; absent while not sharing. */
  overall?: PodTally;
}

export interface PodPerson {
  user_id?: string;
  display_name: string;
  avatar_url?: string;
}

export interface PodHeadToHead {
  a: PodPerson;
  b: PodPerson;
  games: number;
  a_wins: number;
  b_wins: number;
  others: number;
}

export interface PodStats {
  id: string;
  name: string;
  range: StatsRange;
  /** The pod's games: ended games where two or more sharing members sat. */
  games: number;
  min_games: number;
  members: PodStatsMember[];
  win_kinds: KindCount[];
  turn_order: TurnOrderRow[];
  head_to_head: PodHeadToHead[];
}

export interface PodBoardRow extends PodTally {
  user_id: string;
  display_name: string;
  avatar_url?: string;
  you?: boolean;
  /** Absent on an unranked row. */
  rank?: number;
  win_rate: number;
  /** How many more pod games an unranked row needs. */
  needs?: number;
}

export interface PodLeaderboard {
  id: string;
  name: string;
  range: StatsRange;
  games: number;
  min_games: number;
  rows: PodBoardRow[];
  unranked: PodBoardRow[];
}

const base = (id: string) => `/pods/${encodeURIComponent(id)}`;
const query = (range: StatsRange) => `?range=${range}`;

function tidyStats(s: PodStats): PodStats {
  return {
    ...s,
    members: s.members ?? [],
    win_kinds: s.win_kinds ?? [],
    turn_order: s.turn_order ?? [],
    head_to_head: s.head_to_head ?? [],
  };
}

export async function fetchPodStats(id: string, range: StatsRange): Promise<PodStats> {
  const res = await authFetch(`${base(id)}/stats${query(range)}`, { cache: "no-store" });
  return tidyStats((await res.json()) as PodStats);
}

export async function fetchPodLeaderboard(id: string, range: StatsRange): Promise<PodLeaderboard> {
  const res = await authFetch(`${base(id)}/leaderboard${query(range)}`, { cache: "no-store" });
  const raw = (await res.json()) as PodLeaderboard;
  return { ...raw, rows: raw.rows ?? [], unranked: raw.unranked ?? [] };
}

/** The route of a pod's stats page. */
export function podStatsHash(id: string): string {
  return `#/pods/${id}/stats`;
}

// ---- Words ----

/** "5-2-0" — wins, losses, draws. */
export function recordLine(t: Pick<PodTally, "wins" | "losses" | "draws">): string {
  return `${t.wins}-${t.losses}-${t.draws}`;
}

/** "62%" of a record, "-" with no games. */
export function winRateLine(t: Pick<PodTally, "games" | "wins">): string {
  return t.games === 0 ? "-" : pctText(t.wins / t.games);
}

export function placeLine(v: number | null | undefined): string {
  return v == null ? "-" : oneDecimal(v);
}

export function gamesLine(n: number): string {
  return `${n} ${n === 1 ? "game" : "games"}`;
}

/** The leaderboard's caption: who leads, with the count it stands on. */
export function boardCaption(b: PodLeaderboard): string {
  if (b.rows.length === 0) {
    return b.games === 0
      ? "No pod games yet. A pod game is one where two or more sharing members sat."
      : `Nobody has ${b.min_games} pod games yet in this range.`;
  }
  const top = b.rows[0];
  return `${top.display_name} leads with ${pctText(top.win_rate)} over ${gamesLine(top.games)}.`;
}

/** The pod-games line under the page title. */
export function podGamesLine(s: Pick<PodStats, "games" | "range">): string {
  const when = { all: "", "90d": " in the last 90 days", "30d": " in the last 30 days" }[s.range];
  return `${gamesLine(s.games)}${when} where two or more sharing members sat.`;
}

/** The win-kinds caption for the pod. */
export function podWinKindsCaption(kinds: readonly KindCount[]): string {
  const sorted = [...kinds].filter((k) => k.games > 0).sort((a, b) => b.games - a.games);
  const total = sorted.reduce((t, k) => t + k.games, 0);
  if (total === 0) return "No pod games with a known win kind yet.";
  return `Pod games are won most often by ${kindText(sorted[0].kind).toLowerCase()}: ${sorted[0].games} of ${total}.`;
}

/** The first-turn effect's caption. */
export function podTurnOrderCaption(rows: readonly TurnOrderRow[]): string {
  const rated = rows.filter((r) => r.wins.rate != null);
  if (rated.length === 0) return "Turn order needs 3 four-seat pod games at a position.";
  const best = [...rated].sort((a, b) => (b.wins.rate ?? 0) - (a.wins.rate ?? 0))[0];
  return `Acting ${["1st", "2nd", "3rd", "4th"][best.position - 1] ?? best.position} wins most in this pod: ${pctText(best.wins.rate)} of ${gamesLine(best.wins.of)}.`;
}

/** The head-to-head line for one pair: "Ann 4, Bea 2 over 7 games." */
export function headToHeadLine(h: PodHeadToHead): string {
  return `${h.a.display_name} ${h.a_wins}, ${h.b.display_name} ${h.b_wins} over ${gamesLine(h.games)}`;
}

/** The sentence under the sharing switch: it covers every pod. */
export const SHARE_PODS_NOTE =
  "One switch for every pod you are in. While it is on, the members of those pods can see your record and your place on their leaderboards. People you have blocked, or who have blocked you, never see it. Joining a pod shares nothing by itself.";

export const NOT_SHARING_LINE = "Not sharing. Their numbers stay private to them.";
