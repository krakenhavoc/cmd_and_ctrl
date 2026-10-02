// @vitest-environment jsdom
//
// ADR 0110 Delivery PR 7 rendered: the create form for a signed-in
// player with "use my last setup", the tablemate picker that opens on
// the new table with the last table's people first, the deck pickers
// preselecting the last deck, and the Join page pre-filling a guest's
// name.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Lobby from "../routes/Lobby.svelte";
import Join from "../routes/Join.svelte";
import PrebuiltDeckPicker from "./components/PrebuiltDeckPicker.svelte";
import YourDecksPicker from "./components/YourDecksPicker.svelte";
import { GUEST_NAME_KEY } from "./guestName";
import { GUEST_LAST_DECK_KEY } from "./lastDeck";
import { setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const USER = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const BOB = "0b0b0b0b-0000-4000-8000-000000000001";
const CAROL = "0c0c0c0c-0000-4000-8000-000000000002";
const NIL = "00000000-0000-0000-0000-000000000000";
const GAME = "11111111-2222-3333-4444-555555555555";

function sess(role: Session["principal"]["role"], userID: string): Session {
  const seated = role === "player";
  return {
    token: `${role}-${userID.slice(0, 4)}`,
    expiresAt: new Date(Date.now() + 86_400_000).toISOString(),
    principal: {
      role,
      user_id: userID,
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: new Date(Date.now() + 86_400_000).toISOString(),
      game_id: seated ? "other-game" : undefined,
      player_id: seated ? "p1" : undefined,
    },
    gameID: seated ? "other-game" : undefined,
    playerID: seated ? "p1" : undefined,
  };
}

type Route = (method: string, body: unknown) => unknown;
let routes: Record<string, Route> = {};
let posted: Array<{ url: string; body: unknown }> = [];

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      const body = typeof init.body === "string" ? JSON.parse(init.body) : undefined;
      if (method !== "GET") posted.push({ url: input, body });
      const route = routes[`${method} ${input}`];
      const payload = route ? route(method, body) : {};
      return {
        ok: true,
        status: 200,
        statusText: "OK",
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

const createdMeta = {
  id: GAME,
  name: "Friday",
  created_at: new Date().toISOString(),
  players: [],
  state: "lobby",
  is_creator: true,
};

beforeEach(() => {
  routes = {};
  posted = [];
  localStorage.clear();
  location.hash = "#/lobby";
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
  localStorage.clear();
});

describe("the lobby's create flow for a signed-in player", () => {
  it("offers the form with the last setup, creates with it, and opens the tablemates", async () => {
    setSession(sess("identified", USER));
    let created = false;
    routes = {
      "GET /games": () => ({ games: created ? [createdMeta] : [] }),
      "GET /me": () => ({ admin: false }),
      "GET /me/setup": () => ({
        setup: {
          settings: { starting_life: 30 },
          bots: [{ tier: "random", deck_id: "raid", name: "Robo" }],
          tablemates: [CAROL],
        },
      }),
      "GET /bot/options": () => ({
        enabled: true,
        tiers: [],
        decks: [{ id: "raid", name: "Raid and Ransack" }],
      }),
      "POST /games": () => {
        created = true;
        return {
          ...createdMeta,
          invite_token: "inv-player",
          spectator_invite: "inv-spec",
          setup: { settings: true, bots_added: 1, skipped: [] },
        };
      },
      "GET /me/tablemates": () => ({
        tablemates: [
          { user_id: BOB, display_name: "Bob", last_played_at: Date.now() },
          { user_id: CAROL, display_name: "Carol", last_played_at: Date.now() - 86_400_000 },
        ],
      }),
    };
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const form = container.querySelector<HTMLFormElement>("form.create");
    expect(form).not.toBeNull();
    const setupBox = container.querySelector<HTMLInputElement>(".use-setup input");
    expect(setupBox?.checked).toBe(true);
    expect(container.querySelector(".use-setup")?.textContent).toMatch(
      /Use my last setup.*1 bot: Robo \(random, Raid and Ransack\) · starting life 30/s,
    );

    const name = container.querySelector<HTMLInputElement>('input[aria-label="game name"]')!;
    name.value = "Friday";
    name.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    form!.dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();

    expect(posted).toContainEqual({ url: "/games", body: { name: "Friday", setup: "last" } });
    expect(container.textContent).toMatch(/Last setup: table settings applied, 1 bot added\./);
    // The creator's own table: links to share, a seat to take, and the
    // tablemate picker already open, with the last table's people first.
    expect(container.textContent).toMatch(/copy invite/);
    expect(container.textContent).toMatch(/take a seat/);
    const picker = container.querySelector<HTMLDetailsElement>("details.invite-picker");
    expect(picker?.open).toBe(true);
    const mates = [...container.querySelectorAll(".mates li")].map((li) => li.textContent ?? "");
    expect(mates).toHaveLength(2);
    expect(mates[0]).toMatch(/Carol\s*at your last table/);
    expect(mates[1]).toMatch(/Bob/);
    expect(mates[1]).not.toMatch(/at your last table/);
  });

  it("takes the creator's seat with their own session", async () => {
    setSession(sess("identified", USER));
    routes = {
      "GET /games": () => ({ games: [createdMeta] }),
      "GET /me": () => ({ admin: false }),
      "GET /me/setup": () => ({ setup: null }),
      [`GET /games/${GAME}`]: () => ({ ...createdMeta, invite_token: "inv-player" }),
      [`POST /games/${GAME}/join`]: () => ({
        token: "seat-tok",
        expires_at: new Date(Date.now() + 86_400_000).toISOString(),
        principal: { role: "player", user_id: USER, game_id: GAME, player_id: "p9" },
        game: { id: GAME },
        player_id: "p9",
      }),
    };
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    // No last setup: no checkbox, the form still there.
    expect(container.querySelector("form.create")).not.toBeNull();
    expect(container.querySelector(".use-setup")).toBeNull();

    const seat = [...container.querySelectorAll("button")].find((b) =>
      /take a seat/.test(b.textContent ?? ""),
    );
    expect(seat).toBeDefined();
    click(seat!);
    await settle();
    expect(posted).toContainEqual({
      url: `/games/${GAME}/join`,
      body: { invite_token: "inv-player", name: "" },
    });
  });

  it("does not offer a guest the form", async () => {
    setSession(sess("player", NIL));
    routes = { "GET /games": () => ({ games: [] }) };
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    expect(container.querySelector("form.create")).toBeNull();
    expect(container.textContent).toMatch(/ask for an invite link/);
  });
});

describe("the deck pickers preselect the last deck", () => {
  const prebuilt = (id: string, name: string) => ({
    id,
    name,
    card_count: 100,
    coverage: { cards: 10, full: 10, caveats: 0, unreviewed: 0, basics: 1, unregistered: 0 },
  });

  it("checks the last pre-built deck", async () => {
    setSession(sess("player", NIL));
    routes = {
      "GET /decks": () => ({ decks: [prebuilt("raid", "Raid"), prebuilt("deep", "Deep")] }),
    };
    stubServer();
    const { container } = render(
      PrebuiltDeckPicker as never,
      {
        gameID: GAME,
        playerID: "p1",
        lastDeck: { kind: "prebuilt", id: "deep" },
      } as never,
    );
    await settle();
    const checked = container.querySelector<HTMLInputElement>(
      'input[name="prebuilt-deck"]:checked',
    );
    expect(checked?.value).toBe("deep");
  });

  it("checks the last library deck, and ignores one that has gone", async () => {
    setSession(sess("player", USER));
    const deck = (id: string, name: string) => ({
      id,
      name,
      commanders: [],
      card_count: 100,
      updated_at: new Date().toISOString(),
    });
    routes = { "GET /me/decks": () => ({ decks: [deck("lib-a", "A"), deck("lib-b", "B")] }) };
    stubServer();
    const r = render(
      YourDecksPicker as never,
      {
        gameID: GAME,
        playerID: "p1",
        lastDeck: { kind: "library", id: "lib-b" },
      } as never,
    );
    await settle();
    expect(
      r.container.querySelector<HTMLInputElement>('input[name="your-deck"]:checked')?.value,
    ).toBe("lib-b");
    cleanup();

    const gone = render(
      YourDecksPicker as never,
      {
        gameID: GAME,
        playerID: "p1",
        lastDeck: { kind: "library", id: "deleted" },
      } as never,
    );
    await settle();
    expect(
      gone.container.querySelector<HTMLInputElement>('input[name="your-deck"]:checked')?.value,
    ).toBe("lib-a");
  });

  it("remembers a guest's pre-built pick in the browser", async () => {
    setSession(sess("player", NIL));
    routes = {
      "GET /decks": () => ({ decks: [prebuilt("raid", "Raid")] }),
      [`POST /games/${GAME}/decks`]: () => ({
        game: createdMeta,
        deck_name: "Raid",
        card_count: 100,
        deck_id: "raid",
        commanders: [],
      }),
    };
    stubServer();
    const { container } = render(
      PrebuiltDeckPicker as never,
      {
        gameID: GAME,
        playerID: "p1",
      } as never,
    );
    await settle();
    const play = [...container.querySelectorAll("button")].find((b) =>
      /play/.test(b.textContent ?? ""),
    );
    click(play!);
    await settle();
    expect(JSON.parse(localStorage.getItem(GUEST_LAST_DECK_KEY) ?? "null")).toEqual({
      kind: "prebuilt",
      id: "raid",
    });
  });
});

describe("the Join page remembers a guest's name", () => {
  const preview = {
    game: { ...createdMeta, players: [], is_creator: false },
    invite: "player",
    max_seats: 4,
  };

  it("pre-fills the last name and remembers the one joined with", async () => {
    localStorage.setItem(GUEST_NAME_KEY, "Sam");
    routes = {
      [`GET /games/${GAME}/preview?t=inv`]: () => preview,
      "GET /auth/discord/config": () => ({ enabled: false }),
      [`POST /games/${GAME}/join`]: () => ({
        token: "guest-tok",
        expires_at: new Date(Date.now() + 3_600_000).toISOString(),
        principal: { role: "player", user_id: NIL, game_id: GAME, player_id: "p2" },
        game: { id: GAME },
        player_id: "p2",
      }),
    };
    stubServer();
    const { container } = render(
      Join as never,
      {
        gameID: GAME,
        inviteToken: "inv",
        spectator: false,
      } as never,
    );
    await settle();
    const input = container.querySelector<HTMLInputElement>('input[placeholder="your name"]')!;
    expect(input.value).toBe("Sam");

    input.value = "Max";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    container
      .querySelector("form.frow")!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();
    expect(posted).toContainEqual({
      url: `/games/${GAME}/join`,
      body: { invite_token: "inv", name: "Max" },
    });
    expect(localStorage.getItem(GUEST_NAME_KEY)).toBe("Max");
  });

  it("asks a signed-in person for no name at all", async () => {
    localStorage.setItem(GUEST_NAME_KEY, "Sam");
    setSession(sess("identified", USER));
    routes = {
      [`GET /games/${GAME}/preview?t=inv`]: () => preview,
      "GET /auth/discord/config": () => ({ enabled: true }),
    };
    stubServer();
    const { container } = render(
      Join as never,
      {
        gameID: GAME,
        inviteToken: "inv",
        spectator: false,
      } as never,
    );
    await settle();
    expect(container.querySelector('input[placeholder="your name"]')).toBeNull();
    expect(container.textContent).toMatch(/Join as Alice/);
  });
});
