// @vitest-environment jsdom
//
// ADR 0158 §1: Tables has Now and History, and My games is folded in.
// Rendered because what matters is in the markup: a game in play shows
// its turn and life totals, a finished game is a History row with its
// result and a Replay, the filters and "Show N more" work over the list
// the page already has, and #/my-games lands on History.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Lobby from "../routes/Lobby.svelte";
import App from "../App.svelte";
import type { MyGame } from "./myGames";
import { parseHash } from "./router";
import { currentSession, setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

HTMLMediaElement.prototype.play = () => Promise.resolve();
HTMLMediaElement.prototype.load = () => {};
HTMLMediaElement.prototype.pause = () => {};

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const DAY = 86_400_000;

function identity(): Session {
  const exp = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token: "id-tok",
    expiresAt: exp,
    principal: {
      role: "identified",
      user_id: USER,
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: exp,
    },
  };
}

function myGame(i: number, extra: Partial<MyGame> = {}): MyGame {
  return {
    id: `g-${i}`,
    name: `Game ${i}`,
    state: "ended",
    seat: 0,
    winner_seat: 1,
    created_at: Date.now() - (i + 1) * DAY,
    started_at: Date.now() - (i + 1) * DAY,
    ended_at: Date.now() - (i + 1) * DAY,
    archived_at: null,
    others: [{ seat: 1, name: "Bot 1", bot: true }],
    ...extra,
  };
}

const LIVE = {
  id: "live",
  name: "Friday Pod",
  created_at: new Date().toISOString(),
  state: "active",
  turn: 7,
  players: [
    {
      player_id: "p0",
      name: "Slysloth",
      seat: 0,
      deck_uploaded: true,
      deck_name: "Atraxa",
      is_host: true,
      life: 31,
    },
    { player_id: "p1", name: "Alice", seat: 1, deck_uploaded: true, deck_name: "Vivi", life: 28 },
  ],
};

let mine: MyGame[] = [];
let posted: string[] = [];

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      let body: unknown = {};
      if (input === "/games") body = { games: [LIVE] };
      if (input === "/me/games") body = { games: mine };
      if (method === "POST" && input.startsWith("/me/games/")) {
        posted.push(input);
        body = {
          token: "seat-tok",
          expires_at: new Date(Date.now() + 3_600_000).toISOString(),
          principal: { role: "player", user_id: USER, game_id: "live", player_id: "p1" },
          game: { id: "live" },
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

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    for (let j = 0; j < 5; j++) await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

function buttonNamed(root: ParentNode, re: RegExp): HTMLButtonElement | undefined {
  return [...root.querySelectorAll("button")].find((b) => re.test(b.textContent ?? ""));
}

beforeEach(() => {
  posted = [];
  mine = [];
  setSession(identity());
  location.hash = "#/lobby";
});
afterEach(async () => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
  location.hash = "";
  await settle();
});

describe("Now (§1 item 2)", () => {
  it("shows a game in play with its turn, each seat's life, and Resume", async () => {
    mine = [
      myGame(0, {
        id: "live",
        name: "Friday Pod",
        state: "active",
        seat: 1,
        winner_seat: null,
        ended_at: null,
        rejoin: "/me/games/live/session",
      }),
    ];
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const card = container.querySelector("li.tcard")!;
    expect(card.querySelector(".chip")?.textContent?.trim()).toBe("In play · turn 7");
    expect(card.querySelector(".ctx")?.textContent).toBe("Slysloth hosts · your seat 2");
    const lines = [...card.querySelectorAll(".sstat")].map((s) => s.textContent?.trim());
    expect(lines).toEqual(["Atraxa · 31 life", "Vivi · 28 life"]);
    // No open seats on a game in play.
    expect(card.textContent).not.toMatch(/Open seat/);

    const resume = buttonNamed(card, /^\s*Resume\s*$/)!;
    click(resume);
    await settle();
    expect(posted).toEqual(["/me/games/live/session"]);
    expect(currentSession()?.gameID).toBe("live");
    expect(location.hash).toBe("#/games/live");
  });
});

describe("History (§1 item 3)", () => {
  it("lists the finished games, not the ones in Now, newest first, six at a time", async () => {
    mine = [
      myGame(0, {
        id: "live",
        name: "Friday Pod",
        state: "active",
        winner_seat: null,
        rejoin: "/me/games/live/session",
      }),
      ...Array.from({ length: 9 }, (_, i) => myGame(i + 1)),
    ];
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const history = container.querySelector("#history")!;
    expect(history.querySelector("h2")?.textContent).toMatch(/History\s*9 games/);
    const names = () =>
      [...history.querySelectorAll("li.hrow .hname")].map((n) => n.textContent?.trim());
    expect(names()).toEqual(["Game 1", "Game 2", "Game 3", "Game 4", "Game 5", "Game 6"]);
    const row = history.querySelector("li.hrow")!;
    expect(row.querySelector(".result")?.textContent).toBe("Bot 1 won");
    expect(row.querySelector(".hwith")?.textContent).toBe("Bot 1");
    expect(row.querySelector("a[download]")?.getAttribute("href")).toBe(
      "/games/g-1/replay?token=id-tok",
    );

    click(buttonNamed(history, /Show 3 more/)!);
    expect(names()).toHaveLength(9);
    expect(buttonNamed(history, /more/)).toBeUndefined();
  });

  it("filters Won, With friends and Archived in the client", async () => {
    mine = [
      myGame(1, { name: "Won it", winner_seat: 0 }),
      myGame(2, { name: "With Bob", others: [{ seat: 1, name: "Bob" }] }),
      myGame(3, { name: "Put away", archived_at: Date.now() - DAY }),
    ];
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const history = container.querySelector("#history")!;
    const names = () =>
      [...history.querySelectorAll("li.hrow .hname")].map((n) => n.textContent?.trim());
    const calls = () => (fetch as unknown as { mock: { calls: unknown[] } }).mock.calls.length;
    const before = calls();

    click(buttonNamed(history, /^Won$/)!);
    expect(names()).toEqual(["Won it"]);
    click(buttonNamed(history, /^With friends$/)!);
    expect(names()).toEqual(["With Bob"]);
    click(buttonNamed(history, /^Archived$/)!);
    expect(names()).toEqual(["Put away"]);
    expect(history.querySelector(".result.archived")?.textContent).toBe("Archived");
    click(buttonNamed(history, /^All$/)!);
    expect(names()).toHaveLength(3);
    expect(calls()).toBe(before);
  });
});

describe("My games is folded in (§1 item 5)", () => {
  it("redirects #/my-games to Tables' History", async () => {
    expect(parseHash("#/lobby?show=history")).toEqual({ name: "lobby", show: "history" });
    stubServer();
    location.hash = "#/my-games";
    await settle();
    const { container } = render(App as never, {} as never);
    await settle();
    expect(location.hash).toBe("#/lobby?show=history");
    expect(container.querySelector("#history")).not.toBeNull();
    // The header has no My games link any more.
    const nav = container.querySelector('nav[aria-label="Site"]')!;
    expect(nav.textContent).not.toMatch(/My games/);
  });
});

describe("router", () => {
  it("carries user_id through the oauth-complete fragment", () => {
    const r = parseHash(
      "#/oauth-complete?token=t&expires_at=2026-09-20T00%3A00%3A00Z&game=g&player_id=p&user_id=" +
        USER,
    );
    expect(r).toMatchObject({ name: "oauthComplete", gameID: "g", playerID: "p", userID: USER });
  });
});
