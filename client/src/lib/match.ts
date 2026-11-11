// match.ts — a table's match on the client (ADR 0222 §3, #5436).
//
// A match is one or three games between the same two players at one
// table. The server keeps the score and runs the sideboarding window
// between games (GameView.match); this file is what the client needs to
// SAY about it and the arithmetic of the sideboarding screen. Nothing
// here decides anything: a deck the server's check refuses is refused
// whatever this file thinks.
//
// A table with no match (every Commander table) has no `match` on its
// view, and a best-of-one has one with best_of 1. Both read as "no
// match to show" here, so a caller shows a score only for a longer one.

import type { CardView, GameView, MatchPoolView, MatchView, SideboardingView } from "./protocol";

type Seat = { seat: number; name: string };

/** The view's match when it is longer than one game, else null. */
export function matchOf(view: Pick<GameView, "match"> | null | undefined): MatchView | null {
  const m = view?.match;
  return m && m.best_of > 1 ? m : null;
}

/** The open sideboarding window, or null while a game is in play. */
export function sideboardingOf(
  view: Pick<GameView, "match"> | null | undefined,
): SideboardingView | null {
  return matchOf(view)?.sideboarding ?? null;
}

/** The viewer's own pool between games; null for a spectator. */
export function poolOf(
  sb: SideboardingView | null | undefined,
  viewerID: string | null | undefined,
): MatchPoolView | null {
  if (!sb || !viewerID) return null;
  return sb.pools?.find((p) => p.player_id === viewerID) ?? null;
}

/** "1–0", in seat order. */
export function scoreOf(match: MatchView): string {
  return match.wins.join("–");
}

/** "Alice 1–0 Bob", the seats named in seat order. */
export function scoreLine(match: MatchView, seats: readonly Seat[]): string {
  const ordered = [...seats].sort((a, b) => a.seat - b.seat);
  if (ordered.length !== 2) return scoreOf(match);
  const [a, b] = ordered;
  return `${a.name} ${match.wins[a.seat] ?? 0}–${match.wins[b.seat] ?? 0} ${b.name}`;
}

/** "Game 2 of 3", the game in play or the last one played. */
export function gameOfText(match: MatchView): string {
  return `Game ${match.game} of ${match.best_of}`;
}

/** What the last finished game came to: "Alice won game 1." */
export function lastGameText(match: MatchView, seats: readonly Seat[]): string {
  const games = match.games ?? [];
  const last = games[games.length - 1];
  if (!last) return "";
  const n = games.length;
  if (last.winner_seat === undefined) return `Game ${n} was a draw.`;
  const name = seats.find((s) => s.seat === last.winner_seat)?.name ?? "A player";
  return `${name} won game ${n}.`;
}

/**
 * The game-over banner's words for a finished match, after the winner's
 * name: "wins the match 2–1." With no winner, the whole sentence.
 */
export function matchOverText(match: MatchView, seats: readonly Seat[]): string {
  if (match.winner_seat === undefined) return `The match is a draw, ${scoreOf(match)}.`;
  const winner = match.winner_seat;
  const loser = seats.find((s) => s.seat !== winner);
  const won = match.wins[winner] ?? 0;
  const lost = loser ? (match.wins[loser.seat] ?? 0) : 0;
  if (match.conceded_seat !== undefined && loser && match.conceded_seat === loser.seat) {
    return `wins the match. ${loser.name} conceded it at ${won}–${lost}.`;
  }
  return `wins the match ${won}–${lost}.`;
}

/** One line of the sideboarding screen: a card name and its copies. */
export interface PoolLine {
  name: string;
  /** The pool ids of the copies on this side, in pool order. */
  ids: string[];
  /** One copy, for the line's cost and type. */
  card: CardView;
  /**
   * How many of the copies are on the other side in the registered
   * list: brought in, on the deck's side; taken out, on the sideboard's.
   */
  moved: number;
}

/**
 * The cards of one side as lines, one per name, in name order. `main`
 * is the pool ids of the deck as the player has it now; `inDeck` picks
 * the side. `registered` marks what moved.
 */
export function poolLines(
  pool: MatchPoolView,
  main: ReadonlySet<string>,
  inDeck: boolean,
): PoolLine[] {
  const registered = new Set(pool.registered);
  const byName = new Map<string, PoolLine>();
  for (const card of [...pool.deck, ...pool.sideboard]) {
    if (main.has(card.instance_id) !== inDeck) continue;
    let line = byName.get(card.name);
    if (!line) {
      line = { name: card.name, ids: [], card, moved: 0 };
      byName.set(card.name, line);
    }
    line.ids.push(card.instance_id);
    if (registered.has(card.instance_id) !== inDeck) line.moved++;
  }
  return [...byName.values()].sort((a, b) => a.name.localeCompare(b.name));
}

/** The pool ids of the deck as the server has it. */
export function mainOf(pool: MatchPoolView): string[] {
  return pool.deck.map((c) => c.instance_id);
}

/** `main` with one card moved to the other side. */
export function moveCard(main: readonly string[], id: string): string[] {
  return main.includes(id) ? main.filter((x) => x !== id) : [...main, id];
}

/** Whether two decks hold the same cards, whatever their order. */
export function sameDeck(a: readonly string[], b: readonly string[]): boolean {
  if (a.length !== b.length) return false;
  const set = new Set(a);
  return b.every((id) => set.has(id));
}

/**
 * What is wrong with a deck of `deck` cards and a sideboard of `side`,
 * in words for the player, or "" when the format's check would take it.
 * The server's check is the one that counts; this says the same thing
 * before the player asks.
 */
export function deckProblem(deck: number, side: number, sb: SideboardingView): string {
  if (deck < sb.deck_min) {
    const short = sb.deck_min - deck;
    return `Your deck has ${deck} cards. Bring in ${short} more to reach ${sb.deck_min}.`;
  }
  if (side > sb.sideboard_max) {
    const over = side - sb.sideboard_max;
    return `Your sideboard has ${side} cards. Bring in ${over} more, or take fewer out: it holds ${sb.sideboard_max}.`;
  }
  return "";
}

/**
 * Seconds left in the window: the server's count when the frame was
 * built, less the time since this client received it, so a clock that
 * is wrong on this device does not move the deadline.
 */
export function secondsLeft(sb: SideboardingView, receivedAt: number, now: number): number {
  return Math.max(0, Math.ceil(sb.seconds_left - (now - receivedAt) / 1000));
}

/** "2:05". */
export function clockText(seconds: number): string {
  const s = Math.max(0, Math.floor(seconds));
  return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, "0")}`;
}

/** The match length of a table's settings: 3 or 1. */
export function bestOfSetting(n: number | undefined | null): 1 | 3 {
  return n === 3 ? 3 : 1;
}
