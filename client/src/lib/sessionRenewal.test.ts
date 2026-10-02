// @vitest-environment jsdom
//
// ADR 0110 Delivery PR 2 on the client: renewal on use (§1 item 5,
// owner answer 1) and the saved identity (§1 item 6).
//
//   - A signed-in session past half its life is renewed through POST
//     /me/session, at page load and on the half-life timer.
//   - A failed renewal is quiet: the session stays until it expires.
//   - A session with no user installed over a signed-in one keeps the
//     signed-in one aside, and puts it back when it ends.
//
// jsdom for localStorage. The router is mocked (a plain store the test
// drives) and the network is a stub that records every call.

import { afterEach, beforeEach, describe, expect, it, vi, type Mock } from "vitest";
import { get, writable, type Writable } from "svelte/store";
import type { Session } from "./session";
import type { Route } from "./router";

const SESSION_KEY = "cmdctrl.session";
const IDENTITY_KEY = "cmdctrl.identity";
const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const HOUR = 60 * 60 * 1000;
const DAY = 24 * HOUR;

// signedIn is a signed-in session issued `ageMs` ago that lives `lifeMs`.
function signedIn(
  token: string,
  role: Session["principal"]["role"],
  ageMs: number,
  lifeMs: number,
  extra: Partial<Session> = {},
): Session {
  const issued = Date.now() - ageMs;
  const expires = new Date(issued + lifeMs).toISOString();
  return {
    token,
    expiresAt: expires,
    principal: {
      role,
      user_id: USER,
      name: "Alice",
      game_id: role === "identified" ? undefined : "g1",
      player_id: role === "player" ? "p1" : undefined,
      issued_at: new Date(issued).toISOString(),
      expires_at: expires,
    },
    ...extra,
  };
}

// noUser is an admin-token session, which carries no user.
function noUser(token: string, lifeMs: number): Session {
  const expires = new Date(Date.now() + lifeMs).toISOString();
  return {
    token,
    expiresAt: expires,
    principal: { role: "admin", issued_at: new Date().toISOString(), expires_at: expires },
  };
}

interface Call {
  url: string;
  method: string;
  auth: string | null;
}

let calls: Call[] = [];
// reply decides POST /me/session's answer for the bearer it was sent.
let reply: (bearer: string) => Response | Promise<Response>;

function sessionBody(token: string, role: Session["principal"]["role"], lifeMs: number): Response {
  const now = Date.now();
  const expires = new Date(now + lifeMs).toISOString();
  return new Response(
    JSON.stringify({
      token,
      expires_at: expires,
      principal: {
        role,
        user_id: USER,
        name: "Alice",
        game_id: role === "identified" ? undefined : "g1",
        player_id: role === "player" ? "p1" : undefined,
        issued_at: new Date(now).toISOString(),
        expires_at: expires,
      },
    }),
    { status: 200, headers: { "Content-Type": "application/json" } },
  );
}

function stubFetch(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const call = {
        url: input,
        method: init.method ?? "GET",
        auth: new Headers(init.headers).get("Authorization"),
      };
      calls.push(call);
      if (call.url === "/me/session") return reply((call.auth ?? "").replace(/^Bearer /, ""));
      if (call.url === "/logout") return new Response(null, { status: 204 });
      // Anything else answers 401, the server's word that the
      // credential is gone.
      return new Response(JSON.stringify({ error: "session expired" }), { status: 401 });
    }),
  );
}

// load imports the modules fresh over whatever localStorage holds: a
// new page.
async function load() {
  vi.resetModules();
  const route = writable<Route>({ name: "lobby" });
  const navigate = vi.fn();
  vi.doMock("./router", () => ({ route, navigate }));
  const sessionMod = await import("./session");
  const api = await import("./api");
  return {
    ...sessionMod,
    logout: api.logout,
    route: route as Writable<Route>,
    navigate: navigate as Mock,
  };
}

const renewals = () => calls.filter((c) => c.url === "/me/session");

// settle lets the stubbed fetch and the code after it run.
async function settle(): Promise<void> {
  for (let i = 0; i < 5; i++) await Promise.resolve();
  await vi.advanceTimersByTimeAsync(0);
}

beforeEach(() => {
  vi.useFakeTimers();
  localStorage.clear();
  calls = [];
  reply = () => new Response("{}", { status: 500 });
  stubFetch();
});

afterEach(() => {
  vi.useRealTimers();
  vi.unstubAllGlobals();
});

