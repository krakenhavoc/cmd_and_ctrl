// myStats.ts is the pure half of the stats page (ADR 0165 §7, #3366):
// the wire shape of GET /me/stats and the words the page puts on it.
// No DOM and no fetch, so it is unit-tested in the node environment.

/** A count of results over some games. Mirrors gamestats.Tally. */
export interface StatsTally {
  games: number;
  wins: number;
  losses: number;
  draws: number;
  /** Mean place over the games that have one; null when none do. */
  avg_place: number | null;
}

/** A deck or a commander (pair) and its tally. */
export interface StatsGroup extends StatsTally {
  name: string;
}

export type StatsResult = "win" | "loss" | "draw" | "none";

/** One finished game of the caller's. Times are Unix milliseconds. */
export interface RecentStatsGame {
  game_id: string;
  name: string;
  ended_at: number;
  result: StatsResult;
  place: number | null;
  turns: number;
  players: number;
  deck?: string;
  commanders: string[];
}

/** A card the caller cast, and how often over their counted games. */
export interface TopCard {
  name: string;
  casts: number;
}

/** The body of GET /me/stats. */
export interface MyStats extends StatsTally {
  /** Tables that ended with no outcome; they are left out of every count. */
  unfinished: number;
  by_deck: StatsGroup[];
  by_commander: StatsGroup[];
  recent: RecentStatsGame[];
  /** The caller's most-cast cards, at most ten (ADR 0179 §7). */
  top_cards?: TopCard[];
}

/** "4-7-1", or "4-7" when nothing was drawn. */
export function recordText(t: Pick<StatsTally, "wins" | "losses" | "draws">): string {
  return t.draws > 0 ? `${t.wins}-${t.losses}-${t.draws}` : `${t.wins}-${t.losses}`;
}

/** The win rate as a whole percent, or "-" with no games. */
export function winRateText(t: Pick<StatsTally, "games" | "wins">): string {
  if (t.games <= 0) return "-";
  return `${Math.round((t.wins / t.games) * 100)}%`;
}

/** An average place to one decimal ("2.4"), or "-" when unknown. */
export function avgPlaceText(avg: number | null | undefined): string {
  return avg == null ? "-" : avg.toFixed(1);
}

/** 1 -> "1st", 2 -> "2nd", 11 -> "11th". "-" when unknown. */
export function placeText(place: number | null | undefined): string {
  if (place == null || place < 1) return "-";
  const tens = place % 100;
  if (tens >= 11 && tens <= 13) return `${place}th`;
  const suffix = ({ 1: "st", 2: "nd", 3: "rd" } as Record<number, string>)[place % 10] ?? "th";
  return `${place}${suffix}`;
}

/** The result as a player reads it. */
export function resultText(r: StatsResult): string {
  return { win: "Won", loss: "Lost", draw: "Draw", none: "No result" }[r];
}
