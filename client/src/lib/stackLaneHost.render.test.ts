// @vitest-environment jsdom
//
// stackLaneHost.render.test.ts — #3429, ADR 0176: the floating stack lane
// can be dragged by its header in every style, remembers its place per
// seat, resets, and stays on the board. jsdom has no layout, so the board
// is given a size by hand: 1000x600 at the page's corner, with a lane of
// 300x200.

import { describe, it, expect, afterEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import StackLaneHost from "./components/board/StackLaneHost.svelte";
import type { CardView, GameView, PlayerView, StackItemView, StackPositionView } from "./protocol";
import type { StackLaneStyle } from "./stackLane";
import { render, cleanup, flushSync } from "./test/render.svelte";

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
  const isLane = (el: HTMLElement) => el.classList.contains("stack-lane");
  Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
    configurable: true,
    get() {
      return isLane(this) ? 300 : 0;
    },
  });
  Object.defineProperty(HTMLElement.prototype, "offsetHeight", {
    configurable: true,
    get() {
      return isLane(this) ? 200 : 0;
    },
  });
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  const proto = HTMLElement.prototype as unknown as Record<string, unknown>;
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

const zone = (kind: string, owner: string | undefined, cards: CardView[] = []) => ({
  kind,
  owner,
  count: cards.length,
  cards,
});

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
  }) as unknown as PlayerView;

const trigger: StackItemView = {
  id: "trig-1",
  kind: "triggered",
  controller: "me",
  owner: "me",
  source_card_id: "mull",
  label: "Mulldrifter — draw two cards",
};

const view = {
  id: "g1",
  state: "active",
  seats: [seat("me", "Me", 0), seat("opp", "Bot 2", 1)],
  battlefield: zone("battlefield", undefined),
  stack: zone("stack", undefined),
  exile: zone("exile", undefined),
  stack_items: [trigger],
  pending_triggers: [],
  turn: { number: 3, active_seat: 0, priority_holder: 0, phase: "main1", step: "main" },
} as unknown as GameView;

/** A seat slot under the board, as Board.svelte draws one. */
function slot(pos: string, r: DOMRect): HTMLElement {
  const el = document.createElement("div");
  el.className = "slot";
  el.dataset.pos = pos;
  el.getBoundingClientRect = () => r;
  return el;
}

