import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import {
  DEFAULT_STACK_HOLD_MS,
  STACK_HOLD_CHOICES_MS,
  STACK_HOLD_MAX_MS,
  _resetForTests,
  clampStackHoldMs,
  combinedPassDelayMs,
  noteStackSeen,
  setStackHoldStatus,
  stackHoldRemainingMs,
  stackHoldStatus,
  stackHoldStatusText,
  topStackItem,
} from "./stackHold";
import type { CardView, GameView, StackItemView } from "./protocol";

// The viewer is "me"; "opp" is somebody else. A stack is listed bottom
// to top, as the wire sends it.
function item(id: string, controller: string): StackItemView {
  return { id, kind: "spell", controller, owner: controller, source_card_id: id };
}

function view(items: StackItemView[], cards: Partial<CardView>[] = []): GameView {
  return {
    stack_items: items,
    stack: { kind: "stack", owner: "", count: cards.length, cards: cards as CardView[] },
  } as unknown as GameView;
}

function remaining(
  v: GameView,
  firstSeen: Map<string, number>,
  now: number,
  holdMs: unknown = 2000,
  viewerID: string | null = "me",
): number {
  return stackHoldRemainingMs({ view: v, viewerID, firstSeen, holdMs, now });
}

describe("clampStackHoldMs", () => {
  it("keeps the offered choices", () => {
    for (const ms of STACK_HOLD_CHOICES_MS) expect(clampStackHoldMs(ms)).toBe(ms);
    expect(STACK_HOLD_CHOICES_MS).toEqual([0, 1000, 2000, 3000]);
  });

  it("clamps a stored value to 0–3000", () => {
    expect(clampStackHoldMs(-50)).toBe(0);
    expect(clampStackHoldMs(60_000)).toBe(STACK_HOLD_MAX_MS);
    expect(clampStackHoldMs(1499.6)).toBe(1500);
  });

  it("falls back to the default for something that is not a number", () => {
    expect(clampStackHoldMs(undefined)).toBe(DEFAULT_STACK_HOLD_MS);
    expect(clampStackHoldMs("3000")).toBe(DEFAULT_STACK_HOLD_MS);
    expect(clampStackHoldMs(Number.NaN)).toBe(DEFAULT_STACK_HOLD_MS);
    expect(DEFAULT_STACK_HOLD_MS).toBe(2000);
  });
});

describe("topStackItem", () => {
  it("is the last stack item, the one that resolves next", () => {
    expect(topStackItem(view([item("a", "me"), item("b", "opp")]))).toEqual({
      id: "b",
      controller: "opp",
    });
  });

  it("falls back to the stack zone's last card mid-flight", () => {
    expect(topStackItem(view([], [{ instance_id: "c", controller: "opp" }]))).toEqual({
      id: "c",
      controller: "opp",
    });
  });

  it("is null on an empty stack or no view", () => {
    expect(topStackItem(view([]))).toBeNull();
    expect(topStackItem(null)).toBeNull();
  });
});

describe("noteStackSeen", () => {
  it("stamps a new item now and keeps an old item's first time", () => {
    const first = noteStackSeen(new Map(), view([item("a", "opp")]), 1000);
    expect(first.get("a")).toBe(1000);
    const second = noteStackSeen(first, view([item("a", "opp"), item("b", "opp")]), 1800);
    expect(second.get("a")).toBe(1000);
    expect(second.get("b")).toBe(1800);
  });

  it("drops an item that has left the stack", () => {
    const first = noteStackSeen(new Map(), view([item("a", "opp")]), 1000);
    const second = noteStackSeen(first, view([]), 1500);
    expect(second.size).toBe(0);
  });

  it("reads spell cards the item list has not caught up with", () => {
    const seen = noteStackSeen(new Map(), view([], [{ instance_id: "c", controller: "opp" }]), 5);
    expect(seen.get("c")).toBe(5);
  });

  it("does not change the map it was given", () => {
    const prev = new Map([["gone", 1]]);
    noteStackSeen(prev, view([item("a", "opp")]), 2);
    expect([...prev.keys()]).toEqual(["gone"]);
  });
});

