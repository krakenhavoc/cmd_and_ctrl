import { describe, it, expect, beforeEach } from "vitest";

import { hasAnyLegalResponse, owesBlockDecision, _resetCacheForTests } from "./priority";
import type { CardView, GameView, PlayerView, StackItemView, TurnView, ZoneView } from "./protocol";

// Fixture helpers mirror the shape used in timing.test.ts. Kept
// local rather than exported because they'd otherwise ship in the
// prod bundle.
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

function card(name: string, type: string, extras: Partial<CardView> = {}): CardView {
  return {
    instance_id: `c-${name}`,
    name,
    owner: "p0",
    controller: "p0",
    type_line: type,
    ...extras,
  };
}

interface SnapOpts {
  step?: string;
  activeSeat?: number;
  priorityHolder?: number;
  splitSecond?: boolean;
  stackItems?: StackItemView[];
  battlefield?: CardView[];
  hand?: CardView[];
  command?: CardView[];
}

function snap(o: SnapOpts = {}): GameView {
  const seats = [seat("p0", 0), seat("p1", 1)];
  if (o.hand) seats[0].hand = { ...emptyZone("hand", "p0"), cards: o.hand, count: o.hand.length };
  if (o.command)
    seats[0].command = {
      ...emptyZone("command", "p0"),
      cards: o.command,
      count: o.command.length,
    };
  return {
    id: "g",
    state: "active",
    seats,
    battlefield: { ...emptyZone("battlefield"), cards: o.battlefield ?? [] },
    stack: emptyZone("stack"),
    exile: emptyZone("exile"),
    turn: turn({
      step: o.step ?? "precombat_main",
      active_seat: o.activeSeat ?? 0,
      priority_holder: o.priorityHolder ?? 0,
    }),
    mulligans_open: false,
    stack_items: o.stackItems ?? [],
    split_second_active: o.splitSecond ?? false,
  };
}

