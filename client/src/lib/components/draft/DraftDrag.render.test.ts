// @vitest-environment jsdom
//
// #5848 §1 and §2: dragging a draft card with pointer events (pack to
// Main or Sideboard, between the two, mouse and touch) and the large
// preview beside a hovered, focused or held card.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { DraftCard } from "../../draft";
import { HOLD_MS } from "../../draftDrag.svelte";
import { hidePreview, placePreview } from "../../draftPreview.svelte";
import { L } from "../../labels";
import { settings } from "../../settings";
import { cleanup, click, flushSync, render } from "../../test/render.svelte";
import DraftPack from "./DraftPack.svelte";
import DraftSplit from "./DraftSplit.svelte";

const card = (id: string, name: string, over: Partial<DraftCard> = {}): DraftCard => ({
  id,
  name,
  mana_value: 2,
  colors: ["U"],
  type_line: "Creature",
  ...over,
});

const bolt = card("c1", "Lightning Bolt");
const trainer = card("c2", "Thundertrap Trainer");
const picks = [trainer, card("c3", "Mindwhisker")];

/** A pointer event that jsdom will carry the pointer fields on. */
function fire(
  target: Element | Window,
  type: string,
  x = 0,
  y = 0,
  kind: "mouse" | "touch" | "pen" = "mouse",
  id = 1,
): void {
  const ev = new MouseEvent(type, {
    bubbles: true,
    cancelable: true,
    clientX: x,
    clientY: y,
    button: 0,
  });
  Object.defineProperty(ev, "pointerId", { value: id });
  Object.defineProperty(ev, "pointerType", { value: kind });
  target.dispatchEvent(ev);
  flushSync();
}

/** Whatever `elementFromPoint` should find at the next pointer move. */
let hit: Element | null = null;

beforeEach(() => {
  vi.useFakeTimers();
  document.elementFromPoint = () => hit;
  hit = null;
  settings.update((s) => ({ ...s, display: { ...s.display, hoverDelayMs: 300 } }));
});

afterEach(() => {
  hidePreview();
  cleanup();
  vi.runOnlyPendingTimers(); // a drag's click guard is removed on a timer
  vi.useRealTimers();
  vi.restoreAllMocks();
  localStorage.clear();
});

const text = (el: Element | null): string => (el?.textContent ?? "").replace(/\s+/g, " ").trim();
const area = (c: HTMLElement, id: string): HTMLElement =>
  c.querySelector<HTMLElement>(`[data-testid="${id}"]`)!;
const move = (c: HTMLElement, id: string, name: string): HTMLElement =>
  [...area(c, id).querySelectorAll<HTMLElement>("[data-move-id]")].find((b) =>
    b.getAttribute("aria-label")!.includes(`: ${name}`),
  )!;

function split(side: string[] = [], onmove = vi.fn(), images = false) {
  const r = render(
    DraftSplit as never,
    {
      cards: picks,
      side,
      onmove,
      images,
      ids: { main: "m", side: "s" },
    } as never,
  );
  return { ...r, onmove };
}

/** Where the preview is on screen: the card's right edge, with its gap. */
function stubRect(el: Element, left: number, top: number, w = 120, h = 160): void {
  const rect = {
    left,
    top,
    right: left + w,
    bottom: top + h,
    width: w,
    height: h,
    x: left,
    y: top,
  };
  Object.defineProperty(el, "getBoundingClientRect", { value: () => rect, configurable: true });
}

