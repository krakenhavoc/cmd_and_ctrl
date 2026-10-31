// @vitest-environment jsdom
//
// ADR 0212: a lobby table's card as a character select. Four seat tiles,
// the settings summary in the head, the deck picker or the bot picker
// under the seats, and a footer with the readiness line, Open table, the
// Invite menu (both links, Invite by DM and the two new links) and one
// primary that follows the table's state, the pickers' commits included.
// The right-click menu (#3499) carries the same actions plus the admin's.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Lobby from "../routes/Lobby.svelte";
import { DEFAULT_TABLE_SETTINGS } from "./tableSettings";
import { currentSession, setSession, type Session } from "./session";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

const USER = "6f9619ff-8b86-d011-b42d-00c04fc964ff";
const GAME = "11111111-2222-3333-4444-555555555555";
const HOST = "a1a1a1a1-0000-4000-8000-000000000001";
const BOB = "b0b0b0b0-0000-4000-8000-000000000002";

function seated(playerID: string, role: Session["principal"]["role"] = "player"): Session {
  const exp = new Date(Date.now() + 86_400_000).toISOString();
  const game = role === "admin" ? undefined : GAME;
  const player = role === "admin" ? undefined : playerID;
  return {
    token: `tok-${playerID.slice(0, 4)}`,
    expiresAt: exp,
    principal: {
      role,
      user_id: role === "admin" ? undefined : USER,
      name: "Alice",
      issued_at: new Date().toISOString(),
      expires_at: exp,
      game_id: game,
      player_id: player,
    },
    gameID: game,
    playerID: player,
  } as Session;
}

function meta(over: Record<string, unknown> = {}) {
  return {
    id: GAME,
    name: "Friday",
    created_at: new Date().toISOString(),
    state: "lobby",
    is_creator: true,
    host_player_id: HOST,
    players: [
      { player_id: HOST, name: "Alice", seat: 1, deck_uploaded: false, is_host: true },
      { player_id: BOB, name: "Bob", seat: 2, deck_uploaded: true, deck_name: "Atraxa" },
    ],
    ...over,
  };
}

type Route = (body: unknown) => unknown;
let routes: Record<string, Route> = {};
let posted: Array<{ url: string; body: unknown }> = [];
let clipboard: string[] = [];
let clipboardFails = false;

