// "This table was closed." (ADR 0211 §4 "Closed means closed", #5082).
//
// An archived table refuses every way in for anyone but an admin, by
// hand or (ADR 0211) by the idle sweep:
//
//   - HTTP: join, spectate, start, a deck, a bot and the setup route
//     answer 409 with `{"error": <message>, "code": "game_closed"}`
//     (lobby.CodeGameClosed). authFetch swaps the server's sentence for
//     GAME_CLOSED_MESSAGE, so every surface that shows a refusal's
//     message says the same thing without matching on it.
//   - WebSocket: the upgrade is refused with 409, which a browser never
//     sees: it reports a bare 1006, and the reconnect ladder would redial
//     it forever. So after a failed dial the Game route asks GET
//     /games/{id} (api.ts tableIsClosed), which still answers a seat or
//     a spectator of an archived table, with `archived_at` set; ws.ts
//     then stops and the connection banner says GAME_CLOSED_MESSAGE.
//
// The code is what the client matches, never the sentence, as for
// serverFull.ts's `server_full`. This module imports only the labels, so
// session.ts can use it without a cycle.
import { L } from "./labels";

/** The `code` the server sends (lobby.CodeGameClosed). */
export const GAME_CLOSED_CODE = "game_closed";

/** What the player reads, wherever the refusal lands. */
export const GAME_CLOSED_MESSAGE: string = L.gameClosed;

/**
 * isGameClosedBody reports whether a parsed error body is the server's
 * closed-table refusal.
 */
export function isGameClosedBody(body: unknown): boolean {
  return (
    typeof body === "object" &&
    body !== null &&
    (body as { code?: unknown }).code === GAME_CLOSED_CODE
  );
}

/**
 * isGameClosed reports whether err is the closed-table refusal: a
 * LobbyApiError (or anything else carrying the parsed body) whose body's
 * code is GAME_CLOSED_CODE.
 */
export function isGameClosed(err: unknown): boolean {
  if (typeof err !== "object" || err === null) return false;
  return isGameClosedBody((err as { body?: unknown }).body);
}
