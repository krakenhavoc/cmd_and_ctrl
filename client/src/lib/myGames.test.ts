import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { discordLinkHref, fetchMyGames, rejoinMyGame } from "./api";
import {
  canLinkDiscord,
  linkDiscordLabel,
  myGameStatus,
  othersLabel,
  playedWhen,
  signedInUserID,
  sortMyGames,
  type MyGame,
} from "./myGames";
import { currentSession, setSession, type Session } from "./session";

// myGames.test.ts covers the client half of S34 sub-PR 4 (ADR 0051):
// who counts as signed in, when "Link Discord" is offered, the labels
// the My games page renders, and the two calls it makes.

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const NIL = "00000000-0000-0000-0000-000000000000";

function sessionAs(role: Session["principal"]["role"], userID?: string): Session {
  return {
    token: "tok",
    expiresAt: new Date(Date.now() + 3_600_000).toISOString(),
    principal: {
      role,
      user_id: userID,
      issued_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 3_600_000).toISOString(),
    },
  };
}

function game(over: Partial<MyGame> = {}): MyGame {
  return {
    id: "g1",
    name: "Friday",
    state: "active",
    seat: 0,
    winner_seat: null,
    created_at: Date.UTC(2026, 8, 1),
    started_at: null,
    ended_at: null,
    archived_at: null,
    others: [],
    ...over,
  };
}

describe("signedInUserID", () => {
  it("is the user id for an identity or player session that has one", () => {
    expect(signedInUserID(sessionAs("identified", USER))).toBe(USER);
    expect(signedInUserID(sessionAs("player", USER))).toBe(USER);
  });

  it("is null for guests, admins, spectators and no session", () => {
    // A guest's principal spells "no user" as the nil uuid.
    expect(signedInUserID(sessionAs("player", NIL))).toBeNull();
    expect(signedInUserID(sessionAs("player"))).toBeNull();
    expect(signedInUserID(sessionAs("admin", USER))).toBeNull();
    expect(signedInUserID(sessionAs("spectator", USER))).toBeNull();
    expect(signedInUserID(null)).toBeNull();
  });
});

describe("canLinkDiscord", () => {
  it("is offered to a seated human player when Discord is configured", () => {
    expect(canLinkDiscord({ role: "player", discordEnabled: true, isBotSeat: false })).toBe(true);
  });

  it("is withheld without Discord, from bot seats, admins and spectators", () => {
    expect(canLinkDiscord({ role: "player", discordEnabled: false, isBotSeat: false })).toBe(false);
    expect(canLinkDiscord({ role: "player", discordEnabled: true, isBotSeat: true })).toBe(false);
    expect(canLinkDiscord({ role: "admin", discordEnabled: true, isBotSeat: false })).toBe(false);
    expect(canLinkDiscord({ role: "spectator", discordEnabled: true, isBotSeat: false })).toBe(
      false,
    );
    expect(canLinkDiscord({ role: undefined, discordEnabled: true, isBotSeat: false })).toBe(false);
  });

  it("names a relink for a seat that already has an account", () => {
    expect(linkDiscordLabel(false)).toBe("Link Discord");
    expect(linkDiscordLabel(true)).toMatch(/different Discord account/);
  });
});

describe("myGameStatus", () => {
  it("reads the table's state", () => {
    expect(myGameStatus(game({ state: "lobby" })).tone).toBe("lobby");
    expect(myGameStatus(game({ state: "active" }))).toEqual({ label: "in progress", tone: "live" });
    expect(myGameStatus(game({ state: "ended" }))).toEqual({ label: "ended", tone: "ended" });
  });

  it("says who won, and says so to the winner", () => {
    const others = [{ seat: 1, name: "Bob" }];
    expect(myGameStatus(game({ state: "ended", seat: 0, winner_seat: 0, others })).label).toBe(
      "you won",
    );
    expect(myGameStatus(game({ state: "ended", seat: 0, winner_seat: 1, others })).label).toBe(
      "Bob won",
    );
  });

  it("puts archived ahead of the state", () => {
    expect(myGameStatus(game({ state: "active", archived_at: 1 })).tone).toBe("archived");
  });
});

