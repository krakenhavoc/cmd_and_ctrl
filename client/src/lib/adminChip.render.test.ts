// @vitest-environment jsdom
//
// ADR 0112 Delivery PR 4, rendered: the Admin chip in the header's
// account menu ("Admin · until HH:MM" on the accent colour, "Player"
// muted), the token's static "Admin token" badge, the same switch in the
// in-game ⋯ menu, and what a 4001 does at the table: reconnect to a seat
// the person may hold, or go to the Lobby with a notice when the
// connection needed admin mode.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { tick } from "svelte";
import { get } from "svelte/store";

import SiteHeader from "./components/SiteHeader.svelte";
import GameMenu from "./components/board/GameMenu.svelte";
import { adminNotice, PLAYER_MODE_NOT_YOURS, resetAdminChecksForTest } from "./admin";
import type { GameMenuOptions } from "./gameMenu";
import { PROTOCOL_VERSION, type GameView, type PlayerView } from "./protocol";
import { route } from "./router";
import { session, type Session } from "./session";
import { _resetForTests as resetDock } from "./dock";
import { _resetForTests as resetModals } from "./modalLayers";
import { cleanup, click, flushSync, render } from "./test/render.svelte";

vi.mock("./components/board/Board.svelte", async () => ({
  default: (await import("./test/BoardAttentionStub.svelte")).default,
}));

const Game = await import("../routes/Game.svelte").then((m) => m.default);
const Lobby = await import("../routes/Lobby.svelte").then((m) => m.default);

const USER = "5b0d6a3e-8f7f-4e0e-9b1a-0f3c1d2e4a5b";
const GAME = "game-1";
const ME = "p-1";
const HOUR = 3_600_000;

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

class FakeSocket {
  static readonly OPEN = 1;
  static all: FakeSocket[] = [];
  readyState = FakeSocket.OPEN;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};
  constructor(readonly url: string) {
    FakeSocket.all.push(this);
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(): void {}
  close(): void {}
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

// --- a tiny server ---------------------------------------------------

// mode is what the server holds for this person; PUT flips it, /me
// reports it. refuseSwitch makes the next PUT a 403.
let mode: "admin" | "player" = "player";
let refuseSwitch = false;
let puts: unknown[] = [];
const ENDS = new Date(2026, 9, 2, 23, 40).getTime();

function modeBody() {
  const on = mode === "admin";
  return {
    admin: on,
    admin_allowed: true,
    admin_mode: on,
    ...(on ? { admin_mode_ends_at: ENDS } : {}),
  };
}

function json(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    statusText: status === 403 ? "Forbidden" : "OK",
  });
}

async function serve(input: string, init: RequestInit = {}): Promise<Response> {
  const method = init.method ?? "GET";
  if (input === "/me" && method === "GET") return json(modeBody());
  if (input === "/me/admin-mode" && method === "PUT") {
    const body = JSON.parse(String(init.body)) as { on: boolean };
    puts.push(body);
    if (refuseSwitch) return json({ error: "lobby: not an admin" }, 403);
    mode = body.on ? "admin" : "player";
    return json(modeBody());
  }
  if (input === "/auth/discord/config") return json({ enabled: false });
  if (input.startsWith("/games") || input.startsWith("/me/games")) return json({ games: [] });
  return new Response("{}", { status: 404 });
}

