// Removing someone else's seat from an unstarted table (#2785).
//
// The server is the gate: DELETE /games/{id}/seats/{player_id} answers
// 403 to anyone but the table host, the game's creator or an admin, and
// refuses a started table, the host's own seat and a bot. This module
// only decides whether to SHOW the Remove control, so a player is not
// offered a button that always fails, and it owns the confirm copy so
// the sentence can be tested.

import type { GameMeta, SeatInfo } from "./api";

export interface SeatRemovalViewer {
  /** The viewer is an admin (shared token, or an allowlisted person in admin mode). */
  isAdmin: boolean;
  /** The viewer's own seat at this table, or null when they hold none. */
  mySeat: SeatInfo | null;
}

// canRemoveSeat reports whether the Remove control belongs on `seat`
// for this viewer.
//
// It mirrors the server's rule and nothing wider: the viewer is the
// host (their seat carries is_host), the game's creator (the server
// stamps is_creator on the viewer's copy of the meta), or an admin; the
// table has not started; and the seat is a person's, not the host's and
// not a bot's (bots have their own control). Your own seat is left out
// too: nobody is offered a button to remove themselves, and the server
// has no use for that request.
export function canRemoveSeat(g: GameMeta, seat: SeatInfo, viewer: SeatRemovalViewer): boolean {
  if (g.state !== "lobby") return false;
  if (seat.is_bot || seat.is_host) return false;
  if (viewer.mySeat !== null && viewer.mySeat.player_id === seat.player_id) return false;
  return viewer.isAdmin || viewer.mySeat?.is_host === true || g.is_creator === true;
}

// The confirm's two sentences. It names what happens to the person and
// says the invite is unchanged, because the obvious next question is
// "can they just walk back in?" and the answer is yes unless the link is
// rotated.
export function removeSeatQuestion(seatName: string): string {
  return `Remove ${seatName} from this table?`;
}

export const REMOVE_SEAT_CONSEQUENCE =
  "Their seat opens up and they are disconnected. " +
  "The invite link still works, so mint a new one if you do not want them back.";
