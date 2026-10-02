// Construction of the game WebSocket URL, extracted from Game.svelte
// so the seat-selection rules can be pinned by tests. They are the
// kind of rules that look obvious and are easy to get subtly wrong:
// which seat a connection renders decides which hand you can see.

import { isAdmin } from "./admin";
import type { Session } from "./session";

export interface GameWSTargetOpts {
  // baseURL is the ws:// or wss:// endpoint, e.g. "wss://host/ws".
  baseURL: string;
  gameID: string;
  session: Session | null;
  // seatOverride is the develop environment's seat swap (ADR 0023):
  // the seat an admin wants to view and act as, or null for the
  // admin spectator view. Ignored for every non-admin session.
  seatOverride?: string | null;
}

// canSwapSeats reports whether seat swapping is meaningful for this
// session.
//
// Only admin sessions: the shared token, or a signed-in person on the
// server's admin allowlist (isAdmin, ADR 0110 §3). This is not a
// security decision — the server makes that one, and makes it the safe
// way round: WSAuthorizer returns a non-admin player's own PlayerID and
// ignores ?player= entirely, so a player who forged the parameter would
// simply see their own seat. This function exists so we don't render a
// control that silently does nothing.
export function canSwapSeats(session: Session | null): boolean {
  return isAdmin(session);
}

// gameWSURL builds the connection URL for one seat's view of a game.
//
// Seat precedence:
//   1. An admin's seat override, when one is set.
//   2. The session's own bound seat, when it belongs to this game.
//   3. No ?player= at all — the spectator view, where every opponent
//      hand renders as a hidden count.
//
// Binding as a seat means the connection ACTS as that seat too: the
// hub stamps the bound player onto every action as its caller. An
// admin who picks a seat therefore gives up the admin bypass on
// player-scoped actions and gets that seat's real permissions, which
// is what you want when the point is to test what that seat can do.
// The one exception is the admin context menu's card overrides (tap,
// move, counters, position), which an admin connection keeps in any
// seat, its own included (ADR 0110 §3, server actions.overrideCaller).
export function gameWSURL(opts: GameWSTargetOpts): string {
  const { baseURL, gameID, session, seatOverride } = opts;
  const params = new URLSearchParams();
  params.set("game", gameID);
  if (session?.token) params.set("token", session.token);

  const seat =
    canSwapSeats(session) && seatOverride
      ? seatOverride
      : session?.playerID && session.gameID === gameID
        ? session.playerID
        : null;
  if (seat) params.set("player", seat);

  return `${baseURL}?${params.toString()}`;
}
