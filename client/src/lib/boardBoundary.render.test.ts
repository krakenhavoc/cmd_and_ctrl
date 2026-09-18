// @vitest-environment jsdom
//
// #720: the `<svelte:boundary>` around `<Board>` in Game.svelte.
//
// This is the render half of #266's symptom. The store guards stop a
// throwing SUBSCRIBER from stalling every store in the app; they do
// nothing about a throw while the board is being drawn, or inside one
// of its `$effect`s. Before the boundary that surfaced as an uncaught
// error and a table frozen on its last good frame: no message, nothing
// to press, and a bug report that said "it froze".
//
// The test mounts the real Game.svelte with Board.svelte replaced by
// BoardStub.svelte, which throws when asked. What is under test is
// Game.svelte's boundary, its `onerror`, and the fallback's wording and
// controls — the fallback has to be something a player can act on, not
// a blank panel.

import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import { get } from "svelte/store";

import { PROTOCOL_VERSION, type GameView } from "./protocol";
import { recentClientErrors, resetClientErrors } from "./clientErrors";
import { session } from "./session";
import { BOARD_STUB_ERROR, boardStub, resetBoardStub } from "./test/boardStub";
import { render, click, cleanup, flushSync } from "./test/render.svelte";

// Hoisted above the imports by vitest, hence the dynamic import of the
// stub. The path is Board's own module id, so Game.svelte's
// `../lib/components/board/Board.svelte` resolves to the stub too.
vi.mock("./components/board/Board.svelte", async () => ({
  default: (await import("./test/BoardStub.svelte")).default,
}));

const Game = await import("../routes/Game.svelte").then((m) => m.default);

class FakeSocket {
  static readonly OPEN = 1;
  static last: FakeSocket | null = null;

  readyState = FakeSocket.OPEN;
  sent: string[] = [];
  private listeners: Record<string, ((ev: unknown) => void)[]> = {};

  constructor(readonly url: string) {
    FakeSocket.last = this;
  }
  addEventListener(type: string, fn: (ev: unknown) => void): void {
    (this.listeners[type] ??= []).push(fn);
  }
  removeEventListener(): void {}
  send(data: string): void {
    this.sent.push(data);
  }
  close(): void {}
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

const view: GameView = {
  id: "g",
  state: "active",
  seats: [],
  battlefield: { kind: "battlefield", count: 0, cards: [] },
  stack: { kind: "stack", count: 0, cards: [] },
  exile: { kind: "exile", count: 0, cards: [] },
  turn: { number: 4, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
  mulligans_open: false,
} as unknown as GameView;

const snapshotFrame = (seq: number): string =>
  JSON.stringify({
    v: PROTOCOL_VERSION,
    kind: "snapshot",
    id: `frame-${seq}`,
    payload: { seq, game: view },
  });

let realWebSocket: unknown;
let realFetch: unknown;

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  realWebSocket = g.WebSocket;
  realFetch = g.fetch;
  g.WebSocket = FakeSocket;
  // Game.svelte probes /bugreport/config on mount. fetchBugReportConfig
  // swallows every failure into "disabled", which is the state this
  // test wants — but an unstubbed fetch would leave a real request
  // dangling past the end of the test.
  g.fetch = () => Promise.reject(new Error("no network in this test"));
  FakeSocket.last = null;
  resetClientErrors();
  resetBoardStub();
  session.set(null);
});

afterEach(() => {
  cleanup();
  const g = globalThis as Record<string, unknown>;
  g.WebSocket = realWebSocket;
  g.fetch = realFetch;
  vi.restoreAllMocks();
});

// mountGame mounts the route and delivers one snapshot, which is what
// makes `{#if view}` true and the board render at all.
function mountGame(): { container: HTMLElement } {
  const handle = render(Game as never, { gameID: "game-1" } as never);
  const socket = FakeSocket.last;
  expect(socket, "Game.svelte should have opened a socket on mount").toBeTruthy();
  socket!.emit("open", {});
  socket!.emit("message", { data: snapshotFrame(7) });
  flushSync();
  return { container: handle.container };
}

