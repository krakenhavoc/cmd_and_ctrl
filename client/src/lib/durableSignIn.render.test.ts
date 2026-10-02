// @vitest-environment jsdom
//
// ADR 0110 Delivery PR 1, the sign-in fix, on the client: the login page
// offers its code box to a signed-in person who is already at a table,
// joins the next table by code as that person, offers "Sign in with a
// different Discord account" (prompt=consent), and a seat session that
// lasts as long as the sign-in is not dropped at 12 hours.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Login from "../routes/Login.svelte";
import { discordLoginHref, joinByCode } from "./api";
import { MAX_TIMER_MS, currentSession, setSession, type Session } from "./session";
import { cleanup, flushSync, render } from "./test/render.svelte";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const NIL = "00000000-0000-0000-0000-000000000000";
const DAY = 24 * 60 * 60 * 1000;

function sessionAs(role: Session["principal"]["role"], userID?: string, name = "Alice"): Session {
  const expires = new Date(Date.now() + 20 * DAY).toISOString();
  return {
    token: `${role}-tok`,
    expiresAt: expires,
    principal: {
      role,
      user_id: userID,
      name,
      game_id: role === "identified" ? undefined : "g1",
      player_id: role === "player" ? "p1" : undefined,
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
    gameID: role === "identified" ? undefined : "g1",
    playerID: role === "player" ? "p1" : undefined,
  };
}

// posted records the POST bodies; the server's answer to POST /join is
// a seat at the next table that ends with the sign-in, 20 days out.
let posted: Array<{ url: string; body: unknown; auth: string | null }> = [];
const seatExpiry = () => new Date(Date.now() + 20 * DAY).toISOString();

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      let body: unknown = {};
      if (input === "/auth/discord/config") body = { enabled: true };
      if (method === "POST" && input === "/join") {
        posted.push({
          url: input,
          body: JSON.parse(String(init.body)),
          auth: new Headers(init.headers).get("Authorization"),
        });
        const exp = seatExpiry();
        body = {
          token: "next-seat-tok",
          expires_at: exp,
          principal: {
            role: "player",
            user_id: USER,
            game_id: "g2",
            player_id: "p2",
            name: "Alice",
            issued_at: new Date().toISOString(),
            expires_at: exp,
          },
          game: { id: "g2" },
          player_id: "p2",
        };
      }
      return {
        ok: true,
        status: 200,
        statusText: "OK",
        json: async () => body,
        clone() {
          return this;
        },
      } as unknown as Response;
    }),
  );
}

async function settle(): Promise<void> {
  for (let i = 0; i < 5; i++) await Promise.resolve();
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
}

beforeEach(() => {
  posted = [];
  location.hash = "#/login";
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.unstubAllGlobals();
  setSession(null);
});

describe("discordLoginHref", () => {
  it("starts a plain sign-in, or one that asks Discord for its account screen", () => {
    expect(discordLoginHref()).toBe("/auth/discord/start");
    expect(discordLoginHref({ consent: true })).toBe("/auth/discord/start?prompt=consent");
  });
});

describe("Login page for a signed-in person", () => {
  it("offers the code box, the way back and a different account to a signed-in seat", async () => {
    setSession(sessionAs("player", USER));
    stubServer();
    const { container } = render(Login as never, {} as never);
    await settle();

    expect(container.textContent).toMatch(/Join a table/);
    expect(container.textContent).toMatch(/Signed in as Alice/);
    // A signed-in seat needs no name for a code: the server seats the
    // Discord identity.
    const input = container.querySelector<HTMLInputElement>(
      'input[aria-label="invite code or link"]',
    );
    expect(input).not.toBeNull();
    input!.value = "abc123";
    input!.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(container.querySelector('input[aria-label="your name"]')).toBeNull();

    expect(container.querySelector('a[href="#/lobby"]')?.textContent).toMatch(/Back to your table/);
    const other = [...container.querySelectorAll("a")].find((a) =>
      /different Discord account/.test(a.textContent ?? ""),
    );
    expect(other?.getAttribute("href")).toBe("/auth/discord/start?prompt=consent");
    // The plain sign-in button is for people who are not signed in.
    expect(container.textContent).not.toMatch(/Continue with Discord/);
  });

  it("offers the different-account link to a fresh sign-in too, without the way back", async () => {
    setSession(sessionAs("identified", USER));
    stubServer();
    const { container } = render(Login as never, {} as never);
    await settle();
    expect(container.textContent).toMatch(/different Discord account/);
    expect(container.querySelector('a[href="#/lobby"]')).toBeNull();
  });

  it("treats a guest seat as signed out: Discord sign-in, and a name for a code", async () => {
    setSession(sessionAs("player", NIL, "Guest"));
    stubServer();
    const { container } = render(Login as never, {} as never);
    await settle();
    expect(container.textContent).toMatch(/Continue with Discord/);
    expect(container.textContent).toMatch(/Have an invite\?/);
    expect(container.textContent).not.toMatch(/different Discord account/);
  });
});

describe("joining the next table by code", () => {
  it("sends the signed-in seat's own token and installs the new seat", async () => {
    setSession(sessionAs("player", USER));
    stubServer();
    await joinByCode("abc123");
    expect(posted).toEqual([
      { url: "/join", body: { invite_token: "abc123", name: "" }, auth: "Bearer player-tok" },
    ]);
    const s = currentSession();
    expect(s?.token).toBe("next-seat-tok");
    expect(s?.gameID).toBe("g2");
    expect(s?.principal.user_id).toBe(USER);
  });
});

describe("a seat session that lasts as long as the sign-in", () => {
  it("is not dropped at 12 hours, nor at setTimeout's clamp, and expires on time", async () => {
    vi.useFakeTimers();
    const start = new Date("2026-10-02T12:00:00Z");
    vi.setSystemTime(start);
    setSession(sessionAs("identified", USER));
    stubServer();
    await joinByCode("abc123");
    const exp = Date.parse(currentSession()!.expiresAt);
    expect(exp - start.getTime()).toBe(20 * DAY);

    vi.advanceTimersByTime(12 * 3600_000 + 60_000);
    expect(currentSession()?.token).toBe("next-seat-tok");
    vi.advanceTimersByTime(Math.min(MAX_TIMER_MS, exp - Date.now() - 60_000));
    expect(currentSession()?.token).toBe("next-seat-tok");
    vi.advanceTimersByTime(exp - Date.now() + 60_000);
    expect(currentSession()).toBeNull();
  });
});
