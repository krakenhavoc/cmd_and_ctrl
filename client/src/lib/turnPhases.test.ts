// turnPhases.test.ts — ADR 0188 (#3567): the repeated-phase captions,
// the added-phase markers, the "Next:" line and the skip chip.

import { describe, it, expect } from "vitest";

import { addedPhaseMarkers, nextTurnLine, phaseCaption, turnSkipChip } from "./turn";

describe("phaseCaption", () => {
  it("names the nth combat", () => {
    expect(phaseCaption("combat", "Combat", 1)).toBe("Combat");
    expect(phaseCaption("combat", "Combat", 2)).toBe("Combat 2");
    expect(phaseCaption("combat", "Combat", 3)).toBe("Combat 3");
  });
  it("counts main phases as one family", () => {
    expect(phaseCaption("main", "Main 2", 1)).toBe("Main 1");
    expect(phaseCaption("main", "Main 2", 2)).toBe("Main 2");
    expect(phaseCaption("main", "Main 2", 3)).toBe("Main 3");
  });
  it("keeps the printed caption without an ordinal", () => {
    expect(phaseCaption("combat", "Combat", undefined)).toBe("Combat");
    expect(phaseCaption("main", "Main 2", undefined)).toBe("Main 2");
    expect(phaseCaption("other", "End", 2)).toBe("End");
  });
});

const combatSteps = (id: number) =>
  ["begin_combat", "declare_attackers", "end_combat"].map((step) => ({ step, phase_id: id }));

describe("addedPhaseMarkers", () => {
  it("marks two added combats and a main, in order", () => {
    const upcoming = [
      { step: "postcombat_main", phase_id: 4 },
      ...combatSteps(6),
      { step: "postcombat_main", phase_id: 7 },
      ...combatSteps(8),
      { step: "end", phase_id: 5 },
    ];
    expect(addedPhaseMarkers(upcoming, 2)).toEqual(["+ Combat", "+ Main", "+ Combat"]);
  });
  it("shows no marker without added phases", () => {
    expect(addedPhaseMarkers([...combatSteps(3), { step: "end", phase_id: 5 }], 2)).toEqual([]);
    expect(addedPhaseMarkers(undefined, undefined)).toEqual([]);
  });
  it("does not mark the added phase already in progress", () => {
    expect(addedPhaseMarkers(combatSteps(6), 6)).toEqual([]);
  });
});

describe("nextTurnLine", () => {
  const seats = [
    { seat: 0, name: "Alice" },
    { seat: 1, name: "Bob", turn_skips: 1 },
    { seat: 2, name: "Cy" },
  ];
  it("names a queued extra turn", () => {
    expect(nextTurnLine({ active_seat: 0, next_seat: 0, extra_turns: [0] }, seats)).toBe(
      "Next: Alice (extra)",
    );
  });
  it("names the seat after a skip", () => {
    expect(nextTurnLine({ active_seat: 0, next_seat: 2 }, seats)).toBe("Next: Cy (skips Bob)");
  });
  it("has no line in the plain case", () => {
    expect(nextTurnLine({ active_seat: 1, next_seat: 2 }, seats)).toBeNull();
    expect(nextTurnLine({ active_seat: 0 }, seats)).toBeNull();
  });
  it("does not call a departed seat a skip", () => {
    const left = [
      { seat: 0, name: "Alice" },
      { seat: 1, name: "Bob" },
      { seat: 2, name: "Cy" },
    ];
    expect(nextTurnLine({ active_seat: 0, next_seat: 2 }, left)).toBeNull();
  });
});

describe("turnSkipChip", () => {
  it("reads one and many skips", () => {
    expect(turnSkipChip(1)).toBe("Skips next turn");
    expect(turnSkipChip(2)).toBe("Skips next 2 turns");
  });
  it("is absent at zero", () => {
    expect(turnSkipChip(0)).toBeNull();
    expect(turnSkipChip(undefined)).toBeNull();
  });
});
