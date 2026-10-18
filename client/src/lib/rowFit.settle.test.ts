// @vitest-environment jsdom
//
// #4111: a row a pixel short of fitting must settle on one card size.
// fitRow used to read the piles' rounded offsetWidth, so a shrink of
// under 1% (45.67px cards read back as 46px, then 45.33px as 45px) moved
// the scale by more than its 0.005 tolerance every frame, and the row
// flipped between three sizes forever, holding the page's main thread.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fitRow } from "./rowFit";

const CARD_W = 46;
const CARD_H = 64;
const GAP = 8;
// Three 46px cards and two 8px gaps need 154px; the row has 153.
const ROOM = 153;

let frames: FrameRequestCallback[] = [];
let resized: (() => void)[] = [];

class FakeResizeObserver {
  constructor(cb: () => void) {
    resized.push(cb);
  }
  observe(): void {}
  disconnect(): void {}
}

beforeEach(() => {
  frames = [];
  resized = [];
  vi.stubGlobal("ResizeObserver", FakeResizeObserver);
  vi.stubGlobal("requestAnimationFrame", (cb: FrameRequestCallback) => frames.push(cb));
  vi.stubGlobal("cancelAnimationFrame", () => {});
});

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
});

// A row's card list with a stubbed layout: each pile is one card wide at
// the list's --card-w (fractional, as the browser draws it), and
// offsetWidth rounds it, as the browser does.
function stubRow(): HTMLElement {
  const row = document.createElement("div");
  row.className = "row";
  row.dataset.zone = "artifacts";
  const probeWrap = document.createElement("div");
  probeWrap.className = "fit-probe";
  const probe = document.createElement("div");
  probe.className = "fit-probe-card";
  probeWrap.append(probe);
  const list = document.createElement("div");
  list.className = "row-cards";
  row.append(probeWrap, list);
  document.body.append(row);

  const cardW = (): number => parseFloat(list.style.getPropertyValue("--card-w")) || CARD_W;
  const cardH = (): number => parseFloat(list.style.getPropertyValue("--card-h")) || CARD_H;
  for (let i = 0; i < 3; i++) {
    const pile = document.createElement("div");
    Object.defineProperty(pile, "offsetWidth", { get: () => Math.round(cardW()) });
    Object.defineProperty(pile, "offsetHeight", { get: () => Math.round(cardH()) });
    pile.getBoundingClientRect = () => new DOMRect(0, 0, cardW(), cardH());
    list.append(pile);
  }
  Object.defineProperty(probe, "offsetWidth", { get: () => CARD_W });
  Object.defineProperty(probe, "offsetHeight", { get: () => CARD_H });
  Object.defineProperty(list, "clientWidth", { get: () => ROOM });

  const real = window.getComputedStyle.bind(window);
  vi.spyOn(window, "getComputedStyle").mockImplementation((el: Element) => {
    if (el !== list && el !== row) return real(el);
    return {
      columnGap: el === list ? `${GAP}px` : "0px",
      paddingTop: "0px",
      paddingBottom: "0px",
      paddingLeft: "0px",
      paddingRight: "0px",
      getPropertyValue: () => "",
    } as unknown as CSSStyleDeclaration;
  });
  return list;
}

describe("fitRow (#4111)", () => {
  it("settles a row a pixel short on one card size instead of oscillating", () => {
    const list = stubRow();
    const fit = fitRow(list, { budget: 0 });
    const sizes: string[] = [];
    // Run frames as the browser would: a frame measures, and a change of
    // card size resizes the piles, which the ResizeObserver reports.
    for (let i = 0; i < 12 && frames.length > 0; i++) {
      const before = list.style.getPropertyValue("--card-w");
      const run = frames;
      frames = [];
      run.forEach((cb) => cb(0));
      const after = list.style.getPropertyValue("--card-w");
      if (after !== before) {
        sizes.push(after);
        resized.forEach((cb) => cb());
      }
    }
    fit.destroy();

    // Shrunk once, to the size that fits 153px exactly, and no more.
    expect(sizes).toHaveLength(1);
    expect(parseFloat(sizes[0])).toBeCloseTo((CARD_W * (ROOM - 2 * GAP)) / (3 * CARD_W), 2);
    expect(frames).toHaveLength(0);
  });
});
