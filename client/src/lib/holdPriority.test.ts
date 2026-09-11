import { describe, it, expect, beforeEach } from "vitest";
import { get } from "svelte/store";

import {
  holdPriority,
  isHoldingPriority,
  ownsEveryStackItem,
  setHoldPriority,
  toggleHoldPriority,
  _resetForTests,
} from "./holdPriority";
import type { CardView, GameView, PlayerView, StackItemView, TurnView, ZoneView } from "./protocol";

// Fixture helpers mirror priority.test.ts / timing.test.ts. Kept
// local rather than exported because they'd otherwise ship in the
// prod bundle.
function emptyZone(kind: string, owner = ""): ZoneView {
  return { kind, owner, count: 0, cards: [] };
}

function turn(): TurnView {
  return {
    number: 1,
    active_seat: 0,
    priority_holder: 0,
    phase: "precombat_main",
    step: "precombat_main",
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

function stackCard(name: string, controller: string): CardView {
  return {
    instance_id: `c-${name}`,
    name,
    owner: controller,
    controller,
    type_line: "Instant",
  };
}

function item(
  id: string,
  controller: string,
  kind: StackItemView["kind"] = "spell",
): StackItemView {
  return { id, kind, controller, owner: controller, source_card_id: `src-${id}` };
}

interface SnapOpts {
  stackItems?: StackItemView[];
  stackCards?: CardView[];
  pendingTriggers?: StackItemView[];
}

function snap(o: SnapOpts = {}): GameView {
  return {
    id: "g",
    state: "active",
    seats: [seat("p0", 0), seat("p1", 1)],
    battlefield: emptyZone("battlefield"),
    stack: {
      ...emptyZone("stack"),
      cards: o.stackCards ?? [],
      count: (o.stackCards ?? []).length,
    },
    exile: emptyZone("exile"),
    turn: turn(),
    mulligans_open: false,
    stack_items: o.stackItems ?? [],
    pending_triggers: o.pendingTriggers ?? [],
    split_second_active: false,
  };
}

describe("holdPriority store", () => {
  beforeEach(() => _resetForTests());

  it("starts released", () => {
    expect(get(holdPriority)).toBe(false);
    expect(isHoldingPriority()).toBe(false);
  });

  it("toggleHoldPriority flips and reports the new state", () => {
    expect(toggleHoldPriority()).toBe(true);
    expect(get(holdPriority)).toBe(true);
    expect(toggleHoldPriority()).toBe(false);
    expect(get(holdPriority)).toBe(false);
  });

  it("setHoldPriority pins an explicit value", () => {
    setHoldPriority(true);
    expect(isHoldingPriority()).toBe(true);
    setHoldPriority(true);
    expect(isHoldingPriority()).toBe(true);
    setHoldPriority(false);
    expect(isHoldingPriority()).toBe(false);
  });
});

describe("ownsEveryStackItem", () => {
  it("is false for a null snapshot or an unknown viewer", () => {
    expect(ownsEveryStackItem(null, "p0")).toBe(false);
    expect(ownsEveryStackItem(snap({ stackItems: [item("a", "p0")] }), null)).toBe(false);
  });

  it("is false for an empty stack — the caller's own gates own that path", () => {
    expect(ownsEveryStackItem(snap(), "p0")).toBe(false);
  });

  it("is true when every stack item is the viewer's own spell", () => {
    const s = snap({
      stackItems: [item("a", "p0"), item("b", "p0")],
      stackCards: [stackCard("bolt", "p0")],
    });
    expect(ownsEveryStackItem(s, "p0")).toBe(true);
  });

  it("is true for the viewer's own triggered and activated abilities", () => {
    const s = snap({
      stackItems: [item("a", "p0", "triggered"), item("b", "p0", "activated")],
    });
    expect(ownsEveryStackItem(s, "p0")).toBe(true);
  });

  it("is false when an opponent controls any item — the response window is theirs to answer", () => {
    const s = snap({ stackItems: [item("a", "p0"), item("b", "p1")] });
    expect(ownsEveryStackItem(s, "p0")).toBe(false);
  });

  it("is false when the opponent's item sits UNDER the viewer's", () => {
    // Wire order is bottom..top: the opponent cast first, the viewer
    // responded. The opponent's spell is still live and still needs
    // an answer once the viewer's resolves.
    const s = snap({ stackItems: [item("theirs", "p1"), item("mine", "p0")] });
    expect(ownsEveryStackItem(s, "p0")).toBe(false);
  });

  it("is false when a card in the stack zone belongs to an opponent", () => {
    const s = snap({
      stackItems: [item("a", "p0")],
      stackCards: [stackCard("bolt", "p1")],
    });
    expect(ownsEveryStackItem(s, "p0")).toBe(false);
  });

  it("reads the stack zone even when stack_items has not landed yet", () => {
    expect(ownsEveryStackItem(snap({ stackCards: [stackCard("bolt", "p0")] }), "p0")).toBe(true);
    expect(ownsEveryStackItem(snap({ stackCards: [stackCard("bolt", "p1")] }), "p0")).toBe(false);
  });

  it("is false while pending triggers are still draining (CR 603.3b)", () => {
    const s = snap({
      stackItems: [item("a", "p0")],
      pendingTriggers: [item("t", "p1", "triggered")],
    });
    expect(ownsEveryStackItem(s, "p0")).toBe(false);
  });

  it("is false from the opponent's point of view on the viewer's own stack", () => {
    const s = snap({ stackItems: [item("a", "p0")] });
    expect(ownsEveryStackItem(s, "p1")).toBe(false);
  });
});