describe("dragging between Main and the Sideboard", () => {
  it("moves a card dropped on the other area, and lifts, highlights and tidies up as it goes", () => {
    const { container: c, onmove } = split();
    const btn = move(c, "m", "Thundertrap Trainer");
    const side = area(c, "s");
    fire(btn, "pointerdown", 10, 10);
    fire(window, "pointermove", 12, 12); // within the slop: still a click
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).toBeNull();

    hit = side;
    fire(window, "pointermove", 60, 200);
    const ghost = document.querySelector<HTMLElement>('[data-testid="draft-drag-ghost"]')!;
    expect(ghost).not.toBeNull();
    // A copy for the eye: no second button, nothing for a screen reader.
    expect(ghost.getAttribute("aria-hidden")).toBe("true");
    expect(document.querySelectorAll("[data-move-id]").length).toBe(picks.length);
    expect(btn.style.opacity).toBe("0.35");
    expect(side.classList.contains("drop")).toBe(true);
    expect(area(c, "m").classList.contains("drop")).toBe(false);

    fire(window, "pointerup", 60, 200);
    expect(onmove).toHaveBeenCalledExactlyOnceWith("c2", "side");
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).toBeNull();
    expect(btn.style.opacity).toBe("");
    expect(side.classList.contains("drop")).toBe(false);
    expect(text(c.querySelector('[data-testid="draft-split-said"]'))).toBe(
      "Moved Thundertrap Trainer to sideboard.",
    );
  });

  it("moves a card from the Sideboard back to Main", () => {
    const { container: c, onmove } = split(["c3"]);
    const btn = move(c, "s", "Mindwhisker");
    fire(btn, "pointerdown", 10, 10);
    hit = area(c, "m");
    fire(window, "pointermove", 80, 40);
    fire(window, "pointerup", 80, 40);
    expect(onmove).toHaveBeenCalledExactlyOnceWith("c3", "main");
  });

  it("drags image cards on the deckbuilding screen too", () => {
    const { container: c, onmove } = split([], vi.fn(), true);
    const btn = move(c, "m", "Thundertrap Trainer");
    fire(btn, "pointerdown", 5, 5);
    hit = area(c, "s");
    fire(window, "pointermove", 70, 90);
    fire(window, "pointerup", 70, 90);
    expect(onmove).toHaveBeenCalledExactlyOnceWith("c2", "side");
  });

  it("does nothing when a card is dropped where it was, or on no area", () => {
    const { container: c, onmove } = split();
    const btn = move(c, "m", "Thundertrap Trainer");
    fire(btn, "pointerdown", 10, 10);
    hit = area(c, "m");
    fire(window, "pointermove", 40, 40);
    fire(window, "pointerup", 40, 40);
    fire(btn, "pointerdown", 10, 10);
    hit = null;
    fire(window, "pointermove", 400, 400);
    fire(window, "pointerup", 400, 400);
    expect(onmove).not.toHaveBeenCalled();
  });

  it("puts the card back on Escape, and the click after a drag is not a move", () => {
    const { container: c, onmove } = split();
    const btn = move(c, "m", "Thundertrap Trainer");
    fire(btn, "pointerdown", 10, 10);
    hit = area(c, "s");
    fire(window, "pointermove", 90, 90);
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).not.toBeNull();
    window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
    flushSync();
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).toBeNull();
    expect(btn.style.opacity).toBe("");
    fire(window, "pointerup", 90, 90);
    expect(onmove).not.toHaveBeenCalled();

    // A real drop is followed by the browser's click, which must not move it back.
    fire(btn, "pointerdown", 10, 10);
    fire(window, "pointermove", 90, 90);
    fire(window, "pointerup", 90, 90);
    expect(onmove).toHaveBeenCalledTimes(1);
    click(btn);
    expect(onmove).toHaveBeenCalledTimes(1);
  });

  it("still moves a card with a click, and says so", () => {
    const { container: c, onmove } = split();
    const btn = move(c, "m", "Thundertrap Trainer");
    expect(btn.getAttribute("aria-label")).toMatch(/^Move to sideboard: Thundertrap Trainer/);
    fire(btn, "pointerdown", 10, 10);
    fire(window, "pointerup", 10, 10);
    click(btn);
    expect(onmove).toHaveBeenCalledExactlyOnceWith("c2", "side");
    expect(text(c.querySelector('[data-testid="draft-split-said"]'))).toBe(
      "Moved Thundertrap Trainer to sideboard.",
    );
  });

  it("keeps the view in place on a click: focus follows the card with no scroll (#5921)", async () => {
    const onmove = vi.fn();
    const r = split([], onmove);
    onmove.mockImplementation((id: string) => r.setProps({ side: [id] } as never));
    const focus = vi.spyOn(HTMLElement.prototype, "focus");
    click(move(r.container, "m", "Thundertrap Trainer"));
    await vi.runAllTimersAsync();
    flushSync();
    const moved = move(r.container, "s", "Thundertrap Trainer");
    expect(document.activeElement).toBe(moved);
    expect(focus).toHaveBeenCalledWith({ preventScroll: true });
  });

  it("is not draggable with the old HTML5 drag, and cancels the browser's own", () => {
    const { container: c } = split();
    const btn = move(c, "m", "Thundertrap Trainer");
    expect(btn.getAttribute("draggable")).toBeNull();
    const ev = new Event("dragstart", { bubbles: true, cancelable: true });
    btn.dispatchEvent(ev);
    expect(ev.defaultPrevented).toBe(true);
  });
});

