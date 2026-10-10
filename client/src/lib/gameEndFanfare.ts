// gameEndFanfare.ts — #2920: what the game-end overlay says and how it
// feels, decided from the view and the viewer. Pure so the tone table
// is tested without a DOM; GameEndFanfare.svelte only paints it.
//
// Four tones:
//   own-win  the viewer won: the biggest moment
//   loss     the viewer sat at the table and did not win (a draw is "draw")
//   draw     nobody won: neutral, no confetti
//   watch    a spectator (not seated): the winner, without "you"

import { seatColor } from "./colors";
import { gameOverText } from "./gameOutcome";
import type { GameView } from "./protocol";

export type FanfareTone = "own-win" | "loss" | "draw" | "watch";

export interface Fanfare {
  tone: FanfareTone;
  /** Small line above the name. */
  kicker: string;
  /** The big line: the winner's name, or the draw/no-survivor headline. */
  headline: string;
  /** The line below: how it was won, or a consolation. */
  detail: string;
  /** The winner's seat colour, or null when there is no winner. */
  color: string | null;
  /** Whether the confetti burst runs (the viewer's win, or a spectator's). */
  celebrate: boolean;
  /** One sentence for the dialog's accessible description. */
  summary: string;
}

const EFFECT_PREFIX = "wins the game — ";

/** The overlay's content, or null while the game is not over. */
export function fanfareFor(
  view: GameView | null | undefined,
  viewerID: string | null,
): Fanfare | null {
  if (!view || view.state !== "ended") return null;
  const { winner, text } = gameOverText(view);
  const seated = viewerID !== null && (view.seats ?? []).some((s) => s.id === viewerID);
  // gameOverText's tail is "wins the game." or "wins the game — Source".
  const how = text.startsWith(EFFECT_PREFIX) ? text.slice(EFFECT_PREFIX.length) : "";

  if (winner) {
    const own = viewerID !== null && winner.id === viewerID;
    const tone: FanfareTone = own ? "own-win" : seated ? "loss" : "watch";
    let detail: string;
    if (own) detail = how ? `You won with ${how}.` : "You are the last one standing.";
    else if (how) detail = `Won with ${how}.`;
    else detail = seated ? "Better luck next game." : "Last one standing.";
    return {
      tone,
      kicker: own ? "Victory" : "Winner",
      headline: winner.name,
      detail,
      color: seatColor(winner.seat),
      celebrate: tone !== "loss",
      summary: `${winner.name} ${text}`,
    };
  }
  const draw = view.outcome?.kind === "draw";
  return {
    tone: "draw",
    kicker: "Game over",
    headline: draw ? "A draw" : "No survivors",
    detail: draw ? "Everyone left lost at once." : "Nobody is left to name.",
    color: null,
    celebrate: false,
    summary: text,
  };
}
