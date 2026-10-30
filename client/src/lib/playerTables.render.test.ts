// @vitest-environment jsdom
//
// ADR 0110 Delivery PR 7 rendered: the create form for a signed-in
// player with "use my last setup", the tablemate picker that opens on
// the new table with the last table's people first, the deck picker
// (ADR 0212 §3: its tabs, its grids, the last deck preselected, the
// sideboard and Paste a list), and the Join page pre-filling a guest's
// name.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Lobby from "../routes/Lobby.svelte";
import Join from "../routes/Join.svelte";
import DeckPicker from "./components/DeckPicker.svelte";
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
    // The creator's own table: a seat to take as its primary (ADR 0212
    // §2), the links in its Invite menu, and the tablemate picker already
    // open, with the last table's people first.
    // The form closed on success.
    expect(container.querySelector("form.new-table")).toBeNull();
    const foot = container.querySelector<HTMLElement>("li.tcard .tfoot")!;
    expect(foot.querySelector(".btn.primary")?.textContent?.trim()).toBe("Take a seat");
    const invite = [...foot.querySelectorAll<HTMLButtonElement>("button")].find(
      (b) => b.textContent?.trim() === "Invite",
    )!;
    click(invite);
    await settle();
    const menu = document.querySelector('[role="menu"][aria-label="invite to Friday"]');
    expect(menu?.textContent).toMatch(/Copy invite/);
    menu!.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
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

    // Take a seat is the card's primary (ADR 0212 §2).
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

// ADR 0212 §3: one deck picker, a grid behind tabs, in both places that
// seat a deck. These render it alone, with its own button (the table
// route's dialog); tableCard.render.test.ts covers it under a lobby
// card, where the footer's primary commits.
const prebuilt = (id: string, name: string) => ({
  id,
  name,
  card_count: 100,
  coverage: { cards: 10, full: 10, caveats: 0, unreviewed: 0, basics: 1, unregistered: 0 },
});
const saved = (id: string, name: string, extra: Record<string, unknown> = {}) => ({
  id,
  name,
  commanders: [],
  card_count: 100,
  updated_at: new Date().toISOString(),
  ...extra,
});

function renderPicker() {
  return render(DeckPicker as never, { gameID: GAME, playerID: "p1" } as never);
}

const tabNames = (c: HTMLElement) =>
  [...c.querySelectorAll('[role="tab"]')].map((t) => t.textContent?.replace(/\s+/g, " ").trim());
const selectedTab = (c: HTMLElement) =>
  c.querySelector('[role="tab"][aria-selected="true"]')?.textContent?.replace(/\s+/g, " ").trim();
const checked = (c: HTMLElement, name: string) =>
  c.querySelector<HTMLInputElement>(`input[name="${name}"]:checked`)?.value;
const commitButton = (c: HTMLElement) => c.querySelector<HTMLButtonElement>(".row-actions button")!;
function tab(c: HTMLElement, name: RegExp): HTMLButtonElement {
  return [...c.querySelectorAll<HTMLButtonElement>('[role="tab"]')].find((t) =>
    name.test(t.textContent ?? ""),
  )!;
}

describe("the deck picker's tabs (owner decision 7)", () => {
  function setup(userID: string, library: unknown[], lastDeck: unknown = null) {
    setSession(sess("player", userID));
    routes = {
      "GET /decks": () => ({ decks: [prebuilt("raid", "Raid"), prebuilt("deep", "Deep")] }),
      "GET /me/decks": () => ({ decks: library }),
      "GET /me/last-deck": () => ({ last_deck: lastDeck }),
    };
    stubServer();
    return renderPicker();
  }

  it("opens a signed-in person with saved decks on My decks", async () => {
    const { container } = setup(USER, [saved("lib-a", "Mine")]);
    await settle();
    expect(tabNames(container)).toEqual(["My decks 1", "Pre-built 2", "Paste a list"]);
    expect(selectedTab(container)).toBe("My decks 1");
    expect(container.querySelector('[role="radiogroup"][aria-label="your decks"]')).not.toBeNull();
  });

  it("opens on Pre-built when the last deck seated was a pre-built one", async () => {
    const { container } = setup(USER, [saved("lib-a", "Mine")], { kind: "prebuilt", id: "deep" });
    await settle();
    expect(selectedTab(container)).toBe("Pre-built 2");
    expect(checked(container, "prebuilt-deck")).toBe("deep");
  });

  it("gives a guest, or an empty library, no My decks tab, and the note on Paste a list", async () => {
    const r = setup(USER, []);
    await settle();
    expect(tabNames(r.container)).toEqual(["Pre-built 2", "Paste a list"]);
    cleanup();

    const guest = setup(NIL, []);
    await settle();
    expect(tabNames(guest.container)).toEqual(["Pre-built 2", "Paste a list"]);
    expect(selectedTab(guest.container)).toBe("Pre-built 2");
    // A guest is never asked for their library.
    expect(posted).toEqual([]);
    click(tab(guest.container, /Paste a list/));
    await settle();
    expect(guest.container.textContent).toContain(
      "Sign in with Discord to keep the decks you paste.",
    );
    expect(guest.container.querySelector('textarea[aria-label="decklist source"]')).not.toBeNull();
  });

  it("moves between tabs with the arrow keys", async () => {
    const { container } = setup(USER, [saved("lib-a", "Mine")]);
    await settle();
    const list = container.querySelector('[role="tablist"]')!;
    list.dispatchEvent(new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true }));
    flushSync();
    expect(selectedTab(container)).toBe("Pre-built 2");
    expect(document.activeElement?.textContent).toMatch(/Pre-built/);
    list.dispatchEvent(new KeyboardEvent("keydown", { key: "End", bubbles: true }));
    flushSync();
    expect(selectedTab(container)).toBe("Paste a list");
  });
});