function stubServer(): void {
  vi.stubGlobal(
    "fetch",
    vi.fn(async (input: string, init: RequestInit = {}) => {
      const method = init.method ?? "GET";
      const body = typeof init.body === "string" ? JSON.parse(init.body) : undefined;
      if (method !== "GET") posted.push({ url: input, body });
      const route = routes[`${method} ${input}`];
      const payload = route ? route(body) : {};
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
  Object.defineProperty(navigator, "clipboard", {
    configurable: true,
    value: {
      writeText: async (s: string) => {
        if (clipboardFails) throw new Error("denied");
        clipboard.push(s);
      },
    },
  });
}

async function settle(): Promise<void> {
  for (let i = 0; i < 8; i++) {
    for (let j = 0; j < 5; j++) await Promise.resolve();
    await new Promise((r) => setTimeout(r, 0));
    flushSync();
  }
}

const prebuilt = (id: string, name: string, commander: string, colors: string[]) => ({
  id,
  name,
  commander,
  colors,
  card_count: 100,
  coverage: { cards: 60, full: 60, caveats: 0, unreviewed: 0, basics: 2, unregistered: 0 },
});

function hostRoutes(over: Record<string, unknown> = {}): void {
  routes = {
    "GET /games": () => ({ games: [meta(over)] }),
    "GET /decks": () => ({
      decks: [
        prebuilt("raid", "Raid and Ransack", "Gishath, Sun's Avatar", ["R", "G", "W"]),
        prebuilt("deep", "Deep Waters", "Talrand, Sky Summoner", ["U"]),
      ],
    }),
    "GET /me": () => ({ admin: false }),
    "GET /me/setup": () => ({ setup: null }),
    [`GET /games/${GAME}`]: () => ({
      ...meta(over),
      invite_token: "inv-player",
      spectator_invite: "inv-spec",
    }),
    [`GET /games/${GAME}/settings`]: () => ({
      settings: { ...DEFAULT_TABLE_SETTINGS, starting_life: 30, undo_limit: 0 },
    }),
    [`POST /games/${GAME}/invites/rotate`]: (b) => ({
      token: (b as { kind: string }).kind === "player" ? "inv-player-2" : "inv-spec-2",
    }),
  };
}

function card(c: HTMLElement): HTMLElement {
  return c.querySelector<HTMLElement>("li.tcard")!;
}

function button(c: ParentNode, name: RegExp): HTMLButtonElement | undefined {
  return [...c.querySelectorAll<HTMLButtonElement>("button")].find((b) =>
    name.test(b.getAttribute("aria-label") ?? b.textContent ?? ""),
  );
}

function rightClick(el: Element, x = 120, y = 80): MouseEvent {
  const e = new MouseEvent("contextmenu", {
    bubbles: true,
    cancelable: true,
    clientX: x,
    clientY: y,
  });
  el.dispatchEvent(e);
  flushSync();
  return e;
}

function menu(): HTMLElement | null {
  return document.querySelector<HTMLElement>('[role="menu"][aria-label="actions for Friday"]');
}

function menuItems(): string[] {
  return [...(menu()?.querySelectorAll('[role="menuitem"]') ?? [])].map(
    (m) => m.textContent?.trim() ?? "",
  );
}

beforeEach(() => {
  routes = {};
  posted = [];
  clipboard = [];
  clipboardFails = false;
  localStorage.clear();
  location.hash = "#/lobby";
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  setSession(null);
  localStorage.clear();
});

function picker(c: HTMLElement): HTMLElement | null {
  return c.querySelector<HTMLElement>('[role="region"][aria-label="choose your deck"]');
}

function footer(c: HTMLElement): HTMLElement {
  return c.querySelector<HTMLElement>(".tfoot")!;
}

function primary(c: HTMLElement): HTMLButtonElement {
  return footer(c).querySelector<HTMLButtonElement>(".btn.primary")!;
}

function inviteMenu(): HTMLElement | null {
  return document.querySelector<HTMLElement>('[role="menu"][aria-label="invite to Friday"]');
}

function inviteItems(): string[] {
  return [...(inviteMenu()?.querySelectorAll('[role="menuitem"]') ?? [])].map(
    (m) => m.getAttribute("aria-label") ?? m.textContent?.trim() ?? "",
  );
}

async function openInvite(c: HTMLElement): Promise<HTMLButtonElement> {
  const b = button(footer(c), /^\s*Invite\s*$/)!;
  click(b);
  await settle();
  return b;
}

function inviteItem(name: string): HTMLButtonElement {
  return [...inviteMenu()!.querySelectorAll<HTMLButtonElement>('[role="menuitem"]')].find(
    (m) => (m.getAttribute("aria-label") ?? m.textContent?.trim()) === name,
  )!;
}

describe("the lobby card (ADR 0212)", () => {
  it("draws four seat tiles, the settings summary, and a footer with one primary", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const c = card(container);
    expect(c.querySelector('button[aria-label^="more for"]')).toBeNull();
    // No toolbar any more: the links live in the Invite menu.
    expect(c.querySelector('[aria-label="tools for Friday"]')).toBeNull();
    const tiles = [...c.querySelectorAll('ul[aria-label="seats"] > li')].map((li) =>
      li.getAttribute("aria-label"),
    );
    expect(tiles).toEqual(["seat 1: open", "seat 2: Alice", "seat 3: Bob", "seat 4: open"]);
    expect(c.querySelector('li[aria-label="seat 3: Bob"]')?.textContent).toMatch(/Atraxa.*Ready/s);

    // The readiness line counts the seated players and says what it waits on.
    const ready = footer(c).querySelector('[role="group"][aria-label="ready 1 of 2"]');
    expect(ready?.textContent).toMatch(/2 seats open · waiting on your deck/);
    // The picker opened on its own, so the primary seats its deck.
    expect(primary(c).textContent?.trim()).toBe("Sit down with this deck");
    expect(button(footer(c), /^\s*Open table\s*$/)).toBeDefined();

    // The settings summary names life and undos, and opens the panel.
    const rules = c.querySelector<HTMLButtonElement>('button[aria-label^="table settings: "]')!;
    expect(rules.closest(".thead")).not.toBeNull();
    expect(rules.getAttribute("aria-label")).toBe(
      "table settings: 30 life · no undos · standard rules",
    );
    expect(rules.getAttribute("aria-expanded")).toBe("false");
    click(rules);
    await settle();
    expect(rules.getAttribute("aria-expanded")).toBe("true");
    expect(c.querySelector(".tsettings")).not.toBeNull();
  });

  it("opens your deck picker from your seat's tile and from Choose your deck", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    // Seated without a deck: the picker opens on its own.
    const tile = card(container).querySelector<HTMLButtonElement>(
      'button[aria-label="choose a deck for your seat"]',
    )!;
    expect(tile.getAttribute("aria-expanded")).toBe("true");
    expect(picker(container)).not.toBeNull();
    click(tile);
    await settle();
    expect(tile.getAttribute("aria-expanded")).toBe("false");
    expect(picker(container)).toBeNull();
    // Closed, the primary opens it.
    expect(primary(card(container)).textContent?.trim()).toBe("Choose your deck");

    click(primary(card(container)));
    await settle();
    expect(picker(container)).not.toBeNull();
    expect(document.activeElement).toBe(picker(container));
    expect(primary(card(container)).textContent?.trim()).toBe("Sit down with this deck");
  });

  it("previews the selected deck in your seat, and sits down from the footer", async () => {
    setSession(seated(HOST));
    hostRoutes();
    let seatedDeck = "";
    // The answer carries the table with the deck in; the list is read
    // again after it.
    const after = meta({
      players: [
        { ...meta().players[0], deck_uploaded: true, deck_name: "Deep Waters" },
        meta().players[1],
      ],
    });
    routes[`POST /games/${GAME}/decks`] = (b) => {
      seatedDeck = (b as { deck?: string }).deck ?? "";
      routes["GET /games"] = () => ({ games: [after] });
      return { game: after, deck_name: "Deep Waters", card_count: 100, commanders: [] };
    };
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    // A signed-in person with no saved decks gets no My decks tab and
    // opens on Pre-built, the first deck selected.
    const tabs = [...picker(container)!.querySelectorAll('[role="tab"]')].map((t) =>
      t.textContent?.replace(/\s+/g, " ").trim(),
    );
    expect(tabs).toEqual(["Pre-built 2", "Paste a list"]);
    const mine = () =>
      card(container).querySelector<HTMLElement>('li[aria-label="seat 2: Alice"]')!;
    expect(mine().textContent).toMatch(/Preview/);
    expect(mine().textContent).toMatch(/Gishath, Sun's Avatar/);
    expect(mine().textContent).toMatch(/Pre-built · Raid and Ransack/);
    expect(mine().textContent).toMatch(/Not seated yet/);
    expect(mine().textContent).not.toMatch(/Ready/);
    // Nothing is sent by choosing.
    click(picker(container)!.querySelector<HTMLInputElement>('input[value="deep"]')!);
    await settle();
    expect(mine().textContent).toMatch(/Talrand, Sky Summoner/);
    expect(posted).toEqual([]);

    click(primary(card(container)));
    await settle();
    expect(posted.map((p) => p.url)).toEqual([`/games/${GAME}/decks`]);
    expect(seatedDeck).toBe("deep");
    // Seated: the picker closes, and the seat says Ready.
    expect(picker(container)).toBeNull();
    expect(mine().textContent).not.toMatch(/Preview/);
    expect(mine().textContent).toMatch(/Deep Waters.*Ready/s);
  });

  it("makes the primary Upload deck on Paste a list, disabled until there is a list", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const paste = [...picker(container)!.querySelectorAll<HTMLButtonElement>('[role="tab"]')].find(
      (t) => t.textContent?.trim() === "Paste a list",
    )!;
    click(paste);
    await settle();
    expect(primary(card(container)).textContent?.trim()).toBe("Upload deck");
    expect(primary(card(container)).disabled).toBe(true);
    // Nothing selected in the grid now: the seat says it has no deck.
    expect(card(container).querySelector('li[aria-label="seat 2: Alice"]')?.textContent).toMatch(
      /No deck yet/,
    );
    const box = picker(container)!.querySelector<HTMLTextAreaElement>(
      'textarea[aria-label="decklist source"]',
    )!;
    box.value = "https://archidekt.com/decks/1";
    box.dispatchEvent(new Event("input", { bubbles: true }));
    await settle();
    expect(primary(card(container)).textContent?.trim()).toBe("Import deck");
    expect(primary(card(container)).disabled).toBe(false);
  });

  it("offers the host Start game once every deck is in, naming the count", async () => {
    setSession(seated(HOST));
    hostRoutes({
      players: [
        { player_id: HOST, name: "Alice", seat: 0, deck_uploaded: true, is_host: true },
        { player_id: BOB, name: "Bob", seat: 1, deck_uploaded: true, deck_name: "Atraxa" },
      ],
    });
    routes[`POST /games/${GAME}/start`] = () => ({});
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    expect(footer(card(container)).textContent).toMatch(/2 seats open · every deck is in/);
    expect(primary(card(container)).textContent?.trim()).toBe("Start game · 2 players");
    click(primary(card(container)));
    await settle();
    expect(posted).toContainEqual({ url: `/games/${GAME}/start`, body: undefined });
  });

  it("puts both links, Invite by DM and the new links in one Invite menu", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const opener = await openInvite(card(container));
    expect(opener.getAttribute("aria-expanded")).toBe("true");
    expect(inviteItems()).toEqual([
      "Copy invite",
      "Copy spectator link",
      "Invite by DM…",
      "New invite link…",
      "New spectator link…",
    ]);
    expect(
      inviteMenu()!.querySelector('[role="group"][aria-label="Make new links"]'),
    ).not.toBeNull();
    expect(document.activeElement?.getAttribute("role")).toBe("menuitem");
    // An open seat's Invite opens the same menu.
    inviteMenu()!.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(inviteMenu()).toBeNull();
    expect(document.activeElement).toBe(opener);
    const seatInvite = button(
      card(container).querySelector('li[aria-label="seat 4: open"]')!,
      /^\s*Invite\s*$/,
    )!;
    click(seatInvite);
    await settle();
    expect(inviteItems()).toContain("Copy invite");
  });

  it("copies the held link from the menu", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const opener = await openInvite(card(container));
    click(inviteItem("Copy invite"));
    await settle();
    expect(clipboard).toEqual([`${location.origin}/#/games/${GAME}/join?t=inv-player`]);
    expect(footer(card(container)).querySelector('[role="status"]')?.textContent).toBe(
      "Invite copied",
    );
    expect(posted.some((p) => p.url.endsWith("/invites/rotate"))).toBe(false);
    expect(inviteMenu()).toBeNull();
    expect(document.activeElement).toBe(opener);
  });

  it("asks before a new link, then mints and copies it", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    await openInvite(card(container));
    click(inviteItem("New spectator link…"));
    flushSync();
    // Nothing is minted until the confirm.
    expect(posted).toEqual([]);
    expect(card(container).textContent).toMatch(/Mint a new spectator link\?/);
    click(button(card(container), /mint new link/)!);
    await settle();

    expect(posted).toContainEqual({
      url: `/games/${GAME}/invites/rotate`,
      body: { kind: "spectator" },
    });
    expect(clipboard).toHaveLength(1);
    expect(clipboard[0]).toMatch(/inv-spec-2/);
    expect(card(container).textContent).toMatch(/New spectator link copied/);
    // Copied, so no panel to copy it from by hand.
    expect(card(container).querySelector('input[aria-label="new spectator link"]')).toBeNull();
  });

  it("shows the new link to copy by hand when the clipboard refuses", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    clipboardFails = true;
    const { container } = render(Lobby as never, {} as never);
    await settle();

    await openInvite(card(container));
    click(inviteItem("New invite link…"));
    flushSync();
    click(button(card(container), /mint new link/)!);
    await settle();
    const field = card(container).querySelector<HTMLInputElement>(
      'input[aria-label="new player invite"]',
    );
    expect(field?.value).toMatch(/inv-player-2/);
    expect(container.querySelector(".error")).toBeNull();
  });

  it("gives a seat who does not host a read-only version", async () => {
    setSession(seated(BOB));
    hostRoutes({ is_creator: false });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const c = card(container);
    // Their deck is in; the host's is not.
    expect(primary(c).textContent?.trim()).toBe("Waiting for 1 player");
    expect(primary(c).disabled).toBe(true);
    expect(primary(c).title).toBe("waiting on Alice's deck");
    expect(c.querySelector('button[aria-label="change your deck"]')).not.toBeNull();
    // The spectator link, and no new links for someone who may not
    // replace one.
    await openInvite(c);
    expect(inviteItems()).toContain("Copy spectator link");
    expect(inviteItems()).not.toContain("Copy invite");
    expect(inviteItems()).not.toContain("New spectator link…");
    // They read the rules; the panel says only the host changes them.
    const rules = c.querySelector<HTMLButtonElement>('button[aria-label^="table settings: "]')!;
    expect(rules.title).toMatch(/only the host can change them/);
  });
});

