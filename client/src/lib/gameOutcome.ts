// gameOutcome.ts — ADR 0057 (#749): who won, and what the game-over
// banner says about it.
//
// The engine names the result on the wire (GameView.outcome) instead
// of leaving the client to infer it. An effect win (Felidar Sovereign,
// Laboratory Maniac, Thassa's Oracle) ends the game with the other
// seats STILL SEATED, so "the one seat left standing" is only a
// fallback — for a view with state "ended" and no outcome: a replay
// written before the engine recorded one, or a table an admin closed.

import type { GameView, PlayerView } from "./protocol";

/**
 * The winning seat of an ended game, or null for a draw, an active
 * game, or an ended game with nobody to name.
 */
export function winnerOf(view: GameView | null | undefined): PlayerView | null {
  if (!view || view.state !== "ended") return null;
  const seats = view.seats ?? [];
  const outcome = view.outcome;
  if (outcome) {
    if (outcome.kind !== "win" || !outcome.winner) return null;
    return seats.find((s) => s.id === outcome.winner) ?? null;
  }
  const survivors = seats.filter((s) => !s.eliminated);
  return survivors.length === 1 ? survivors[0] : null;
}

/**
 * The game-over banner's sentence (decided on 2026-09-17, option (b)):
 * the winner and, for an effect win, what won it.
 *
 *   "Alice wins the game — Felidar Sovereign"   an effect win
 *   "Alice wins the game."                       the last player standing
 *   "The game is a draw."                        everyone left lost at once
 *   "Game ended — no survivors."                 an ended view with nobody to name
 *
 * The winner's name is left to the caller, which renders it in bold
 * beside the seat colour; this returns the text AFTER the name, or the
 * whole sentence when there is no winner.
 */
export function gameOverText(view: GameView | null | undefined): {
  winner: PlayerView | null;
  text: string;
} {
  const winner = winnerOf(view);
  const outcome = view?.outcome;
  if (winner) {
    if (outcome?.cause === "effect" && outcome.source_name) {
      return { winner, text: `wins the game — ${outcome.source_name}` };
    }
    return { winner, text: "wins the game." };
  }
  if (outcome?.kind === "draw") return { winner: null, text: "The game is a draw." };
  return { winner: null, text: "Game ended — no survivors." };
}

/**
 * Which cue the viewer hears as the game ends: "win" for the winner,
 * "loss" for every other seat — a draw is not a win — and nothing for
 * a spectator (no viewer ID), whose outcome it isn't.
 */
export function endCueFor(
  view: GameView | null | undefined,
  viewerID: string | null,
): "win" | "loss" | null {
  if (!viewerID) return null;
  return winnerOf(view)?.id === viewerID ? "win" : "loss";
}
