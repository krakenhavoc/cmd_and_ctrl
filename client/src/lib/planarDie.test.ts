import { describe, expect, it } from "vitest";

import { planarDieButton, planarDieMove } from "./planarDie";
import type { GameView, LegalMoveView } from "./protocol";

// ADR 0216 §4, §10: the dock's planar die button reads the server's own
// move and cost. It is live exactly when legal_moves carry the move, and
// otherwise disabled with a reason, never hidden.

const ALICE = "alice";
const BOB = "bob";

const MOVE: LegalMoveView = {
  type: "special_action",
  player: ALICE,
  kind: "special_action",
  label: "Roll the planar die {0}",
  params: { kind: "planar_die", strict: true, auto_tap: true },
};

function view(over: Partial<GameView> = {}): GameView {
  return {
    state: "active",
    seats: [
      { id: ALICE, seat: 0, name: "Alice" },
      { id: BOB, seat: 1, name: "Bob" },
    ],
    turn: { active_seat: 0, priority_holder: 0, step: "precombat_main", number: 1 },
    stack: { count: 0, cards: [] },
    planechase: { deck_count: 70, face_up: [], die_rolls: 0, die_cost: "{0}" },
    legal_moves: [MOVE],
    ...over,
  } as unknown as GameView;
}

describe("planarDieButton", () => {
  it("is absent at a table that does not play Planechase, and for a spectator", () => {
    expect(planarDieButton(view({ planechase: undefined }), ALICE)).toBeNull();
    expect(planarDieButton(view(), null)).toBeNull();
    expect(planarDieButton(view(), "someone-else")).toBeNull();
  });

  it("sends the server's own move, with its cost", () => {
    const b = planarDieButton(view(), ALICE)!;
    expect(b.enabled).toBe(true);
    expect(b.cost).toBe("{0}");
    expect(b.params).toEqual(MOVE.params);
    expect(b.title).toMatch(/free/);
  });

  it("shows the next roll's cost", () => {
    const pc = { deck_count: 70, face_up: [], die_rolls: 2, die_cost: "{2}" };
    const b = planarDieButton(view({ planechase: pc }), ALICE)!;
    expect(b.cost).toBe("{2}");
    expect(b.title).toMatch(/\{2\}/);
  });

  it("is disabled, with a reason, outside the window", () => {
    const other = planarDieButton(
      view({
        legal_moves: [],
        turn: { active_seat: 1, priority_holder: 1, step: "precombat_main" } as never,
      }),
      ALICE,
    )!;
    expect(other.enabled).toBe(false);
    expect(other.params).toBeNull();
    expect(other.title).toMatch(/whose turn/);

    const combat = planarDieButton(
      view({
        legal_moves: [],
        turn: { active_seat: 0, priority_holder: 0, step: "declare_attackers" } as never,
      }),
      ALICE,
    )!;
    expect(combat.title).toMatch(/main phase/);

    const stacked = planarDieButton(
      view({ legal_moves: [], stack_items: [{ id: "x" }] as never }),
      ALICE,
    )!;
    expect(stacked.title).toMatch(/stack is empty/);

    const broke = planarDieButton(
      view({
        legal_moves: [],
        planechase: { deck_count: 1, face_up: [], die_rolls: 1, die_cost: "{1}" },
      }),
      ALICE,
    )!;
    expect(broke.title).toMatch(/can't pay \{1\}/);
  });

  it("never takes another seat's move or another special action", () => {
    const foretell: LegalMoveView = { ...MOVE, params: { kind: "foretell", card_id: "c" } };
    expect(planarDieMove(view({ legal_moves: [foretell] }), ALICE)).toBeNull();
    expect(planarDieMove(view({ legal_moves: [{ ...MOVE, player: BOB }] }), ALICE)).toBeNull();
  });
});
