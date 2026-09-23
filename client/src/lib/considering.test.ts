import { describe, it, expect } from "vitest";

import { isResponseWindowFor, responseWindowKey } from "./considering";
import type { GameView, PendingChoiceView, PlayerView, TurnView, ZoneView } from "./protocol";

// Fixture helpers mirror the shape used in priority.test.ts.

function emptyZone(kind: string, owner = ""): ZoneView {
  return { kind, owner, count: 0, cards: [] };
}

function turn(opts: Partial<TurnView> = {}): TurnView {
  return {
    number: 1,
    active_seat: 0,
    priority_holder: 0,
    phase: "precombat_main",
    step: "precombat_main",
    ...opts,
  };
}

function seat(id: string, idx: number): PlayerView {
  return {
    id,
    name: `seat ${idx}`,
    seat: idx,
    life: 40,
    library: emptyZone("library", id),
    hand: emptyZone("hand", id),
    graveyard: emptyZone("graveyard", id),
    command: emptyZone("command", id),
    commander_damage: {},
    life_history: [],
  };
}

function pendingChoice(): PendingChoiceView {
  return { id: "ch-1", kind: "confirm", chooser: "p0", from_player: "p0", count: 1 };
}

interface SnapOpts {
  step?: string;
  turnNumber?: number;
  activeSeat?: number;
  priorityHolder?: number;
  stackDepth?: number;
  blockDecisionSeats?: number[];
  pendingChoices?: PendingChoiceView[];
}

function snap(o: SnapOpts = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat("p0", 0), seat("p1", 1)],
    battlefield: emptyZone("battlefield"),
    stack: { ...emptyZone("stack"), cards: [] },
    exile: emptyZone("exile"),
    turn: turn({
      number: o.turnNumber ?? 1,
      step: o.step ?? "precombat_main",
      active_seat: o.activeSeat ?? 0,
      priority_holder: o.priorityHolder ?? 0,
      block_decision_seats: o.blockDecisionSeats,
    }),
    mulligans_open: false,
    // stack_items carries the count for a non-empty stack, exactly as
    // Board's own reads of it do.
    stack_items: o.stackDepth
      ? Array.from({ length: o.stackDepth }, (_, i) => ({
          id: `si-${i}`,
          kind: "spell" as const,
          controller: "p0",
          owner: "p0",
          source_card_id: `c-${i}`,
        }))
      : [],
    pending_choices: o.pendingChoices,
  };
}

describe("isResponseWindowFor", () => {
  it("returns false for a null view", () => {
    expect(isResponseWindowFor(null, 0)).toBe(false);
  });

  it("is true when the stack is non-empty and seat holds priority", () => {
    expect(isResponseWindowFor(snap({ stackDepth: 1, priorityHolder: 0 }), 0)).toBe(true);
  });

  it("is false with an empty stack when the holder is the active seat", () => {
    expect(isResponseWindowFor(snap({ activeSeat: 0, priorityHolder: 0 }), 0)).toBe(false);
  });

  it("is true with an empty stack when the holder isn't the active seat", () => {
    expect(isResponseWindowFor(snap({ activeSeat: 0, priorityHolder: 1 }), 1)).toBe(true);
  });

  it("is false when a pending choice is open, even on a non-empty stack", () => {
    expect(
      isResponseWindowFor(
        snap({ stackDepth: 1, priorityHolder: 0, pendingChoices: [pendingChoice()] }),
        0,
      ),
    ).toBe(false);
  });

  it("is false when block_decision_seats is non-empty", () => {
    expect(
      isResponseWindowFor(snap({ stackDepth: 1, priorityHolder: 0, blockDecisionSeats: [0] }), 0),
    ).toBe(false);
  });

  it("is false when the queried seat doesn't hold priority", () => {
    expect(isResponseWindowFor(snap({ stackDepth: 1, priorityHolder: 0 }), 1)).toBe(false);
  });
});

describe("responseWindowKey", () => {
  it("is empty for a null view", () => {
    expect(responseWindowKey(null)).toBe("");
  });

  it("changes when the turn number changes", () => {
    const a = responseWindowKey(snap({ turnNumber: 1 }));
    const b = responseWindowKey(snap({ turnNumber: 2 }));
    expect(a).not.toBe(b);
  });

  it("changes when the step changes", () => {
    const a = responseWindowKey(snap({ step: "precombat_main" }));
    const b = responseWindowKey(snap({ step: "declare_attackers" }));
    expect(a).not.toBe(b);
  });

  it("changes when the priority holder changes", () => {
    const a = responseWindowKey(snap({ priorityHolder: 0 }));
    const b = responseWindowKey(snap({ priorityHolder: 1 }));
    expect(a).not.toBe(b);
  });

  it("changes when the stack depth changes", () => {
    const a = responseWindowKey(snap({ stackDepth: 0 }));
    const b = responseWindowKey(snap({ stackDepth: 1 }));
    expect(a).not.toBe(b);
  });

  it("is stable across snapshots with no relevant change", () => {
    const a = responseWindowKey(snap({ stackDepth: 1, priorityHolder: 0 }));
    const b = responseWindowKey(snap({ stackDepth: 1, priorityHolder: 0 }));
    expect(a).toBe(b);
  });
});
