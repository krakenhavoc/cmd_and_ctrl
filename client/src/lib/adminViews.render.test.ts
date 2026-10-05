// @vitest-environment jsdom
//
// adminViews.render.test.ts — ADR 0124 §7 and §8, rendered: each admin
// view from a fixture response (Live now, Games, one table, Accounts,
// one account), the linked actions going to the existing routes, the
// header's Admin link, and the gate: an allowlisted person in player
// mode and a non-admin see a message, no row, and send no request; a
// stale admin that hears 403 sees the same message and asks /me again.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";

import Admin from "../routes/Admin.svelte";
import SiteHeader from "./components/SiteHeader.svelte";
import { resetAdminChecksForTest } from "./admin";
import type {
  AdminAccountResponse,
  AdminAccountsResponse,
  AdminGameResponse,
  AdminGamesResponse,
  AdminView,
  LiveNowResponse,
} from "./adminViews";
import { session, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const NOW = Date.UTC(2026, 9, 5, 12, 0, 0);
const MIN = 60_000;
const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const BOB = "7c1e2d3f-0a0b-4c0d-8e0f-1a2b3c4d5e6f";
const GAME = "9c2f0e4c-1d1e-4c3b-9d7e-2f8a1b2c3d4e";

const LIVE: LiveNowResponse = {
  generated_at: NOW,
  tables: [
    {
      id: GAME,
      name: "Friday night",
      state: "active",
      practice: false,
      archived: false,
      seats: [
        {
          seat: 0,
          kind: "human",
          guest_name: "Alice",
          deck_name: "Mono Red",
          host: true,
          connected: 1,
          since: NOW - 4 * MIN,
        },
        {
          seat: 1,
          kind: "human",
          account: { id: BOB, name: "Bobby", avatar_url: "/avatars/1/h.png" },
          host: false,
          connected: 0,
        },
        {
          seat: 2,
          kind: "bot",
          guest_name: "Bot 1",
          bot_tier: "heuristic",
          host: false,
          connected: 0,
        },
        {
          seat: 3,
          kind: "agent",
          guest_name: "Claude",
          agent_client: "claude-code",
          host: false,
          connected: 1,
          since: NOW - MIN,
        },
      ],
      spectators: [{ account: { id: USER, name: "Carol" }, since: NOW - 2 * MIN }, { since: NOW }],
      admins: [{ as_seat: 0, since: NOW - 3 * MIN }],
    },
  ],
  unbound_sockets: 0,
  totals: { players_connected: 2, spectators: 2, bot_seats: 1, practice_tables: 0, admin_views: 1 },
};

const ROW = {
  id: GAME,
  name: "Friday night",
  state: "ended",
  created_at: NOW - 3 * 3_600_000,
  started_at: NOW - 170 * MIN,
  ended_at: NOW - 60 * MIN,
  outcome: "win",
  winner_seat: 1,
  creator: { id: USER, name: "Ann" },
  practice: false,
  loaded: true,
  seats: [
    {
      seat: 0,
      kind: "human",
      account: { id: USER, name: "Ann" },
      deck_name: "Atraxa",
      host: true,
      connected: 1,
    },
    { seat: 1, kind: "human", guest_name: "Gus", host: false, connected: 0 },
  ],
  spectators_connected: 0,
};

const GAMES: AdminGamesResponse = {
  generated_at: NOW,
  games: [ROW, { ...ROW, id: "g-2", name: "Lobby table", state: "lobby", outcome: undefined }],
  next_cursor: "c2",
};

const GAME_DETAIL: AdminGameResponse = {
  ...ROW,
  state: "active",
  outcome: undefined,
  ended_at: undefined,
  generated_at: NOW,
  connections: [
    { kind: "seat", seat: 0, account: { id: USER, name: "Ann" }, since: NOW - 5 * MIN },
    { kind: "admin", since: NOW - MIN },
  ],
};

const ACCOUNTS: AdminAccountsResponse = {
  generated_at: NOW,
  truncated: false,
  accounts: [
    {
      id: USER,
      name: "zed",
      first_seen_at: NOW - 30 * 86_400_000,
      last_sign_in_at: NOW - 86_400_000,
      games_played: 12,
      last_played_at: NOW - 3_600_000,
      playing_now: true,
    },
    { id: BOB, name: "Ann", first_seen_at: NOW - 86_400_000, games_played: 0, playing_now: false },
  ],
};

const ACCOUNT: AdminAccountResponse = {
  generated_at: NOW,
  account: ACCOUNTS.accounts[0],
  sign_in: {
    last_sign_in_at: NOW - 86_400_000,
    discord_linked_at: NOW - 30 * 86_400_000,
    revoke_path: `/admin/users/${USER}/revoke-sessions`,
  },
  games: [{ ...ROW, their_seat: 0 }],
  games_truncated: false,
  decks: [
    {
      id: "d1",
      name: "Atraxa superfriends",
      format: "moxfield",
      source_url: "https://moxfield.com/decks/abc",
      commanders: ["Atraxa, Praetors' Voice"],
      card_count: 100,
      updated_at: NOW - 86_400_000,
    },
  ],
  deck_requests: [
    {
      deck_key: "moxfield:abc",
      asked_at: NOW - 86_400_000,
      issue_number: 1234,
      issue_url: "https://github.com/krakenhavoc/cmd_and_ctrl/issues/1234",
    },
  ],
  deck_requests_truncated: false,
};

interface Call {
  url: string;
  method: string;
}
let calls: Call[];
let forbid: boolean;

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), { status });
}

