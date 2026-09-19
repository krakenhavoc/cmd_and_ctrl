import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";

import { logoutEverywhere } from "./api";
import {
  LobbyApiError,
  MAX_TIMER_MS,
  canSignOutEverywhere,
  currentSession,
  expiryNotice,
  sessionFromOAuth,
  setSession,
  type Session,
} from "./session";

// signOutEverywhere.test.ts covers the client half of S34 sub-PR 7
// (ADR 0051 decisions 3 and 6): a Discord sign-in now lives 30 days
// and must not be dropped early by the expiry timer, and "sign out
// everywhere" is offered, and behaves, only for a session with a user.

const USER = "0b7c9d7e-2f11-4c43-9f5e-8a8e2c1d4b6a";
const DAY = 24 * 60 * 60 * 1000;

function identitySession(expiresAt: string, userID?: string): Session {
  return sessionFromOAuth({ token: "id-tok", expiresAt, displayName: "Alice", userID });
}

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
  setSession(null);
  expiryNotice.set("");
});

describe("canSignOutEverywhere", () => {
  it("is true only for a session tied to a user", () => {
    expect(canSignOutEverywhere(null)).toBe(false);
    expect(canSignOutEverywhere(identitySession("2099-01-01T00:00:00Z", USER))).toBe(true);
    // A Discord sign-in on a server with no user database has no id.
    expect(canSignOutEverywhere(identitySession("2099-01-01T00:00:00Z"))).toBe(false);
    const admin: Session = {
      token: "a",
      expiresAt: "2099-01-01T00:00:00Z",
      principal: { role: "admin", issued_at: "", expires_at: "2099-01-01T00:00:00Z" },
    };
    expect(canSignOutEverywhere(admin)).toBe(false);
  });
});

describe("sessionFromOAuth", () => {
  it("builds the identity-only session with its user id", () => {
    const now = new Date("2026-09-19T12:00:00Z");
    const s = sessionFromOAuth(
      { token: "t", expiresAt: "2026-10-19T12:00:00Z", displayName: "Alice", userID: USER },
      now,
    );
    expect(s.principal).toEqual({
      role: "identified",
      user_id: USER,
      game_id: undefined,
      player_id: undefined,
      name: "Alice",
      issued_at: now.toISOString(),
      expires_at: "2026-10-19T12:00:00Z",
    });
  });

  it("is a player session when the seat was claimed in the callback", () => {
    const s = sessionFromOAuth({
      token: "t",
      expiresAt: "2026-09-20T00:00:00Z",
      gameID: "g",
      playerID: "p",
      userID: USER,
    });
    expect(s.principal.role).toBe("player");
    expect(s.gameID).toBe("g");
    expect(s.playerID).toBe("p");
    expect(s.principal.user_id).toBe(USER);
  });
});

describe("expiry timer", () => {
  it("keeps a 30-day identity session past setTimeout's clamp, then expires it on time", () => {
    vi.useFakeTimers();
    const start = new Date("2026-09-19T12:00:00Z");
    vi.setSystemTime(start);
    const expires = new Date(start.getTime() + 30 * DAY);
    setSession(identitySession(expires.toISOString(), USER));

    // The clamped timer fires at ~23 days. The session is still good,
    // so it must survive.
    vi.advanceTimersByTime(MAX_TIMER_MS + 1000);
    expect(currentSession()).not.toBeNull();
    expect(get(expiryNotice)).toBe("");

    // A minute before expiry: still there. (Fake timers move Date.now
    // along with them.)
    vi.advanceTimersByTime(expires.getTime() - 60_000 - Date.now());
    expect(currentSession()).not.toBeNull();

    // Past expiry: gone, with the notice.
    vi.advanceTimersByTime(2 * 60_000);
    expect(currentSession()).toBeNull();
    expect(get(expiryNotice)).toMatch(/expired/);
  });

  it("still expires a 12h session at 12h", () => {
    vi.useFakeTimers();
    const start = new Date("2026-09-19T12:00:00Z");
    vi.setSystemTime(start);
    setSession(identitySession(new Date(start.getTime() + 12 * 3600_000).toISOString()));
    vi.advanceTimersByTime(12 * 3600_000 - 1000);
    expect(currentSession()).not.toBeNull();
    vi.advanceTimersByTime(2000);
    expect(currentSession()).toBeNull();
  });
});

describe("logoutEverywhere", () => {
  let calls: { url: string; method: string; auth: string | null }[] = [];

  function stubFetch(status: number, body: unknown = {}): void {
    vi.stubGlobal(
      "fetch",
      vi.fn(async (input: string, init: RequestInit = {}) => {
        const headers = new Headers(init.headers);
        calls.push({
          url: input,
          method: init.method ?? "GET",
          auth: headers.get("Authorization"),
        });
        return {
          ok: status >= 200 && status < 300,
          status,
          statusText: "stub",
          json: async () => body,
        } as unknown as Response;
      }),
    );
  }

  beforeEach(() => {
    calls = [];
    setSession(identitySession("2099-01-01T00:00:00Z", USER));
  });

  it("posts to /logout/everywhere with the session and drops it", async () => {
    stubFetch(204);
    await logoutEverywhere();
    expect(calls).toEqual([{ url: "/logout/everywhere", method: "POST", auth: "Bearer id-tok" }]);
    expect(currentSession()).toBeNull();
  });

  it("drops a session the server says is already dead", async () => {
    stubFetch(401, { error: "session revoked" });
    await logoutEverywhere();
    expect(currentSession()).toBeNull();
  });

  it("reports a failure and keeps the session, since the other browsers are still signed in", async () => {
    stubFetch(503, { error: "session revocation needs the user database" });
    const err = await logoutEverywhere().catch((e: unknown) => e);
    expect(err).toBeInstanceOf(LobbyApiError);
    expect((err as LobbyApiError).status).toBe(503);
    expect((err as LobbyApiError).message).toMatch(/user database/);
    expect(currentSession()).not.toBeNull();
  });
});
