// #759: the tap-another cost (station, CR 702.184a) greys its menu
// row on the same terms as a return-to-hand cost — the server's option
// list already holds only the untapped creatures that could pay, so
// its length against `min` is the whole of CR 118.3.

import { describe, expect, it } from "vitest";
import { abilityBlocked, tapOthersShortfall } from "./contextMenu.logic";

describe("tapOthersShortfall", () => {
  it("is empty when the ability has no tap-another cost", () => {
    expect(tapOthersShortfall(undefined)).toBe("");
  });

  it("is empty when an untapped creature can pay", () => {
    expect(tapOthersShortfall({ cards: ["bear"], min: 1, max: 1 })).toBe("");
  });

  it("names the clause when nothing can pay", () => {
    expect(
      tapOthersShortfall({ cards: [], min: 1, max: 1 }, "another untapped creature you control"),
    ).toBe("nothing to tap (another untapped creature you control)");
  });

  it("counts against the clause's floor, not against one", () => {
    expect(tapOthersShortfall({ cards: ["a", "b"], min: 3, max: 3 })).not.toBe("");
  });

  it("requires one useful payment for a count announced as X", () => {
    const variable = { min: 0, max: 0, count_from_x: true };
    expect(tapOthersShortfall({ ...variable, cards: [] }, "X untapped Foods you control")).toBe(
      "nothing to tap (X untapped Foods you control)",
    );
    expect(tapOthersShortfall({ ...variable, cards: ["food"] })).toBe("");
  });
});

describe("abilityBlocked with a tap-another cost", () => {
  const station = {
    tap_others_label: "another untapped creature you control",
    tap_others_options: { cards: [] as string[], min: 1, max: 1 },
  };

  it("greys the row when no creature can be tapped", () => {
    expect(abilityBlocked(station, false, false)).toContain("nothing to tap");
  });

  it("leaves the row live when one can", () => {
    const payable = { ...station, tap_others_options: { cards: ["bear"], min: 1, max: 1 } };
    expect(abilityBlocked(payable, false, false)).toBe("");
  });

  it("does not care that the SOURCE is summoning sick — the cost is not {T}", () => {
    const payable = { ...station, tap_others_options: { cards: ["bear"], min: 1, max: 1 } };
    expect(abilityBlocked(payable, false, true)).toBe("");
  });
});

// #5669: "Untap N tapped creatures you control" (Halo Fountain) rides
// the same option list, with the tapped permanents on it.
describe("an untap-tapped-creatures cost", () => {
  const fountain = {
    tap_cost: true,
    tap_others_untap: true,
    tap_others_label: "two tapped creatures you control",
    tap_others_options: { cards: ["a"], min: 2, max: 2 },
  };

  it("names the clause with the verb untap when too few are tapped", () => {
    expect(tapOthersShortfall(fountain.tap_others_options, fountain.tap_others_label, true)).toBe(
      "nothing to untap (two tapped creatures you control)",
    );
    expect(abilityBlocked(fountain, false, false)).toBe(
      "nothing to untap (two tapped creatures you control)",
    );
  });

  it("is open when enough tapped creatures can pay", () => {
    const payable = { ...fountain, tap_others_options: { cards: ["a", "b"], min: 2, max: 2 } };
    expect(abilityBlocked(payable, false, false)).toBe("");
  });

  it("greys a tapped source, unless naming it for the clause untaps it first", () => {
    const payable = { ...fountain, tap_others_options: { cards: ["a", "b"], min: 2, max: 2 } };
    expect(abilityBlocked(payable, true, false)).toBe("already tapped");
    expect(abilityBlocked({ ...payable, tap_others_frees_source: true }, true, false)).toBe("");
  });
});