async function serve(url: string, init: RequestInit = {}): Promise<Response> {
  const method = init.method ?? "GET";
  calls.push({ url, method });
  if (url === "/me") return json({ admin: false, admin_allowed: true, admin_mode: false });
  if (url === "/auth/discord/config") return json({ enabled: false });
  if (url.startsWith("/admin/") && forbid) return json({ error: "admin only" }, 403);
  if (url === "/admin/live") return json(LIVE);
  if (url.startsWith("/admin/games/")) return json(GAME_DETAIL);
  if (url.startsWith("/admin/games")) return json(GAMES);
  if (url.endsWith("/revoke-sessions")) {
    return json({ user_id: USER, sessions_invalid_before: "x", sockets_closed: 2 });
  }
  if (url.startsWith("/admin/users/")) return json(ACCOUNT);
  if (url.startsWith("/admin/users")) return json(ACCOUNTS);
  if (url.endsWith("/archive")) return json({ id: GAME, name: "Friday night", players: [] });
  return json({ error: "nope" }, 404);
}

function token(): Session {
  const expiresAt = new Date(NOW + 86_400_000).toISOString();
  return {
    token: "admin-tok",
    expiresAt,
    principal: {
      role: "admin",
      name: "admin",
      issued_at: new Date(NOW).toISOString(),
      expires_at: expiresAt,
    },
  };
}

function person(admin: boolean, allowed = true): Session {
  const expiresAt = new Date(NOW + 86_400_000).toISOString();
  return {
    token: "person-tok",
    expiresAt,
    principal: {
      role: "identified",
      user_id: USER,
      name: "Ann",
      issued_at: new Date(NOW).toISOString(),
      expires_at: expiresAt,
    },
    admin,
    admin_allowed: allowed,
    admin_mode: admin,
    ...(admin ? { admin_mode_ends_at: NOW + 3_600_000 } : {}),
  };
}

beforeEach(() => {
  calls = [];
  forbid = false;
  resetAdminChecksForTest();
  vi.useFakeTimers({ toFake: ["Date"] });
  vi.setSystemTime(NOW);
  vi.stubGlobal("fetch", vi.fn(serve));
  sessionStorage.clear();
});

