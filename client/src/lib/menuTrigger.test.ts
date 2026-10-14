// @vitest-environment jsdom
//
// #3499: the custom context menu's three ways in (mouse, keyboard,
// touch), and the places it leaves the browser's own menu alone.

import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import {
  LONG_PRESS_MS,
  contextMenuTrigger,
  keepsNativeMenu,
  menuPoint,
  type MenuPoint,
} from "./menuTrigger";

function pointer(type: string, pointerType: string, x: number, y: number): Event {
  const e = new MouseEvent(type, { bubbles: true, cancelable: true, clientX: x, clientY: y });
  Object.defineProperty(e, "pointerType", { value: pointerType });
  return e;
}

describe("menuPoint", () => {
  it("opens at the cursor", () => {
    expect(menuPoint({ clientX: 40, clientY: 70 }, { left: 0, top: 0 })).toEqual({ x: 40, y: 70 });
  });

  it("opens inside the focused element for the Menu key or Shift+F10", () => {
    expect(menuPoint({ clientX: 0, clientY: 0 }, { left: 100, top: 200 })).toEqual({
      x: 116,
      y: 216,
    });
  });
});

describe("keepsNativeMenu", () => {
  it("keeps the browser's menu on text fields and links only", () => {
    document.body.innerHTML = `
      <div id="card">
        <input id="i" /><textarea id="t"></textarea>
        <a id="a" href="/x"><span id="in-a">x</span></a>
        <button id="b">b</button>
      </div>`;
    for (const id of ["i", "t", "a", "in-a"]) {
      expect(keepsNativeMenu(document.getElementById(id)), id).toBe(true);
    }
    expect(keepsNativeMenu(document.getElementById("b"))).toBe(false);
    expect(keepsNativeMenu(document.getElementById("card"))).toBe(false);
    expect(keepsNativeMenu(null)).toBe(false);
  });
});

describe("contextMenuTrigger", () => {
  let node: HTMLElement;
  let opened: Array<{ at: MenuPoint; opener: HTMLElement | null }>;
  let destroy: () => void;

  beforeEach(() => {
    vi.useFakeTimers();
    document.body.innerHTML = `<div id="card"><button id="b">b</button><input id="i" /></div>`;
    node = document.getElementById("card")!;
    opened = [];
    destroy = contextMenuTrigger(node, (at, opener) => opened.push({ at, opener })).destroy;
  });
  afterEach(() => {
    destroy();
    vi.useRealTimers();
  });

  it("opens on a right-click and takes the browser's menu away", () => {
    const e = new MouseEvent("contextmenu", {
      bubbles: true,
      cancelable: true,
      clientX: 30,
      clientY: 40,
    });
    document.getElementById("b")!.dispatchEvent(e);
    expect(e.defaultPrevented).toBe(true);
    expect(opened).toEqual([{ at: { x: 30, y: 40 }, opener: document.getElementById("b") }]);
  });

  it("names the focused control as the opener, for the keyboard", () => {
    const b = document.getElementById("b")!;
    b.focus();
    node.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
    expect(opened[0].opener).toBe(b);
  });

  it("leaves a text field's menu alone", () => {
    const e = new MouseEvent("contextmenu", { bubbles: true, cancelable: true });
    document.getElementById("i")!.dispatchEvent(e);
    expect(e.defaultPrevented).toBe(false);
    expect(opened).toEqual([]);
  });

  it("opens on a held touch, and swallows the click the lift makes", () => {
    const b = document.getElementById("b")!;
    b.dispatchEvent(pointer("pointerdown", "touch", 50, 60));
    vi.advanceTimersByTime(LONG_PRESS_MS - 1);
    expect(opened).toEqual([]);
    vi.advanceTimersByTime(1);
    expect(opened).toEqual([{ at: { x: 50, y: 60 }, opener: b }]);

    b.dispatchEvent(pointer("pointerup", "touch", 50, 60));
    const clicked = vi.fn();
    b.addEventListener("click", clicked);
    b.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    expect(clicked).not.toHaveBeenCalled();

    // Android's own contextmenu for the same press does not open it twice.
    node.dispatchEvent(new MouseEvent("contextmenu", { bubbles: true, cancelable: true }));
    expect(opened).toHaveLength(1);
  });

  it("opens once when Android's contextmenu comes before our timer", () => {
    node.dispatchEvent(pointer("pointerdown", "touch", 50, 60));
    vi.advanceTimersByTime(LONG_PRESS_MS - 100);
    node.dispatchEvent(
      new MouseEvent("contextmenu", { bubbles: true, cancelable: true, clientX: 50, clientY: 60 }),
    );
    vi.advanceTimersByTime(LONG_PRESS_MS);
    expect(opened).toHaveLength(1);
    const clicked = vi.fn();
    node.addEventListener("click", clicked);
    node.dispatchEvent(new MouseEvent("click", { bubbles: true, cancelable: true }));
    expect(clicked).not.toHaveBeenCalled();
  });

  it("is a scroll, not a press, once the touch moves", () => {
    node.dispatchEvent(pointer("pointerdown", "touch", 50, 60));
    node.dispatchEvent(pointer("pointermove", "touch", 50, 90));
    vi.advanceTimersByTime(LONG_PRESS_MS * 2);
    expect(opened).toEqual([]);
  });

  it("ignores a held mouse button", () => {
    node.dispatchEvent(pointer("pointerdown", "mouse", 50, 60));
    vi.advanceTimersByTime(LONG_PRESS_MS * 2);
    expect(opened).toEqual([]);
  });
});