function person(
  over: Partial<Session> & { role?: Session["principal"]["role"]; gameID?: string } = {},
): Session {
  const { role = "player", gameID = GAME, ...rest } = over;
  const expiresAt = new Date(Date.now() + 24 * HOUR).toISOString();
  const playerID = role === "player" ? ME : undefined;
  return {
    token: "tok",
    expiresAt,
    principal: {
      role,
      user_id: USER,
      name: "Owner",
      game_id: role === "identified" ? undefined : gameID,
      player_id: playerID,
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
    gameID: role === "identified" ? undefined : gameID,
    playerID,
    ...rest,
  };
}

const playerMode: Partial<Session> = { admin: false, admin_allowed: true, admin_mode: false };
const adminMode: Partial<Session> = {
  admin: true,
  admin_allowed: true,
  admin_mode: true,
  admin_mode_ends_at: ENDS,
};

let realWebSocket: unknown;
let realFetch: unknown;

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  realWebSocket = g.WebSocket;
  realFetch = g.fetch;
  g.WebSocket = FakeSocket;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  g.fetch = vi.fn(serve);
  FakeSocket.all = [];
  mode = "player";
  refuseSwitch = false;
  puts = [];
  resetAdminChecksForTest();
  resetDock();
  resetModals();
  adminNotice.set("");
  vi.useFakeTimers({ toFake: ["Date"] });
  vi.setSystemTime(new Date(2026, 9, 2, 12, 0));
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
});

afterEach(() => {
  cleanup();
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
  session.set(null);
  vi.useRealTimers();
  vi.restoreAllMocks();
});

async function settle(): Promise<void> {
  for (let i = 0; i < 4; i++) {
    await tick();
    await new Promise((r) => setTimeout(r, 0));
  }
  flushSync();
}

async function openAccount(container: HTMLElement): Promise<HTMLElement> {
  click(container.querySelector<HTMLButtonElement>('button[aria-controls="account-menu"]')!);
  await settle();
  return container.querySelector<HTMLElement>("#account-menu")!;
}

const chipOf = (menu: ParentNode) =>
  menu.querySelector<HTMLElement>("button[aria-pressed], .mode-chip");

describe("the Admin chip in the header's account menu", () => {
  it("reads Player, muted and unpressed, in player mode", async () => {
    session.set(person(playerMode));
    const { container } = render(SiteHeader as never, {} as never);
    const menu = await openAccount(container);
    const chip = chipOf(menu)!;
    expect(chip.tagName).toBe("BUTTON");
    expect(chip.textContent?.trim()).toBe("Player");
    expect(chip.getAttribute("aria-pressed")).toBe("false");
    expect(chip.getAttribute("aria-label")).toBe("admin mode: Player");
    expect(chip.classList.contains("on")).toBe(false);
    expect(chip.getAttribute("title")).toMatch(/Switch to admin mode/);
    // First in the menu, before Settings.
    const first = menu.querySelector(".acct-mode, .acct-item")!;
    expect(first.classList.contains("acct-mode")).toBe(true);
  });

  it("reads Admin · until HH:MM on the accent colour in admin mode", async () => {
    session.set(person(adminMode));
    const { container } = render(SiteHeader as never, {} as never);
    const chip = chipOf(await openAccount(container))!;
    expect(chip.textContent?.trim()).toBe("Admin · until 23:40");
    expect(chip.getAttribute("aria-pressed")).toBe("true");
    expect(chip.classList.contains("on")).toBe(true);
    expect(chip.getAttribute("title")).toMatch(/Switch to player mode/);
  });

  it("switches with PUT /me/admin-mode and shows the answer", async () => {
    session.set(person(playerMode));
    const { container } = render(SiteHeader as never, {} as never);
    const menu = await openAccount(container);
    click(chipOf(menu)!);
    await settle();
    expect(puts).toEqual([{ on: true }]);
    expect(chipOf(menu)!.textContent?.trim()).toBe("Admin · until 23:40");
    expect(get(session)?.admin_mode).toBe(true);

    click(chipOf(menu)!);
    await settle();
    expect(puts).toEqual([{ on: true }, { on: false }]);
    expect(chipOf(menu)!.textContent?.trim()).toBe("Player");
  });

  it("leaves the chip as it was on a refusal, and shows the server's message", async () => {
    session.set(person(playerMode));
    refuseSwitch = true;
    const { container } = render(SiteHeader as never, {} as never);
    const menu = await openAccount(container);
    click(chipOf(menu)!);
    await settle();
    expect(chipOf(menu)!.textContent?.trim()).toBe("Player");
    expect(menu.querySelector('[role="alert"]')?.textContent).toMatch(/not an admin/);
  });

  it("turns back to Player when admin mode lapses", async () => {
    session.set(person(adminMode));
    const { container } = render(SiteHeader as never, {} as never);
    const menu = await openAccount(container);
    expect(chipOf(menu)!.textContent?.trim()).toBe("Admin · until 23:40");
    // The lapse timer (App.svelte) writes the session; here, directly.
    vi.setSystemTime(ENDS + 1000);
    session.update((s) => (s ? { ...s } : s));
    flushSync();
    expect(chipOf(menu)!.textContent?.trim()).toBe("Player");
  });

  it("is the static Admin token badge for the shared token", async () => {
    const exp = new Date(Date.now() + HOUR).toISOString();
    session.set({
      token: "admin-tok",
      expiresAt: exp,
      principal: { role: "admin", name: "admin", issued_at: "", expires_at: exp },
    });
    const { container } = render(SiteHeader as never, {} as never);
    const menu = await openAccount(container);
    const badge = menu.querySelector(".mode-chip")!;
    expect(badge.tagName).toBe("SPAN");
    expect(badge.textContent?.trim()).toBe("Admin token");
    expect(menu.querySelector("button[aria-pressed]")).toBeNull();
  });

  it("is absent for a person who is not on the list, and for a guest", async () => {
    session.set(person({ admin: false, admin_allowed: false, admin_mode: false }));
    const one = render(SiteHeader as never, {} as never);
    expect(chipOf(await openAccount(one.container))).toBeNull();
    one.destroy();

    const exp = new Date(Date.now() + HOUR).toISOString();
    session.set({
      token: "guest",
      expiresAt: exp,
      principal: { role: "player", game_id: GAME, player_id: ME, issued_at: "", expires_at: exp },
      gameID: GAME,
      playerID: ME,
      admin_allowed: true,
    });
    const two = render(SiteHeader as never, {} as never);
    expect(chipOf(await openAccount(two.container))).toBeNull();
  });
});

