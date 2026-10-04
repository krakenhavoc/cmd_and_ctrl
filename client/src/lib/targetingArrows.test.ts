// ADR 0119 §4 — the arrows drawn while the viewer chooses targets, as
// pure functions: which picks get an arrow, where the arrows start
// (the source, or the dock when the source is not on screen), and the
// follow arrow, which only a pointer that hovers gets.

import { describe, it, expect } from "vitest";

import {
  FOLLOW_ARROW_COLOR,
  FOLLOW_SNAP_COLOR,
  PICK_ARROW_COLOR,
  followsPointer,
  pickOpen,
  planPickArrows,
  planTargetingArrows,
  type PickZones,
} from "./targetingArrows";
import type { Box } from "./stackArrows";
import type { TargetingState } from "./targeting";

const zones: PickZones = {
  battlefield: new Set(["birds", "vivi"]),
  stack: new Set(["bolt"]),
  players: new Set(["me", "opp"]),
};

const board = { width: 1400, height: 900 };
const box = (left: number, top: number, width = 100, height = 140): Box => ({
  left,
  top,
  width,
  height,
});

function prompt(over: Partial<TargetingState> = {}): TargetingState {
  return {
    card: { instance_id: "shock", name: "Shock" },
    mode: "any",
    min: 1,
    max: 1,
    picked: [],
    steps: [],
    step: 0,
    done: [],
    ...over,
  } as TargetingState;
}

describe("planPickArrows", () => {
  it("gives a player, a permanent and a stack item one arrow each, in pick order", () => {
    const plans = planPickArrows(
      [
        { kind: "player", id: "opp" },
        { kind: "card", id: "birds" },
        { kind: "card", id: "bolt" },
      ],
      zones,
    );
    expect(plans.map((p) => [p.kind, p.targetID])).toEqual([
      ["player", "opp"],
      ["permanent", "birds"],
      ["stack", "bolt"],
    ]);
  });

  it("gives a graveyard or exile pick no arrow", () => {
    // "gy-card" is in no zone the table draws: a graveyard or an exiled card.
    expect(planPickArrows([{ kind: "card", id: "gy-card" }], zones)).toEqual([]);
  });

  it("draws one arrow for a target two clauses both picked", () => {
    const plans = planPickArrows(
      [
        { kind: "card", id: "vivi", slot: 0 },
        { kind: "card", id: "vivi", slot: 1 },
      ],
      zones,
    );
    expect(plans).toHaveLength(1);
  });

  it("skips a player who is not at the table", () => {
    expect(planPickArrows([{ kind: "player", id: "ghost" }], zones)).toEqual([]);
  });
});

describe("pickOpen", () => {
  it("is open until a counted clause is full", () => {
    expect(pickOpen(prompt({ max: 2, picked: [{ kind: "card", id: "birds" }] }))).toBe(true);
    expect(
      pickOpen(
        prompt({
          max: 2,
          picked: [
            { kind: "card", id: "birds" },
            { kind: "card", id: "vivi" },
          ],
        }),
      ),
    ).toBe(false);
  });

  it("is always open for an unbounded clause", () => {
    expect(
      pickOpen(
        prompt({
          max: 0,
          picked: [
            { kind: "card", id: "birds" },
            { kind: "card", id: "vivi" },
          ],
        }),
      ),
    ).toBe(true);
  });

  it("closes a divided clause once every point has a target", () => {
    expect(pickOpen(prompt({ max: 0, divide: 1, picked: [{ kind: "card", id: "birds" }] }))).toBe(
      false,
    );
  });
});

describe("followsPointer", () => {
  it("is a mouse or a pen, never touch or a keyboard", () => {
    expect(followsPointer("mouse")).toBe(true);
    expect(followsPointer("pen")).toBe(true);
    expect(followsPointer("touch")).toBe(false);
    expect(followsPointer("")).toBe(false);
    expect(followsPointer(null)).toBe(false);
  });
});

