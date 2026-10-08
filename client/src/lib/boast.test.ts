import { describe, it, expect } from "vitest";
import {
  abilityBlocked,
  ACTIVATION_CONDITION_UNMET,
  BOAST_NOT_ATTACKED,
  BOAST_USED,
} from "./contextMenu.logic";

// boast.test.ts — the client half of #2697 (CR 702.142a). The server
// ships `boast_blocked` on a boast ability row it will refuse, naming
// the failing half of "Activate only if this creature attacked this turn
// and only once each turn", and the menu greys the row with a reason
// that tells the player what to do about it.

describe("abilityBlocked and boast_blocked", () => {
  it("says the creature has not attacked", () => {
    expect(abilityBlocked({ boast_blocked: "not_attacked" }, false, false)).toBe(
      BOAST_NOT_ATTACKED,
    );
  });

  it("says the boast has been used", () => {
    expect(abilityBlocked({ boast_blocked: "used" }, false, false)).toBe(BOAST_USED);
  });

  it("keeps the two reasons distinct, because they recover differently", () => {
    expect(BOAST_NOT_ATTACKED).not.toBe(BOAST_USED);
  });

  it("is silent for a boast row nothing objects to", () => {
    expect(abilityBlocked({}, false, false)).toBe("");
  });

  it("beats the generic condition on a row that carries both", () => {
    expect(abilityBlocked({ boast_blocked: "used", condition_unmet: true }, false, false)).toBe(
      BOAST_USED,
    );
    expect(abilityBlocked({ condition_unmet: true }, false, false)).toBe(
      ACTIVATION_CONDITION_UNMET,
    );
  });
});