describe("the deck picker preselects the last deck", () => {
  it("checks the last pre-built deck, from the browser for a guest", async () => {
    setSession(sess("player", NIL));
    localStorage.setItem(GUEST_LAST_DECK_KEY, JSON.stringify({ kind: "prebuilt", id: "deep" }));
    routes = {
      "GET /decks": () => ({ decks: [prebuilt("raid", "Raid"), prebuilt("deep", "Deep")] }),
    };
    stubServer();
    const { container } = renderPicker();
    await settle();
    expect(checked(container, "prebuilt-deck")).toBe("deep");
  });

  it("flags only the tile whose deck has gaps", async () => {
    setSession(sess("player", NIL));
    const gappy = {
      ...prebuilt("gap", "Gappy"),
      coverage: { ...prebuilt("x", "x").coverage, unregistered: 2 },
    };
    routes = { "GET /decks": () => ({ decks: [prebuilt("raid", "Raid"), gappy] }) };
    stubServer();
    const { container } = renderPicker();
    await settle();
    const flags = [...container.querySelectorAll(".gapflag")].map((e) => e.textContent?.trim());
    expect(flags).toEqual(["2 not implemented"]);
  });

  it("checks the last library deck, and ignores one that has gone", async () => {
    setSession(sess("player", USER));
    routes = {
      "GET /me/decks": () => ({ decks: [saved("lib-a", "A"), saved("lib-b", "B")] }),
      "GET /me/last-deck": () => ({ last_deck: { kind: "library", id: "lib-b" } }),
    };
    stubServer();
    const r = renderPicker();
    await settle();
    expect(checked(r.container, "your-deck")).toBe("lib-b");
    cleanup();

    routes["GET /me/last-deck"] = () => ({ last_deck: { kind: "library", id: "deleted" } });
    const gone = renderPicker();
    await settle();
    expect(checked(gone.container, "your-deck")).toBe("lib-a");
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
    const { container } = renderPicker();
    await settle();
    expect(commitButton(container).textContent?.trim()).toBe("Sit down with this deck");
    click(commitButton(container));
    await settle();
    expect(posted).toContainEqual({
      url: `/games/${GAME}/decks`,
      body: { player_id: "p1", deck: "raid" },
    });
    expect(JSON.parse(localStorage.getItem(GUEST_LAST_DECK_KEY) ?? "null")).toEqual({
      kind: "prebuilt",
      id: "raid",
    });
    expect(container.textContent).toContain("Playing Raid");
  });
});

describe("the deck picker stays one screen with many decks", () => {
  const deck = (i: number) =>
    saved(`lib-${i}`, i === 7 ? "Inverse Nissa" : `Deck ${i}`, {
      commanders: [i === 7 ? "Nissa, Leyline Tamer" : `Commander ${i}`],
    });

  function setup(n: number) {
    setSession(sess("player", USER));
    routes = {
      "GET /me/decks": () => ({ decks: Array.from({ length: n }, (_, i) => deck(i)) }),
      "GET /me/last-deck": () => ({ last_deck: null }),
    };
    stubServer();
    return renderPicker();
  }

  const radios = (c: HTMLElement) =>
    [...c.querySelectorAll<HTMLInputElement>('input[name="your-deck"]')].map((r) => r.value);

  it("offers no filter for a short list", async () => {
    const { container } = setup(3);
    await settle();
    expect(container.querySelector('input[type="search"]')).toBeNull();
    expect(radios(container)).toHaveLength(3);
  });

  it("filters a long list by name or commander, and keeps the pick", async () => {
    const { container } = setup(30);
    await settle();
    const filter = container.querySelector<HTMLInputElement>('input[type="search"]')!;
    expect(filter.getAttribute("aria-label")).toBe("filter your decks");
    expect(radios(container)).toHaveLength(30);

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
    expect(commitButton(container).textContent?.trim()).toBe("Sit down with this deck");
  });

  it("sorts by Recent or Name", async () => {
    setSession(sess("player", USER));
    routes = {
      "GET /me/decks": () => ({
        decks: [saved("z", "Zed"), saved("a", "Atraxa"), saved("m", "Meren")],
      }),
      "GET /me/last-deck": () => ({ last_deck: null }),
    };
    stubServer();
    const { container } = renderPicker();
    await settle();
    expect(radios(container)).toEqual(["z", "a", "m"]);
    click(
      [...container.querySelectorAll<HTMLButtonElement>("button")].find(
        (b) => b.textContent === "Name",
      )!,
    );
    flushSync();
    expect(radios(container)).toEqual(["a", "m", "z"]);
  });
});