// --- the ⋯ menu ------------------------------------------------------

function mountMenu(over: Partial<GameMenuOptions>) {
  const noop = () => {};
  return render(
    GameMenu as never,
    {
      seated: true,
      canManage: false,
      spawnAvailable: false,
      myGames: true,
      voteOpen: false,
      onDraw: noop,
      onUntapAll: noop,
      onShuffle: noop,
      onMulligan: noop,
      onLifeHistory: noop,
      onTableSettings: noop,
      onSpawn: noop,
      onMyGames: noop,
      onBack: noop,
      onConcede: noop,
      onStartVote: noop,
      ...over,
    } as never,
  );
}

const menuItemNames = (root: ParentNode) =>
  [...root.querySelectorAll('[role="menu"] [role="menuitem"]')].map((el) =>
    (el.textContent ?? "").replace(/\s+/g, " ").trim(),
  );

async function openMore(root: HTMLElement): Promise<void> {
  click(root.querySelector<HTMLButtonElement>('button[aria-label="more actions"]')!);
  await settle();
}

describe("the in-game ⋯ menu's switch", () => {
  it("offers the other mode to an allowlisted person, and calls back", async () => {
    const calls: number[] = [];
    const off = mountMenu({ adminMode: { on: false }, onAdminMode: () => calls.push(1) });
    await openMore(off.container);
    const names = menuItemNames(off.container);
    expect(names).toContain("Switch to admin mode player");
    const switchItem = [...off.container.querySelectorAll<HTMLElement>('[role="menuitem"]')].find(
      (el) => el.textContent?.includes("Switch to admin mode"),
    )!;
    click(switchItem);
    expect(calls).toEqual([1]);
    off.destroy();

    const on = mountMenu({ adminMode: { on: true }, onAdminMode: () => {} });
    await openMore(on.container);
    expect(menuItemNames(on.container)).toContain("Switch to player mode admin");
  });

  it("is not there for anyone else", async () => {
    const m = mountMenu({ adminMode: null });
    await openMore(m.container);
    expect(menuItemNames(m.container).some((n) => /^Switch to/.test(n))).toBe(false);
  });
});