describe("dragging with a finger", () => {
  it("treats a swipe before the hold as a scroll", () => {
    const { container: c, onmove } = split();
    const btn = move(c, "m", "Thundertrap Trainer");
    fire(btn, "pointerdown", 10, 10, "touch");
    vi.advanceTimersByTime(HOLD_MS / 2);
    hit = area(c, "s");
    fire(window, "pointermove", 10, 60, "touch");
    vi.advanceTimersByTime(HOLD_MS);
    fire(window, "pointermove", 10, 120, "touch");
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).toBeNull();
    fire(window, "pointerup", 10, 120, "touch");
    expect(onmove).not.toHaveBeenCalled();
  });

  it("lifts after the hold, shows the preview, and then drags", () => {
    const { container: c, onmove } = split();
    const btn = move(c, "m", "Thundertrap Trainer");
    stubRect(btn, 20, 100);
    fire(btn, "pointerdown", 30, 110, "touch");
    vi.advanceTimersByTime(HOLD_MS - 20);
    expect(document.querySelector('[data-testid="draft-preview"]')).toBeNull();
    vi.advanceTimersByTime(40);
    flushSync();
    expect(document.querySelector('[data-testid="draft-preview"]')).not.toBeNull();

    hit = area(c, "s");
    fire(window, "pointermove", 40, 200, "touch");
    // Dragging hides the preview and shows the ghost.
    expect(document.querySelector('[data-testid="draft-preview"]')).toBeNull();
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).not.toBeNull();
    fire(window, "pointerup", 40, 200, "touch");
    expect(onmove).toHaveBeenCalledExactlyOnceWith("c2", "side");
  });

  it("a held card released in place is not a pick, and the preview goes", () => {
    const { container: c, onmove } = split();
    const btn = move(c, "m", "Thundertrap Trainer");
    fire(btn, "pointerdown", 30, 110, "touch");
    vi.advanceTimersByTime(HOLD_MS + 10);
    flushSync();
    expect(document.querySelector('[data-testid="draft-preview"]')).not.toBeNull();
    fire(window, "pointerup", 30, 110, "touch");
    expect(document.querySelector('[data-testid="draft-preview"]')).toBeNull();
    click(btn);
    expect(onmove).not.toHaveBeenCalled();
  });
});

