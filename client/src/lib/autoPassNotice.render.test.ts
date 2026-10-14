// @vitest-environment jsdom
//
// autoPassNotice.render.test.ts — #3343: the dock chip "Passed for you:
// <reason>" and its one-tap Undo, mounted beside a real ActionDock. The
// Undo goes through a real GameClient over a fake socket so the test sees
// the frame that leaves: it must be the ordinary `undo` action.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

import AutoPassNotice from "./components/board/AutoPassNotice.svelte";
import { _resetForTests as resetDock } from "./dock";
import { L } from "./labels";
import { PASS_NOTICE_MS, type PassNotice, type PassNoticeFacts } from "./passReason";
import type { PlayerView } from "./protocol";
import { GameClient } from "./ws";
import { render, click, cleanup, flushSync } from "./test/render.svelte";
import { nameOf } from "./test/dockView";
import DockHarness from "./test/DockHarness.svelte";

class FakeSocket {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSING = 2;
  static readonly CLOSED = 3;
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
  close(): void {
    this.readyState = FakeSocket.CLOSED;
  }
  emit(type: string, ev: unknown): void {
    for (const fn of this.listeners[type] ?? []) fn(ev);
  }
}

class FakeObserver {
  observe(): void {}
  unobserve(): void {}
  disconnect(): void {}
}

let realWebSocket: unknown;

beforeEach(() => {
  const g = globalThis as Record<string, unknown>;
  g.ResizeObserver ??= FakeObserver;
  g.IntersectionObserver ??= FakeObserver;
  realWebSocket = g.WebSocket;
  g.WebSocket = FakeSocket;
  FakeSocket.last = null;
  resetDock();
});
afterEach(() => {
  cleanup();
  vi.useRealTimers();
  (globalThis as Record<string, unknown>).WebSocket = realWebSocket;
});

const seat = (over: Partial<PlayerView> = {}): PlayerView =>
  ({ id: "me", seat: 0, undos_remaining: 3, ...over }) as unknown as PlayerView;

const notice = (over: Partial<PassNotice> = {}): PassNotice => ({
  frame: "pass-frame",
  reason: "nothing to respond with",
  actionsSent: 1,
  ...over,
});

const facts = (over: Partial<PassNoticeFacts> = {}): PassNoticeFacts => ({
  ack: { id: "pass-frame", seq: 7 },
  lastSeq: 7,
  actionsSent: 1,
  seat: seat(),
  undoBlocked: null,
  ...over,
});

function mount(over: Record<string, unknown> = {}) {
  const client = new GameClient("ws://test/ws");
  client.connect();
  const socket = FakeSocket.last!;
  socket.emit("open", {});
  const r = render(
    DockHarness as never,
    {
      component: AutoPassNotice,
      props: {
        notice: notice(),
        facts: facts(),
        refusedFrame: null,
        live: true,
        onUndo: () => client.sendAction("undo"),
        ...over,
      },
    } as never,
  );
  return { ...r, client, socket };
}

const chip = (): HTMLElement | null =>
  document.querySelector<HTMLElement>(`[role="dialog"][aria-label="${L.autoPassNotice}"]`);
const undoButton = (): HTMLButtonElement | null =>
  [...document.querySelectorAll<HTMLButtonElement>("button")].find(
    (b) => b.getAttribute("aria-label") === L.undoAutoPass,
  ) ?? null;
const text = (el: Element | null | undefined): string =>
  el?.textContent?.replace(/\s+/g, " ").trim() ?? "";

describe("the auto-pass chip (#3343)", () => {
  it("renders 'Passed for you: <reason>' in the dock", () => {
    mount();
    expect(chip()).not.toBeNull();
    expect(text(chip())).toContain("Passed for you: nothing to respond with");
  });

  it("offers Undo, which sends the ordinary undo action and closes the chip", () => {
    const { socket } = mount();
    const btn = undoButton()!;
    expect(btn).not.toBeNull();
    expect(text(btn)).toBe("Undo");
    // The dock's own Undo ("Undo (N left)") is the one /^Undo/ counts.
    expect(nameOf(btn)).not.toMatch(/^Undo/);

    click(btn);
    flushSync();

    // (Connecting also sends a visibility frame; only actions matter.)
    const frames = socket.sent.map((s) => JSON.parse(s)).filter((f) => f.kind === "action");
    expect(frames).toHaveLength(1);
    expect(frames[0].kind).toBe("action");
    expect(frames[0].payload.type).toBe("undo");
    expect(chip()).toBeNull();
  });

  it("shows the reason alone when the existing undo would not take the pass back", () => {
    // Someone else acted after the pass: the stack's top is not ours.
    mount({ facts: facts({ undoBlocked: "the last action was not yours" }) });
    expect(text(chip())).toContain("Passed for you: nothing to respond with");
    expect(undoButton()).toBeNull();
  });

  it("drops Undo the moment the viewer acts, and gains it when the ack lands", () => {
    const r = mount({ facts: facts({ ack: null }) });
    // Sent but not yet acknowledged: no Undo yet.
    expect(undoButton()).toBeNull();
    // DockHarness hands `props` to the notice whole, so replace it whole.
    const withFacts = (f: PassNoticeFacts) =>
      r.setProps({
        props: { notice: notice(), facts: f, refusedFrame: null, live: true, onUndo: () => {} },
      } as never);
    withFacts(facts());
    flushSync();
    expect(undoButton()).not.toBeNull();
    withFacts(facts({ actionsSent: 2 }));
    flushSync();
    expect(undoButton()).toBeNull();
    expect(text(chip())).toContain("Passed for you");
  });

  it("goes away when the server refused the pass", () => {
    mount({ refusedFrame: "pass-frame" });
    expect(chip()).toBeNull();
  });

  it("never shows over a replay frame", () => {
    mount({ live: false });
    expect(chip()).toBeNull();
  });

  it("clears itself after a few seconds", () => {
    vi.useFakeTimers();
    mount();
    expect(chip()).not.toBeNull();
    vi.advanceTimersByTime(PASS_NOTICE_MS + 50);
    flushSync();
    expect(chip()).toBeNull();
  });
});
