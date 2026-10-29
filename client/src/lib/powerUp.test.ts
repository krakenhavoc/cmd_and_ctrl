import { describe, it, expect } from "vitest";
import {
  abilityBlocked,
  ABILITY_EXHAUSTED,
  ACTIVATION_CONDITION_UNMET,
  POWER_UP_USED,
} from "./contextMenu.logic";
import { L } from "./labels";

// powerUp.test.ts — the client half of #4527. The server ships
// `power_up_used` on a power-up ability row this permanent has already
// activated (CR 702.193a: "Activate each power-up ability only once"),
// and the menu greys the row with its own reason, a registered label.

describe("abilityBlocked and power_up_used", () => {
  it("names the power-up for a used ability", () => {
    expect(abilityBlocked({ power_up_used: true }, false, false)).toBe(POWER_UP_USED);
    expect(POWER_UP_USED).toBe(L.powerUpUsed);
  });

  it("is silent for a power-up ability that has not been used", () => {
    expect(abilityBlocked({}, false, false)).toBe("");
  });

  it("is not the exhaust reason, and beats condition_unmet", () => {
    expect(POWER_UP_USED).not.toBe(ABILITY_EXHAUSTED);
    expect(abilityBlocked({ power_up_used: true, condition_unmet: true }, false, false)).toBe(
      POWER_UP_USED,
    );
    expect(abilityBlocked({ condition_unmet: true }, false, false)).toBe(
      ACTIVATION_CONDITION_UNMET,
    );
  });
});