describe("othersLabel", () => {
  it("names the rest of the table in seat order, marking bots", () => {
    const g = game({
      others: [
        { seat: 3, name: "Ghoul", bot: true },
        { seat: 1, name: "Bob" },
        { seat: 2, name: "Carol" },
      ],
    });
    expect(othersLabel(g)).toBe("with Bob, Carol and Ghoul (bot)");
    expect(othersLabel(game({ others: [{ seat: 1, name: "Bob" }] }))).toBe("with Bob");
    expect(othersLabel(game())).toBe("nobody else yet");
  });
});

describe("playedWhen", () => {
  const now = new Date(2026, 8, 19, 15, 0, 0).getTime();
  it("is relative while recent, from the latest thing that happened", () => {
    expect(playedWhen(game({ created_at: new Date(2026, 8, 19, 9).getTime() }), now)).toBe("today");
    expect(playedWhen(game({ created_at: new Date(2026, 8, 18, 22).getTime() }), now)).toBe(
      "yesterday",
    );
    expect(
      playedWhen(
        game({
          created_at: new Date(2026, 7, 1).getTime(),
          ended_at: new Date(2026, 8, 16, 20).getTime(),
        }),
        now,
      ),
    ).toBe("3 days ago");
  });

  it("is a date after a week", () => {
    expect(playedWhen(game({ created_at: new Date(2026, 6, 4).getTime() }), now)).toMatch(/2026/);
  });
});

describe("sortMyGames", () => {
  it("is newest first", () => {
    const a = game({ id: "a", created_at: 1 });
    const b = game({ id: "b", created_at: 3 });
    const c = game({ id: "c", created_at: 2 });
    expect(sortMyGames([a, b, c]).map((g) => g.id)).toEqual(["b", "c", "a"]);
  });
});

// --- the two calls ---------------------------------------------------

interface Call {
  url: string;
  method: string;
  auth: string | null;
}
let calls: Call[] = [];

function stubFetch(status: number, body: unknown): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      calls.push({
        url: input,
        method: init.method ?? "GET",
        auth: new Headers(init.headers).get("Authorization"),
      });
      return {
        ok: status >= 200 && status < 300,
        status,
        statusText: "stub",
        json: async () => body,
        clone() {
          return this;
        },
      } as unknown as Response;
    }),
  );
}

describe("My games API", () => {
  beforeEach(() => {
    calls = [];
    setSession(sessionAs("identified", USER));
  });
  afterEach(() => {
    vi.unstubAllGlobals();
    setSession(null);
  });

  it("fetchMyGames reads GET /me/games with the session", async () => {
    stubFetch(200, { games: [game({ id: "x" })] });
    const games = await fetchMyGames();
    expect(games.map((g) => g.id)).toEqual(["x"]);
    expect(calls).toEqual([{ url: "/me/games", method: "GET", auth: "Bearer tok" }]);
  });

  it("rejoinMyGame posts to the rejoin path and installs the seat session", async () => {
    stubFetch(200, {
      token: "seat-tok",
      expires_at: "2026-09-20T00:00:00Z",
      principal: {
        role: "player",
        user_id: USER,
        game_id: "g1",
        player_id: "p1",
        issued_at: "2026-09-19T12:00:00Z",
        expires_at: "2026-09-20T00:00:00Z",
      },
      game: { id: "g1" },
      player_id: "p1",
    });
    await rejoinMyGame("/me/games/g1/session");
    expect(calls[0]).toMatchObject({ url: "/me/games/g1/session", method: "POST" });
    const s = currentSession();
    expect(s?.token).toBe("seat-tok");
    expect(s?.gameID).toBe("g1");
    expect(s?.playerID).toBe("p1");
    expect(signedInUserID(s)).toBe(USER);
  });

  it("discordLinkHref names the table being linked", () => {
    expect(discordLinkHref("g 1")).toBe("/auth/discord/link?game=g%201");
  });
});
