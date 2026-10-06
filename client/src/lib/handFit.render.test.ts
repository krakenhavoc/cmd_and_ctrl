// @vitest-environment jsdom
//
// handFit.render.test.ts — #2395. handFan.test.ts pins fitFan's
// arithmetic; this file pins that the viewer's own hand measures its
// row and one card and draws what fitFan answers: the overlap it
// publishes, the tilt on each slot, and the scroll class.
//
// jsdom lays nothing out, so the row and the card are stubbed: the
// strip's clientWidth is `row`, and a slot is 168×235 (a 1440×900
// card). A ResizeObserver stand-in reports a resize on demand.

import { describe, it, expect, afterEach, beforeEach, vi } from "vitest";

vi.mock("./sounds", () => ({ play: () => {} }));

import Hand from "./components/board/Hand.svelte";
import type { CardView } from "./protocol";
import { render, cleanup, flushSync } from "./test/render.svelte";

const ME = "me";

const cardsOf = (n: number): CardView[] =>
  Array.from({ length: n }, (_, i) => ({
    instance_id: `c${i}`,
    name: `Card ${i}`,
    owner: ME,
    controller: ME,
    type_line: "Creature — Elf",
  }));

let row = 2000;
const observers: (() => void)[] = [];
const restore: (() => void)[] = [];

function stub<K extends "clientWidth" | "offsetWidth" | "offsetHeight">(
  key: K,
  get: (el: HTMLElement) => number,
): void {
  const was = Object.getOwnPropertyDescriptor(HTMLElement.prototype, key);
  Object.defineProperty(HTMLElement.prototype, key, {
    configurable: true,
    get(this: HTMLElement) {
      return get(this);
    },
  });
  restore.push(() => {
    if (was) Object.defineProperty(HTMLElement.prototype, key, was);
  });
}

beforeEach(() => {
  row = 2000;
  observers.length = 0;
  stub("clientWidth", (el) => (el.classList.contains("hand") ? row : 0));
  stub("offsetWidth", (el) => (el.classList.contains("hand-slot") ? 168 : 0));
  stub("offsetHeight", (el) => (el.classList.contains("hand-slot") ? 235 : 0));
  vi.stubGlobal(
    "ResizeObserver",
    class {
      constructor(private readonly cb: () => void) {
        observers.push(() => this.cb());
      }
      observe(): void {}
      disconnect(): void {}
    },
  );
});

afterEach(() => {
  cleanup();
  for (const r of restore.splice(0)) r();
  vi.unstubAllGlobals();
});

function mount(n: number, isSelf = true) {
  const r = render(
    Hand as never,
    {
      hand: { kind: "hand", owner: ME, count: n, cards: isSelf ? cardsOf(n) : [] },
      isSelf,
      viewerID: ME,
    } as never,
  );
  flushSync();
  return r;
}

function resize(width: number): void {
  row = width;
  for (const o of observers) o();
  flushSync();
}

const strip = (c: HTMLElement) => c.querySelector<HTMLElement>(".hand")!;
const overlapOf = (c: HTMLElement) => Number(strip(c).style.getPropertyValue("--hand-overlap"));
const transforms = (c: HTMLElement) =>
  [...c.querySelectorAll<HTMLElement>(".hand-slot")].map((s) => s.style.transform);

describe("the hand fits its row (#2395)", () => {
  it("keeps the resting fan in a row with room", () => {
    const { container } = mount(7);
    expect(overlapOf(container)).toBe(0.5);
    expect(transforms(container)[0]).toContain("rotate(-21deg)");
    expect(strip(container).classList.contains("scroll")).toBe(false);
  });

  it("tightens and flattens beside the coach card, and loosens again when it goes", () => {
    const { container } = mount(7);
    // 1440×900 beside the coach card: 460px of row, 4px padding a side
    // in the browser (jsdom computes none, so all 460 is room here).
    resize(460);
    expect(overlapOf(container)).toBeGreaterThan(0.7);
    expect(transforms(container)[0]).toContain("rotate(0deg)");
    // The fan is no wider than its row: one card, and the rest's slivers.
    expect(168 + 6 * 168 * (1 - overlapOf(container))).toBeLessThanOrEqual(460.001);
    expect(strip(container).classList.contains("scroll")).toBe(false);

    resize(2000);
    expect(overlapOf(container)).toBe(0.5);
    expect(transforms(container)[0]).toContain("rotate(-21deg)");
  });

  it("scrolls a hand too long for its row even flat", () => {
    const { container } = mount(30);
    resize(460);
    expect(strip(container).classList.contains("scroll")).toBe(true);
  });

  it("leaves an opponent's face-down fan at its resting overlap", () => {
    const { container } = mount(7, false);
    resize(200);
    expect(overlapOf(container)).toBe(0.62);
    expect(strip(container).classList.contains("scroll")).toBe(false);
  });
});
