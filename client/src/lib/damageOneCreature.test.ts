// damageOneCreature.test.ts — #6192, Cunning Giant: "If this creature is
// unblocked, you may have it assign its combat damage to a creature
// defending player controls." The prompt offers the defending player's
// creatures and the player; the answer is all of the damage on one
// creature or all of it on the player, never divided.

import { describe, expect, it } from "vitest";
import { canDealDamage, oneCreatureSplit } from "./damageAssignment";
import type { DamageShares } from "./damageAssignment";
import type { DamageAssignmentView } from "./protocol";

const giant: DamageAssignmentView = {
  attacker_card_id: "giant",
  blocker_card_ids: ["bear", "wall"],
  attacker_power: 4,
  as_though_unblocked: true,
  one_creature: true,
};

const shares = (bear: number, wall: number, trample: number): DamageShares => ({
  amounts: { bear, wall },
  trample,
});

describe("one creature or the player (#6192)", () => {
  it("all of it on one creature deals", () => {
    expect(oneCreatureSplit(giant, shares(4, 0, 0))).toBe(false);
    expect(canDealDamage(giant, shares(4, 0, 0))).toBe(true);
  });

  it("all of it on the player deals", () => {
    expect(canDealDamage(giant, shares(0, 0, 4))).toBe(true);
  });

  it("divided between two creatures does not", () => {
    expect(oneCreatureSplit(giant, shares(2, 2, 0))).toBe(true);
    expect(canDealDamage(giant, shares(2, 2, 0))).toBe(false);
  });

  it("split between a creature and the player does not", () => {
    expect(canDealDamage(giant, shares(2, 0, 2))).toBe(false);
  });

  it("an ordinary frame is not held to it", () => {
    const wolverine = { ...giant, one_creature: false };
    expect(oneCreatureSplit(wolverine, shares(2, 2, 0))).toBe(false);
  });
});