describe("dragging from the pack", () => {
  const pack = [bolt, card("c9", "Counterspell")];
  function pickPack(take = 1) {
    const onpick = vi.fn();
    const onaside = vi.fn();
    const r = render(DraftPack as never, { pack, take, onpick, onaside } as never);
    // The picks' areas, stood in for by two drop targets.
    const main = document.createElement("div");
    main.dataset.dropZone = "main";
    const side = document.createElement("div");
    side.dataset.dropZone = "side";
    document.body.append(main, side);
    const btn = (name: string): HTMLElement =>
      [...r.container.querySelectorAll<HTMLElement>("button.pick-card")].find((b) =>
        b.getAttribute("aria-label")!.startsWith(name),
      )!;
    return { ...r, onpick, onaside, main, side, btn };
  }
  const drag = (el: HTMLElement, to: Element): void => {
    fire(el, "pointerdown", 10, 10);
    hit = to;
    fire(window, "pointermove", 300, 300);
    fire(window, "pointerup", 300, 300);
  };

  it("picks a card dragged onto the Sideboard, and sets it aside first", () => {
    const { onpick, onaside, side, btn, container } = pickPack();
    drag(btn("Lightning Bolt"), side);
    expect(onaside).toHaveBeenCalledExactlyOnceWith(["c1"]);
    expect(onpick).toHaveBeenCalledExactlyOnceWith(["c1"]);
    expect(onaside.mock.invocationCallOrder[0]).toBeLessThan(onpick.mock.invocationCallOrder[0]);
    expect(text(container.querySelector('[data-testid="draft-pack-said"]'))).toBe(
      `Picked Lightning Bolt into ${L.draftSideboardZone}.`,
    );
  });

  it("picks a card dragged onto Main with no set-aside", () => {
    const { onpick, onaside, main, btn } = pickPack();
    drag(btn("Counterspell"), main);
    expect(onaside).not.toHaveBeenCalled();
    expect(onpick).toHaveBeenCalledExactlyOnceWith(["c9"]);
  });

  it("picks nothing when it is dropped on neither, and a click still picks into Main's flow", () => {
    const { onpick, btn } = pickPack();
    drag(btn("Lightning Bolt"), document.body);
    expect(onpick).not.toHaveBeenCalled();
    vi.advanceTimersByTime(100); // the browser's click right after a drag is eaten; a later one is not
    click(btn("Lightning Bolt"));
    expect(btn("Lightning Bolt").getAttribute("aria-pressed")).toBe("true");
    const primary = [...document.querySelectorAll<HTMLButtonElement>("button.primary")][0];
    click(primary);
    expect(onpick).toHaveBeenCalledExactlyOnceWith(["c1"]);
  });

  it("with two to take, holds the first card dropped and sends both on the second drop, no confirm (#5921)", () => {
    const { onpick, onaside, side, main, btn, container } = pickPack(2);
    drag(btn("Lightning Bolt"), side);
    // The server takes the pair in one pick: the first waits, chosen for its pile.
    expect(onpick).not.toHaveBeenCalled();
    vi.advanceTimersByTime(100);
    expect(btn("Lightning Bolt").getAttribute("aria-pressed")).toBe("true");
    expect(text(container.querySelector(".prompt"))).toContain(
      `Lightning Bolt to ${L.draftSideboardZone}`,
    );
    expect(text(container.querySelector('[data-testid="draft-pack-said"]'))).toBe(
      `Chose Lightning Bolt for ${L.draftSideboardZone}. Drag or choose 1 more.`,
    );
    drag(btn("Counterspell"), main);
    expect(onaside).toHaveBeenCalledExactlyOnceWith(["c1"]);
    expect(onpick).toHaveBeenCalledExactlyOnceWith(["c1", "c9"]);
    expect(onaside.mock.invocationCallOrder[0]).toBeLessThan(onpick.mock.invocationCallOrder[0]);
    expect(text(container.querySelector('[data-testid="draft-pack-said"]'))).toBe(
      `Picked Lightning Bolt into ${L.draftSideboardZone} and Counterspell into ${L.draftDeckZone}.`,
    );
  });

  it("with two to take, a drop that completes a pair chosen by a click sends it at once (#5921)", () => {
    const { onpick, onaside, side, btn } = pickPack(2);
    click(btn("Counterspell"));
    expect(onpick).not.toHaveBeenCalled();
    drag(btn("Lightning Bolt"), side);
    expect(onaside).toHaveBeenCalledExactlyOnceWith(["c1"]);
    expect(onpick).toHaveBeenCalledExactlyOnceWith(["c9", "c1"]);
  });

  it("with one to take, a drop replaces a card chosen by a click and picks the dropped one (#5921)", () => {
    const { onpick, onaside, main, btn } = pickPack();
    click(btn("Lightning Bolt"));
    drag(btn("Counterspell"), main);
    expect(onaside).not.toHaveBeenCalled();
    expect(onpick).toHaveBeenCalledExactlyOnceWith(["c9"]);
  });

  it("keeps click then confirm for clicks with two to take", () => {
    const { onpick, btn, container } = pickPack(2);
    click(btn("Counterspell"));
    click(btn("Lightning Bolt"));
    expect(onpick).not.toHaveBeenCalled();
    click(container.querySelector<HTMLElement>("button.primary")!);
    expect(onpick).toHaveBeenCalledExactlyOnceWith(["c9", "c1"]);
  });

  it("cannot be dragged while a pick is on its way", () => {
    const onpick = vi.fn();
    const r = render(DraftPack as never, { pack, take: 1, busy: true, onpick } as never);
    const side = document.createElement("div");
    side.dataset.dropZone = "side";
    document.body.append(side);
    const btn = r.container.querySelector<HTMLElement>("button.pick-card")!;
    fire(btn, "pointerdown", 10, 10);
    hit = side;
    fire(window, "pointermove", 200, 200);
    fire(window, "pointerup", 200, 200);
    expect(onpick).not.toHaveBeenCalled();
  });

  it("puts a drag away if its card leaves the pack mid-drag", () => {
    const { btn, container, setProps, side } = pickPack();
    const b = btn("Lightning Bolt");
    fire(b, "pointerdown", 10, 10);
    hit = side;
    fire(window, "pointermove", 200, 200);
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).not.toBeNull();
    setProps({ pack: [pack[1]] } as never);
    expect(container.querySelectorAll("button.pick-card").length).toBe(1);
    expect(document.querySelector('[data-testid="draft-drag-ghost"]')).toBeNull();
  });
});