describe("the card's right-click menu", () => {
  it("opens at the cursor with the card's actions, and Escape gives focus back", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const invite = button(footer(card(container)), /^\s*Invite\s*$/)!;
    invite.focus();
    const e = rightClick(card(container).querySelector('ul[aria-label="seats"]')!);
    expect(e.defaultPrevented).toBe(true);
    await settle();
    expect(menu()).not.toBeNull();
    expect(menuItems()).toEqual(
      expect.arrayContaining([
        "Sit down with this deck",
        "Open table",
        "Copy invite",
        "New invite link…",
        "Copy spectator link",
        "New spectator link…",
        "Invite by DM…",
        "Table settings",
      ]),
    );
    // Not an admin: none of the admin's items.
    expect(menuItems()).not.toContain("Archive…");
    expect(menuItems().some((m) => m.startsWith("Seat link"))).toBe(false);
    expect(document.activeElement?.getAttribute("role")).toBe("menuitem");

    menu()!.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(menu()).toBeNull();
    expect(document.activeElement).toBe(invite);
  });

  // The e2e caught this: the browser scrolls the card into view, and the
  // scroll event lands after the menu has mounted.
  it("stays open for a late scroll event, and closes once the page moves", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    rightClick(card(container));
    await settle();
    window.dispatchEvent(new Event("scroll"));
    flushSync();
    expect(menu()).not.toBeNull();

    Object.defineProperty(window, "scrollY", { configurable: true, value: 300 });
    try {
      window.dispatchEvent(new Event("scroll"));
      flushSync();
      expect(menu()).toBeNull();
    } finally {
      Object.defineProperty(window, "scrollY", { configurable: true, value: 0 });
    }
  });

  it("runs an item and closes", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    rightClick(card(container));
    await settle();
    click(button(menu()!, /^Table settings$/)!);
    await settle();
    expect(menu()).toBeNull();
    expect(card(container).querySelector(".tsettings")).not.toBeNull();
  });

  it("adds the admin's items for an admin, a seat link per person among them", async () => {
    setSession(seated(HOST, "admin"));
    hostRoutes({ is_creator: false });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    // The seat link moved off the tiles into this menu (ADR 0212 §1).
    expect(card(container).querySelector('button[aria-label^="seat link"]')).toBeNull();
    rightClick(card(container));
    await settle();
    expect(menuItems()).toEqual(
      expect.arrayContaining([
        "Seat link for Alice (seat 2)",
        "Seat link for Bob (seat 3)",
        "Archive…",
      ]),
    );
    // The admin, not seated, at a table that cannot start: Open table.
    expect(primary(card(container)).textContent?.trim()).toBe("Open table");
  });

  it("leaves the browser's own menu on a text field", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    // The deck picker is open while the deck is pending; Paste a list is
    // where someone pastes one.
    click(
      [...card(container).querySelectorAll<HTMLButtonElement>('[role="tab"]')].find(
        (t) => t.textContent?.trim() === "Paste a list",
      )!,
    );
    await settle();
    const field = card(container).querySelector("textarea")!;
    expect(field).not.toBeNull();
    const e = rightClick(field);
    expect(e.defaultPrevented).toBe(false);
    expect(menu()).toBeNull();
  });
});

