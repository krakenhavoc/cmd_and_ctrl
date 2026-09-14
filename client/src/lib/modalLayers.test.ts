import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";
import { modalDepth, modalOpen, isModalOpen, pushModalLayer, _resetForTests } from "./modalLayers";

describe("modalLayers", () => {
  beforeEach(() => _resetForTests());

  it("starts closed", () => {
    expect(get(modalOpen)).toBe(false);
    expect(get(modalDepth)).toBe(0);
    expect(isModalOpen()).toBe(false);
  });

  it("opens while any layer is registered", () => {
    const off = pushModalLayer();
    expect(get(modalOpen)).toBe(true);
    off();
    expect(get(modalOpen)).toBe(false);
  });

  it("stays open until the last layer drops, in any unmount order", () => {
    // A cost modal on top of a choice prompt on top of settings; the
    // outer one can finish first when a snapshot tears the stack down.
    const a = pushModalLayer();
    const b = pushModalLayer();
    const c = pushModalLayer();
    expect(get(modalDepth)).toBe(3);
    a();
    expect(get(modalOpen)).toBe(true);
    c();
    expect(get(modalOpen)).toBe(true);
    b();
    expect(get(modalOpen)).toBe(false);
  });

  it("survives a double unregister without stranding the keymap off", () => {
    // Svelte can run a cleanup twice across a keyed re-mount. With a
    // bare counter that drives the depth negative and every shortcut
    // stays dead for the rest of the session.
    const a = pushModalLayer();
    const b = pushModalLayer();
    a();
    a();
    a();
    expect(get(modalDepth)).toBe(1);
    expect(get(modalOpen)).toBe(true);
    b();
    expect(get(modalDepth)).toBe(0);
    expect(get(modalOpen)).toBe(false);
  });

  it("notifies subscribers on every transition", () => {
    const seen: boolean[] = [];
    const stop = modalOpen.subscribe((v) => seen.push(v));
    const off = pushModalLayer();
    off();
    stop();
    expect(seen).toEqual([false, true, false]);
  });
});
