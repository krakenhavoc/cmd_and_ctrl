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
import DeckUploadForm from "./components/DeckUploadForm.svelte";
import PrebuiltDeckPicker from "./components/PrebuiltDeckPicker.svelte";
import YourDecksPicker from "./components/YourDecksPicker.svelte";
import { GUEST_NAME_KEY } from "./guestName";
import { SERVER_FULL_MESSAGE } from "./serverFull";
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

    // The form opens from + New table (ADR 0158 §1 item 1).
    expect(container.querySelector("form.new-table")).toBeNull();
    click(container.querySelector<HTMLButtonElement>('button[aria-label="new table"]')!);
    flushSync();
    const form = container.querySelector<HTMLFormElement>('form[aria-label="create game"]');
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
    expect(container.textContent).toMatch(/Copy invite/);
    // The form closed on success.
    expect(container.querySelector("form.new-table")).toBeNull();
    click(container.querySelector<HTMLButtonElement>('button[aria-label="more for Friday"]')!);
    flushSync();
    expect(container.querySelector('[role="menu"]')?.textContent).toMatch(/Take a seat/);
    expect(container.querySelector(".invite-picker")).not.toBeNull();
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
    click(container.querySelector<HTMLButtonElement>('button[aria-label="new table"]')!);
    flushSync();
    expect(container.querySelector("form.new-table")).not.toBeNull();
    expect(container.querySelector(".use-setup")).toBeNull();

    // Take a seat is in the card's ⋯ menu.
    click(container.querySelector<HTMLButtonElement>('button[aria-label="more for Friday"]')!);
    flushSync();
    const seat = [...container.querySelectorAll("button")].find((b) =>
      /Take a seat/.test(b.textContent ?? ""),
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
    expect(container.querySelector('button[aria-label="new table"]')).toBeNull();
    expect(container.textContent).toMatch(/Ask for an invite link/);
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

describe("the deck pickers stay one screen with many decks", () => {
  const deck = (i: number) => ({
    id: `lib-${i}`,
    name: i === 7 ? "Inverse Nissa" : `Deck ${i}`,
    commanders: [i === 7 ? "Nissa, Leyline Tamer" : `Commander ${i}`],
    card_count: 100,
    updated_at: new Date().toISOString(),
  });

  function setup(n: number) {
    setSession(sess("player", USER));
    routes = {
      "GET /me/decks": () => ({ decks: Array.from({ length: n }, (_, i) => deck(i)) }),
    };
    stubServer();
    return render(YourDecksPicker as never, { gameID: GAME, playerID: "p1" } as never);
  }

  const radios = (c: HTMLElement) =>
    [...c.querySelectorAll<HTMLInputElement>('input[name="your-deck"]')].map((r) => r.value);
  const playButton = (c: HTMLElement) =>
    [...c.querySelectorAll("button")].find((b) => /^play /.test(b.textContent ?? ""));

  it("offers no filter for a short list", async () => {
    const { container } = setup(3);
    await settle();
    expect(container.querySelector('input[type="search"]')).toBeNull();
    expect(radios(container)).toHaveLength(3);
  });

  it("filters a long list by name or commander, and names the deck on the button", async () => {
    const { container } = setup(30);
    await settle();
    const filter = container.querySelector<HTMLInputElement>('input[type="search"]')!;
    expect(filter.getAttribute("aria-label")).toBe("filter your decks");
    expect(radios(container)).toHaveLength(30);
    expect(playButton(container)?.textContent?.trim()).toBe("play Deck 0");

    filter.value = "leyline";
    filter.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(radios(container)).toEqual(["lib-7"]);

    filter.value = "nothing like this";
    filter.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(radios(container)).toEqual([]);
    expect(container.textContent).toContain("No deck matches");
    // The pick survives a filter that hides it.
    expect(playButton(container)?.textContent?.trim()).toBe("play Deck 0");
  });
});

describe("the pre-built decks fold away behind your own", () => {
  const pb = (id: string) => ({
    id,
    name: id,
    card_count: 100,
    coverage: { cards: 10, full: 10, caveats: 0, unreviewed: 0, basics: 1, unregistered: 0 },
  });
  const saved = {
    id: "lib-a",
    name: "Mine",
    commanders: [],
    card_count: 100,
    updated_at: new Date().toISOString(),
  };

  function setup(userID: string, library: unknown[], lastDeck: unknown = null) {
    setSession(sess("player", userID));
    routes = {
      "GET /decks": () => ({ decks: [pb("raid"), pb("deep")] }),
      "GET /me/decks": () => ({ decks: library }),
      "GET /me/last-deck": () => ({ last_deck: lastDeck }),
    };
    stubServer();
    return render(DeckUploadForm as never, { gameID: GAME, playerID: "p1" } as never);
  }

  const prebuiltRadios = (c: HTMLElement) => c.querySelectorAll('input[name="prebuilt-deck"]');
  const toggle = (c: HTMLElement) =>
    [...c.querySelectorAll<HTMLButtonElement>("button")].find((b) =>
      /pre-built deck \(2\)/.test(b.textContent ?? ""),
    );

  it("starts as one line for a player with saved decks, and opens on a click", async () => {
    const { container } = setup(USER, [saved]);
    await settle();
    expect(prebuiltRadios(container)).toHaveLength(0);
    const t = toggle(container)!;
    expect(t.getAttribute("aria-expanded")).toBe("false");
    click(t);
    flushSync();
    expect(t.getAttribute("aria-expanded")).toBe("true");
    expect(prebuiltRadios(container)).toHaveLength(2);
  });

  it("stays open when the last deck seated was a pre-built one", async () => {
    const { container } = setup(USER, [saved], { kind: "prebuilt", id: "deep" });
    await settle();
    expect(toggle(container)).toBeUndefined();
    expect(prebuiltRadios(container)).toHaveLength(2);
  });

  it("stays open with an empty library, and for a guest", async () => {
    const r = setup(USER, []);
    await settle();
    expect(toggle(r.container)).toBeUndefined();
    expect(prebuiltRadios(r.container)).toHaveLength(2);
    cleanup();

    const guest = setup(NIL, []);
    await settle();
    expect(toggle(guest.container)).toBeUndefined();
    expect(prebuiltRadios(guest.container)).toHaveLength(2);
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

  // ADR 0148 / #3103: the Discord callback sends the browser back to the
  // invite's own page when the server is at its connection cap. The page
  // says so in the one sentence every other refusal uses, and offers the
  // Discord button again for the retry.
  it("opens with the server-full message after the Discord callback was refused at the cap", async () => {
    routes = {
      [`GET /games/${GAME}/preview?t=inv`]: () => preview,
      "GET /auth/discord/config": () => ({ enabled: true }),
    };
    stubServer();
    const { container } = render(
      Join as never,
      { gameID: GAME, inviteToken: "inv", spectator: false, serverFull: true } as never,
    );
    await settle();
    const alert = container.querySelector('[role="alert"]');
    expect(alert?.textContent).toBe(SERVER_FULL_MESSAGE);
    expect(container.querySelector("a.discord-btn")).not.toBeNull();
  });

  it("shows no error on a plain visit", async () => {
    routes = {
      [`GET /games/${GAME}/preview?t=inv`]: () => preview,
      "GET /auth/discord/config": () => ({ enabled: true }),
    };
    stubServer();
    const { container } = render(
      Join as never,
      { gameID: GAME, inviteToken: "inv", spectator: false } as never,
    );
    await settle();
    expect(container.querySelector('[role="alert"]')).toBeNull();
  });
});
