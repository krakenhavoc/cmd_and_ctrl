// @vitest-environment jsdom
//
// #1952: "Report a bug or idea" is a button on the game's main toolbar,
// next to settings, not an item in the three-dot menu. It shows only
// when the server says bug reports are enabled, and it opens the same
// BugReportModal the menu item used to.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";

import { PROTOCOL_VERSION, type GameView } from "./protocol";
import { session, type Session } from "./session";
import { resetBoardStub } from "./test/boardStub";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

vi.mock("./components/board/Board.svelte", async () => ({
  default: (await import("./test/BoardStub.svelte")).default,
}));

const Game = await import("../routes/Game.svelte").then((m) => m.default);

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;
  readyState = FakeSocket.OPEN;
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};
  constructor(readonly url: string) {
    FakeSocket.last = this;
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

const view = {
  id: "g",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { number: 4, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  mulligans_open: false,
} as unknown as GameView;

function playerSession(): Session {
  const expiresAt = new Date(Date.now() + 3_600_000).toISOString();
  return {
    token: "tok",
    expiresAt,
    principal: {
      role: "player",
      game_id: "game-1",
      player_id: "p-1",
      issued_at: new Date().toISOString(),
      expires_at: expiresAt,
    },
    playerID: "p-1",
    gameID: "game-1",
  } as Session;
}

let realWebSocket: unknown;
let realFetch: unknown;

function stubConfig(enabled: boolean): void {
  (globalThis as Record<string, unknown>).fetch = (url: string) =>
    Promise.resolve(
      url.includes("/bugreport/config")
        ? new Response(JSON.stringify({ enabled, attachments: false }), { status: 200 })
        : new Response("{}", { status: 404 }),
    );
}

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  realWebSocket = g.WebSocket;
  realFetch = g.fetch;
  g.WebSocket = FakeSocket;
  FakeSocket.last = null;
  resetBoardStub();
  session.set(playerSession());
});

afterEach(() => {
  cleanup();
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
  session.set(null);
  vi.restoreAllMocks();
});

async function mountGame(): Promise<HTMLElement> {
  const handle = render(Game as never, { gameID: "game-1" } as never);
  const socket = FakeSocket.last!;
  socket.emit("open", {});
  socket.emit("message", {
    data: JSON.stringify({
      v: PROTOCOL_VERSION,
      kind: "snapshot",
      id: "frame-1",
      payload: { seq: 1, game: view },
    }),
  });
  // Let fetchBugReportConfig resolve.
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
  return handle.container;
}

const bugButtons = (c: ParentNode): HTMLElement[] => [
  ...c.querySelectorAll<HTMLElement>('button[aria-label="Report a bug or idea"]'),
];

describe("Report a bug or idea on the game toolbar", () => {
  it("is a toolbar button next to settings when reports are enabled, and opens the modal", async () => {
    stubConfig(true);
    const c = await mountGame();

    const buttons = bugButtons(c);
    expect(buttons).toHaveLength(1);
    const btn = buttons[0]!;
    expect(btn.title).toBe("Report a bug or idea");
    expect(btn.closest(".bar-icons")).toBeTruthy();
    expect(btn.previousElementSibling?.getAttribute("aria-label")).toBe("open settings");

    expect(c.querySelector('[role="dialog"]')).toBeNull();
    click(btn);
    flushSync();
    expect(document.body.textContent ?? "").toMatch(/bug|idea/i);
    expect(
      document.querySelector('[role="dialog"], .modal, [aria-modal="true"]'),
      "the report modal should be open",
    ).toBeTruthy();
  });

  it("is absent when reports are disabled", async () => {
    stubConfig(false);
    const c = await mountGame();
    expect(bugButtons(c)).toHaveLength(0);
  });

  it("is no longer an item in the three-dot menu", async () => {
    stubConfig(true);
    const c = await mountGame();
    const more = c.querySelector<HTMLElement>('button[aria-label="more actions"]');
    expect(more, "the three-dot menu should exist for a seated player").toBeTruthy();
    click(more!);
    flushSync();
    const items = [...c.querySelectorAll<HTMLElement>('[role="menuitem"]')];
    expect(items.length).toBeGreaterThan(0);
    expect(items.some((i) => /bug or idea/i.test(i.textContent ?? ""))).toBe(false);
  });
});

// ADR 0111 PR 2: Pass turn left the command bar for the action dock,
// which Game mounts beside the board, in the play area.
describe("the command bar and the action dock", () => {
  it("has no Pass turn on the toolbar; the dock in the play area has it, disabled", async () => {
    stubConfig(false);
    const c = await mountGame();
    const bar = c.querySelector("header.bar")!;
    expect(
      [...bar.querySelectorAll("button")].some((b) => /pass turn/i.test(b.textContent ?? "")),
    ).toBe(false);
    // The icons keep their order: mute, log, settings, then the ⋯ menu.
    const icons = [...bar.querySelectorAll(".bar-icons > button, .bar-icons > .more > button")];
    expect(icons.map((b) => b.getAttribute("aria-label"))).toEqual([
      "mute sound effects",
      "open game log",
      "open settings",
      "more actions",
    ]);

    const dock = c.querySelector<HTMLElement>('.play-area > section[aria-label="actions"]');
    expect(dock, "the dock should be a child of the play area").toBeTruthy();
    const passTurn = [...dock!.querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === "Pass turn",
    );
    expect(passTurn?.disabled).toBe(true);

    // It publishes its size where the board, the zoom and the log drawer
    // can all read it.
    const root = c.querySelector<HTMLElement>("section.has-dock")!;
    expect(root.style.getPropertyValue("--dock-w")).toMatch(/px$/);
    expect(root.style.getPropertyValue("--dock-h")).toMatch(/px$/);
  });
});
