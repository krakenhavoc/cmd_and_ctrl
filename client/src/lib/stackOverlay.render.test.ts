// @vitest-environment jsdom
//
// stackOverlay.render.test.ts — #3429, ADR 0176: the docked stack card can
// be dragged by its header, remembers its place per seat, resets, and stays
// on the board. jsdom has no layout, so the board, the strip and the card
// are given sizes by hand: the board is 1000x600 at the page's corner, the
// strip (the card's offset parent) starts 100px in and 10px down, and the
// card is 300x200.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import StackOverlay from "./components/board/StackOverlay.svelte";
import type { StackPositionView, ZoneView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const stack: ZoneView = { kind: "stack", count: 1, cards: [] };

let resizeCallbacks: Array<() => void> = [];

function install() {
  resizeCallbacks = [];
  vi.stubGlobal(
    "ResizeObserver",
    class {
      constructor(cb: () => void) {
        resizeCallbacks.push(cb);
      }
      observe() {}
      disconnect() {}
    },
  );
  Object.defineProperty(HTMLElement.prototype, "offsetParent", {
    configurable: true,
    get() {
      return this.parentElement;
    },
  });
  Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
    configurable: true,
    get() {
      return this.hasAttribute("data-stack-overlay") ? 300 : 0;
    },
  });
  Object.defineProperty(HTMLElement.prototype, "offsetHeight", {
    configurable: true,
    get() {
      return this.hasAttribute("data-stack-overlay") ? 200 : 0;
    },
  });
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  const proto = HTMLElement.prototype as unknown as Record<string, unknown>;
  delete proto.offsetParent;
  delete proto.offsetWidth;
  delete proto.offsetHeight;
});

function rect(left: number, top: number, width: number, height: number): DOMRect {
  return {
    left,
    top,
    width,
    height,
    right: left + width,
    bottom: top + height,
    x: left,
    y: top,
  } as DOMRect;
}

function mountCard(extra: Record<string, unknown> = {}) {
  install();
  const board = document.createElement("div");
  let boardRect = rect(0, 0, 1000, 600);
  board.getBoundingClientRect = () => boardRect;
  document.body.appendChild(board);
  const r = render(StackOverlay, {
    stack,
    stackItems: [],
    pendingTriggers: [],
    seats: [],
    viewerHasPriority: false,
    splitSecondActive: false,
    onCounter: () => {},
    boardEl: board,
    position: null,
    ...extra,
  } as never);
  // The strip is the card's parent: 100px in, 10px down from the board.
  r.container.getBoundingClientRect = () => rect(100, 10, 512, 0);
  // Measure again now the rects exist.
  for (const cb of resizeCallbacks) cb();
  flushSync();
  return {
    r,
    resizeBoard(w: number, h: number) {
      boardRect = rect(0, 0, w, h);
      for (const cb of resizeCallbacks) cb();
      flushSync();
    },
  };
}

const card = (c: HTMLElement) => c.querySelector<HTMLElement>("[data-stack-overlay]")!;
const handle = (c: HTMLElement) => c.querySelector<HTMLElement>("[data-stack-handle]")!;

function pointer(el: Element, type: string, x: number, y: number) {
  const e = new MouseEvent(type, {
    clientX: x,
    clientY: y,
    button: 0,
    bubbles: true,
    cancelable: true,
  });
  Object.defineProperty(e, "pointerId", { value: 1 });
  el.dispatchEvent(e);
  flushSync();
}

function drag(c: HTMLElement, from: [number, number], to: [number, number]) {
  pointer(handle(c), "pointerdown", ...from);
  pointer(handle(c), "pointermove", ...to);
  pointer(handle(c), "pointerup", ...to);
}