describe("renewal on use", () => {
  it("renews a signed-in session past half its life at page load", async () => {
    const old = signedIn("old-seat", "player", 16 * DAY, 30 * DAY, {
      gameID: "g1",
      playerID: "p1",
    });
    localStorage.setItem(SESSION_KEY, JSON.stringify(old));
    reply = () => sessionBody("new-seat", "player", 30 * DAY);

    const m = await load();
    await settle();

    expect(renewals()).toHaveLength(1);
    expect(renewals()[0].method).toBe("POST");
    expect(renewals()[0].auth).toBe("Bearer old-seat");
    const now = m.currentSession();
    expect(now?.token).toBe("new-seat");
    // The same seat: the fields the client stored for it are kept.
    expect(now?.gameID).toBe("g1");
    expect(now?.playerID).toBe("p1");
    expect(JSON.parse(localStorage.getItem(SESSION_KEY) ?? "{}").token).toBe("new-seat");
  });

  it("does not call before half-life, and calls when the half-life timer fires", async () => {
    const fresh = signedIn("id-tok", "identified", 1 * DAY, 30 * DAY);
    localStorage.setItem(SESSION_KEY, JSON.stringify(fresh));
    reply = () => sessionBody("renewed", "identified", 30 * DAY);

    const m = await load();
    await settle();
    expect(renewals()).toHaveLength(0);
    expect(m.pastHalfLife(fresh)).toBe(false);

    await vi.advanceTimersByTimeAsync(13 * DAY);
    expect(renewals()).toHaveLength(0);
    await vi.advanceTimersByTimeAsync(1 * DAY + 1000);
    await settle();
    expect(renewals()).toHaveLength(1);
    expect(m.currentSession()?.token).toBe("renewed");
  });

  it("never renews a session with no user", async () => {
    localStorage.setItem(SESSION_KEY, JSON.stringify(noUser("admin-tok", 12 * HOUR)));
    const m = await load();
    await vi.advanceTimersByTimeAsync(11 * HOUR);
    await settle();
    expect(renewals()).toHaveLength(0);
    expect(m.currentSession()?.token).toBe("admin-tok");
  });

  it("a failed renewal is quiet and is retried an hour later", async () => {
    const old = signedIn("old-tok", "identified", 20 * DAY, 30 * DAY);
    localStorage.setItem(SESSION_KEY, JSON.stringify(old));
    let n = 0;
    reply = () => {
      n++;
      if (n === 1) throw new TypeError("network down");
      if (n === 2)
        return new Response(JSON.stringify({ error: "session revoked" }), { status: 401 });
      return sessionBody("renewed", "identified", 30 * DAY);
    };

    const m = await load();
    await settle();
    expect(renewals()).toHaveLength(1);
    expect(m.currentSession()?.token).toBe("old-tok");
    expect(get(m.expiryNotice)).toBe("");

    // A 401 on the renewal itself is still quiet: the session lasts
    // until it expires, and nothing signs the tab out.
    await vi.advanceTimersByTimeAsync(m.RENEW_RETRY_MS);
    await settle();
    expect(renewals()).toHaveLength(2);
    expect(m.currentSession()?.token).toBe("old-tok");
    expect(get(m.expiryNotice)).toBe("");

    await vi.advanceTimersByTimeAsync(m.RENEW_RETRY_MS);
    await settle();
    expect(renewals()).toHaveLength(3);
    expect(m.currentSession()?.token).toBe("renewed");
  });

  it("a no-op answer (the same token) asks again an hour later", async () => {
    const old = signedIn("same-tok", "identified", 16 * DAY, 30 * DAY);
    localStorage.setItem(SESSION_KEY, JSON.stringify(old));
    reply = () => sessionBody("same-tok", "identified", 14 * DAY);
    const m = await load();
    await settle();
    expect(renewals()).toHaveLength(1);
    expect(m.currentSession()?.token).toBe("same-tok");
    await vi.advanceTimersByTimeAsync(m.RENEW_RETRY_MS);
    await settle();
    expect(renewals()).toHaveLength(2);
  });

  it("does not adopt another session the cookie holds (another tab's seat)", async () => {
    const old = signedIn("my-id", "identified", 16 * DAY, 30 * DAY);
    localStorage.setItem(SESSION_KEY, JSON.stringify(old));
    // The server read the cookie, which another tab moved to a seat.
    reply = () => sessionBody("other-tabs-seat", "player", 20 * DAY);
    const m = await load();
    await settle();
    expect(renewals()).toHaveLength(1);
    expect(m.currentSession()?.token).toBe("my-id");
  });

  it("a request sent during a renewal waits for it, and goes out on the renewed token", async () => {
    const old = signedIn("old-tok", "identified", 16 * DAY, 30 * DAY);
    localStorage.setItem(SESSION_KEY, JSON.stringify(old));
    let answer: (r: Response) => void = () => {};
    reply = () => new Promise<Response>((resolve) => (answer = resolve));
    const m = await load();
    await settle();
    expect(renewals()).toHaveLength(1);

    const pending = m.authFetch("/me").catch(() => {});
    await settle();
    expect(calls.filter((c) => c.url === "/me")).toHaveLength(0);

    answer(sessionBody("new-tok", "identified", 30 * DAY));
    await settle();
    await pending;
    const me = calls.filter((c) => c.url === "/me");
    expect(me).toHaveLength(1);
    expect(me[0].auth).toBe("Bearer new-tok");
  });
});