afterEach(() => {
  cleanup();
  session.set(null);
  vi.unstubAllGlobals();
  vi.useRealTimers();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 6; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

async function show(view: AdminView, s: Session = token()): Promise<HTMLElement> {
  session.set(s);
  const { container } = render(Admin as never, { view } as never);
  await settle();
  return container;
}

const adminCalls = (): Call[] => calls.filter((c) => c.url.startsWith("/admin/"));
const text = (el: Element | null | undefined): string =>
  (el?.textContent ?? "").replace(/\s+/g, " ");
const button = (c: ParentNode, name: string): HTMLButtonElement | undefined =>
  [...c.querySelectorAll<HTMLButtonElement>("button")].find((b) => b.textContent?.trim() === name);

describe("the tab strip", () => {
  it("is navigation 'admin views' with Live now, Games and Accounts, the current one marked", async () => {
    const c = await show({ view: "game", id: GAME });
    const nav = c.querySelector('nav[aria-label="admin views"]')!;
    const links = [...nav.querySelectorAll("a")];
    expect(links.map((a) => a.textContent?.trim())).toEqual(["Live now", "Games", "Accounts"]);
    expect(links.map((a) => a.getAttribute("href"))).toEqual([
      "#/admin/live",
      "#/admin/games",
      "#/admin/accounts",
    ]);
    expect(links[1].getAttribute("aria-current")).toBe("page");
  });
});

describe("Live now", () => {
  it("shows each table, its seats, spectators and admins, and the totals", async () => {
    const c = await show({ view: "live" });
    expect(adminCalls()).toEqual([{ url: "/admin/live", method: "GET" }]);
    const region = c.querySelector('section[aria-label="live now"]')!;
    expect(region).not.toBeNull();
    const t = text(region);
    expect(t).toContain("Players connected 2");
    expect(t).toContain("Bot seats 1");
    expect(t).toContain("Friday night");
    expect(t).toContain("Alice");
    expect(t).toContain("connected · since 4 min ago");
    expect(t).toContain("Bobby");
    expect(t).toContain("not connected");
    expect(t).toContain("bot · heuristic");
    expect(t).toContain("agent · claude-code");
    expect(t).toContain("Carol");
    expect(t).toContain("guest spectator");
    expect(t).toContain("admin token");
    expect(t).toContain("as seat 1");
    // The table opens its detail, and an account opens its own.
    expect(region.querySelector(`a[href="#/admin/games/${GAME}"]`)).not.toBeNull();
    expect(region.querySelector(`a[href="#/admin/accounts/${BOB}"]`)).not.toBeNull();
    // The avatar carries the session token, as every avatar does.
    expect(region.querySelector("img")?.getAttribute("src")).toBe(
      "/avatars/1/h.png?token=admin-tok",
    );
  });
});

describe("Games", () => {
  it("lists the tables in table 'games' and asks with the hash's filters", async () => {
    const c = await show({ view: "games", filter: { state: "ended", archived: "false" } });
    expect(adminCalls()).toEqual([
      { url: "/admin/games?state=ended&archived=false", method: "GET" },
    ]);
    const table = c.querySelector('table[aria-label="games"]')!;
    const rows = table.querySelectorAll("tbody tr");
    expect(rows).toHaveLength(2);
    expect(text(rows[0])).toContain("Friday night");
    expect(text(rows[0])).toContain("Won by Gus");
    expect(text(rows[0])).toContain("Ann (account), Gus (guest)");
    expect(text(rows[1])).toContain("Waiting to start");
    expect(c.querySelector(`a[href="#/admin/games/${GAME}"]`)).not.toBeNull();
    // The next page is the server's cursor, in the hash.
    const older = [...c.querySelectorAll("a")].find((a) => a.textContent === "Older tables");
    expect(older?.getAttribute("href")).toBe("#/admin/games?state=ended&archived=false&cursor=c2");
  });
});

describe("one table", () => {
  it("shows its seats and connections, and archives a running table after a confirm that says it ends", async () => {
    const c = await show({ view: "game", id: GAME });
    expect(adminCalls()).toEqual([{ url: `/admin/games/${GAME}`, method: "GET" }]);
    const t = text(c);
    expect(t).toContain("Friday night");
    expect(t).toContain("In progress");
    expect(t).toContain("Connected now · 2");
    expect(c.querySelector(`a[href="#/games/${GAME}"]`)?.textContent).toContain("Open table");
    // A running table has no replay yet.
    expect(t).not.toContain("Replay");

    click(button(c, "Archive")!);
    expect(text(c.querySelector(".confirm"))).toContain("ends the game");
    click(button(c, "archive and end")!);
    await settle();
    expect(calls).toContainEqual({ url: `/games/${GAME}/archive`, method: "POST" });
    expect(text(c)).toContain("Archived.");
  });
});

describe("Accounts", () => {
  it("lists every account in table 'accounts', sorted here, never sending sort", async () => {
    const c = await show({ view: "accounts", filter: { played: "7d", sort: "name" } });
    expect(adminCalls()).toEqual([{ url: "/admin/users?played=7d", method: "GET" }]);
    const rows = c.querySelectorAll('table[aria-label="accounts"] tbody tr');
    expect([...rows].map((r) => r.querySelector(".nm")?.textContent)).toEqual(["Ann", "zed"]);
    expect(text(rows[1])).toContain("playing now");
    expect(text(c)).toContain("2 accounts played in the last 7d");
  });

  it("filters the list as you type", async () => {
    const c = await show({ view: "accounts", filter: {} });
    const input = c.querySelector<HTMLInputElement>('input[aria-label="find an account"]')!;
    input.value = "ze";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    const rows = c.querySelectorAll('table[aria-label="accounts"] tbody tr');
    expect(rows).toHaveLength(1);
    expect(text(rows[0])).toContain("zed");
  });
});

describe("one account", () => {
  it("shows sign-in state, tables, decks and deck requests, and revokes after naming the person", async () => {
    const c = await show({ view: "account", id: USER });
    expect(adminCalls()).toEqual([{ url: `/admin/users/${USER}`, method: "GET" }]);
    const t = text(c);
    expect(t).toContain("zed");
    expect(t).toContain("no list on the server");
    expect(t).toContain("Atraxa superfriends");
    expect(t).toContain("moxfield:abc");
    expect(
      c.querySelector('a[href="https://github.com/krakenhavoc/cmd_and_ctrl/issues/1234"]'),
    ).not.toBeNull();
    expect(c.querySelector(`a[href="#/admin/games?user=${USER}"]`)).not.toBeNull();
    expect(text(c.querySelector('table[aria-label="their tables"]'))).toContain("seat 1");

    click(button(c, "Revoke sessions")!);
    expect(text(c.querySelector(".confirm"))).toContain("Sign zed out everywhere?");
    click(button(c, "sign them out everywhere")!);
    await settle();
    expect(calls).toContainEqual({ url: `/admin/users/${USER}/revoke-sessions`, method: "POST" });
    expect(text(c)).toContain("2 connections closed");
  });
});

describe("the gate", () => {
  it("shows player mode its message and switch, no rows, and asks nothing", async () => {
    const c = await show({ view: "live" }, person(false));
    expect(text(c)).toContain("Admin mode is off. Switch it on to see this page.");
    expect(c.querySelector('button[aria-label^="admin mode"]')).not.toBeNull();
    expect(c.querySelector('nav[aria-label="admin views"]')).toBeNull();
    expect(c.querySelector('section[aria-label="live now"]')).toBeNull();
    expect(c.querySelectorAll("table, tr")).toHaveLength(0);
    expect(adminCalls()).toEqual([]);
  });

  it("tells anyone else the page is for admins, and asks nothing", async () => {
    const c = await show({ view: "accounts", filter: {} }, person(false, false));
    expect(text(c)).toContain("This page is for admins.");
    expect(c.querySelector('button[aria-label^="admin mode"]')).toBeNull();
    expect(c.querySelectorAll("table, tr")).toHaveLength(0);
    expect(adminCalls()).toEqual([]);
  });

  it("shows the page to an allowlisted person in admin mode", async () => {
    const c = await show({ view: "accounts", filter: {} }, person(true));
    expect(c.querySelector('table[aria-label="accounts"]')).not.toBeNull();
  });

  it("treats a 403 as a stale answer: the message, no rows, and /me asked again", async () => {
    forbid = true;
    const c = await show({ view: "games", filter: {} }, person(true));
    expect(adminCalls()).toEqual([{ url: "/admin/games", method: "GET" }]);
    expect(calls.some((x) => x.url === "/me")).toBe(true);
    // /me said player mode, so the session drops admin and the page says so.
    expect(text(c)).toContain("Admin mode is off");
    expect(c.querySelectorAll("table, tr")).toHaveLength(0);
  });
});

describe("the header's Admin link", () => {
  const adminLink = (c: HTMLElement) =>
    [...c.querySelectorAll('nav[aria-label="Site"] a')].find((a) => a.textContent === "Admin");

  it("is last in the site nav for the token and for admin mode, going to Live now", async () => {
    for (const s of [token(), person(true)]) {
      session.set(s);
      const { container, destroy } = render(SiteHeader as never, {} as never);
      await settle();
      const links = [...container.querySelectorAll('nav[aria-label="Site"] a')];
      expect(links.at(-1)?.textContent).toBe("Admin");
      expect(adminLink(container)?.getAttribute("href")).toBe("#/admin/live");
      destroy();
    }
  });

  it("is absent in player mode and for everyone else", async () => {
    for (const s of [person(false), person(false, false), null]) {
      session.set(s);
      const { container, destroy } = render(SiteHeader as never, {} as never);
      await settle();
      expect(adminLink(container)).toBeUndefined();
      destroy();
    }
  });
});