describe("the deck picker shows a saved deck's sideboard (#3833)", () => {
  const printing = (name: string, count: number, board: string, entry: number) => ({
    entry,
    name,
    count,
    board,
    printing_id: "p",
    set: "s",
    collector_number: "1",
  });

  function setup() {
    setSession(sess("player", USER));
    routes = {
      "GET /me/decks": () => ({ decks: [saved("lib-a", "Wishes"), saved("lib-b", "Plain")] }),
      "GET /me/last-deck": () => ({ last_deck: null }),
      "GET /me/decks/lib-a/printings": () => ({
        revision: "r",
        cards: [
          printing("Forest", 99, "Mainboard", 0),
          printing("Living Wish", 1, "Sideboard", 1),
          printing("Burning Wish", 1, "Sideboard", 2),
        ],
      }),
      "GET /me/decks/lib-b/printings": () => ({
        revision: "r",
        cards: [printing("Forest", 99, "Mainboard", 0)],
      }),
      [`POST /games/${GAME}/decks/library`]: () => ({}),
    };
    stubServer();
    return renderPicker();
  }

  const sideButton = (c: HTMLElement) =>
    c.querySelector<HTMLButtonElement>('[data-testid="your-deck-sideboard"]')!;
  const list = (c: HTMLElement) => c.querySelector('[data-testid="deck-sideboard"]');

  it("opens the selected deck's sideboard read only, and closes it again", async () => {
    const { container } = setup();
    await settle();
    expect(list(container)).toBeNull();
    expect(sideButton(container).getAttribute("aria-expanded")).toBe("false");
    click(sideButton(container));
    await settle();
    const box = list(container)!;
    expect(box.textContent).toContain("Sideboard");
    expect(box.textContent).toContain("(2)");
    expect(box.textContent).toContain("Burning Wish");
    expect(box.textContent).toContain("Living Wish");
    expect(box.textContent).not.toContain("Forest");
    expect(sideButton(container).getAttribute("aria-expanded")).toBe("true");
    // Read only: its one control is Close.
    expect([...box.querySelectorAll("button")].map((b) => b.textContent?.trim())).toEqual([
      "Close",
    ]);
    click(box.querySelector("button")!);
    flushSync();
    expect(list(container)).toBeNull();
  });

  it("closes when another deck is selected, and says a deck has none", async () => {
    const { container } = setup();
    await settle();
    click(sideButton(container));
    await settle();
    expect(list(container)).not.toBeNull();
    click(container.querySelector<HTMLInputElement>('input[value="lib-b"]')!);
    await settle();
    expect(list(container)).toBeNull();
    click(sideButton(container));
    await settle();
    expect(list(container)?.textContent).toContain("This deck has no sideboard.");
  });
});

describe("Paste a list", () => {
  it("imports a link, then lists what the engine does not carry out", async () => {
    setSession(sess("player", USER));
    routes = {
      "GET /me/decks": () => ({ decks: [] }),
      "GET /me/last-deck": () => ({ last_deck: null }),
      [`POST /games/${GAME}/decks`]: () => ({
        game: createdMeta,
        deck_name: "Linked",
        card_count: 100,
        commanders: ["Atraxa, Praetors' Voice"],
        unimplemented: ["Oko", "Hullbreacher"],
      }),
    };
    let more: boolean | undefined;
    stubServer();
    const { container } = render(
      DeckPicker as never,
      {
        gameID: GAME,
        playerID: "p1",
        onSuccess: (_res: unknown, m: boolean) => (more = m),
      } as never,
    );
    await settle();
    // No library and no pre-built decks: Paste a list is all there is.
    expect(tabNames(container)).toEqual(["Paste a list"]);
    expect(commitButton(container).textContent?.trim()).toBe("Upload deck");
    expect(commitButton(container).disabled).toBe(true);
    // Signed in: no sign-in note.
    expect(container.textContent).not.toContain("Sign in with Discord");
    const box = container.querySelector<HTMLTextAreaElement>("textarea")!;
    box.value = "https://archidekt.com/decks/1";
    box.dispatchEvent(new Event("input", { bubbles: true }));
    flushSync();
    expect(commitButton(container).textContent?.trim()).toBe("Import deck");
    click(commitButton(container));
    await settle();
    expect(posted).toContainEqual({
      url: `/games/${GAME}/decks`,
      body: { player_id: "p1", source: "https://archidekt.com/decks/1" },
    });
    expect(container.textContent).toContain("Uploaded Linked: 100 cards");
    expect(container.querySelector(".unimpl summary")?.textContent).toMatch(/2 of these cards/);
    expect(more).toBe(true);
    expect(box.value).toBe("");
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
