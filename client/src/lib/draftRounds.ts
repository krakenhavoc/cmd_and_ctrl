// draftRounds.ts — reading a draft's rounds (ADR 0222 §6, #5440): who
// the viewer plays, how each match went, and the win-loss table.
//
// Pure functions over the `play` part of a draft view. Seats are named
// by number on the wire; a seat's name comes from `players`.

import type { DraftMatch, DraftPlayer, DraftStanding, DraftView } from "./draft";

/** A seat's name, with "(bot)" left to the caller's own marking. */
export function seatName(view: Pick<DraftView, "players">, seat: number | undefined): string {
  if (seat === undefined) return "";
  return view.players.find((p) => p.seat === seat)?.name ?? `Seat ${seat + 1}`;
}

function player(view: Pick<DraftView, "players">, seat: number): DraftPlayer | undefined {
  return view.players.find((p) => p.seat === seat);
}

/** The seat the viewer plays in a match, or undefined for a bye. */
export function opponentSeat(match: DraftMatch, you: number): number | undefined {
  if (match.bye || match.b === undefined) return undefined;
  return match.a === you ? match.b : match.a;
}

/** "win", "loss" or "draw" for one seat of a finished match; "" otherwise. */
export function resultFor(match: DraftMatch, seat: number): "win" | "loss" | "draw" | "" {
  if (!match.done || match.bye) return "";
  if (match.winner === undefined) return "draw";
  return match.winner === seat ? "win" : "loss";
}

/** A seat's record as "2–1", with draws as a third number when it has any. */
export function recordText(s: Pick<DraftStanding, "wins" | "losses" | "draws">): string {
  return s.draws > 0 ? `${s.wins}–${s.losses}–${s.draws}` : `${s.wins}–${s.losses}`;
}

/** "Best of three" or "Best of one". */
export function matchLengthName(bestOf: number | undefined): string {
  return bestOf === 3 ? "Best of three" : "Best of one";
}

/** One line for one match, as everyone at the draft may read it. */
export function matchLine(view: Pick<DraftView, "players">, m: DraftMatch): string {
  const a = seatName(view, m.a);
  if (m.bye || m.b === undefined) return `${a} sits this round out`;
  const b = seatName(view, m.b);
  if (!m.done) return `${a} against ${b}: playing`;
  const score = `${m.wins_a}–${m.wins_b}`;
  const how = m.coin ? ", decided by a coin toss" : "";
  if (m.winner === undefined) return `${a} and ${b} drew ${score}${how}`;
  const [w, l, ws] = m.winner === m.a ? [a, b, score] : [b, a, `${m.wins_b}–${m.wins_a}`];
  return `${w} beat ${l} ${ws}${how}`;
}

export type YourMatchState = "bye" | "opening" | "ready" | "won" | "lost" | "drew" | "none";

/** Where the viewer's match of the round in play stands. */
export function yourMatchState(view: DraftView): YourMatchState {
  const mine = view.play?.yours;
  if (!mine || view.phase !== "rounds") return "none";
  if (mine.bye) return "bye";
  if (mine.done) {
    const r = resultFor(mine, view.you);
    return r === "win" ? "won" : r === "loss" ? "lost" : "drew";
  }
  return view.play?.table ? "ready" : "opening";
}

/** What the viewer is told about their match of the round in play. */
export function yourMatchLine(view: DraftView): string {
  const mine = view.play?.yours;
  const state = yourMatchState(view);
  if (!mine || state === "none") return "";
  if (state === "bye") {
    return "You sit this round out: the draft has an odd number of seats. It counts for nothing.";
  }
  const seat = opponentSeat(mine, view.you);
  const who = seatName(view, seat);
  const bot = seat !== undefined && player(view, seat)?.bot ? ", a bot" : "";
  switch (state) {
    case "opening":
      return `You play ${who}${bot}. Your table is being set.`;
    case "ready":
      return `You play ${who}${bot}. Your table is ready.`;
    case "won":
      return `You beat ${who}. The next round starts when the other matches finish.`;
    case "lost":
      return `${who} beat you. The next round starts when the other matches finish.`;
    default:
      return `You and ${who} drew. The next round starts when the other matches finish.`;
  }
}

export interface StandingRow extends DraftStanding {
  name: string;
  bot: boolean;
  you: boolean;
}

/**
 * The win-loss table: most wins first, then fewest losses, then seat
 * order. There are no standings beyond this and no prizes (ADR 0222 §6).
 */
export function standingRows(view: DraftView): StandingRow[] {
  const rows = (view.play?.standings ?? []).map((s) => ({
    ...s,
    name: seatName(view, s.seat),
    bot: player(view, s.seat)?.bot === true,
    you: s.seat === view.you,
  }));
  return rows.sort((a, b) => b.wins - a.wins || a.losses - b.losses || a.seat - b.seat);
}

/** What a screen reader hears when the rounds move on; "" for no news. */
export function roundsAnnouncement(prev: DraftView | null, next: DraftView): string {
  if (next.phase === "building" && prev?.phase !== "building") {
    return "The picks are done. Build your deck.";
  }
  if (next.phase === "finished" && prev?.phase !== "finished") {
    return "The draft is finished.";
  }
  if (next.phase !== "rounds" || !next.play) return "";
  if (prev?.phase !== "rounds" || prev.play?.round !== next.play.round) {
    return `Round ${next.play.round} of ${next.play.rounds}. ${yourMatchLine(next)}`;
  }
  if (!prev.play?.table && next.play.table) return "Your table is ready.";
  return "";
}
