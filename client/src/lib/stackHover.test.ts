import { describe, expect, it, vi } from "vitest";
import { get, writable } from "svelte/store";

import type { CardView } from "./protocol";
import { createStackHover } from "./stackHover";

const c = (id: string) => ({ instance_id: id, name: id }) as CardView;

describe("createStackHover (#322, shared by the docked card and the lane)", () => {
  it("previews at once with no dwell, and clears on leave", () => {
    const store = writable<CardView | null>(null);
    const h = createStackHover(store);
    h.enter("item", c("a"), 0);
    expect(get(store)?.instance_id).toBe("a");
    h.leave("item");
    expect(get(store)).toBeNull();
  });

  it("waits out the dwell, and a leave before it fires shows nothing", () => {
    vi.useFakeTimers({ toFake: ["setTimeout", "clearTimeout"] });
    try {
      const store = writable<CardView | null>(null);
      const h = createStackHover(store);
      h.enter("item", c("a"), 300);
      expect(get(store)).toBeNull();
      vi.advanceTimersByTime(300);
      expect(get(store)?.instance_id).toBe("a");

      h.leave("item");
      h.enter("item", c("b"), 300);
      h.leave("item");
      vi.advanceTimersByTime(300);
      expect(get(store)).toBeNull();
    } finally {
      vi.useRealTimers();
    }
  });

  it("never clears a preview someone else wrote after it", () => {
    const store = writable<CardView | null>(null);
    const h = createStackHover(store);
    h.enter("item", c("a"), 0);
    store.set(c("battlefield-card"));
    h.leave("item");
    expect(get(store)?.instance_id).toBe("battlefield-card");
  });

  it("drops the preview when its row resolves out from under the cursor", () => {
    const store = writable<CardView | null>(null);
    const h = createStackHover(store);
    h.enter("item", c("a"), 0);
    h.sync(["item", "other"]);
    expect(get(store)?.instance_id).toBe("a");
    h.sync(["other"]);
    expect(get(store)).toBeNull();
  });

  it("ignores a row with nothing to preview", () => {
    const store = writable<CardView | null>(c("x"));
    const h = createStackHover(store);
    h.enter("item", null, 0);
    h.destroy();
    expect(get(store)?.instance_id).toBe("x");
  });
});
