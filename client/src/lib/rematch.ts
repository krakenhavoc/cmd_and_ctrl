// rematch.ts — #3285, ADR 0164: what a finished table offers each viewer.
//
// The host of a finished table can offer the same game again. The offer
// rides GameView.rematch, reduced by the server to the viewer's own part;
// this module decides, from the view alone, what the results screen shows:
//
//   offer       the viewer may offer one and nobody has: a Rematch button
//   open-table  the viewer offered one: a way back to the new table
//   ask         the host offered and the offer is open: Join / No thanks
//   none        nothing (a spectator, a bot's seat, a practice table, an
//               offer already answered or closed)
//
// "May offer" is the server's rule, mirrored (lobby.rematchHostLocked): the
// table's host, or, when every human lost and the bots won so that nobody
// is left to host, the first human seat. Only that seat can offer, so a
// button shown to anyone else would only fail.

import type { GameView, PlayerView } from "./protocol";

export type RematchPlan =
  | { kind: "none" }
  | { kind: "offer" }
  | { kind: "open-table"; gameID: string }
  | { kind: "ask"; hostName: string; gameID: string };

/** Whether a seat is a person who plays for themselves. */
function isHuman(s: PlayerView): boolean {
  return !s.is_bot && !s.is_agent;
}

/**
 * The seat that may offer the rematch: the table's host, else the first
 * human seat. Null when the table has no human seat.
 */
export function rematchHostSeat(view: GameView | null | undefined): PlayerView | null {
  const seats = view?.seats ?? [];
  const host = seats.find((s) => s.is_host && isHuman(s));
  if (host) return host;
  const humans = seats.filter(isHuman).sort((a, b) => a.seat - b.seat);
  return humans[0] ?? null;
}

/**
 * What the results screen offers `viewerID`. `practice` is the tutorial's
 * table, which has nothing to play again.
 */
export function rematchPlan(
  view: GameView | null | undefined,
  viewerID: string | null,
  practice = false,
): RematchPlan {
  if (practice || !view || view.state !== "ended" || !viewerID) return { kind: "none" };
  const viewer = (view.seats ?? []).find((s) => s.id === viewerID);
  if (!viewer || !isHuman(viewer)) return { kind: "none" };
  const offer = view.rematch;
  if (offer) {
    if (offer.you === "host" || offer.host === viewerID) {
      return { kind: "open-table", gameID: offer.game_id };
    }
    if (offer.open && offer.you === "undecided") {
      return { kind: "ask", hostName: offer.host_name || "The host", gameID: offer.game_id };
    }
    return { kind: "none" };
  }
  return rematchHostSeat(view)?.id === viewerID ? { kind: "offer" } : { kind: "none" };
}
