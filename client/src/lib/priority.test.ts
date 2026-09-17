import { describe, it, expect, beforeEach } from "vitest";

import {
  autopassSuspended,
  hasAnyLegalResponse,
  hasDeclaredAttackers,
  loopNoticeText,
  owesBlockDecision,
  _resetCacheForTests,
} from "./priority";
import type {
  CardView,
  GameView,
  LegalMoveView,
  PlayerView,
  StackItemView,
  TurnView,
  ZoneView,
} from "./protocol";

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
  // S31: the seat's enumerated move list, which is now the whole
  // input to hasAnyLegalResponse. Absent means the server said
  // nothing — a distinct state from "enumerated, nothing to do".
  moves?: LegalMoveView[];
}

// pass / play are the two move shapes these tests care about: a list
// holding only `pass` means "yield is all you can do", and anything
// else means the window is live.
const pass: LegalMoveView = {
  type: "pass_priority",
  player: "p0",
  kind: "pass",
  label: "Pass priority",
};

function play(kind: LegalMoveView["kind"], source = "c-something"): LegalMoveView {
  return { type: "cast_spell", player: "p0", kind, label: `${kind} ${source}`, source };
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
    legal_moves: o.moves,
  };
}

describe("hasAnyLegalResponse", () => {
  beforeEach(() => _resetCacheForTests());

  it("returns false for null snap or viewer", () => {
    expect(hasAnyLegalResponse(null, "p0")).toBe(false);
    expect(hasAnyLegalResponse(snap(), null)).toBe(false);
  });

  it("returns false when viewer does not hold priority", () => {
    const s = snap({ priorityHolder: 1, moves: [pass, play("cast", "c-bolt")] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(false);
  });

  it("returns false during no-priority steps (Untap / Cleanup sentinel)", () => {
    const s = snap({ step: "untap", priorityHolder: -1, moves: [pass, play("cast", "c-bolt")] });
    expect(hasAnyLegalResponse(s, "p0")).toBe(false);
  });

  it("is true for any non-pass move the server enumerated", () => {
    for (const kind of ["land", "cast", "activate", "mana", "attack", "block", "choice"] as const) {
      const s = snap({ moves: [pass, play(kind)] });
      expect(hasAnyLegalResponse(s, "p0"), `kind=${kind}`).toBe(true);
    }
  });

  it("is false when yielding is the only move", () => {
    expect(hasAnyLegalResponse(snap({ moves: [pass] }), "p0")).toBe(false);
  });

  it("is false when the server enumerated nothing at all", () => {
    expect(hasAnyLegalResponse(snap({ moves: [] }), "p0")).toBe(false);
  });

  // The affordability case the old card-walk could not see. It read
  // type lines and said "you hold an instant, therefore you have a
  // response", with no idea whether the mana was there. The
  // enumerator pays for what it offers, so an unaffordable hand is
  // now correctly a skippable window.
  it("is false for a hand it cannot pay for, which the old card walk called a response", () => {
    const s = snap({
      hand: [card("bolt", "Instant"), card("wrath", "Sorcery")],
      moves: [pass],
    });
    expect(hasAnyLegalResponse(s, "p0")).toBe(false);
  });

  // Compatibility: no move list on a frame where the viewer holds
  // priority means a pre-S31 server, or a field we dropped. Stopping
  // costs one click; skipping costs the player a window.
  it("errs toward stopping when the server shipped no move list", () => {
    expect(hasAnyLegalResponse(snap(), "p0")).toBe(true);
    expect(hasAnyLegalResponse(snap({ priorityHolder: 1 }), "p0")).toBe(false);
  });

  it("is stable across repeat calls and across seq changes", () => {
    // The per-seq memo is gone — the answer is one array scan of a
    // field the server computed — so the parameter is inert and the
    // answer tracks the snapshot, never a stale cache entry.
    const live = snap({ moves: [pass, play("cast", "c-bolt")] });
    expect(hasAnyLegalResponse(live, "p0", 1)).toBe(true);
    expect(hasAnyLegalResponse(live, "p0", 1)).toBe(true);
    const quiet = snap({ moves: [pass] });
    expect(hasAnyLegalResponse(quiet, "p0", 1)).toBe(false);
    expect(hasAnyLegalResponse(quiet, "p0", 2)).toBe(false);
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
    // The server enumerated and found nothing but pass — every
    // priority-shaped branch says "nothing to do", which is precisely
    // how smart-skip would eat the window.
    const s = snap({
      step: "declare_blockers",
      activeSeat: 1,
      priorityHolder: 0,
      battlefield: [card("enemy", "Creature", { controller: "p1" })],
      moves: [pass],
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
      moves: [pass],
    });
    s.turn.block_decision_seats = [1];
    expect(hasAnyLegalResponse(s, "p0", 502)).toBe(false);
  });

  // The two signals must not be able to disagree: a `block` move in
  // the list and `block_decision_seats` come off the same server-side
  // eligibility test (#328), and either one alone is enough to hold
  // the window open. The defender may be reading a frame where the
  // active player still holds priority, which is why the seat list
  // exists at all.
  it("holds the window on a block move even before the seat list arrives", () => {
    const s = snap({
      step: "declare_blockers",
      activeSeat: 1,
      priorityHolder: 0,
      moves: [pass, play("block", "c-my-bear")],
    });
    expect(hasAnyLegalResponse(s, "p0", 503)).toBe(true);
  });
});

// #599 — the attacking half of the #328 guard. A bulk declaration
// taps every attacker it declares, so the enumerator has nothing left
// to offer and smart auto-pass would close the window the undo lives
// in.
describe("hasAnyLegalResponse — #599 declare-attackers review window", () => {
  beforeEach(() => _resetCacheForTests());

  const attacker = (name: string, extras = {}) =>
    card(name, "Creature — Bear", { controller: "p0", ...extras });

  it("holds the window once the viewer has declared an attacker", () => {
    const s = snap({
      step: "declare_attackers",
      activeSeat: 0,
      priorityHolder: 0,
      battlefield: [attacker("declared", { attacking_target: "p1", tapped: true })],
      moves: [pass],
    });
    expect(hasAnyLegalResponse(s, "p0", 600)).toBe(true);
  });

  it("passes a declare-attackers window with nothing declared", () => {
    // Nothing swung, nothing to review: the seat with no attack move
    // left is a seat with nothing to do, which is what smart-skip is
    // for.
    const s = snap({
      step: "declare_attackers",
      activeSeat: 0,
      priorityHolder: 0,
      battlefield: [attacker("idle", { tapped: true })],
      moves: [pass],
    });
    expect(hasAnyLegalResponse(s, "p0", 601)).toBe(false);
  });

  it("does not hold the window for a seat that is not attacking", () => {
    // The defender sees the attack; the review window belongs to the
    // player who declared it.
    const s = snap({
      step: "declare_attackers",
      activeSeat: 1,
      priorityHolder: 0,
      battlefield: [
        card("theirs", "Creature — Bear", { controller: "p1", attacking_target: "p0" }),
      ],
      moves: [pass],
    });
    expect(hasAnyLegalResponse(s, "p0", 602)).toBe(false);
  });
});

describe("hasDeclaredAttackers", () => {
  it("is false outside declare_attackers, even with attackers on the board", () => {
    const s = snap({
      step: "declare_blockers",
      activeSeat: 0,
      priorityHolder: 0,
      battlefield: [card("mine", "Creature — Bear", { controller: "p0", attacking_target: "p1" })],
    });
    expect(hasDeclaredAttackers(s, "p0")).toBe(false);
  });

  it("is false for a null snapshot or viewer", () => {
    expect(hasDeclaredAttackers(null, "p0")).toBe(false);
    expect(hasDeclaredAttackers(snap({ step: "declare_attackers" }), null)).toBe(false);
  });
});

// #628 (CR 726) — the loop breaker's client half. The autopass gate
// and the banner both read these two, so they are the only thing
// between a server that has spotted a loop and a browser that would
// otherwise keep feeding it.
describe("autopassSuspended", () => {
  it("is false on a quiet table", () => {
    expect(autopassSuspended(snap())).toBe(false);
  });

  it("is true whenever the server ships a loop notice", () => {
    const s = snap();
    s.loop_notice = { label: "Mirror Engine — create a Spark", count: 25 };
    expect(autopassSuspended(s)).toBe(true);
  });

  it("is false for a missing snapshot rather than throwing", () => {
    expect(autopassSuspended(null)).toBe(false);
    expect(autopassSuspended(undefined)).toBe(false);
  });
});

describe("loopNoticeText", () => {
  it("is empty when nothing is suspected", () => {
    expect(loopNoticeText(snap())).toBe("");
    expect(loopNoticeText(null)).toBe("");
  });

  it("reads the label straight through, since it already names card and ability", () => {
    const s = snap();
    s.loop_notice = { label: "Mirror Engine — create a Spark", count: 25 };
    expect(loopNoticeText(s)).toBe(
      "Mirror Engine — create a Spark has resolved 25 times this turn.",
    );
  });

  it("still says something useful when the ability has no label", () => {
    const s = snap();
    s.loop_notice = { label: "", count: 40 };
    expect(loopNoticeText(s)).toBe("An ability has resolved 40 times this turn.");
  });
});
