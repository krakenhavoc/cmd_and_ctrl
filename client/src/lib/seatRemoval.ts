// Removing someone else's seat from a table (#2785, #3093).
//
// The server is the gate: DELETE /games/{id}/seats/{player_id} answers
// 403 to anyone but the table host, the game's creator or an admin. An
// unstarted table takes all three; a started game takes the admin
// alone, and the removal concedes the seat out (CR 800.4a). The host
// and the creator get a 409 there. An ended game, a bot, and the host's
// seat (unless the host is leaving it, #3494) are refused. This module
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
// It mirrors the server's rule and nothing wider. On an unstarted table
// the viewer is the host (their seat carries is_host), the game's
// creator (the server stamps is_creator on the viewer's copy of the
// meta), or an admin. On a started one (state "active") only an admin.
// An ended game offers nothing. The seat is a person's, not a bot's
// (bots have their own route and gate). The host's seat is offered to
// the host alone, as leaving it (#3494): a host who sat down by default
// and meant only to run the table.
export function canRemoveSeat(g: GameMeta, seat: SeatInfo, viewer: SeatRemovalViewer): boolean {
  if (g.state !== "lobby" && g.state !== "active") return false;
  if (seat.is_bot) return false;
  if (isOwnSeat(seat, viewer)) {
    if (g.state !== "lobby") return false;
    return viewer.isAdmin || seat.is_host === true || g.is_creator === true;
  }
  if (seat.is_host) return false;
  if (g.state === "active") return viewer.isAdmin;
  return viewer.isAdmin || viewer.mySeat?.is_host === true || g.is_creator === true;
}

/** True when `seat` is the viewer's own. */
export function isOwnSeat(seat: SeatInfo, viewer: SeatRemovalViewer): boolean {
  return viewer.mySeat !== null && viewer.mySeat.player_id === seat.player_id;
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
  "The invite link still works unless you block them below, so mint a new one if you do not want them back.";

// A started game keeps the seat in the list: the player leaves the game
// the way a conceding player does (CR 800.4a), and the engine records it
// as a concession. They cannot reopen the table afterwards.
export const REMOVE_STARTED_SEAT_CONSEQUENCE =
  "This concedes the player out of the game. " +
  "Everything they own leaves with them, and they are disconnected and cannot come back to the table.";

/** The consequence sentence for the confirm on the table `g`. */
export function removeSeatConsequence(g: GameMeta): string {
  return g.state === "active" ? REMOVE_STARTED_SEAT_CONSEQUENCE : REMOVE_SEAT_CONSEQUENCE;
}

// #3494: every seat's × asks first, a bot's and your own included, so a
// misclick costs nothing.
export const LEAVE_SEAT_QUESTION = "Leave your seat at this table?";

/**
 * What leaving your own seat does. The creator keeps the table: no one
 * else is made host, and they take it back by sitting down again.
 * Another host's leaving passes hosting on, as a concession would.
 */
export function leaveSeatConsequence(g: GameMeta, isHost: boolean): string {
  if (g.is_creator) {
    return "Your seat opens up. The table stays yours: you can still invite people, manage the seats and sit back down with Take a seat.";
  }
  if (isHost) return "Your seat opens up, and hosting passes to the next player at the table.";
  return "Your seat opens up. You can sit down again with the invite link.";
}

export const REMOVE_BOT_CONSEQUENCE =
  "The bot leaves and its seat opens up again. You can add another bot.";