describe("the saved identity", () => {
  it("keeps a signed-in session aside under an admin session, and reinstalls it at expiry", async () => {
    const m = await load();
    const identity = signedIn("id-tok", "identified", 1 * DAY, 30 * DAY);
    m.setSession(identity);
    m.setSession(noUser("admin-tok", 1 * HOUR));
    expect(JSON.parse(localStorage.getItem(IDENTITY_KEY) ?? "{}").token).toBe("id-tok");

    m.route.set({ name: "game", gameID: "someone-elses-game" });
    reply = (bearer) =>
      bearer === "id-tok"
        ? sessionBody("id-tok", "identified", 29 * DAY)
        : new Response("{}", { status: 401 });
    await vi.advanceTimersByTimeAsync(1 * HOUR + 1000);
    await settle();

    expect(renewals()).toHaveLength(1);
    expect(renewals()[0].auth).toBe("Bearer id-tok");
    expect(m.currentSession()?.token).toBe("id-tok");
    expect(m.currentSession()?.principal.user_id).toBe(USER);
    expect(get(m.expiryNotice)).toBe("");
    // Back on the person's own session, so off the admin's table.
    expect(m.navigate).toHaveBeenCalledWith("#/lobby");
    // It is the live session again, not a saved one.
    expect(localStorage.getItem(IDENTITY_KEY)).toBeNull();
  });

  it("reinstalls it when the server answers 401 for the session on top", async () => {
    const m = await load();
    m.setSession(signedIn("id-tok", "identified", 1 * DAY, 30 * DAY));
    m.setSession(noUser("admin-tok", 12 * HOUR));
    reply = () => sessionBody("id-tok", "identified", 29 * DAY);

    await expect(m.authFetch("/games")).rejects.toThrow("session expired");
    await settle();
    expect(m.currentSession()?.token).toBe("id-tok");
  });

  it("signs out to the login page when the saved identity is refused", async () => {
    const m = await load();
    m.setSession(signedIn("id-tok", "identified", 1 * DAY, 30 * DAY));
    m.setSession(noUser("admin-tok", 1 * HOUR));
    reply = () => new Response(JSON.stringify({ error: "session revoked" }), { status: 401 });

    await vi.advanceTimersByTimeAsync(1 * HOUR + 1000);
    await settle();
    expect(m.currentSession()).toBeNull();
    expect(get(m.expiryNotice)).toMatch(/expired/);
    expect(localStorage.getItem(IDENTITY_KEY)).toBeNull();
  });

  it("with nothing saved, an expired session goes to the login page as before", async () => {
    const m = await load();
    m.setSession(noUser("admin-tok", 1 * HOUR));
    await vi.advanceTimersByTimeAsync(1 * HOUR + 1000);
    await settle();
    expect(renewals()).toHaveLength(0);
    expect(m.currentSession()).toBeNull();
    expect(get(m.expiryNotice)).toMatch(/expired/);
  });

  it("is not kept for a session with a user, nor from an expired one", async () => {
    const m = await load();
    m.setSession(signedIn("id-tok", "identified", 1 * DAY, 30 * DAY));
    m.setSession(signedIn("seat-tok", "player", 0, 29 * DAY));
    expect(localStorage.getItem(IDENTITY_KEY)).toBeNull();

    m.setSession(signedIn("stale", "identified", 31 * DAY, 30 * DAY));
    m.setSession(noUser("admin-tok", 1 * HOUR));
    expect(localStorage.getItem(IDENTITY_KEY)).toBeNull();
  });

  it("signing out forgets it", async () => {
    const m = await load();
    m.setSession(signedIn("id-tok", "identified", 1 * DAY, 30 * DAY));
    m.setSession(noUser("admin-tok", 1 * HOUR));
    expect(localStorage.getItem(IDENTITY_KEY)).not.toBeNull();
    await m.logout();
    expect(localStorage.getItem(IDENTITY_KEY)).toBeNull();
    expect(m.currentSession()).toBeNull();
  });
});
