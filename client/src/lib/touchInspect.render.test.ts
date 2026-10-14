// @vitest-environment jsdom
//
// #3338: on a tablet a player could not read a card without acting on
// it. The zoom panel is fed by hover, a finger has none, and the tap
// runs the card's click (cast, tap, select). A long-press now pins the
// zoom and eats its own click, a quick tap still acts, and the Z key
// pins it for a keyboard. touchInspect.ts owns the pin; Card.svelte
// owns the gestures, which only a DOM can check.

import { describe, it, expect, afterEach, vi } from "vitest";
import { get } from "svelte/store";

import Card from "./components/board/Card.svelte";
import { hoveredCard } from "./cardTypes";
import { LONG_PRESS_MS, isZoomPinned, unpinZoom } from "./touchInspect";
import type { CardView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

afterEach(() => {
  unpinZoom();
  hoveredCard.set(null);
  cleanup();
  vi.useRealTimers();
});

const bolt = (): CardView =>
  ({
    instance_id: "c1",
    name: "Lightning Bolt",
    owner: "me",
    controller: "me",
    scryfall_id: "aaaa-bbbb",
    type_line: "Instant",
  }) as unknown as CardView;

// jsdom has no PointerEvent: a MouseEvent carrying the pointer fields
// the card reads stands in for one.
function pointer(
  el: Element,
  type: string,
  pointerType: "touch" | "mouse",
  at: { x: number; y: number } = { x: 10, y: 10 },
): void {
  const ev = new MouseEvent(type, { bubbles: true, clientX: at.x, clientY: at.y });
  Object.defineProperty(ev, "pointerType", { value: pointerType });
  Object.defineProperty(ev, "pointerId", { value: 1 });
  el.dispatchEvent(ev);
  flushSync();
}

function mount(): { root: HTMLElement; clicks: CardView[] } {
  // Only setTimeout is faked; Svelte flushes on a microtask.
  vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
  const clicks: CardView[] = [];
  const view = render(
    Card as never,
    { card: bolt(), onClick: (c: CardView) => clicks.push(c) } as never,
  );
  return { root: view.container.querySelector<HTMLElement>(".card")!, clicks };
}

/** A finger down, held for ms, then lifted: the events a tablet sends. */
function press(root: HTMLElement, ms: number, drift = 0): void {
  pointer(root, "pointerenter", "touch");
  pointer(root, "pointerdown", "touch");
  if (drift) pointer(root, "pointermove", "touch", { x: 10 + drift, y: 10 });
  vi.advanceTimersByTime(ms);
  flushSync();
  pointer(root, "pointerup", "touch");
  pointer(root, "pointerleave", "touch");
  root.click();
  flushSync();
}

describe("reading a card on a tablet (#3338)", () => {
  it("a long-press pins the zoom and does not play the card", () => {
    const { root, clicks } = mount();
    press(root, LONG_PRESS_MS + 50);
    expect(get(hoveredCard)?.instance_id).toBe("c1");
    expect(isZoomPinned("c1")).toBe(true);
    expect(clicks).toHaveLength(0);
  });

  it("a long-press held well past the threshold still does not play the card", () => {
    const { root, clicks } = mount();
    press(root, 3000);
    expect(clicks).toHaveLength(0);
    expect(isZoomPinned("c1")).toBe(true);
  });

  it("a quick tap still plays the card, and shows no zoom", () => {
    const { root, clicks } = mount();
    press(root, 100);
    expect(clicks).toHaveLength(1);
    expect(get(hoveredCard)).toBeNull();
  });

  it("a finger that drifts is a scroll: no zoom", () => {
    const { root } = mount();
    press(root, LONG_PRESS_MS + 50, 40);
    expect(isZoomPinned()).toBe(false);
  });

  it("the next tap anywhere closes the zoom without acting", () => {
    const { root, clicks } = mount();
    press(root, LONG_PRESS_MS + 50);
    vi.advanceTimersByTime(1); // the pin arms its dismiss listeners
    // The dismissing tap lands on the card itself.
    press(root, 80);
    expect(isZoomPinned()).toBe(false);
    expect(get(hoveredCard)).toBeNull();
    expect(clicks).toHaveLength(0);
  });

  it("a mouse keeps its hover zoom and its click", () => {
    const { root, clicks } = mount();
    pointer(root, "pointerenter", "mouse");
    vi.advanceTimersByTime(1000);
    flushSync();
    expect(get(hoveredCard)?.instance_id).toBe("c1");
    pointer(root, "pointerdown", "mouse");
    vi.advanceTimersByTime(1000);
    root.click();
    flushSync();
    expect(clicks).toHaveLength(1);
    expect(isZoomPinned()).toBe(false);
  });

  it("Z on the focused card pins the zoom, and Escape closes it", () => {
    const { root, clicks } = mount();
    root.dispatchEvent(new KeyboardEvent("keydown", { key: "z", bubbles: true }));
    flushSync();
    expect(isZoomPinned("c1")).toBe(true);
    expect(clicks).toHaveLength(0);
    vi.advanceTimersByTime(1);
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    expect(isZoomPinned()).toBe(false);
    expect(get(hoveredCard)).toBeNull();
  });
});