// --- at the table ----------------------------------------------------

const zone = (kind: string, owner?: string) => ({ kind, owner, count: 0, cards: [] });
const seat = (id: string, name: string, n: number): PlayerView =>
  ({
    id,
    name,
    seat: n,
    life: 40,
    library: zone("library", id),
    hand: zone("hand", id),
    graveyard: zone("graveyard", id),
    command: zone("command", id),
    commander_damage: {},
    life_history: [],
    mana_pool: [],
    hand_kept: true,
    undos_remaining: 1,
  }) as unknown as PlayerView;

function tableView(): GameView {
  return {
    id: GAME,
    state: "active",
    seats: [seat(ME, "Me", 0), seat("p-2", "Opp", 1)],
    battlefield: zone("battlefield"),
    stack: zone("stack"),
    exile: zone("exile"),
    stack_items: [],
    pending_triggers: [],
    pending_choices: [],
    turn: {
      seq: 3,
      number: 3,
      active_seat: 0,
      priority_holder: 0,
      phase: "precombat_main",
      step: "precombat_main",
    },
    mulligans_open: false,
  } as unknown as GameView;
}

let seq = 1;
function snapshot(sock: FakeSocket): void {
  sock.emit("open", {});
  sock.emit("message", {
    data: JSON.stringify({
      v: PROTOCOL_VERSION,
      kind: "snapshot",
      id: `frame-${seq}`,
      payload: { seq: seq++, game: tableView() },
    }),
  });
  flushSync();
}

describe("switching at the table", () => {
  it("switches from the ⋯ menu, and comes back to the seat after the 4001", async () => {
    session.set(person(playerMode));
    const { container } = render(Game as never, { gameID: GAME } as never);
    snapshot(FakeSocket.all.at(-1)!);
    await settle();

    await openMore(container);
    const sw = [...container.querySelectorAll<HTMLElement>('[role="menuitem"]')].find((el) =>
      el.textContent?.includes("Switch to admin mode"),
    )!;
    expect(sw).toBeDefined();
    click(sw);
    await settle();
    expect(puts).toEqual([{ on: true }]);

    // The server rebinds this socket: 4001.
    const before = FakeSocket.all.length;
    FakeSocket.all.at(-1)!.emit("close", { code: 4001, reason: "admin mode changed" });
    await settle();
    expect(FakeSocket.all.length).toBe(before + 1);
    expect(new URL(FakeSocket.all.at(-1)!.url).searchParams.get("player")).toBe(ME);
    expect(get(route).name).not.toBe("lobby");
  });

  it("goes to the Lobby, with a notice, when the connection needed admin mode", async () => {
    // An admin in admin mode watching someone else's table, seatless.
    window.location.hash = `#/games/other-table`;
    mode = "admin";
    session.set(person({ role: "identified", ...adminMode }));
    render(Game as never, { gameID: "other-table" } as never);
    snapshot(FakeSocket.all.at(-1)!);
    await settle();
    expect(FakeSocket.all).toHaveLength(1);

    // Switched off in another tab; the server closes this socket.
    mode = "player";
    FakeSocket.all.at(-1)!.emit("close", { code: 4001, reason: "admin mode changed" });
    await settle();
    expect(FakeSocket.all).toHaveLength(1);
    expect(window.location.hash).toBe("#/lobby");
    expect(get(adminNotice)).toBe(PLAYER_MODE_NOT_YOURS);

    // The Lobby shows it, once.
    const lobby = render(Lobby as never, {} as never);
    await settle();
    expect(lobby.container.textContent).toContain(PLAYER_MODE_NOT_YOURS);
    expect(get(adminNotice)).toBe("");
  });
});
