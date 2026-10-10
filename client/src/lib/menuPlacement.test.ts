// @vitest-environment jsdom
//
// #2960: the ability menu is a fixed box placed from the card's rect:
// above the card, flipped below when there is no room, clamped to the
// viewport, and moved out of the card into the board's popover host.

import { afterEach, describe, expect, it } from "vitest";
import { MENU_GUTTER, POPOVER_HOST_ATTR, anchoredMenu, placeMenu } from "./menuPlacement";

const rect = (left: number, top: number, w = 88, h = 123) => ({
  left,
  top,
  right: left + w,
  bottom: top + h,
});

describe("placeMenu", () => {
  it("opens above a card with room above it", () => {
    const p = placeMenu(rect(600, 500), 200, 160, 1280, 720);
    expect(p.side).toBe("above");
    expect(p.top + 160).toBeLessThanOrEqual(500);
  });

  it("flips below when the card is too near the top edge", () => {
    const p = placeMenu(rect(600, 20), 200, 160, 1280, 720);
    expect(p.side).toBe("below");
    expect(p.top).toBeGreaterThanOrEqual(20 + 123);
  });

  it("stays inside the viewport on every edge", () => {
    for (const [vw, vh] of [
      [1280, 720],
      [1998, 716],
      [390, 844],
    ]) {
      for (const [x, y] of [
        [0, 0],
        [vw - 88, 0],
        [0, vh - 123],
        [vw - 88, vh - 123],
        [vw / 2, vh / 2],
      ]) {
        const p = placeMenu(rect(x, y), 220, 300, vw, vh);
        expect(p.left).toBeGreaterThanOrEqual(MENU_GUTTER);
        expect(p.left + 220).toBeLessThanOrEqual(vw - MENU_GUTTER);
        expect(p.top).toBeGreaterThanOrEqual(MENU_GUTTER);
        expect(p.top + Math.min(300, p.maxHeight)).toBeLessThanOrEqual(vh - MENU_GUTTER);
      }
    }
  });

  it("caps a menu taller than the viewport so it scrolls inside itself", () => {
    const p = placeMenu(rect(600, 300), 220, 900, 1280, 400);
    expect(p.maxHeight).toBe(400 - 2 * MENU_GUTTER);
    expect(p.top).toBe(MENU_GUTTER);
  });

  it("opens a land at the foot of a 1998x716 screen above it, fully visible", () => {
    const p = placeMenu(rect(900, 540), 240, 170, 1998, 716);
    expect(p.side).toBe("above");
    expect(p.top).toBeGreaterThanOrEqual(MENU_GUTTER);
    expect(p.top + 170).toBeLessThanOrEqual(540);
  });
});

describe("anchoredMenu", () => {
  afterEach(() => {
    document.body.innerHTML = "";
  });

  function mount(withHost: boolean) {
    const host = document.createElement("div");
    host.setAttribute(POPOVER_HOST_ATTR, "");
    if (withHost) document.body.appendChild(host);
    const card = document.createElement("div");
    card.getBoundingClientRect = () =>
      ({ ...rect(600, 540), x: 600, y: 540, width: 88, height: 123 }) as DOMRect;
    const node = document.createElement("div");
    Object.defineProperty(node, "offsetWidth", { value: 200 });
    Object.defineProperty(node, "scrollHeight", { value: 150 });
    card.appendChild(node);
    document.body.appendChild(card);
    return { host, card, node };
  }

  it("moves the menu into the popover host and places it above the card", () => {
    const { host, node } = mount(true);
    const a = anchoredMenu(node);
    expect(node.parentElement).toBe(host);
    expect(node.dataset.side).toBe("above");
    expect(parseFloat(node.style.top) + 150).toBeLessThanOrEqual(540);
    a.destroy();
    expect(node.isConnected).toBe(false);
  });

  it("stays in the card when there is no host, still placed", () => {
    const { card, node } = mount(false);
    const a = anchoredMenu(node);
    expect(node.parentElement).toBe(card);
    expect(node.style.left).not.toBe("");
    a.destroy();
  });
});
