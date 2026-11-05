import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

import {
  addBotSeat,
  applyLastSetup,
  joinGame,
  seatLibraryDeck,
  spectateGame,
  startGame,
  tableIsClosed,
  uploadDeck,
} from "./api";
import { GAME_CLOSED_CODE, GAME_CLOSED_MESSAGE, isGameClosed } from "./gameClosed";
import { L } from "./labels";
import { lobbyErrorMessage } from "./serverFull";
import { LobbyApiError, session, type Session } from "./session";

// #5082, ADR 0211 §4 "Closed means closed": every way into an archived
// table answers 409 with code "game_closed", and the client says "This
// table was closed." for it, wherever the refusal lands. The server's
// own sentence ("lobby: game is archived") never reaches a player.

const GAME = "11111111-2222-3333-4444-555555555555";
const SERVER_SENTENCE = "lobby: game is archived";

function closedResponse(): Response {
  return new Response(JSON.stringify({ error: SERVER_SENTENCE, code: GAME_CLOSED_CODE }), {
    status: 409,
    statusText: "Conflict",
    headers: { "Content-Type": "application/json" },
  });
}

function jsonResponse(status: number, body: unknown): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function seatSession(): Session {
  const expiresAt = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token: "seat-token",
    expiresAt,
    principal: {
      role: "player",
      game_id: GAME,
      player_id: "p-1",
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
    playerID: "p-1",
    gameID: GAME,
  };
}

beforeEach(() => {
  session.set(null);
});

afterEach(() => {
  session.set(null);
  vi.unstubAllGlobals();
});

describe("the closed-table message", () => {
  it("is the registered label", () => {
    expect(GAME_CLOSED_MESSAGE).toBe(L.gameClosed);
    expect(GAME_CLOSED_MESSAGE).toBe("This table was closed.");
  });
});

// One case per HTTP path ADR 0211 §4 closes: join, spectate, start, a
// deck (pasted and from the library), a bot and the setup route.
describe("an HTTP refusal with code game_closed", () => {
  const paths: [string, () => Promise<unknown>][] = [
    ["join", () => joinGame(GAME, "invite", "Bob")],
    ["spectate", () => spectateGame(GAME, "spectator-invite", "watcher")],
    ["start", () => startGame(GAME)],
    ["deck upload", () => uploadDeck(GAME, "p-1", "1 Sol Ring")],
    ["library deck", () => seatLibraryDeck(GAME, "deck-1")],
    ["bot", () => addBotSeat(GAME, { tier: "random" })],
    ["setup", () => applyLastSetup(GAME)],
  ];

  for (const [name, call] of paths) {
    it(`${name}: says "${L.gameClosed}"`, async () => {
      session.set(seatSession());
      vi.stubGlobal(
        "fetch",
        vi.fn(async () => closedResponse()),
      );
      const err = await call().then(
        () => null,
        (e: unknown) => e,
      );
      expect(err).toBeInstanceOf(LobbyApiError);
      const apiErr = err as LobbyApiError;
      expect(apiErr.status).toBe(409);
      expect(apiErr.message).toBe(L.gameClosed);
      expect(isGameClosed(apiErr)).toBe(true);
      // What Join, the bot picker and the create form render.
      expect(lobbyErrorMessage(apiErr, "join failed")).toBe(L.gameClosed);
      // A 409 is not a dead session: the tab stays signed in.
      expect(get(session)?.token).toBe("seat-token");
    });
  }

  it("leaves any other 409 to the server's own sentence", async () => {
    session.set(seatSession());
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(409, { error: "the game has already started" })),
    );
    const err = (await startGame(GAME).catch((e: unknown) => e)) as LobbyApiError;
    expect(err.message).toBe("the game has already started");
    expect(isGameClosed(err)).toBe(false);
  });
});

// The WebSocket half asks GET /games/{id} after a failed dial, because
// the browser only ever sees a refused upgrade as a 1006.
describe("tableIsClosed", () => {
  it("is true when the table carries archived_at", async () => {
    session.set(seatSession());
    const fetchMock = vi.fn(async () =>
      jsonResponse(200, { id: GAME, archived_at: "2026-11-05T10:00:00Z" }),
    );
    vi.stubGlobal("fetch", fetchMock);
    expect(await tableIsClosed(GAME)).toBe(true);
    const [url, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit];
    expect(url).toBe(`/games/${GAME}`);
    expect(new Headers(init.headers).get("Authorization")).toBe("Bearer seat-token");
  });

  it("is false for a table that is still open", async () => {
    session.set(seatSession());
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(200, { id: GAME, state: "active" })),
    );
    expect(await tableIsClosed(GAME)).toBe(false);
  });

  it("is true for a 409 game_closed, should the read ever refuse", async () => {
    session.set(seatSession());
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => closedResponse()),
    );
    expect(await tableIsClosed(GAME)).toBe(true);
  });

  it("is false for a deleted table, a server error and a network error", async () => {
    session.set(seatSession());
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(404, { error: "game not found" })),
    );
    expect(await tableIsClosed(GAME)).toBe(false);
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => new Response("bad gateway", { status: 502 })),
    );
    expect(await tableIsClosed(GAME)).toBe(false);
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => {
        throw new TypeError("network error");
      }),
    );
    expect(await tableIsClosed(GAME)).toBe(false);
  });

  // The dead-session probe owns "sign this tab out"; this question must
  // not answer it on the side.
  it("is false on a 401 and leaves the session alone", async () => {
    session.set(seatSession());
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => jsonResponse(401, { error: "session expired" })),
    );
    expect(await tableIsClosed(GAME)).toBe(false);
    expect(get(session)?.token).toBe("seat-token");
  });
});