describe("stackHoldRemainingMs", () => {
  it("holds an opponent's top item for the rest of the hold", () => {
    const v = view([item("a", "opp")]);
    const seen = noteStackSeen(new Map(), v, 1000);
    expect(remaining(v, seen, 1000)).toBe(2000);
    expect(remaining(v, seen, 2600)).toBe(400);
    expect(remaining(v, seen, 3000)).toBe(0);
    expect(remaining(v, seen, 9000)).toBe(0);
  });

  it("never holds the viewer's own top item", () => {
    const v = view([item("a", "me")]);
    const seen = noteStackSeen(new Map(), v, 1000);
    expect(remaining(v, seen, 1000)).toBe(0);
  });

  it("never holds when the viewer's own item is on top of an opponent's", () => {
    const v = view([item("a", "opp"), item("b", "me")]);
    const seen = noteStackSeen(new Map(), v, 1000);
    expect(remaining(v, seen, 1000)).toBe(0);
  });

  it("holds an opponent's item on top of the viewer's own", () => {
    const v = view([item("a", "me"), item("b", "opp")]);
    const seen = noteStackSeen(new Map(), v, 1000);
    expect(remaining(v, seen, 1500)).toBe(1500);
  });

  it("never holds with the setting at 0", () => {
    const v = view([item("a", "opp")]);
    const seen = noteStackSeen(new Map(), v, 1000);
    expect(remaining(v, seen, 1000, 0)).toBe(0);
  });

  it("never holds an empty stack, or for a spectator", () => {
    expect(remaining(view([]), new Map(), 1000)).toBe(0);
    const v = view([item("a", "opp")]);
    expect(remaining(v, noteStackSeen(new Map(), v, 1000), 1000, 2000, null)).toBe(0);
  });

  it("clamps the setting where it reads it", () => {
    const v = view([item("a", "opp")]);
    const seen = noteStackSeen(new Map(), v, 0);
    expect(remaining(v, seen, 0, 60_000)).toBe(STACK_HOLD_MAX_MS);
  });

  it("counts an item missing from the map as seen now", () => {
    expect(remaining(view([item("a", "opp")]), new Map(), 1000)).toBe(2000);
  });

  // A new frame does not restart the hold for an item already seen; a
  // new item on top starts its own.
  it("measures across frames from the first one, and afresh for a new top item", () => {
    const f1 = view([item("a", "opp")]);
    let seen = noteStackSeen(new Map(), f1, 1000);
    // A later frame with the same item: still measured from 1000.
    const f2 = view([item("a", "opp")]);
    seen = noteStackSeen(seen, f2, 2500);
    expect(remaining(f2, seen, 2500)).toBe(500);
    // Someone responds: the new top item gets the whole hold.
    const f3 = view([item("a", "opp"), item("b", "opp")]);
    seen = noteStackSeen(seen, f3, 2900);
    expect(remaining(f3, seen, 2900)).toBe(2000);
    // It resolves; the item under it was up long enough already.
    const f4 = view([item("a", "opp")]);
    seen = noteStackSeen(seen, f4, 5000);
    expect(remaining(f4, seen, 5000)).toBe(0);
  });
});

describe("combinedPassDelayMs", () => {
  it("is the longer of a timed bluff and the hold, not their sum", () => {
    expect(combinedPassDelayMs(1500, 2000)).toBe(2000);
    expect(combinedPassDelayMs(3500, 2000)).toBe(3500);
    expect(combinedPassDelayMs(1500, 0)).toBe(1500);
    expect(combinedPassDelayMs(0, 0)).toBe(0);
  });
});

describe("stackHoldStatus", () => {
  beforeEach(() => _resetForTests());

  it("renders the countdown in tenths", () => {
    expect(stackHoldStatusText({ passesAt: 2400 }, 1000)).toBe("auto-pass in 1.4 s");
    expect(stackHoldStatusText({ passesAt: 1000 }, 1500)).toBe("auto-pass in 0.0 s");
    expect(stackHoldStatusText(null, 0)).toBe("");
  });

  it("is a store the dock reads", () => {
    expect(get(stackHoldStatus)).toBeNull();
    setStackHoldStatus({ passesAt: 5 });
    expect(get(stackHoldStatus)).toEqual({ passesAt: 5 });
    setStackHoldStatus(null);
    expect(get(stackHoldStatus)).toBeNull();
  });
});