const fallback = (c: HTMLElement): HTMLElement | null => c.querySelector(".board-failed");
const boardEl = (c: HTMLElement): HTMLElement | null =>
  c.querySelector('[data-testid="board-stub"]');
const buttonSaying = (c: HTMLElement, re: RegExp): HTMLElement | undefined =>
  [...c.querySelectorAll<HTMLElement>("button")].find((b) => re.test(b.textContent ?? ""));

describe("the <svelte:boundary> around Board", () => {
  it("renders the board normally when nothing throws", () => {
    const { container } = mountGame();

    expect(boardEl(container), "the board should be rendered").toBeTruthy();
    expect(fallback(container)).toBeNull();
    expect(boardStub.renders).toBeGreaterThan(0);
  });

  it("survives a throw from inside the board", () => {
    boardStub.throwOnNextRender = true;
    const { container } = mountGame();

    // The throw happened, and did not take the test (or the route) with
    // it.
    expect(boardStub.throws).toBe(1);
    expect(boardEl(container)).toBeNull();
    expect(fallback(container), "the fallback panel should be showing").toBeTruthy();
  });

  it("announces the failure and says the game is not lost", () => {
    boardStub.throwOnNextRender = true;
    const { container } = mountGame();

    const panel = fallback(container)!;
    // Announced, not just drawn: the board it replaced is gone, so
    // nothing else would tell a screen reader anything happened.
    expect(panel.getAttribute("role")).toBe("alert");
    const text = panel.textContent!.replace(/\s+/g, " ");
    expect(text).toContain("The board stopped drawing");
    // Deliberately about drawing rather than the connection: #519's
    // stale-connection banner is the other half of #266's symptom, and
    // sending a player to check their network when the socket is fine
    // wastes their time.
    expect(text).toMatch(/game itself/i);
    expect(text).not.toMatch(/connection|offline|network/i);
  });

  it("shows the thrown error rather than making the player find it", () => {
    boardStub.throwOnNextRender = true;
    const { container } = mountGame();

    expect(fallback(container)!.textContent).toContain(BOARD_STUB_ERROR);
  });

  it("records the failure so a bug report carries it", () => {
    boardStub.throwOnNextRender = true;
    mountGame();

    const texts = recentClientErrors().map((e) => e.text);
    expect(
      texts.some((t) => t.includes("board render failed") && t.includes(BOARD_STUB_ERROR)),
      `errors were: ${texts.join(" | ")}`,
    ).toBe(true);
  });

  it("offers a redraw, a reload and a way back to the lobby", () => {
    boardStub.throwOnNextRender = true;
    const { container } = mountGame();

    // A fallback with nothing to press is a blank panel with extra
    // words. All three of these have to be there.
    expect(buttonSaying(container, /Redraw the board/)).toBeTruthy();
    expect(buttonSaying(container, /Reload the page/)).toBeTruthy();
    expect(buttonSaying(container, /Back to lobby/)).toBeTruthy();
  });

  it("points at the bug report without asking the player to copy the error out", () => {
    boardStub.throwOnNextRender = true;
    const { container } = mountGame();

    expect(fallback(container)!.textContent).toMatch(/log a bug report attaches/i);
  });

  it("brings the board back when the player redraws", () => {
    boardStub.throwOnNextRender = true;
    const { container } = mountGame();
    expect(fallback(container)).toBeTruthy();

    // Not a reload: `reset()` re-renders the subtree in place, so a
    // transient failure costs the player a click and nothing else.
    click(buttonSaying(container, /Redraw the board/)!);

    expect(fallback(container), "the fallback should be gone").toBeNull();
    expect(boardEl(container), "the board should be back").toBeTruthy();
    expect(boardStub.throws).toBe(1);
  });

  it("keeps the socket and the rest of the route alive behind the fallback", () => {
    boardStub.throwOnNextRender = true;
    mountGame();

    // The boundary is around the table only. Frames keep arriving and
    // keep being applied, which is what makes the redraw above worth
    // offering — the state it redraws from is current.
    const socket = FakeSocket.last!;
    socket.emit("message", { data: snapshotFrame(8) });
    flushSync();
    expect(socket.sent.length).toBeGreaterThanOrEqual(0);
    expect(get(session)).toBeNull();
  });
});
