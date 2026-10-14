// @vitest-environment jsdom
//
// #3499: a lobby table card with no ⋯. The host toolbar (copy / ↻
// split buttons for both links, the deck and Invite by DM), the rules
// line every seat can read, and the right-click menu that carries the
// same actions plus the admin's.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import Lobby from "../routes/Lobby.svelte";
import { DEFAULT_TABLE_SETTINGS } from "./tableSettings";
import { setSession, type Session } from "./session";
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

function hostRoutes(over: Record<string, unknown> = {}): void {
  routes = {
    "GET /games": () => ({ games: [meta(over)] }),
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

describe("the host toolbar", () => {
  it("puts both links, the deck and Invite by DM on the card, with no ⋯", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    const c = card(container);
    expect(c.querySelector('button[aria-label^="more for"]')).toBeNull();
    const tools = c.querySelector('[aria-label="tools for Friday"]')!;
    expect(tools.textContent).toMatch(/Copy invite/);
    expect(tools.textContent).toMatch(/Copy spectator link/);
    expect(tools.textContent).toMatch(/Choose your deck/);
    expect(tools.textContent).toMatch(/Invite by DM/);
    expect(tools.querySelector('button[aria-label="new invite link"]')).not.toBeNull();
    expect(tools.querySelector('button[aria-label="new spectator link"]')).not.toBeNull();

    // The rules line names only what differs, and opens the panel.
    const rules = c.querySelector<HTMLButtonElement>('button[aria-label^="table settings: "]')!;
    expect(rules.getAttribute("aria-label")).toBe("table settings: 30 life · no undos");
    expect(rules.getAttribute("aria-expanded")).toBe("false");
    click(rules);
    await settle();
    expect(rules.getAttribute("aria-expanded")).toBe("true");
    expect(c.querySelector(".tsettings")).not.toBeNull();
  });

  it("copies the held link from the wide part", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    click(button(card(container), /^\s*Copy invite/)!);
    await settle();
    expect(clipboard).toEqual([`${location.origin}/#/games/${GAME}/join?t=inv-player`]);
    expect(card(container).textContent).toMatch(/Invite copied/);
    expect(posted.some((p) => p.url.endsWith("/invites/rotate"))).toBe(false);
  });

  it("asks before ↻, then mints and copies the new link", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    click(button(card(container), /^new spectator link$/)!);
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

    click(button(card(container), /^new invite link$/)!);
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
    expect(c.textContent).not.toMatch(/Copy invite/);
    expect(c.textContent).toMatch(/Copy spectator link/);
    // Copy only: no ↻ for someone who may not replace the link.
    expect(c.querySelector('button[aria-label="new spectator link"]')).toBeNull();
    expect(c.textContent).toMatch(/Change deck · Atraxa/);
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

    const copy = button(card(container), /^\s*Copy invite/)!;
    copy.focus();
    const e = rightClick(card(container).querySelector(".seats")!);
    expect(e.defaultPrevented).toBe(true);
    await settle();
    expect(menu()).not.toBeNull();
    expect(menuItems()).toEqual(
      expect.arrayContaining([
        "Copy invite",
        "New invite link…",
        "Copy spectator link",
        "New spectator link…",
        "Choose your deck",
        "Invite by DM",
        "Table settings",
      ]),
    );
    // Not an admin: none of the admin's items.
    expect(menuItems()).not.toContain("Archive…");
    expect(document.activeElement?.getAttribute("role")).toBe("menuitem");

    menu()!.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    flushSync();
    expect(menu()).toBeNull();
    expect(document.activeElement).toBe(copy);
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

  it("adds the admin's items for an admin", async () => {
    setSession(seated(HOST, "admin"));
    hostRoutes({ is_creator: false });
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    rightClick(card(container));
    await settle();
    expect(menuItems()).toContain("Archive…");
  });

  it("leaves the browser's own menu on a text field", async () => {
    setSession(seated(HOST));
    hostRoutes();
    stubServer();
    const { container } = render(Lobby as never, {} as never);
    await settle();

    // The deck panel is open while the deck is pending; its fields are
    // where someone pastes a list.
    const field = card(container).querySelector("textarea, input")!;
    expect(field).not.toBeNull();
    const e = rightClick(field);
    expect(e.defaultPrevented).toBe(false);
    expect(menu()).toBeNull();
  });
});