// Owner decision 8: Add bot opens the bot picker in the deck picker's
// place under the seats, and the footer's primary adds the bot.
describe("the bot picker", () => {
  const options = {
    enabled: true,
    tiers: [
      { tier: "random", label: "Random", description: "Any legal move.", available: true },
      { tier: "heuristic", label: "Heuristic", description: "Scores the board.", available: true },
      {
        tier: "strong",
        label: "Strong",
        description: "A model.",
        available: false,
        reason: "this server has no model API key",
      },
    ],
    decks: [
      { id: "raid", name: "Raid and Ransack", commander: "Gishath, Sun's Avatar", colors: ["R"] },
      { id: "deep", name: "Deep Waters", commander: "Talrand, Sky Summoner", colors: ["U"] },
    ],
  };

  it("titles the seat it fills, offers the tiers and the pre-built grid, and adds from the footer", async () => {
    setSession(seated(HOST));
    hostRoutes({
      players: [
        { player_id: HOST, name: "Alice", seat: 0, deck_uploaded: true, is_host: true },
        { player_id: BOB, name: "Bob", seat: 1, deck_uploaded: true, deck_name: "Atraxa" },
      ],
    });
    routes["GET /bot/options"] = () => options;
    routes[`POST /games/${GAME}/seats/bot`] = () => ({
      game: meta(),
      player_id: "bot",
      deck_name: "",
    });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const c = card(container);
    expect(primary(c).textContent?.trim()).toBe("Start game · 2 players");
    click(button(c.querySelector('li[aria-label="seat 4: open"]')!, /Add bot/)!);
    await settle();

    // The first open seat is the one a bot takes.
    const region = c.querySelector<HTMLElement>("section[aria-labelledby]")!;
    expect(region.querySelector("h4")?.textContent).toBe("Add a bot to seat 3");
    expect(picker(container)).toBeNull();
    const tiers = region.querySelector('[role="radiogroup"][aria-label="bot tier"]')!;
    const radio = (v: string) => tiers.querySelector<HTMLInputElement>(`input[value="${v}"]`)!;
    expect(radio("heuristic").checked).toBe(true);
    expect(radio("strong").disabled).toBe(true);
    expect(region.textContent).toContain("Strong: this server has no model API key");
    // No <select> any more: the decks are the pre-built grid.
    expect(region.querySelector("select")).toBeNull();
    const grid = region.querySelector('[role="radiogroup"][aria-label="pre-built decks"]')!;
    click(grid.querySelector<HTMLInputElement>('input[value="deep"]')!);
    click(radio("random"));
    await settle();
    expect(region.textContent).toMatch(/not an opponent/i);

    // One primary: the footer's adds the bot.
    expect(primary(c).textContent?.trim()).toBe("Add bot");
    click(primary(c));
    await settle();
    expect(posted).toContainEqual({
      url: `/games/${GAME}/seats/bot`,
      body: { tier: "random", deck: "deep" },
    });
    expect(c.querySelector("section[aria-labelledby]")).toBeNull();
  });

  it("closes the deck picker to open, and Cancel gives the region back", async () => {
    setSession(seated(HOST));
    hostRoutes();
    routes["GET /bot/options"] = () => options;
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const c = card(container);
    expect(picker(container)).not.toBeNull();
    click(button(c.querySelector('li[aria-label="seat 1: open"]')!, /Add bot/)!);
    await settle();
    expect(picker(container)).toBeNull();
    expect(c.querySelector("section[aria-labelledby] h4")?.textContent).toBe("Add a bot to seat 1");
    click(button(c.querySelector("section[aria-labelledby]")!, /^\s*Cancel\s*$/)!);
    await settle();
    expect(c.querySelector("section[aria-labelledby]")).toBeNull();
    // Your deck is still to choose.
    expect(primary(c).textContent?.trim()).toBe("Choose your deck");
  });
});

