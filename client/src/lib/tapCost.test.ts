import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import {
  applyCastChoices,
  begin,
  cancel,
  hasXCost,
  tapCostLimit,
  tapCostOf,
  targeting,
} from "./targeting";
import type { CardView } from "./protocol";

function card(extras: Partial<CardView> = {}): CardView {
  return { instance_id: "spell", name: "Spell", owner: "p0", controller: "p0", ...extras };
}

// The Wandering Rescuer: convoke, five symbols, colour clause on.
const rescuer = card({
  instance_id: "rescuer",
  name: "The Wandering Rescuer",
  type_line: "Legendary Creature — Human Samurai Noble",
  mana_cost: "{3}{W}{W}",
  tap_cost: {
    key: "convoke",
    label: "Convoke",
    color_clause: true,
    max: 5,
    options: { cards: ["soldier-a", "soldier-b"] },
  },
});

// Waterbender's Restoration: waterbend {X}, no colour clause, and a
// target count that is the announced X. The server can't size either
// when it builds the snapshot, so both arrive as "ask X first".
const restoration = card({
  instance_id: "restoration",
  name: "Waterbender's Restoration",
  type_line: "Instant — Lesson",
  mana_cost: "{U}{U}",
  target_mode: "creature",
  legal_targets: { cards: ["bear-a", "bear-b", "bear-c"], min: 0, max: 0, count_from_x: true },
  tap_cost: {
    key: "waterbend",
    label: "Waterbend {X}",
    demands_x: true,
    options: { cards: ["bear-a", "bear-b", "bear-c"] },
  },
});

describe("tapCostOf", () => {
  it("is undefined for the vast majority of cards", () => {
    expect(tapCostOf(card())).toBeUndefined();
  });

  it("reads the clause off a card that has one", () => {
    expect(tapCostOf(rescuer)?.key).toBe("convoke");
    expect(tapCostOf(restoration)?.label).toBe("Waterbend {X}");
  });
});

describe("tapCostLimit", () => {
  it("uses the server's cap when it sent one", () => {
    expect(tapCostLimit(tapCostOf(rescuer)!, undefined)).toBe(5);
    // A convoke cap does not move with an X the caster typed.
    expect(tapCostLimit(tapCostOf(rescuer)!, 9)).toBe(5);
  });

  it("falls back to the announced X for a waterbend {X}", () => {
    expect(tapCostLimit(tapCostOf(restoration)!, 3)).toBe(3);
  });

  it("caps at nothing when no X has been announced", () => {
    expect(tapCostLimit(tapCostOf(restoration)!, undefined)).toBe(0);
  });
});

describe("hasXCost", () => {
  it("still fires on an {X} in the printed cost", () => {
    expect(hasXCost(card({ mana_cost: "{X}{B}{B}" }))).toBe(true);
  });

  it("fires for a waterbend {X} on a card whose printed cost has none", () => {
    expect(restoration.mana_cost).not.toContain("{X}");
    expect(hasXCost(restoration)).toBe(true);
  });

  it("does not fire for convoke, which adds no cost of its own", () => {
    expect(hasXCost(rescuer)).toBe(false);
  });
});

describe("applyCastChoices", () => {
  it("sends the tapped permanents as tap_ids", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { tapIDs: ["a", "b"] });
    expect(params.tap_ids).toEqual(["a", "b"]);
  });

  it("omits the field when nothing was tapped — tapping none is legal", () => {
    const params: Record<string, unknown> = {};
    applyCastChoices(params, { tapIDs: [] });
    expect(params).not.toHaveProperty("tap_ids");
  });
});

describe("a clause counted by X", () => {
  it("takes its target count from the announced X, not from min / max", () => {
    cancel();
    begin(restoration, "creature", { xValue: 2, tapIDs: ["bear-c"] });
    const state = get(targeting);
    expect(state?.min).toBe(2);
    expect(state?.max).toBe(2);
    cancel();
  });

  it("collapses to no targets at X=0 rather than reading max 0 as unbounded", () => {
    cancel();
    begin(restoration, "creature", { xValue: 0 });
    const state = get(targeting);
    expect(state?.min).toBe(0);
    expect(state?.max).toBe(0);
    cancel();
  });

  it("leaves an ordinary clause's printed count alone", () => {
    cancel();
    begin(card({ legal_targets: { cards: ["a"], min: 1, max: 3 } }), "creature", { xValue: 7 });
    const state = get(targeting);
    expect(state?.min).toBe(1);
    expect(state?.max).toBe(3);
    cancel();
  });
});
