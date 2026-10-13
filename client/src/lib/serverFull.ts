// "The server is full right now" (ADR 0148, #2977).
//
// The server holds itself to caps on its load, its WebSocket
// connections and the rate tables are created at. When a request would
// cross one it answers with a stable code instead of a raw error:
//
//   - HTTP: 503 with `{"error": <message>, "code": "server_full"}` and a
//     Retry-After, on creating a table, opening a practice table,
//     seating a bot, and joining or spectating while the connection cap
//     is full;
//   - WebSocket: an `error` frame with the same code, then close 1013,
//     to a newcomer with no seat (a browser cannot read the body of a
//     refused handshake, so the upgrade is accepted for the frame).
//
// The code is what the client matches, never the sentence. This module
// owns the sentence the player reads, so a proxy that swallows the body
// still gets the friendly message, and the four places that show a
// create / join failure cannot word it four ways.
//
// A table that already exists is never refused: a seated player
// reconnecting always gets in, and none of this reaches a running game.
import type { Writable } from "svelte/store";
import { guardedWritable } from "./guardedStore";
import { LobbyApiError } from "./session";

/** The `code` the server sends. */
export const SERVER_FULL_CODE = "server_full";

/** What the player reads. Keep in step with capacity.Message (Go). */
export const SERVER_FULL_MESSAGE =
  "The server is full right now. Please try again in a few minutes.";

/** The WebSocket close code of a refusal: 1013, "try again later". */
export const CLOSE_TRY_AGAIN_LATER = 1013;

/** isServerFull reports whether err is the server's capacity refusal. */
export function isServerFull(err: unknown): boolean {
  if (!(err instanceof LobbyApiError)) return false;
  const body = err.body as { code?: unknown } | undefined;
  return body?.code === SERVER_FULL_CODE;
}

/**
 * lobbyErrorMessage is the text to show for a failed create / join:
 * the friendly sentence for a capacity refusal, the server's own
 * message for any other lobby error, `fallback` for anything else.
 */
export function lobbyErrorMessage(err: unknown, fallback: string): string {
  if (isServerFull(err)) return SERVER_FULL_MESSAGE;
  return err instanceof LobbyApiError ? err.message : fallback;
}

/**
 * serverFull is true from the moment a socket is refused at the
 * connection cap until a snapshot arrives. GameClient sets and clears
 * it; the connection banner words itself from it.
 */
export const serverFull: Writable<boolean> = guardedWritable(false, "serverFull");