// #3494: every seat's × asks first, a bot's and the host's own included.
describe("removing a seat", () => {
  const BOT = "c0c0c0c0-0000-4000-8000-000000000003";
  const withBot = {
    players: [
      { player_id: HOST, name: "Alice", seat: 1, deck_uploaded: true, is_host: true },
      { player_id: BOT, name: "Robo", seat: 2, deck_uploaded: true, is_bot: true },
    ],
  };

  it("asks before removing a bot, then uses the bot route", async () => {
    setSession(seated(HOST));
    hostRoutes(withBot);
    routes["GET /bot/options"] = () => ({
      enabled: true,
      tiers: [{ tier: "random", available: true }],
      decks: [{ id: "raid", name: "Raid" }],
    });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    click(button(card(container), /^remove Robo$/)!);
    flushSync();
    expect(posted).toEqual([]);
    expect(card(container).textContent).toMatch(/Remove Robo from this table\?/);
    // No "block" for a bot.
    expect(card(container).querySelector(".block-opt")).toBeNull();
    click(button(card(container), /remove from table/)!);
    await settle();
    expect(posted).toContainEqual({ url: `/games/${GAME}/seats/bot/${BOT}`, body: undefined });
  });

  it("lets the host leave their own seat, and keeps them signed in", async () => {
    setSession(seated(HOST));
    hostRoutes(withBot);
    const exp = new Date(Date.now() + 86_400_000).toISOString();
    routes[`DELETE /games/${GAME}/seats/${HOST}`] = () => ({
      ...meta(),
      players: [],
      session: {
        token: "identity-tok",
        expires_at: exp,
        principal: { role: "identified", user_id: USER, name: "Alice", expires_at: exp },
      },
    });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    click(button(card(container), /^leave your seat$/)!);
    flushSync();
    expect(card(container).textContent).toMatch(/Leave your seat at this table\?/);
    expect(card(container).textContent).toMatch(/The table stays yours/);
    click(button(card(container), /leave my seat/)!);
    await settle();

    expect(posted).toContainEqual({ url: `/games/${GAME}/seats/${HOST}`, body: undefined });
    const now = currentSession();
    expect(now?.token).toBe("identity-tok");
    expect(now?.playerID).toBeUndefined();
  });
});

