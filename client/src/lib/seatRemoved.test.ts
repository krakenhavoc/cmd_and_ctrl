// @vitest-environment jsdom
//
// #2785, the removed player's side: the news arrives as a 401 whose body
// says "removed from table" (the lobby's next request) and as a 1000
// close with that reason (the table's socket, covered in
// ws.reconnect.test.ts). Either one must raise the notice; an ordinary
// 401 must not.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { get, writable } from "svelte/store";
import type { Session } from "./session";
import type { Route } from "./router";

let answer: Response;

function liveSession(): Session {
  const expires = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token: "seat-token",
    expiresAt: expires,
    principal: {
      role: "player",
      game_id: "g1",
      player_id: "p1",
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
    playerID: "p1",
    gameID: "g1",
  };
}

async function load() {
  vi.resetModules();
  vi.doMock("./router", () => ({ route: writable<Route>({ name: "lobby" }), navigate: vi.fn() }));
  return import("./session");
}

beforeEach(() => {
  localStorage.clear();
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => answer.clone()),
  );
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("a 401 for a removed seat", () => {
  it("raises the notice, drops the session and says why on the login page", async () => {
    const m = await load();
    answer = new Response(JSON.stringify({ error: "removed from table" }), { status: 401 });
    m.setSession(liveSession());

    await expect(m.authFetch("/games")).rejects.toMatchObject({
      status: 401,
      message: m.SEAT_REMOVED_MESSAGE,
    });

    expect(get(m.removedNotice)).toBe(m.REMOVED_NOTICE);
    expect(m.currentSession()).toBeNull();
    expect(get(m.expiryNotice)).toBe(m.REMOVED_NOTICE);
  });

  it("is not raised by an ordinary expired session", async () => {
    const m = await load();
    answer = new Response(JSON.stringify({ error: "session expired" }), { status: 401 });
    m.setSession(liveSession());

    await expect(m.authFetch("/games")).rejects.toThrow("session expired");

    expect(get(m.removedNotice)).toBe("");
  });

  it("is not raised by a 401 with a body that is not JSON", async () => {
    const m = await load();
    answer = new Response("nope", { status: 401 });
    m.setSession(liveSession());

    await expect(m.authFetch("/games")).rejects.toThrow("session expired");

    expect(get(m.removedNotice)).toBe("");
  });
});
