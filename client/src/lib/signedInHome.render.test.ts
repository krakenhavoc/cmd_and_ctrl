// @vitest-environment jsdom
//
// ADR 0112 Delivery PR 3, the signed-in home, rendered: the router sends
// every session on #/login to the Lobby and never bounces an invite,
// spectator, reclaim or Discord link; the Discord round trip from the
// login page lands on the Lobby; the Lobby's "Join a table" card, its
// "your tables" list and its empty state; the header's account menu and
// wordmark; the login page with no token card; and Home's join card.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { anchorOf, emptyContext } from "./hints/hint";
import { resolveAnchor } from "./tutorialAnchor";
import CREATE_HINT from "../routes/Lobby.hint";

import App from "../App.svelte";
import Lobby from "../routes/Lobby.svelte";
import Login from "../routes/Login.svelte";
import Home from "../routes/Home.svelte";
import SiteHeader from "./components/SiteHeader.svelte";
import { currentSession, setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

// App arms the ambient music on the first keydown anywhere, and jsdom's
// media elements have no working play().
HTMLMediaElement.prototype.play = () => Promise.resolve();
HTMLMediaElement.prototype.load = () => {};
HTMLMediaElement.prototype.pause = () => {};

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const NIL = "00000000-0000-0000-0000-000000000000";
const DAY = 24 * 60 * 60 * 1000;
const G1 = "11111111-1111-4111-8111-111111111111";
const G2 = "22222222-2222-4222-8222-222222222222";
const G3 = "33333333-3333-4333-8333-333333333333";
const G4 = "44444444-4444-4444-8444-444444444444";

type Role = Session["principal"]["role"];

function sess(role: Role, userID?: string, opts: { gameID?: string; name?: string } = {}): Session {
  const gameID = role === "identified" || role === "admin" ? undefined : (opts.gameID ?? G1);
  const expires = new Date(Date.now() + 20 * DAY).toISOString();
  return {
    token: `${role}-tok`,
    expiresAt: expires,
    principal: {
      role,
      user_id: userID,
      name: opts.name ?? (role === "admin" ? "admin" : "Alice"),
      game_id: gameID,
      player_id: role === "player" ? "p1" : undefined,
      issued_at: new Date().toISOString(),
      expires_at: expires,
    },
    gameID,
    playerID: role === "player" ? "p1" : undefined,
  };
}

function meta(id: string, extra: Record<string, unknown> = {}) {
  return {
    id,
    name: `table ${id.slice(0, 2)}`,
    created_at: new Date().toISOString(),
    players: [],
    state: "lobby",
    ...extra,
  };
}

function seatSession(gameID: string, playerID: string) {
  const exp = new Date(Date.now() + 20 * DAY).toISOString();
  return {
    token: `seat-${gameID.slice(0, 2)}`,
    expires_at: exp,
    principal: {
      role: "player",
      user_id: USER,
      game_id: gameID,
      player_id: playerID,
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: exp,
    },
    game: { id: gameID },
    player_id: playerID,
  };
}

type Handler = (body: unknown) => unknown;
let routes: Record<string, Handler> = {};
let calls: Array<{ method: string; url: string; body: unknown; auth: string | null }> = [];

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      const body = typeof init.body === "string" ? JSON.parse(init.body) : undefined;
      calls.push({
        method,
        url: input,
        body,
        auth: new Headers(init.headers).get("Authorization"),
      });
      const handler: Handler | undefined = routes[`${method} ${input}`];
      const found = handler !== undefined;
      const payload = found ? handler(body) : { error: "not found" };
      return {
        ok: found,
        status: found ? 200 : 404,
        statusText: found ? "OK" : "Not Found",
        json: async () => payload,
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

// goTo sets the hash and waits for the router's hashchange listener.
async function goTo(hash: string): Promise<void> {
  location.hash = hash;
  await settle();
}

function buttonNamed(root: ParentNode, re: RegExp): HTMLButtonElement | undefined {
  return [...root.querySelectorAll("button")].find((b) => re.test(b.textContent ?? ""));
}

function type(input: HTMLInputElement, value: string): void {
  input.value = value;
  input.dispatchEvent(new Event("input", { bubbles: true }));
  flushSync();
}

function submit(form: HTMLFormElement): void {
  form.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
}

beforeEach(() => {
  routes = { "GET /games": () => ({ games: [] }), "GET /me/games": () => ({ games: [] }) };
  calls = [];
  localStorage.clear();
});
afterEach(async () => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
  localStorage.clear();
  location.hash = "";
  await settle();
});

describe("the router (§1 items 1 and 2)", () => {
  it("sends a signed-in person on #/login to the Lobby", async () => {
    setSession(sess("identified", USER));
    stubServer();
    await goTo("#/login");
    render(App as never, {} as never);
    await settle();
    expect(location.hash).toBe("#/lobby");
  });

  it("sends a signed-in seat on #/login to the Lobby too", async () => {
    setSession(sess("player", USER));
    stubServer();
    await goTo("#/login");
    render(App as never, {} as never);
    await settle();
    expect(location.hash).toBe("#/lobby");
  });

  it("keeps a signed-out visitor on #/login, with no token form", async () => {
    stubServer();
    await goTo("#/login");
    const { container } = render(App as never, {} as never);
    await settle();
    expect(location.hash).toBe("#/login");
    expect(container.textContent).toMatch(/Have an invite\?/);
    expect(container.textContent).not.toMatch(/Admin log in/);
    expect(container.querySelector('input[placeholder="admin token"]')).toBeNull();
  });

  it("lands the Discord sign-in from the login page on the Lobby", async () => {
    stubServer();
    await goTo(
      `#/oauth-complete?token=id-tok&expires_at=${encodeURIComponent(new Date(Date.now() + DAY).toISOString())}&name=Alice&user_id=${USER}`,
    );
    render(App as never, {} as never);
    await settle();
    expect(location.hash).toBe("#/lobby");
    expect(currentSession()?.principal.role).toBe("identified");
    expect(currentSession()?.token).toBe("id-tok");
  });

  it("lands the Discord sign-in from an invite on its table", async () => {
    stubServer();
    // The table route would dial a socket; this only asks where the
    // handoff sends the browser, so it stops at the hash.
    vi.stubGlobal(
      "WebSocket",
      class {
        close(): void {}
        addEventListener(): void {}
      },
    );
    await goTo(
      `#/oauth-complete?token=seat-tok&expires_at=${encodeURIComponent(new Date(Date.now() + DAY).toISOString())}&game=${G2}&player_id=p2&user_id=${USER}`,
    );
    render(App as never, {} as never);
    await settle();
    expect(location.hash).toBe(`#/games/${G2}`);
    expect(currentSession()?.gameID).toBe(G2);
    expect(currentSession()?.playerID).toBe("p2");
  });

  // The ways in that do not go through Login keep working whatever the
  // browser holds: each opens its own page and is not bounced.
  const holders: Array<[string, Session | null]> = [
    ["signed out", null],
    ["a signed-in person", sess("identified", USER)],
    ["a signed-in seat", sess("player", USER)],
    ["a guest seat", sess("player", NIL)],
    ["the admin token", sess("admin")],
  ];
  for (const [who, s] of holders) {
    it(`opens an invite link for ${who}`, async () => {
      if (s) setSession(s);
      stubServer();
      const hash = `#/games/${G2}/join?t=invite-tok`;
      await goTo(hash);
      const { container } = render(App as never, {} as never);
      await settle();
      expect(location.hash).toBe(hash);
      expect(container.querySelector("h1")?.textContent).toMatch(/join/i);
    });

    it(`opens a spectator link for ${who}`, async () => {
      if (s) setSession(s);
      stubServer();
      const hash = `#/games/${G2}/join?t=spec-tok&spectator=1`;
      await goTo(hash);
      const { container } = render(App as never, {} as never);
      await settle();
      expect(location.hash).toBe(hash);
      expect(container.textContent).toMatch(/spectat|watch/i);
    });

    it(`opens a reclaim link for ${who}`, async () => {
      if (s) setSession(s);
      routes[`POST /games/${G2}/seats/reclaim`] = () => {
        throw new Error("not reached");
      };
      stubServer();
      const hash = `#/games/${G2}/reclaim?t=ticket`;
      await goTo(hash);
      const { container } = render(App as never, {} as never);
      await settle();
      expect(location.hash).not.toMatch(/#\/(login|lobby)$/);
      expect(container.textContent).toMatch(/welcome back/i);
    });
  }
});

describe("the lobby.create hint", () => {
  it("anchors on the create form a signed-in person sees", async () => {
    setSession(sess("player", USER));
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const a = anchorOf(CREATE_HINT, { ...emptyContext("lobby"), signedIn: true });
    expect(resolveAnchor(a!, container)).toBe(
      container.querySelector('form[aria-label="create game"]'),
    );
  });
});

describe("the Lobby's Join a table card (§1 item 3)", () => {
  it("joins a signed-in person by code, as themselves", async () => {
    setSession(sess("player", USER));
    routes["POST /join"] = () => seatSession(G2, "p2");
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const input = container.querySelector<HTMLInputElement>(
      'input[aria-label="invite code or link"]',
    );
    expect(input).not.toBeNull();
    type(input!, "abc123");
    submit(input!.form!);
    await settle();

    expect(calls).toContainEqual({
      method: "POST",
      url: "/join",
      body: { invite_token: "abc123", name: "" },
      auth: "Bearer player-tok",
    });
    expect(currentSession()?.gameID).toBe(G2);
    expect(input!.value).toBe("");
    // No name is asked of a signed-in person.
    expect(container.querySelector('input[aria-label="your name"]')).toBeNull();
  });

  it("opens a pasted link's Join page", async () => {
    setSession(sess("identified", USER));
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const input = container.querySelector<HTMLInputElement>(
      'input[aria-label="invite code or link"]',
    )!;
    type(input, `https://cmd.example/#/games/${G2}/join?t=xyz&spectator=1`);
    submit(input.form!);
    await settle();
    expect(location.hash).toBe(`#/games/${G2}/join?t=xyz&spectator=1`);
    expect(calls.some((c) => c.url === "/join")).toBe(false);
  });

  it("gives a guest a link box, and answers a bare code before sending it", async () => {
    setSession(sess("player", NIL, { name: "Guest" }));
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    expect(container.querySelector('input[aria-label="invite code or link"]')).toBeNull();
    const input = container.querySelector<HTMLInputElement>('input[aria-label="invite link"]');
    expect(input).not.toBeNull();
    type(input!, "abc123");
    submit(input!.form!);
    await settle();
    expect(container.querySelector('[role="alert"]')?.textContent).toBe(
      "This browser is seated as a guest. Open the invite link instead, or link Discord from your table's menu.",
    );
    expect(calls.some((c) => c.url === "/join")).toBe(false);
  });

  it("gives the admin token no join box", async () => {
    setSession(sess("admin"));
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    expect(container.querySelector('input[aria-label="invite code or link"]')).toBeNull();
    expect(container.querySelector('input[aria-label="invite link"]')).toBeNull();
    // The create form is still there.
    expect(container.querySelector('input[aria-label="game name"]')).not.toBeNull();
  });

  it("has no command bar of its own", async () => {
    setSession(sess("player", USER));
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    const lobby = container.querySelector("section.lobby")!;
    expect(buttonNamed(lobby, /log out|join with a code|my games/)).toBeUndefined();
    expect(lobby.querySelector('[aria-label="open settings"]')).toBeNull();
    // The page title stays.
    expect(container.querySelector('h1[aria-label="cmd_and_ctrl · lobby"]')?.textContent).toBe(
      "Tables",
    );
  });
});

describe("the Lobby lists your tables (§1 item 6, owner answer 3)", () => {
  const room = () => [
    meta(G1, { name: "Created", is_creator: true }),
    meta(G2, { name: "Other pod" }),
    meta(G3, { name: "Seat elsewhere", state: "active" }),
    meta(G4, { name: "Seat in the lobby" }),
  ];

  it("shows a signed-in person their own tables, and opens a seat held elsewhere", async () => {
    setSession(sess("identified", USER));
    routes["GET /games"] = () => ({ games: room() });
    routes["GET /me/games"] = () => ({
      games: [
        { id: G3, name: "Seat elsewhere", state: "active", rejoin: `/me/games/${G3}/session` },
        { id: G4, name: "Seat in the lobby", state: "lobby", rejoin: `/me/games/${G4}/session` },
        // An ended table that is no longer open has no rejoin.
        { id: G2, name: "Other pod", state: "ended" },
      ],
    });
    routes[`POST /me/games/${G3}/session`] = () => seatSession(G3, "p3");
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const names = [...container.querySelectorAll(".tname")].map((n) => n.textContent);
    expect(names).toEqual(["Created", "Seat elsewhere", "Seat in the lobby"]);
    expect(container.textContent).not.toMatch(/Other pod/);

    const card = [...container.querySelectorAll("li.tcard")].find((li) =>
      /Seat elsewhere/.test(li.textContent ?? ""),
    )!;
    const open = buttonNamed(card, /^\s*Open\s*$/);
    expect(open).toBeDefined();
    // Open is the only way in: no "open table", which would watch as a
    // stranger.
    expect(buttonNamed(card, /open table|enter table/)).toBeUndefined();
    click(open!);
    await settle();
    expect(calls).toContainEqual(
      expect.objectContaining({ method: "POST", url: `/me/games/${G3}/session` }),
    );
    expect(currentSession()?.gameID).toBe(G3);
    expect(location.hash).toBe(`#/games/${G3}`);
  });

  it("opens a held seat at a table still in the lobby on the Lobby, with its deck panel", async () => {
    setSession(sess("identified", USER));
    routes["GET /games"] = () => ({
      games: [meta(G4, { players: [{ player_id: "p4", name: "Alice", seat: 0 }] })],
    });
    routes["GET /me/games"] = () => ({
      games: [{ id: G4, name: "t", state: "lobby", rejoin: `/me/games/${G4}/session` }],
    });
    routes[`POST /me/games/${G4}/session`] = () => seatSession(G4, "p4");
    stubServer();
    location.hash = "#/lobby";
    const { container } = render(Lobby as never, {} as never);
    await settle();
    click(buttonNamed(container, /^\s*Open\s*$/)!);
    await settle();
    expect(currentSession()?.gameID).toBe(G4);
    expect(location.hash).toBe("#/lobby");
    expect(buttonNamed(container, /^\s*Open\s*$/)).toBeUndefined();
    expect(container.querySelector("details.deck-upload")).not.toBeNull();
  });

  it("shows the admin token every table", async () => {
    setSession(sess("admin"));
    routes["GET /games"] = () => ({ games: room() });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    expect(container.querySelectorAll("li.tcard")).toHaveLength(4);
    expect(calls.some((c) => c.url === "/me/games")).toBe(false);
  });

  it("shows a signed-in person with no tables the empty state, not the room", async () => {
    setSession(sess("identified", USER));
    routes["GET /games"] = () => ({ games: [meta(G2, { name: "Other pod" })] });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    expect(container.querySelectorAll("li.tcard")).toHaveLength(0);
    const empty = container.querySelector(".empty-card")!;
    expect(empty.querySelector("p")?.textContent?.replace(/\s+/g, " ").trim()).toBe(
      "You're not at a table yet. Paste an invite code above, or create a table and invite your pod.",
    );
    const links = [...empty.querySelectorAll("a")].map((a) => [
      a.textContent?.trim(),
      a.getAttribute("href"),
    ]);
    expect(links).toEqual([
      ["Practice against bots", "#/practice"],
      ["Check a deck", "#/decks"],
      ["My games", "#/my-games"],
    ]);
    // The Join and Create cards are above it.
    expect(container.querySelector('input[aria-label="invite code or link"]')).not.toBeNull();
    expect(container.querySelector('input[aria-label="game name"]')).not.toBeNull();
  });
});

describe("the header's account menu and wordmark (§1 items 4 and 5)", () => {
  function accountButton(container: HTMLElement): HTMLButtonElement {
    return container.querySelector<HTMLButtonElement>('button[aria-controls="account-menu"]')!;
  }

  it("sends the wordmark to the Lobby for a session, and home without one", async () => {
    stubServer();
    const out = render(SiteHeader as never, {} as never);
    await settle();
    expect(out.container.querySelector(".wordmark")?.getAttribute("href")).toBe("#/home");
    expect(out.container.querySelector('a[href="#/login"]')?.textContent).toMatch(/Sign in/);
    out.destroy();

    setSession(sess("player", USER));
    const inn = render(SiteHeader as never, {} as never);
    await settle();
    expect(inn.container.querySelector(".wordmark")?.getAttribute("href")).toBe("#/lobby");
  });

  it("opens a menu with settings, a different account and both sign-outs for a person", async () => {
    setSession(sess("player", USER));
    routes["GET /auth/discord/config"] = () => ({ enabled: true });
    stubServer();
    const { container } = render(SiteHeader as never, {} as never);
    await settle();

    const btn = accountButton(container);
    expect(btn.getAttribute("aria-label")).toBe("account menu: Alice");
    expect(btn.getAttribute("aria-expanded")).toBe("false");
    expect(container.querySelector("#account-menu")).toBeNull();
    click(btn);
    expect(btn.getAttribute("aria-expanded")).toBe("true");

    const menu = container.querySelector("#account-menu")!;
    const items = [...menu.querySelectorAll(".acct-item")].map((el) => el.textContent?.trim());
    expect(items).toEqual([
      "Settings",
      "Sign in with a different Discord account",
      "Sign out",
      "Sign out everywhere",
    ]);
    expect(menu.querySelector("a")?.getAttribute("href")).toBe(
      "/auth/discord/start?prompt=consent",
    );

    // Escape closes it.
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(container.querySelector("#account-menu")).toBeNull();
  });

  it("offers a guest only settings and sign out", async () => {
    setSession(sess("player", NIL, { name: "Guest" }));
    routes["GET /auth/discord/config"] = () => ({ enabled: true });
    stubServer();
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    click(accountButton(container));
    const items = [...container.querySelectorAll("#account-menu .acct-item")].map((el) =>
      el.textContent?.trim(),
    );
    expect(items).toEqual(["Settings", "Sign out"]);
    expect(container.querySelector("#account-menu")?.textContent).toMatch(/guest seat/);
  });

  it("signs out from the menu", async () => {
    setSession(sess("identified", USER));
    stubServer();
    const { container } = render(SiteHeader as never, {} as never);
    await settle();
    click(accountButton(container));
    click(buttonNamed(container.querySelector("#account-menu")!, /^Sign out$/)!);
    await settle();
    expect(calls.some((c) => c.method === "POST" && c.url === "/logout")).toBe(true);
    expect(currentSession()).toBeNull();
    expect(location.hash).toBe("#/login");
  });
});

describe("the login page and #/admin", () => {
  it("has no token card for a signed-out visitor", async () => {
    stubServer();
    const { container } = render(Login as never, {} as never);
    await settle();
    expect(container.textContent).not.toMatch(/Admin log in/);
    expect(container.textContent).not.toMatch(/token/i);
    expect(container.querySelector('input[aria-label="invite code or link"]')).not.toBeNull();
  });

  it("shows #/admin's token form signed out, with no note", async () => {
    stubServer();
    const { container } = render(Login as never, { admin: true } as never);
    await settle();
    expect(container.querySelector('input[placeholder="admin token"]')).not.toBeNull();
    expect(container.querySelector('[role="note"]')).toBeNull();
    expect(container.querySelector('input[aria-label="invite code or link"]')).toBeNull();
  });

  it("tells a signed-in person on #/admin that the token replaces their session", async () => {
    setSession(sess("identified", USER));
    stubServer();
    const { container } = render(Login as never, { admin: true } as never);
    await settle();
    expect(container.querySelector('input[placeholder="admin token"]')).not.toBeNull();
    expect(container.querySelector('[role="note"]')?.textContent).toMatch(
      /replaces this browser's session.*set aside/s,
    );
  });

  it("tells a guest the same, without a sign-in to set aside", async () => {
    setSession(sess("player", NIL, { name: "Guest" }));
    stubServer();
    const { container } = render(Login as never, { admin: true } as never);
    await settle();
    const note = container.querySelector('[role="note"]')?.textContent ?? "";
    expect(note).toMatch(/replaces this browser's session/);
    expect(note).not.toMatch(/set aside/);
  });
});

describe("Home's join card", () => {
  it("links the Lobby for a session, and sign-in without one", async () => {
    stubServer();
    const out = render(Home as never, {} as never);
    await settle();
    const join = () =>
      [...document.querySelectorAll("a.tile")].find((a) =>
        /Join with an invite/.test(a.textContent ?? ""),
      );
    expect(join()?.getAttribute("href")).toBe("#/login");
    out.destroy();

    setSession(sess("player", NIL, { name: "Guest" }));
    render(Home as never, {} as never);
    await settle();
    expect(join()?.getAttribute("href")).toBe("#/lobby");
  });
});
