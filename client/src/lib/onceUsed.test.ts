import { describe, it, expect } from "vitest";
import {
  abilityBlocked,
  ABILITY_EXHAUSTED,
  ACTIVATION_CONDITION_UNMET,
  ONCE_USED,
  POWER_UP_USED,
} from "./contextMenu.logic";
import { L } from "./labels";

// onceUsed.test.ts — the client half of #5221. The server ships
// `once_used` on an ability row that says "Activate only once" without
// the exhaust keyword (CR 602.5b) once this permanent has activated it,
// and the menu greys the row with its own reason, a registered label.

describe("abilityBlocked and once_used", () => {
  it("names the spent ability", () => {
    expect(abilityBlocked({ once_used: true }, false, false)).toBe(ONCE_USED);
    expect(ONCE_USED).toBe(L.onceUsed);
  });

  it("is silent for an ability that has not been used", () => {
    expect(abilityBlocked({}, false, false)).toBe("");
  });

  it("is neither keyword's reason, and beats condition_unmet", () => {
    expect(ONCE_USED).not.toBe(ABILITY_EXHAUSTED);
    expect(ONCE_USED).not.toBe(POWER_UP_USED);
    expect(abilityBlocked({ once_used: true, condition_unmet: true }, false, false)).toBe(
      ONCE_USED,
    );
    expect(abilityBlocked({ condition_unmet: true }, false, false)).toBe(
      ACTIVATION_CONDITION_UNMET,
    );
  });
});
