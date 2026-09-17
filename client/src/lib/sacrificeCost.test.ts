import { describe, expect, it } from "vitest";

import { abilityBlocked } from "./contextMenu.logic";
import type { CardView } from "./protocol";
import {
  canConfirmSacrifice,
  chooseForMeState,
  chooseSacrificeForMe,
  keepAvailablePicks,
  orderSacrificeOptions,
  sacrificeCount,
  sacrificeShortfall,
  toggleSacrificePick,
} from "./sacrificeCost";

// #747: a sacrifice cost of N permanents. The count is the view's
// sacrifice_options.max; the picker confirms at exactly N; "Choose for
// me" takes the first N of the server's payment order and never
// confirms.

function card(id: string, name = id): CardView {
  return { instance_id: id, name, owner: "p0", controller: "p0" } as CardView;
}

describe("sacrificeCount", () => {
  it("reads max, and treats a missing or zero count as one", () => {
    expect(sacrificeCount({ cards: [], min: 3, max: 3 })).toBe(3);
    expect(sacrificeCount({ cards: ["a"] })).toBe(1);
    expect(sacrificeCount({ cards: ["a"], min: 0, max: 0 })).toBe(1);
    expect(sacrificeCount(undefined)).toBe(1);
  });
});

describe("sacrificeShortfall", () => {
  it("keeps the one-permanent wording", () => {
    expect(sacrificeShortfall({ cards: [], min: 1, max: 1 }, "a creature")).toBe(
      "nothing to sacrifice (a creature)",
    );
    expect(sacrificeShortfall({ cards: ["a"], min: 1, max: 1 }, "a creature")).toBe("");
  });

  it("names the count when there are fewer options than it", () => {
    expect(sacrificeShortfall({ cards: ["f1", "f2"], min: 3, max: 3 }, "three Foods")).toBe(
      "needs three Foods (you have 2)",
    );
    expect(sacrificeShortfall({ cards: [], min: 2, max: 2 }, "two artifacts")).toBe(
      "needs two artifacts (you have 0)",
    );
    expect(sacrificeShortfall({ cards: ["f1", "f2", "f3"], min: 3, max: 3 }, "three Foods")).toBe(
      "",
    );
  });

  it("says nothing for an ability with no sacrifice component", () => {
    expect(sacrificeShortfall(undefined, "a creature")).toBe("");
  });
});

describe("the multi-select picker", () => {
  it("replaces the pick at a count of one", () => {
    expect(toggleSacrificePick([], "a", 1)).toEqual(["a"]);
    expect(toggleSacrificePick(["a"], "b", 1)).toEqual(["b"]);
    expect(toggleSacrificePick(["a"], "a", 1)).toEqual(["a"]);
  });

  it("adds and removes picks at a count of N, and never exceeds N", () => {
    let chosen: string[] = [];
    chosen = toggleSacrificePick(chosen, "a", 2);
    chosen = toggleSacrificePick(chosen, "b", 2);
    expect(chosen).toEqual(["a", "b"]);
    expect(toggleSacrificePick(chosen, "c", 2)).toEqual(["a", "b"]);
    expect(toggleSacrificePick(chosen, "a", 2)).toEqual(["b"]);
  });

  it("confirms at exactly N distinct picks, not below or above", () => {
    expect(canConfirmSacrifice([], 2)).toBe(false);
    expect(canConfirmSacrifice(["a"], 2)).toBe(false);
    expect(canConfirmSacrifice(["a", "b"], 2)).toBe(true);
    expect(canConfirmSacrifice(["a", "b", "c"], 2)).toBe(false);
    expect(canConfirmSacrifice(["a", "a"], 2)).toBe(false);
    expect(canConfirmSacrifice(["a"], 1)).toBe(true);
  });

  it("Choose for me takes the first N of the server's order and only selects", () => {
    const order = ["treasure-1", "treasure-2", "ornithopter", "source", "sol-ring"];
    const picks = chooseSacrificeForMe(order, 3);
    expect(picks).toEqual(["treasure-1", "treasure-2", "ornithopter"]);
    // The player can still change a pick before confirming.
    expect(toggleSacrificePick(picks, "ornithopter", 3)).toEqual(["treasure-1", "treasure-2"]);
    expect(chooseSacrificeForMe(["a"], 3)).toEqual(["a"]);
    expect(canConfirmSacrifice(chooseSacrificeForMe(["a"], 3), 3)).toBe(false);
  });

  it("Choose for me replaces a partial selection, and a changed pick is what confirms", () => {
    const order = ["t1", "t2", "cheap", "dear"];
    let chosen = toggleSacrificePick([], "dear", 2);
    chosen = chooseSacrificeForMe(order, 2);
    expect(chosen).toEqual(["t1", "t2"]);
    // Choose for me only selects: the confirm gate is a separate step.
    expect(canConfirmSacrifice(chosen, 2)).toBe(true);
    chosen = toggleSacrificePick(chosen, "t2", 2);
    chosen = toggleSacrificePick(chosen, "cheap", 2);
    expect(canConfirmSacrifice(chosen, 2)).toBe(true);
    expect(chosen).toEqual(["t1", "cheap"]);
  });

  it("shows Choose for me only for a count of two or more, disabled when short", () => {
    expect(chooseForMeState(1, 5)).toEqual({ shown: false, disabled: false });
    expect(chooseForMeState(2, 5)).toEqual({ shown: true, disabled: false });
    expect(chooseForMeState(5, 4)).toEqual({ shown: true, disabled: true });
  });

  it("lists the options in the server's order, not board order", () => {
    const board = [card("rock"), card("bear"), card("food")];
    expect(
      orderSacrificeOptions(board, ["food", "rock", "gone"]).map((c) => c.instance_id),
    ).toEqual(["food", "rock"]);
    expect(orderSacrificeOptions(board, undefined)).toEqual([]);
  });

  it("drops a pick that left the options while the picker was open", () => {
    // Two of three Treasures picked, then one is destroyed in response.
    expect(keepAvailablePicks(["t1", "t2"], ["t2", "t3"])).toEqual(["t2"]);
    const chosen = ["t1", "t2"];
    expect(keepAvailablePicks(chosen, ["t1", "t2", "t3"])).toBe(chosen);
    expect(keepAvailablePicks([], [])).toEqual([]);
  });
});

describe("abilityBlocked for a sacrifice-N cost", () => {
  it("greys the row when the pool is short and names the shortfall", () => {
    const samwise = {
      sacrifice_label: "three Foods",
      sacrifice_options: { cards: ["f1", "f2"], min: 3, max: 3 },
    };
    expect(abilityBlocked(samwise, false, false)).toBe("needs three Foods (you have 2)");
    const paid = { ...samwise, sacrifice_options: { cards: ["f1", "f2", "f3"], min: 3, max: 3 } };
    expect(abilityBlocked(paid, false, false)).toBe("");
  });

  it("keeps the one-permanent reason", () => {
    const bombardment = { sacrifice_label: "a creature", sacrifice_options: { cards: [] } };
    expect(abilityBlocked(bombardment, false, false)).toBe("nothing to sacrifice (a creature)");
  });
});
