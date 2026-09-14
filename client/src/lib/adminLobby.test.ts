import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  archiveGame,
  deleteGame,
  listGames,
  mintSeatReclaim,
  redeemSeatReclaim,
  unarchiveGame,
} from "./api";
import { LobbyApiError, currentSession, setSession } from "./session";

// adminLobby.test.ts covers the two admin-lobby capabilities on the
// client side of the wire: archiving a table instead of destroying
// it, and minting / redeeming a seat-reclaim link.
//
// The reclaim half is the one worth testing carefully. Its redeem
// call is deliberately NOT authFetch — the player redeeming has no
// session, so a 401 means "this ticket is no good", and treating it
// as an expired session (clearing the store, bouncing to login)
// would hide the real message at the exact moment it matters.

const GAME = "11111111-1111-1111-1111-111111111111";
const SEAT = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa";

interface Call {
  url: string;
  method: string;
  body: unknown;
}

let calls: Call[] = [];

function stubFetch(response: { status?: number; body?: unknown }): void {
  const status = response.status ?? 200;
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      calls.push({
        url: input,
        method: init.method ?? "GET",
        body: typeof init.body === "string" ? JSON.parse(init.body) : undefined,
      });
      return {
        ok: status >= 200 && status < 300,
        status,
        statusText: "stub",
        json: async () => response.body ?? {},
        clone() {
          return this;
        },
      } as unknown as Response;
    }),
  );
}

beforeEach(() => {
  calls = [];
  setSession({
    token: "admin-tok",
    expiresAt: "2099-01-01T00:00:00Z",
    principal: { role: "admin", issued_at: "", expires_at: "2099-01-01T00:00:00Z" },
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
  setSession(null);
});

describe("archiving", () => {
  it("lists active tables by default and archived ones on request", async () => {
    stubFetch({ body: { games: [] } });
    await listGames();
    await listGames({ archived: true });
    expect(calls.map((c) => c.url)).toEqual(["/games", "/games?archived=1"]);
  });

  it("archives with POST and restores with DELETE on the same path", async () => {
    stubFetch({ body: {} });
    await archiveGame(GAME);
    await unarchiveGame(GAME);
    expect(calls).toMatchObject([
      { url: `/games/${GAME}/archive`, method: "POST" },
      { url: `/games/${GAME}/archive`, method: "DELETE" },
    ]);
  });

  it("keeps hard delete on a different route from archive", async () => {
    stubFetch({ status: 204, body: {} });
    await deleteGame(GAME);
    expect(calls).toMatchObject([{ url: `/games/${GAME}`, method: "DELETE" }]);
  });

  it("surfaces the server's reason rather than a generic failure", async () => {
    stubFetch({ status: 404, body: { error: "lobby: game not found" } });
    await expect(archiveGame(GAME)).rejects.toThrow("lobby: game not found");
  });
});

describe("seat reclaim", () => {
  it("mints against the seat, carrying the admin session", async () => {
    stubFetch({
      body: {
        ticket: "s3cret",
        game_id: GAME,
        player_id: SEAT,
        seat: 0,
        player_name: "Alice",
        expires_at: "2026-09-14T12:15:00Z",
        ttl_seconds: 900,
        single_use: true,
      },
    });
    const ticket = await mintSeatReclaim(GAME, SEAT);
    expect(calls[0]).toMatchObject({
      url: `/games/${GAME}/seats/${SEAT}/reclaim`,
      method: "POST",
    });
    expect(ticket.single_use).toBe(true);
    expect(ticket.ttl_seconds).toBe(900);
  });

  it("redeems with only the ticket and installs the returned seat session", async () => {
    // The redeeming player has no session — that is the whole point.
    setSession(null);
    stubFetch({
      body: {
        token: "seat-tok",
        expires_at: "2099-01-01T00:00:00Z",
        principal: {
          role: "player",
          game_id: GAME,
          player_id: SEAT,
          issued_at: "",
          expires_at: "2099-01-01T00:00:00Z",
        },
        player_id: SEAT,
        game: { id: GAME, name: "FNM", created_at: "", players: [], state: "active" },
      },
    });

    const s = await redeemSeatReclaim(GAME, "s3cret");
    expect(calls[0]).toMatchObject({
      url: `/games/${GAME}/reclaim`,
      method: "POST",
      body: { ticket: "s3cret" },
    });
    // No name, no invite token: the ticket names the seat by itself.
    expect(Object.keys(calls[0].body as object)).toEqual(["ticket"]);
    expect(s.playerID).toBe(SEAT);
    expect(s.gameID).toBe(GAME);
    expect(currentSession()?.token).toBe("seat-tok");
  });

  it("reports a spent or expired ticket as the server described it", async () => {
    setSession(null);
    stubFetch({ status: 401, body: { error: "lobby: invalid or expired reclaim link" } });
    await expect(redeemSeatReclaim(GAME, "used-already")).rejects.toThrow(
      "lobby: invalid or expired reclaim link",
    );
  });

  it("does not clear an existing session when a ticket is refused", async () => {
    // A host testing a link in their own browser must not get logged
    // out by a 401 that was about the ticket, not about them.
    stubFetch({ status: 401, body: { error: "lobby: invalid or expired reclaim link" } });
    await expect(redeemSeatReclaim(GAME, "nope")).rejects.toBeInstanceOf(LobbyApiError);
    expect(currentSession()?.token).toBe("admin-tok");
  });
});