describe("creating a table seats its creator", () => {
  it("installs the seat session the create answers with", async () => {
    const exp = new Date(Date.now() + 86_400_000).toISOString();
    setSession({
      token: "identity",
      expiresAt: exp,
      principal: { role: "identified", user_id: USER, name: "Alice", expires_at: exp },
    } as Session);
    const one = { players: [meta().players[0]] };
    let created = false;
    routes = {
      "GET /games": () => ({ games: created ? [meta(one)] : [] }),
      "GET /me": () => ({ admin: false }),
      "GET /me/setup": () => ({ setup: null }),
      "POST /games": () => {
        created = true;
        return {
          ...meta(one),
          invite_token: "inv-player",
          spectator_invite: "inv-spec",
          seat: {
            token: "seat-tok",
            expires_at: exp,
            principal: { role: "player", user_id: USER, game_id: GAME, player_id: HOST },
            game: { id: GAME },
            player_id: HOST,
          },
        };
      },
    };
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();
    click(container.querySelector<HTMLButtonElement>('button[aria-label="new table"]')!);
    flushSync();
    container
      .querySelector<HTMLFormElement>('form[aria-label="create game"]')!
      .dispatchEvent(new Event("submit", { bubbles: true, cancelable: true }));
    await settle();

    expect(currentSession()?.token).toBe("seat-tok");
    expect(currentSession()?.playerID).toBe(HOST);
    // Seated already: no Take a seat, and no second call to join.
    expect(card(container).textContent).not.toMatch(/Take a seat/);
    expect(posted.some((p) => p.url.endsWith("/join"))).toBe(false);
  });
});
