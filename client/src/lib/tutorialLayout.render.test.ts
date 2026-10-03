// @vitest-environment jsdom
//
// tutorialLayout.render.test.ts — the coach card's place in the Game route
// (ADR 0076 §2.3 as amended 2026-10-02, #1079). The card is mounted on this
// tab's practice table only; there it publishes --coach-w / --coach-h and
// the board keeps its cell clear. Everywhere else — every real game, and
// the practice table once the coach is hidden — the layout is exactly what
// it is without a tutorial: no class, no variables, no spacer.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { writable } from "svelte/store";

import { PROTOCOL_VERSION, type GameView } from "./protocol";
import { session, type Session } from "./session";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

const practice = vi.hoisted(() => ({ store: null as unknown }));

vi.mock("./components/board/Board.svelte", async () => ({
  default: (await import("./test/CoachedBoardStub.svelte")).default,
}));
vi.mock("./practiceTable", async (orig) => {
  const actual = (await orig()) as Record<string, unknown>;
  const store = writable<{ gameID: string } | null>(null);
  practice.store = store;
  return { ...actual, practiceTable: store };
});

const Game = await import("../routes/Game.svelte").then((m) => m.default);
const practiceStore = () =>
  practice.store as ReturnType<typeof writable<{ gameID: string } | null>>;

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
  id: "game-1",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { number: 1, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
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
const sizeProps = ["offsetWidth", "offsetHeight"] as const;
const realSize: Partial<Record<(typeof sizeProps)[number], PropertyDescriptor | undefined>> = {};

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  realWebSocket = g.WebSocket;
  realFetch = g.fetch;
  g.WebSocket = FakeSocket;
  g.fetch = () => Promise.resolve(new Response("{}", { status: 404 }));
  FakeSocket.last = null;
  session.set(playerSession());
  practiceStore().set(null);
  // jsdom lays nothing out: give the coach card the size it has on a
  // 1280-wide screen, and everything else none.
  for (const p of sizeProps) {
    realSize[p] = Object.getOwnPropertyDescriptor(HTMLElement.prototype, p);
    Object.defineProperty(HTMLElement.prototype, p, {
      configurable: true,
      get(this: HTMLElement) {
        if (!this.classList.contains("coach-slot")) return 0;
        return p === "offsetWidth" ? 307 : 193;
      },
    });
  }
});

afterEach(() => {
  cleanup();
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
  session.set(null);
  for (const p of sizeProps) {
    const d = realSize[p];
    if (d) Object.defineProperty(HTMLElement.prototype, p, d);
  }
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
  await new Promise((r) => setTimeout(r, 0));
  flushSync();
  return handle.container;
}

const routeRoot = (c: HTMLElement) => c.querySelector<HTMLElement>("section")!;
const coach = (c: HTMLElement) => c.querySelector('[aria-label="tutorial coach"]');
const board = (c: HTMLElement) => c.querySelector('[data-testid="board-stub"]')!;

describe("the coach card in the Game route", () => {
  it("is absent from a real game, which keeps its layout exactly", async () => {
    const c = await mountGame();
    expect(c.querySelector('[aria-label="actions"]')).not.toBeNull();
    expect(coach(c)).toBeNull();
    const root = routeRoot(c);
    expect(root.classList.contains("has-coach")).toBe(false);
    expect(root.style.getPropertyValue("--coach-w")).toBe("");
    expect(root.style.getPropertyValue("--coach-h")).toBe("");
    expect(board(c).getAttribute("data-coached")).toBe("no");
  });

  it("is absent from another tab's practice table", async () => {
    practiceStore().set({ gameID: "some-other-game" });
    const c = await mountGame();
    expect(coach(c)).toBeNull();
    expect(routeRoot(c).classList.contains("has-coach")).toBe(false);
  });

  it("shows on this tab's practice table and has the board keep its cell clear", async () => {
    practiceStore().set({ gameID: "game-1" });
    const c = await mountGame();
    expect(coach(c)).not.toBeNull();
    const root = routeRoot(c);
    expect(root.classList.contains("has-coach")).toBe(true);
    expect(root.style.getPropertyValue("--coach-w")).toBe("307px");
    expect(root.style.getPropertyValue("--coach-h")).toBe("193px");
    expect(board(c).getAttribute("data-coached")).toBe("yes");
  });

  it("gives the space back when the player skips the tutorial", async () => {
    practiceStore().set({ gameID: "game-1" });
    const c = await mountGame();
    const skip = [...c.querySelectorAll("button")].find(
      (b) => b.textContent?.trim() === "Skip tutorial",
    )!;
    click(skip);
    flushSync();
    expect(coach(c)).toBeNull();
    const root = routeRoot(c);
    expect(root.classList.contains("has-coach")).toBe(false);
    expect(root.style.getPropertyValue("--coach-w")).toBe("");
    expect(board(c).getAttribute("data-coached")).toBe("no");
  });
});
