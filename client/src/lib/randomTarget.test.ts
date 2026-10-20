import { describe, it, expect } from "vitest";
import { get } from "svelte/store";

import { beginForAbility, cancel, stepsFor, targeting } from "./targeting";
import type { ActivatedAbilityView, CardView, LegalTargetsView } from "./protocol";

// #4051: a clause "chosen at random" (Goblin Test Pilot) is drawn by the
// server as the ability is activated. The walk skips it, so an ability
// whose only clause is random opens no walk and is activated at once.

const random: LegalTargetsView = {
  players: ["p0", "p1"],
  cards: ["c1"],
  min: 1,
  max: 1,
  at_random: true,
};

describe("stepsFor — a random clause is not a step", () => {
  it("is an empty walk when the only clause is drawn at random", () => {
    expect(stepsFor("any", random, undefined, 0)).toEqual([]);
  });
});

describe("beginForAbility — Goblin Test Pilot", () => {
  it("leaves no walk open for an ability whose only clause is random", () => {
    cancel();
    const ability = {
      index: 0,
      label: "{T}: This creature deals 2 damage to any target chosen at random.",
      target_mode: "any",
      legal_targets: random,
    } as unknown as ActivatedAbilityView;
    const card: CardView = {
      instance_id: "src",
      name: "Goblin Test Pilot",
      owner: "p0",
      controller: "p0",
    };
    beginForAbility(card, ability, []);
    expect(get(targeting)).toBeNull();
  });
});
