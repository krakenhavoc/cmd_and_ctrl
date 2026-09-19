// @vitest-environment jsdom
//
// S34 sub-PR 4 (ADR 0051 decision 4): the My games page, and the two
// router entries it depends on. Rendered because what matters is in the
// markup: an open table has an Open button that swaps the session for
// the seat's and lands on the right page, a finished one has none, and
// a guest is told why the page is empty instead of shown an error.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import MyGames from "../routes/MyGames.svelte";
import type { MyGame } from "./myGames";
import { parseHash } from "./router";
import { currentSession, setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";

function identity(userID?: string): Session {
  return {
    token: "id-tok",
    expiresAt: new Date(Date.now() + 3_600_000).toISOString(),
    principal: {
      role: "identified",
      user_id: userID,
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 3_600_000).toISOString(),
    },
  };
}

const live: MyGame = {
  id: "live-1",
  name: "Friday Night",
  state: "active",
  seat: 1,
  winner_seat: null,
  created_at: Date.now(),
  started_at: Date.now(),
  ended_at: null,
  archived_at: null,
  others: [{ seat: 0, name: "Bob" }],
  rejoin: "/me/games/live-1/session",
};
const old: MyGame = {
  ...live,
  id: "old-1",
  name: "Last week",
  state: "ended",
  winner_seat: 1,
  created_at: Date.now() - 7 * 86_400_000,
  rejoin: undefined,
};

let posted: string[] = [];

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      let body: unknown = {};
      if (input === "/me/games" && method === "GET") body = { games: [old, live] };
      if (method === "POST") {
        posted.push(input);
        body = {
          token: "seat-tok",
          expires_at: new Date(Date.now() + 3_600_000).toISOString(),
          principal: { role: "player", user_id: USER, game_id: "live-1", player_id: "p1" },
          game: { id: "live-1" },
          player_id: "p1",
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

// Lets the component's fetch promise chain settle, then renders.
async function settle(): Promise<void> {
  for (let i = 0; i < 5; i++) await Promise.resolve();
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
}

beforeEach(() => {
  posted = [];
  location.hash = "#/my-games";
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
});

describe("MyGames page", () => {
  it("lists the user's games newest first, with Open only on the open one", async () => {
    setSession(identity(USER));
    stubServer();
    const { container } = render(MyGames as never, {} as never);
    await settle();

    const rows = [...container.querySelectorAll("li.game")];
    expect(rows.map((r) => r.querySelector(".name")?.textContent)).toEqual([
      "Friday Night",
      "Last week",
    ]);
    expect(rows[0].textContent).toMatch(/in progress/);
    expect(rows[0].textContent).toMatch(/with Bob/);
    expect(rows[1].textContent).toMatch(/you won/);
    expect(rows[0].querySelector("button")).not.toBeNull();
    expect(rows[1].querySelector("button")).toBeNull();
  });

  it("Open swaps in the seat session and goes to the board", async () => {
    setSession(identity(USER));
    stubServer();
    const { container } = render(MyGames as never, {} as never);
    await settle();

    const open = container.querySelector<HTMLButtonElement>(
      'button[aria-label="open Friday Night"]',
    );
    expect(open).not.toBeNull();
    click(open!);
    await settle();

    expect(posted).toEqual(["/me/games/live-1/session"]);
    expect(currentSession()?.token).toBe("seat-tok");
    expect(location.hash).toBe("#/games/live-1");
  });

  it("tells a session with no user to sign in, and asks the server nothing", async () => {
    setSession(identity(undefined));
    stubServer();
    const { container } = render(MyGames as never, {} as never);
    await settle();

    expect(container.textContent).toMatch(/Sign in with Discord/);
    expect(fetch).not.toHaveBeenCalled();
  });
});

describe("router", () => {
  it("routes #/my-games", () => {
    expect(parseHash("#/my-games")).toEqual({ name: "myGames" });
  });

  it("carries user_id through the oauth-complete fragment", () => {
    const r = parseHash(
      "#/oauth-complete?token=t&expires_at=2026-09-20T00%3A00%3A00Z&game=g&player_id=p&user_id=" +
        USER,
    );
    expect(r).toMatchObject({ name: "oauthComplete", gameID: "g", playerID: "p", userID: USER });
    expect(parseHash("#/oauth-complete?token=t&expires_at=x")).toMatchObject({
      name: "oauthComplete",
      userID: undefined,
    });
  });
});