describe("hasAnyLegalResponse", () => {
  beforeEach(() => _resetCacheForTests());

  it("returns false for null snap or viewer", () => {
    expect(hasAnyLegalResponse(null, "p0")).toBe(false);
    expect(hasAnyLegalResponse(snap(), null)).toBe(false);
  });

  it("returns false when viewer does not hold priority", () => {
    const s = snap({ priorityHolder: 1, hand: [card("bolt", "Instant")] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(false);
  });

  it("returns false during no-priority steps (Untap / Cleanup sentinel)", () => {
    const s = snap({ step: "untap", priorityHolder: -1, hand: [card("bolt", "Instant")] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(false);
  });

  it("returns true when the viewer holds priority with a castable instant in hand", () => {
    const s = snap({ hand: [card("bolt", "Instant")] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(true);
  });

  it("returns true when the viewer holds priority on their main phase with a sorcery", () => {
    const s = snap({ hand: [card("wrath", "Sorcery")] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(true);
  });

  it("returns false for a sorcery on an opponent's turn (viewer holds priority)", () => {
    const s = snap({
      activeSeat: 1,
      priorityHolder: 0,
      step: "upkeep",
      hand: [card("wrath", "Sorcery")],
    });
    expect(hasAnyLegalResponse(s, "p0")).toBe(false);
  });

  it("returns true for an instant on an opponent's turn (any priority window)", () => {
    const s = snap({
      activeSeat: 1,
      priorityHolder: 0,
      step: "upkeep",
      hand: [card("bolt", "Instant")],
    });
    expect(hasAnyLegalResponse(s, "p0")).toBe(true);
  });

  it("returns false while split-second is active, even with instants in hand", () => {
    const s = snap({ hand: [card("bolt", "Instant")], splitSecond: true });
    expect(hasAnyLegalResponse(s, "p0")).toBe(false);
  });

  it("returns true when the viewer controls a battlefield card (potential activation)", () => {
    const s = snap({ battlefield: [card("bird", "Creature", { controller: "p0" })] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(true);
  });

  it("ignores battlefield cards the viewer does not control", () => {
    const s = snap({ battlefield: [card("enemy", "Creature", { controller: "p1" })] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(false);
  });

  it("checks the command zone for castable commanders", () => {
    const s = snap({ command: [card("cmdr", "Legendary Creature")] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(true);
  });

  it("returns false when hand + command are empty and battlefield has no viewer-controlled cards", () => {
    expect(hasAnyLegalResponse(snap(), "p0")).toBe(false);
  });

  it("memoises within a snapshot seq so repeat calls don't re-scan", () => {
    // A large battlefield would normally scale linearly; the cache
    // short-circuits that on the second call. We can't directly
    // observe the cache from the outside, but we can at least prove
    // the behaviour is stable across repeated calls.
    const s = snap({
      battlefield: Array.from({ length: 50 }, (_, i) =>
        card(`c${i}`, "Creature", { controller: "p0", instance_id: `c-${i}` }),
      ),
    });
    const first = hasAnyLegalResponse(s, "p0", 100);
    const second = hasAnyLegalResponse(s, "p0", 100);
    expect(first).toBe(true);
    expect(second).toBe(true);
  });

  it("resets the memo cache when snap seq advances", () => {
    const s1 = snap({ hand: [card("bolt", "Instant")] });
    expect(hasAnyLegalResponse(s1, "p0", 1)).toBe(true);
    // Different snap, same viewer, new seq — cache clears and the
    // answer re-computes against the new state.
    const s2 = snap({ hand: [] });
    expect(hasAnyLegalResponse(s2, "p0", 2)).toBe(false);
  });
});

// #328 — the defending player's declare-blockers window. Blocking is
// a turn-based action, not a priority response, so it was invisible
// to every auto-pass gate: the player in the bug report had their one
// chance to block passed for them and took eight unblocked damage
// with an untapped creature on the table.
describe("owesBlockDecision", () => {
  beforeEach(() => _resetCacheForTests());

  // blockSnap builds a declare-blockers snapshot on the opponent's
  // turn where the viewer (p0, seat 0) holds priority — the exact
  // shape of replay seq 307 in the bug report.
  function blockSnap(seats: number[] | undefined): GameView {
    const s = snap({
      step: "declare_blockers",
      activeSeat: 1,
      priorityHolder: 0,
    });
    s.turn.block_decision_seats = seats;
    return s;
  }

  it("is true when the server says the viewer's seat owes a decision", () => {
    expect(owesBlockDecision(blockSnap([0]), "p0")).toBe(true);
  });

  it("is false when only another seat owes a decision", () => {
    expect(owesBlockDecision(blockSnap([1]), "p0")).toBe(false);
  });

  it("is false when the field is absent or empty", () => {
    expect(owesBlockDecision(blockSnap(undefined), "p0")).toBe(false);
    expect(owesBlockDecision(blockSnap([]), "p0")).toBe(false);
  });

  it("is false for a null snap, a null viewer, or an unseated viewer", () => {
    expect(owesBlockDecision(null, "p0")).toBe(false);
    expect(owesBlockDecision(blockSnap([0]), null)).toBe(false);
    expect(owesBlockDecision(blockSnap([0]), "nobody")).toBe(false);
  });

  it("does not require the viewer to hold priority", () => {
    // The active player holds priority first on entering the step.
    // The defender still owes the decision, and the guard must hold
    // when priority reaches them a moment later.
    const s = blockSnap([0]);
    s.turn.priority_holder = 1;
    expect(owesBlockDecision(s, "p0")).toBe(true);
  });
});

describe("hasAnyLegalResponse — #328 blocking window", () => {
  beforeEach(() => _resetCacheForTests());

  it("is true when the viewer owes a block decision and has nothing else to do", () => {
    // Empty hand, empty command, and not one battlefield card the
    // viewer controls — every pre-existing branch says "nothing to
    // do", which is precisely how smart-skip would eat the window.
    const s = snap({
      step: "declare_blockers",
      activeSeat: 1,
      priorityHolder: 0,
      battlefield: [card("enemy", "Creature", { controller: "p1" })],
    });
    expect(hasAnyLegalResponse(s, "p0", 500)).toBe(false);
    s.turn.block_decision_seats = [0];
    expect(hasAnyLegalResponse(s, "p0", 501)).toBe(true);
  });

  it("still passes a declare-blockers window the viewer cannot block in", () => {
    const s = snap({
      step: "declare_blockers",
      activeSeat: 1,
      priorityHolder: 0,
      battlefield: [card("enemy", "Creature", { controller: "p1" })],
    });
    s.turn.block_decision_seats = [1];
    expect(hasAnyLegalResponse(s, "p0", 502)).toBe(false);
  });
});
