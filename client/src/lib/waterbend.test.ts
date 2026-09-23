import { describe, it, expect } from "vitest";

import { payUnlessAnswer, shouldAskAbilityWaterbend, waterbendLimit } from "./waterbend";
import type { ActivatedAbilityView, PendingChoiceView } from "./protocol";

// Aang, Swift Savior's "Waterbend {8}: Transform Aang" (#1310).
const aang: ActivatedAbilityView = {
  index: 0,
  label: "Waterbend {8}: Transform Aang.",
  mana_cost: "{8}",
  waterbend: {
    key: "waterbend",
    label: "Waterbend {8}",
    max: 8,
    options: { cards: ["aang", "ally-a", "ally-b"] },
  },
};

// Katara, Water Tribe's Hope's "Waterbend {X}", sized by the X.
const katara: ActivatedAbilityView = {
  index: 0,
  label: "Waterbend {X}: …",
  mana_cost: "{X}",
  demands_x: true,
  min_x: 1,
  waterbend: {
    key: "waterbend",
    label: "Waterbend {X}",
    demands_x: true,
    options: { cards: ["a", "b", "c", "d"] },
  },
};

describe("waterbendLimit", () => {
  it("is the server's cap, bounded by what there is to tap", () => {
    expect(waterbendLimit(aang.waterbend!, undefined)).toBe(3);
  });

  it("is the announced X for a Waterbend {X}", () => {
    expect(waterbendLimit(katara.waterbend!, 2)).toBe(2);
    expect(waterbendLimit(katara.waterbend!, 9)).toBe(4);
    expect(waterbendLimit(katara.waterbend!, undefined)).toBe(0);
  });
});

describe("shouldAskAbilityWaterbend", () => {
  it("asks when something could help pay", () => {
    expect(shouldAskAbilityWaterbend(aang, undefined)).toBe(true);
    expect(shouldAskAbilityWaterbend(katara, 3)).toBe(true);
  });

  it("does not ask for an ability without the clause", () => {
    expect(shouldAskAbilityWaterbend({ index: 0, mana_cost: "{2}" }, undefined)).toBe(false);
  });

  it("does not ask when nothing could help", () => {
    const empty: ActivatedAbilityView = {
      ...aang,
      waterbend: { ...aang.waterbend!, options: { cards: [] } },
    };
    expect(shouldAskAbilityWaterbend(empty, undefined)).toBe(false);
  });
});

describe("payUnlessAnswer", () => {
  const ward: PendingChoiceView = {
    id: "c1",
    kind: "pay_unless",
    chooser: "p1",
    count: 1,
    pay_cost: "{4}",
    tap_cost: { key: "waterbend", label: "Waterbend {4}", max: 4, options: { cards: ["x", "y"] } },
  } as PendingChoiceView;

  it("sends the taps beside apply on a Pay (#1311)", () => {
    expect(payUnlessAnswer(ward, true, ["x", "y"])).toEqual({
      choice_id: "c1",
      apply: true,
      tap_ids: ["x", "y"],
    });
  });

  it("never sends taps on a decline", () => {
    expect(payUnlessAnswer(ward, false, ["x"])).toEqual({ choice_id: "c1", apply: false });
  });

  it("sends no tap_ids for an ordinary pay-unless", () => {
    const plain = { ...ward, tap_cost: undefined } as PendingChoiceView;
    expect(payUnlessAnswer(plain, true, ["x"])).toEqual({ choice_id: "c1", apply: true });
  });
});