function mountLane(
  style: StackLaneStyle,
  extra: Record<string, unknown> = {},
  slots: HTMLElement[] = [],
) {
  install();
  const board = document.createElement("div");
  let boardRect = rect(0, 0, 1000, 600);
  board.getBoundingClientRect = () => boardRect;
  for (const s of slots) board.appendChild(s);
  document.body.appendChild(board);
  const r = render(StackLaneHost, {
    view,
    viewerID: "me",
    boardEl: board,
    style,
    onCounter: () => {},
    position: null,
    ...extra,
  } as never);
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

const host = (c: HTMLElement) => c.querySelector<HTMLElement>(".lane-host")!;
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

const dblclick = (c: HTMLElement) => {
  handle(c).dispatchEvent(new MouseEvent("dblclick", { bubbles: true }));
  flushSync();
};

// The default pile and one of the three centred styles share one saved
// spot, so every behaviour is pinned on both.
for (const style of ["pile", "fan"] as const) {
  describe(`movable ${style} lane`, () => {
    it("keeps its own placement, with no reset button, when nothing is saved", () => {
      const { r } = mountLane(style);
      expect(host(r.container).classList.contains("placed")).toBe(false);
      expect(host(r.container).style.left).toBe("");
      expect(r.container.querySelector(".reset-pos")).toBeNull();
      // Items stay measurable for StackLinger and CastFlight.
      expect(r.container.querySelector("[data-stack-item-id]")).not.toBeNull();
    });

    it("applies the saved position on load, relative to the board", () => {
      const { r } = mountLane(style, { position: { x: 0.5, y: 0.25 } as StackPositionView });
      const el = host(r.container);
      expect(el.classList.contains("placed")).toBe(true);
      expect(el.style.left).toBe("500px");
      expect(el.style.top).toBe("150px");
      expect(r.container.querySelector(".reset-pos")).not.toBeNull();
    });

    it("a drag moves the lane and saves the fractions on release", () => {
      const onSave = vi.fn();
      const { r } = mountLane(style, { onSave });
      // A zero-sized rect in jsdom puts the lane's corner at 0,0, so the
      // grab is 50px right and 20px below it.
      drag(r.container, [50, 20], [350, 220]);
      expect(onSave).toHaveBeenCalledTimes(1);
      const saved = onSave.mock.calls[0][0] as StackPositionView;
      expect(saved.x).toBeCloseTo(0.3);
      expect(saved.y).toBeCloseTo(200 / 600);
      const el = host(r.container);
      expect(el.classList.contains("placed")).toBe(true);
      expect(el.style.left).toBe("300px");
      expect(el.style.top).toBe("200px");
    });

    it("a click on the header with no real movement saves nothing", () => {
      const onSave = vi.fn();
      const { r } = mountLane(style, { onSave });
      drag(r.container, [50, 20], [51, 21]);
      expect(onSave).not.toHaveBeenCalled();
      expect(host(r.container).classList.contains("placed")).toBe(false);
    });

    it("keeps the lane on the board when dropped past an edge", () => {
      const onSave = vi.fn();
      const { r } = mountLane(style, { onSave });
      drag(r.container, [0, 0], [5000, 5000]);
      const saved = onSave.mock.calls[0][0] as StackPositionView;
      expect(saved.x).toBeCloseTo(0.7);
      expect(saved.y).toBeCloseTo(400 / 600);
    });

    it("clamps to the board when the window shrinks, and recovers when it grows", () => {
      const m = mountLane(style, { position: { x: 0.7, y: 0.5 } });
      expect(host(m.r.container).style.left).toBe("700px");
      m.resizeBoard(800, 400);
      // 0.7 * 800 = 560 would overhang (800 - 300 = 500).
      expect(host(m.r.container).style.left).toBe("500px");
      expect(host(m.r.container).style.top).toBe("200px");
      m.resizeBoard(1000, 600);
      expect(host(m.r.container).style.left).toBe("700px");
    });

    it("double-clicking the header resets and clears the saved spot", () => {
      const onReset = vi.fn();
      const { r } = mountLane(style, { position: { x: 0.5, y: 0.25 }, onReset });
      dblclick(r.container);
      expect(onReset).toHaveBeenCalledTimes(1);
      expect(host(r.container).classList.contains("placed")).toBe(false);
    });

    it("the reset button is the keyboard route", () => {
      const onReset = vi.fn();
      const { r } = mountLane(style, { position: { x: 0.5, y: 0.25 }, onReset });
      const btn = r.container.querySelector<HTMLButtonElement>(".reset-pos")!;
      expect(btn.getAttribute("aria-label")).toBe("Reset stack position");
      btn.click();
      flushSync();
      expect(onReset).toHaveBeenCalledTimes(1);
      expect(r.container.querySelector(".reset-pos")).toBeNull();
    });

    it("a spectator can drag it, and nothing is saved", () => {
      const { r } = mountLane(style, { onSave: undefined, onReset: undefined });
      drag(r.container, [0, 0], [200, 100]);
      expect(host(r.container).classList.contains("placed")).toBe(true);
      dblclick(r.container);
      expect(host(r.container).classList.contains("placed")).toBe(false);
    });

    it("takes the server's spot over its own once it arrives", () => {
      const { r } = mountLane(style, { onSave: () => {} });
      drag(r.container, [0, 0], [200, 100]);
      r.setProps({ position: { x: 0.1, y: 0.1 } } as never);
      expect(host(r.container).style.left).toBe("100px");
      expect(host(r.container).style.top).toBe("60px");
    });
  });
}

// #3380, the ADR 0176 amendment: with nothing saved the pile docks on the
// seam between the top row's seats. The board is 1000x600 with an 8px gap.
describe("the pile's default spot", () => {
  const fourSeats = () => [
    slot("across", rect(0, 0, 496, 200)),
    slot("across_next", rect(504, 0, 496, 200)),
    slot("next", rect(0, 208, 496, 392)),
    slot("self", rect(504, 208, 496, 392)),
  ];

  it("is centred on the seam between the top seats of a four-seat table", () => {
    const { r } = mountLane("pile", {}, fourSeats());
    const el = host(r.container);
    // The seam is at 500; the 300px pile is centred on it.
    expect(el.style.left).toBe("350px");
    // Centred on the row seam (204) and kept inside the board.
    expect(el.style.top).toBe("104px");
    expect(el.classList.contains("placed")).toBe(false);
  });

  it("keeps the left edge on a two-seat table", () => {
    const { r } = mountLane("pile", {}, [
      slot("across", rect(0, 0, 1000, 296)),
      slot("self", rect(0, 304, 1000, 296)),
    ]);
    expect(host(r.container).style.left).toBe("");
  });

  it("gives way to a saved spot", () => {
    const { r } = mountLane("pile", { position: { x: 0.1, y: 0.1 } }, fourSeats());
    expect(host(r.container).style.left).toBe("100px");
    expect(host(r.container).style.top).toBe("60px");
    expect(host(r.container).classList.contains("placed")).toBe(true);
  });

  it("leaves the centred styles where they were", () => {
    const { r } = mountLane("fan", {}, fourSeats());
    expect(host(r.container).style.left).toBe("");
  });
});

describe("the lane header handle", () => {
  it("without a board it shows no drag affordance", () => {
    const { r } = mountLane("pile", { boardEl: null });
    expect(handle(r.container).classList.contains("movable")).toBe(false);
  });

  it("a press on the collapse button does not start a drag", () => {
    const onSave = vi.fn();
    const { r } = mountLane("pile", { onSave });
    const btn = r.container.querySelector(".collapse")!;
    pointer(btn, "pointerdown", 10, 10);
    pointer(handle(r.container), "pointermove", 400, 300);
    pointer(handle(r.container), "pointerup", 400, 300);
    expect(onSave).not.toHaveBeenCalled();
  });
});

describe("the folded pile tab", () => {
  const fold = (c: HTMLElement) => {
    c.querySelector<HTMLButtonElement>(".collapse")!.click();
    flushSync();
  };
  const tab = (c: HTMLElement) => c.querySelector<HTMLElement>(".stack-tab")!;

  it("drags and saves, and stays folded", () => {
    const onSave = vi.fn();
    const { r } = mountLane("pile", { onSave });
    fold(r.container);
    pointer(tab(r.container), "pointerdown", 50, 20);
    pointer(tab(r.container), "pointermove", 350, 220);
    pointer(tab(r.container), "pointerup", 350, 220);
    // The click a browser sends after the release must not unfold it.
    tab(r.container).click();
    flushSync();
    expect(onSave).toHaveBeenCalledTimes(1);
    const saved = onSave.mock.calls[0][0] as StackPositionView;
    expect(saved.x).toBeCloseTo(0.3);
    expect(saved.y).toBeCloseTo(200 / 600);
    expect(tab(r.container)).not.toBeNull();
    expect(host(r.container).style.left).toBe("300px");
    expect(host(r.container).style.top).toBe("200px");
  });

  it("a plain click unfolds it and saves nothing", () => {
    const onSave = vi.fn();
    const { r } = mountLane("pile", { onSave });
    fold(r.container);
    pointer(tab(r.container), "pointerdown", 50, 20);
    pointer(tab(r.container), "pointerup", 51, 21);
    tab(r.container).click();
    flushSync();
    expect(onSave).not.toHaveBeenCalled();
    expect(r.container.querySelector(".stack-tab")).toBeNull();
    expect(handle(r.container)).not.toBeNull();
  });

  it("sits at the saved spot, and reset is on the unfolded header", () => {
    const { r } = mountLane("pile", { position: { x: 0.5, y: 0.25 } });
    fold(r.container);
    expect(host(r.container).style.left).toBe("500px");
    tab(r.container).click();
    flushSync();
    expect(r.container.querySelector(".reset-pos")).not.toBeNull();
  });
});
