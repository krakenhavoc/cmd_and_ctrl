import { describe, it, expect } from "vitest";
import { abilityBlocked, ABILITY_EXHAUSTED, ACTIVATION_CONDITION_UNMET } from "./contextMenu.logic";

// exhaust.test.ts — the client half of #1181. The server ships
// `exhausted` on an activated-ability row it has already spent
// ("Activate each exhaust ability only once"), and the menu greys the
// row with its own reason.
//
// Its own reason and not ACTIVATION_CONDITION_UNMET because the two
// recover differently: a condition may hold again next turn, an
// exhaust only if the permanent becomes a new object.

describe("abilityBlocked and exhausted", () => {
  it("names the exhaust for a spent ability", () => {
    expect(abilityBlocked({ exhausted: true }, false, false)).toBe(ABILITY_EXHAUSTED);
  });

  it("is silent for an exhaust ability that has not been used", () => {
    expect(abilityBlocked({}, false, false)).toBe("");
  });

  // Bitter Work prints both gates. "Already activated" is the one that
  // will still be true tomorrow, so it wins.
  it("beats condition_unmet on a row that carries both", () => {
    expect(abilityBlocked({ exhausted: true, condition_unmet: true }, false, false)).toBe(
      ABILITY_EXHAUSTED,
    );
    expect(abilityBlocked({ condition_unmet: true }, false, false)).toBe(
      ACTIVATION_CONDITION_UNMET,
    );
  });

  // A tapped source is the more immediate truth, exactly as it is for
  // every other reason in this function.
  it("keeps the tap reason for a tapped source", () => {
    expect(abilityBlocked({ tap_cost: true, exhausted: true }, true, false)).toBe("already tapped");
  });
});