describe("planTargetingArrows", () => {
  const birds = { plan: { id: "permanent:birds", kind: "permanent" as const, targetID: "birds" } };
  const opp = { plan: { id: "player:opp", kind: "player" as const, targetID: "opp" } };

  it("starts every arrow at the source when it is on screen", () => {
    const source = box(600, 700);
    const plan = planTargetingArrows({
      board,
      source,
      dock: box(1200, 820, 180, 40),
      picks: [
        { ...birds, box: box(300, 100) },
        { ...opp, box: box(900, 20, 160, 40) },
      ],
      follow: null,
    });
    expect(plan.from).toBe("source");
    expect(plan.arrows.map((a) => a.id)).toEqual(["permanent:birds", "player:opp"]);
    for (const a of plan.arrows) {
      expect(a.kind).toBe("pick");
      expect(a.color).toBe(PICK_ARROW_COLOR);
      // On the source's top edge, which faces both targets.
      expect(a.y1).toBe(700);
      expect(a.x1).toBeGreaterThanOrEqual(600);
      expect(a.x1).toBeLessThanOrEqual(700);
    }
  });

  it("starts from the dock's question line when the source is not on screen", () => {
    const plan = planTargetingArrows({
      board,
      source: null,
      dock: box(1200, 820, 180, 40),
      picks: [{ ...birds, box: box(300, 100) }],
      follow: null,
    });
    expect(plan.from).toBe("dock");
    expect(plan.arrows).toHaveLength(1);
    const a = plan.arrows[0];
    expect(a.x1).toBeGreaterThanOrEqual(1200);
    expect(a.y1).toBeGreaterThanOrEqual(820);
  });

  it("draws nothing with neither a source nor a dock to start from", () => {
    const plan = planTargetingArrows({
      board,
      source: null,
      dock: null,
      picks: [{ ...birds, box: box(300, 100) }],
      follow: { pointer: { x: 50, y: 50 }, snap: null },
    });
    expect(plan).toEqual({ from: null, arrows: [] });
  });

  it("gives a pick that is not on screen, or is the source itself, no arrow", () => {
    const plan = planTargetingArrows({
      board,
      source: box(600, 700),
      dock: null,
      picks: [
        { ...birds, box: null },
        { ...opp, box: box(600, 700), isSource: true },
      ],
      follow: null,
    });
    expect(plan.arrows).toEqual([]);
  });

  it("ends the follow arrow at the pointer, gold, and snaps it green to a legal target", () => {
    const free = planTargetingArrows({
      board,
      source: box(600, 700),
      dock: null,
      picks: [],
      follow: { pointer: { x: 400, y: 300 }, snap: null },
    });
    const f = free.arrows.find((a) => a.kind === "follow")!;
    expect(f.snapped).toBe(false);
    expect(f.color).toBe(FOLLOW_ARROW_COLOR);
    // A gap short of the pointer, so the head never sits under it.
    expect(Math.hypot(f.x2 - 400, f.y2 - 300)).toBeCloseTo(6, 5);

    const snapped = planTargetingArrows({
      board,
      source: box(600, 700),
      dock: null,
      picks: [],
      follow: { pointer: { x: 340, y: 160 }, snap: box(300, 100) },
    });
    const s = snapped.arrows.find((a) => a.kind === "follow")!;
    expect(s.snapped).toBe(true);
    expect(s.color).toBe(FOLLOW_SNAP_COLOR);
    // On the target's edge, not at the pointer inside it.
    expect(s.y2).toBeGreaterThan(240);
  });

  it("draws no follow arrow while the pointer is over the source", () => {
    const plan = planTargetingArrows({
      board,
      source: box(600, 700),
      dock: null,
      picks: [],
      follow: { pointer: { x: 650, y: 760 }, snap: null },
    });
    expect(plan.arrows).toEqual([]);
  });

  it("puts the follow arrow after the picks", () => {
    const plan = planTargetingArrows({
      board,
      source: box(600, 700),
      dock: null,
      picks: [{ ...birds, box: box(300, 100) }],
      follow: { pointer: { x: 1000, y: 300 }, snap: null },
    });
    expect(plan.arrows.map((a) => a.kind)).toEqual(["pick", "follow"]);
  });
});