describe("movable stack card", () => {
  it("sits docked, with no reset button, when nothing is saved", () => {
    const { r } = mountCard();
    expect(card(r.container).classList.contains("placed")).toBe(false);
    expect(r.container.querySelector(".reset-pos")).toBeNull();
  });

  it("applies the saved position on load, relative to the board", () => {
    const { r } = mountCard({ position: { x: 0.5, y: 0.25 } as StackPositionView });
    const el = card(r.container);
    expect(el.classList.contains("placed")).toBe(true);
    // 0.5 * 1000 and 0.25 * 600, less the strip's offset from the board.
    expect(el.style.left).toBe("400px");
    expect(el.style.top).toBe("140px");
    expect(r.container.querySelector(".reset-pos")).not.toBeNull();
  });

  it("a drag updates the position and saves the fractions on release", () => {
    const onSave = vi.fn();
    const { r } = mountCard({ onSave });
    // Grab 50px right and 20px below the card's corner (a zero-sized rect
    // in jsdom puts the corner at 0,0), and drop at 350, 220.
    drag(r.container, [50, 20], [350, 220]);
    expect(onSave).toHaveBeenCalledTimes(1);
    const saved = onSave.mock.calls[0][0] as StackPositionView;
    expect(saved.x).toBeCloseTo(0.3);
    expect(saved.y).toBeCloseTo(200 / 600);
    const el = card(r.container);
    expect(el.classList.contains("placed")).toBe(true);
    expect(el.style.left).toBe("200px");
    expect(el.style.top).toBe("190px");
  });

  it("a click on the header with no real movement saves nothing", () => {
    const onSave = vi.fn();
    const { r } = mountCard({ onSave });
    drag(r.container, [50, 20], [51, 21]);
    expect(onSave).not.toHaveBeenCalled();
    expect(card(r.container).classList.contains("placed")).toBe(false);
  });

  it("keeps the card on the board when dropped past an edge", () => {
    const onSave = vi.fn();
    const { r } = mountCard({ onSave });
    drag(r.container, [0, 0], [5000, 5000]);
    const saved = onSave.mock.calls[0][0] as StackPositionView;
    expect(saved.x).toBeCloseTo(0.7);
    expect(saved.y).toBeCloseTo(400 / 600);
  });

  it("double-clicking the header resets and clears the saved spot", () => {
    const onReset = vi.fn();
    const { r } = mountCard({ position: { x: 0.5, y: 0.25 }, onReset });
    handle(r.container).dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));
    flushSync();
    expect(onReset).toHaveBeenCalledTimes(1);
    expect(card(r.container).classList.contains("placed")).toBe(false);
  });

  it("the reset button is the keyboard route", () => {
    const onReset = vi.fn();
    const { r } = mountCard({ position: { x: 0.5, y: 0.25 }, onReset });
    const btn = r.container.querySelector<HTMLButtonElement>(".reset-pos")!;
    expect(btn.getAttribute("aria-label")).toBe("Reset stack position");
    btn.click();
    flushSync();
    expect(onReset).toHaveBeenCalledTimes(1);
    expect(r.container.querySelector(".reset-pos")).toBeNull();
  });

  it("a spectator can drag it, and nothing is saved", () => {
    const { r } = mountCard({ onSave: undefined, onReset: undefined });
    drag(r.container, [0, 0], [200, 100]);
    expect(card(r.container).classList.contains("placed")).toBe(true);
    handle(r.container).dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));
    flushSync();
    expect(card(r.container).classList.contains("placed")).toBe(false);
  });

  it("takes the server's spot over its own once it arrives", () => {
    const { r } = mountCard({ onSave: () => {} });
    drag(r.container, [0, 0], [200, 100]);
    r.setProps({ position: { x: 0.1, y: 0.1 } } as never);
    expect(card(r.container).style.left).toBe("0px");
    expect(card(r.container).style.top).toBe("50px");
  });

  it("clamps to the board when the window shrinks, and recovers when it grows", () => {
    const m = mountCard({ position: { x: 0.7, y: 0.5 } });
    expect(card(m.r.container).style.left).toBe("600px");
    m.resizeBoard(800, 400);
    // 0.7 * 800 = 560 would overhang (800 - 300 = 500), so it stops at 500.
    expect(card(m.r.container).style.left).toBe("400px");
    expect(card(m.r.container).style.top).toBe("190px");
    m.resizeBoard(1000, 600);
    expect(card(m.r.container).style.left).toBe("600px");
  });

  it("without a board it stays docked and shows no drag affordance", () => {
    const { r } = mountCard({ boardEl: null });
    expect(handle(r.container).classList.contains("movable")).toBe(false);
  });
});
