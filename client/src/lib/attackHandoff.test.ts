// attackHandoff.test.ts — #4154: the active player's half of an attack
// declaration another player chose (Master Warcraft).

import { describe, expect, it, vi } from "vitest";

import {
  attackHandoffRequest,
  attackTargetsOwed,
  chosenAttacks,
  chosenAttacksCost,
  confirmAttacksParams,
} from "./attackHandoff";
import { owesChosenDeclaration } from "./priority";
import type { GameView } from "./protocol";

const ALICE = "alice";
const BOB = "bob";
const CAROL = "carol";

function frame(extra: Partial<GameView> = {}, taxed = false): GameView {
  return {
    id: "g",
    state: "active",
    seats: [
      { id: ALICE, name: "Alice", seat: 0 },
      { id: BOB, name: "Bob", seat: 1 },
      { id: CAROL, name: "Carol", seat: 2 },
    ],
    turn: {
      seq: 3,
      number: 3,
      active_seat: 1,
      priority_holder: 1,
      step: "declare_attackers",
      attack_targets: [
        { kind: "player", id: ALICE },
        { kind: "player", id: CAROL, ...(taxed ? { tax: "{2}" } : {}) },
      ],
    },
    battlefield: {
      cards: [
        { instance_id: "bear", name: "Bear", controller: BOB, attacking_target: CAROL },
        { instance_id: "elk", name: "Elk", controller: BOB },
        { instance_id: "wall", name: "Wall", controller: CAROL },
      ],
    },
    attack_handoff: { stage: "targets", declarer: ALICE, source_name: "Master Warcraft" },
    ...extra,
  } as unknown as GameView;
}

describe("the active player's half", () => {
  it("is owed by the active player only, while the stage is targets", () => {
    expect(attackTargetsOwed(frame(), BOB)).toBe(true);
    expect(attackTargetsOwed(frame(), ALICE)).toBe(false);
    expect(attackTargetsOwed(frame({ attack_handoff: { declarer: ALICE } }), BOB)).toBe(false);
    expect(attackTargetsOwed(frame({ attack_handoff: undefined }), BOB)).toBe(false);
  });

  it("confirms every chosen attacker at its target, priced by the server", () => {
    const v = frame({}, true);
    const attacks = chosenAttacks(v, BOB);
    expect(attacks).toEqual([{ attacker: "bear", target: CAROL }]);
    expect(chosenAttacksCost(v, attacks)).toBe("{2}");
    expect(chosenAttacksCost(frame(), attacks)).toBe("");
    expect(confirmAttacksParams(attacks)).toEqual({ attackers: attacks, auto_tap: true });
  });

  it("offers Confirm, and Decline only when something costs", () => {
    const confirm = vi.fn();
    const decline = vi.fn();
    const free = attackHandoffRequest(frame(), BOB, confirm, decline)!;
    expect(free.question).toBe("Alice chose 1 attacker for you");
    expect(free.row?.map((a) => a.label)).toEqual(["Confirm attacks"]);
    free.row![0].onPress();
    expect(confirm).toHaveBeenCalledWith([{ attacker: "bear", target: CAROL }]);

    const taxed = attackHandoffRequest(frame({}, true), BOB, confirm, decline)!;
    expect(taxed.row?.map((a) => a.label)).toEqual([
      "Confirm attacks (pay {2})",
      "Decline the costs",
    ]);
    taxed.row![1].onPress();
    expect(decline).toHaveBeenCalled();

    expect(attackHandoffRequest(frame(), ALICE, confirm, decline)).toBeNull();
  });
});

describe("owesChosenDeclaration", () => {
  it("holds autopass for a seat whose declaration is chosen, and in the active player's half", () => {
    const declarer = frame({
      attack_handoff: undefined,
      player_controls: [{ controller: ALICE, controlled: BOB, until: "declarations" }],
    });
    expect(owesChosenDeclaration(declarer, BOB)).toBe(true);
    expect(owesChosenDeclaration(declarer, ALICE)).toBe(false);
    expect(owesChosenDeclaration(frame(), BOB)).toBe(true);
    expect(owesChosenDeclaration(frame(), CAROL)).toBe(false);
    expect(
      owesChosenDeclaration(frame({ attack_handoff: { stage: "chosen", declarer: ALICE } }), BOB),
    ).toBe(false);
  });
});