describe("the large preview", () => {
  const pv = (): HTMLElement | null => document.querySelector('[data-testid="draft-preview"]');
  const showing = (): string | null => pv()?.getAttribute("data-card-id") ?? null;
  const pack = [bolt, card("c9", "Counterspell")];
  function open() {
    const r = render(DraftPack as never, { pack, take: 1, onpick: vi.fn() } as never);
    // In the order of `pack`; the pack itself is shown sorted by name.
    const all = [...r.container.querySelectorAll<HTMLElement>("button.pick-card")];
    const btns = pack.map(
      (p) => all.find((b) => b.getAttribute("aria-label")!.startsWith(p.name))!,
    );
    stubRect(btns[0], 100, 200, 200, 280);
    stubRect(btns[1], 400, 200, 200, 280);
    return { ...r, btns };
  }

  it("appears after the delay, beside the card, and goes when the pointer leaves", () => {
    const { btns } = open();
    fire(btns[0], "pointerenter", 150, 250);
    vi.advanceTimersByTime(299);
    flushSync();
    expect(pv()).toBeNull();
    vi.advanceTimersByTime(2);
    flushSync();
    expect(showing()).not.toBeNull();
    // To the right: the card's right edge (300) and the 14px gap, not on the card.
    expect(pv()!.style.transform).toMatch(/^translate\(314px, /);
    fire(btns[0], "pointerleave", 700, 700);
    vi.advanceTimersByTime(200);
    flushSync();
    expect(pv()).toBeNull();
  });

  it("uses the Hover preview delay setting", () => {
    settings.update((s) => ({ ...s, display: { ...s.display, hoverDelayMs: 900 } }));
    const { btns } = open();
    fire(btns[0], "pointerenter");
    vi.advanceTimersByTime(800);
    flushSync();
    expect(pv()).toBeNull();
    vi.advanceTimersByTime(150);
    flushSync();
    expect(pv()).not.toBeNull();
  });

  it("swaps at once from card to card, across the gap between them", () => {
    const { btns } = open();
    fire(btns[0], "pointerenter");
    vi.advanceTimersByTime(300);
    flushSync();
    const first = showing();
    fire(btns[0], "pointerleave");
    vi.advanceTimersByTime(20); // the gap: not yet gone
    fire(btns[1], "pointerenter");
    flushSync();
    expect(showing()).not.toBe(first);
    expect(showing()).toBe("c9"); // no second delay
    expect(pv()!.style.transform).toMatch(/^translate\(614px, /);
  });

  it("flips to the left when there is no room on the right", () => {
    const was = Object.getOwnPropertyDescriptor(HTMLElement.prototype, "offsetWidth");
    Object.defineProperty(HTMLElement.prototype, "offsetWidth", {
      configurable: true,
      get: () => 340,
    });
    try {
      const { btns } = open();
      stubRect(btns[0], 800, 200, 200, 280);
      fire(btns[0], "pointerenter");
      vi.advanceTimersByTime(300);
      flushSync();
      // 800 - 14 - 340
      expect(pv()!.style.transform).toMatch(/^translate\(446px, /);
    } finally {
      if (was) Object.defineProperty(HTMLElement.prototype, "offsetWidth", was);
    }
  });

  it("shows for keyboard focus with no delay, and hides on blur", () => {
    const { btns } = open();
    btns[0].focus();
    flushSync();
    expect(showing()).toBe("c1");
    btns[0].blur();
    flushSync();
    expect(pv()).toBeNull();
  });

  it("hides on Escape and on scroll", () => {
    const { btns } = open();
    for (const how of ["escape", "scroll"]) {
      fire(btns[0], "pointerenter");
      vi.advanceTimersByTime(300);
      flushSync();
      expect(pv(), how).not.toBeNull();
      if (how === "escape") window.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape" }));
      else document.dispatchEvent(new Event("scroll"));
      flushSync();
      expect(pv(), how).toBeNull();
    }
  });

  it("is hidden while dragging and comes back for the next hover", () => {
    const { btns } = open();
    const zone = document.createElement("div");
    zone.dataset.dropZone = "side";
    document.body.append(zone);
    fire(btns[0], "pointerenter");
    vi.advanceTimersByTime(300);
    flushSync();
    expect(pv()).not.toBeNull();
    fire(btns[0], "pointerdown", 150, 250);
    hit = zone;
    fire(window, "pointermove", 400, 600);
    expect(pv()).toBeNull();
    fire(btns[1], "pointerenter"); // over another card mid-drag: still nothing
    vi.advanceTimersByTime(500);
    flushSync();
    expect(pv()).toBeNull();
    fire(window, "pointerup", 400, 600);
    fire(btns[1], "pointerenter");
    vi.advanceTimersByTime(300);
    flushSync();
    expect(showing()).toBe("c9");
  });

  it("has a place under the image for what a surface says about the card", () => {
    const r = render(
      DraftSplit as never,
      {
        cards: picks,
        side: [],
        onmove: vi.fn(),
        ids: { main: "m", side: "s" },
        previewExtra: undefined,
      } as never,
    );
    const btn = move(r.container, "m", "Thundertrap Trainer");
    fire(btn, "pointerenter");
    vi.advanceTimersByTime(300);
    flushSync();
    expect(pv()).not.toBeNull();
    // No snippet given: no empty box.
    expect(document.querySelector('[data-testid="draft-preview-extra"]')).toBeNull();
  });

  it("ignores a touch's pointerenter: a touch shows it by holding", () => {
    const { btns } = open();
    fire(btns[0], "pointerenter", 0, 0, "touch");
    vi.advanceTimersByTime(1000);
    flushSync();
    expect(pv()).toBeNull();
  });
});

describe("placePreview", () => {
  const anchor = { left: 100, top: 200, right: 300, bottom: 480 };
  const size = { w: 340, h: 600 };

  it("puts the preview to the right, then to the left when it does not fit", () => {
    expect(placePreview(anchor, size, { vw: 1440, vh: 900 })).toEqual({ x: 314, y: 40 });
    expect(
      placePreview({ ...anchor, left: 1000, right: 1200 }, size, { vw: 1440, vh: 900 }),
    ).toEqual({ x: 646, y: 40 });
  });

  it("keeps it on screen", () => {
    const r = placePreview({ left: 0, top: -50, right: 200, bottom: 100 }, size, {
      vw: 1440,
      vh: 900,
    });
    expect(r.y).toBe(8);
    expect(
      placePreview({ left: 0, top: 800, right: 200, bottom: 900 }, size, { vw: 1440, vh: 900 }).y,
    ).toBe(292);
  });

  it("on a phone, centres it above the card, or below when the top has no room", () => {
    const small = { w: 300, h: 300 };
    expect(placePreview(anchor, small, { vw: 400, vh: 800 })).toEqual({ x: 50, y: 492 });
    expect(
      placePreview({ left: 20, top: 500, right: 140, bottom: 660 }, small, { vw: 400, vh: 800 }),
    ).toEqual({
      x: 50,
      y: 186,
    });
  });
});
