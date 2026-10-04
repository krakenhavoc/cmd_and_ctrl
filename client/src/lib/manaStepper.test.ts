// ADR 0117 §4: the per-colour stepper's logic, as pure functions. The
// rendered stepper is manaStepper.render.test.ts.

import { afterEach, describe, expect, it } from "vitest";

import {
  canAdd,
  canRemove,
  fixedShare,
  forgetSplits,
  rememberSplit,
  rememberedSplit,
  splitAssignment,
  splitColors,
  splitComplete,
  splitFeasible,
  splitMemoryKey,
  splitStart,
  stepperSlots,
} from "./manaStepper";
import type { ManaAbilityView } from "./protocol";

afterEach(() => forgetSplits());

const UR = ["U", "R"];
const FIVE = ["W", "U", "B", "R", "G"];
const times = (n: number, l: string[]) => Array.from({ length: n }, () => [...l]);

// Whether `colors` is one colour per slot, each one its slot offers.
function respects(lists: string[][], colors: string[] | null): boolean {
  return !!colors && colors.length === lists.length && colors.every((c, i) => lists[i].includes(c));
}

const tally = (colors: string[]) =>
  colors.reduce<Record<string, number>>((m, c) => ({ ...m, [c]: (m[c] ?? 0) + 1 }), {});

describe("stepperSlots: when the stepper answers an ability", () => {
  const ability = (extra: Partial<ManaAbilityView>): ManaAbilityView => ({ index: 0, ...extra });

  it("takes two or more lists with a real choice (Vivi at power 3)", () => {
    const lists = times(3, UR);
    expect(stepperSlots(ability({ produced: "{U|R}{U|R}{U|R}", color_options: lists }))).toEqual(
      lists,
    );
  });

  it("takes a fixed slot beside a wide one", () => {
    const lists = [["G"], FIVE];
    expect(stepperSlots(ability({ color_options: lists }))).toEqual(lists);
  });

  it("leaves one picking slot to its buttons, OneColorOfAmount included", () => {
    expect(stepperSlots(ability({ produced: "{W|U|B|R|G}", color_options: [FIVE] }))).toBeNull();
    expect(
      stepperSlots(ability({ produced: "{W3|U3|B3|R3|G3}", color_options: [FIVE] })),
    ).toBeNull();
  });

  it("is not needed when every list has one option (no choice at all)", () => {
    expect(stepperSlots(ability({ color_options: [["G"], ["G"]] }))).toBeNull();
  });

  it("leaves a counted slot beside a second slot to buttons (#742)", () => {
    expect(
      stepperSlots(
        ability({
          produced: "{W2|U2}{W|U}",
          color_options: [
            ["W", "U"],
            ["W", "U"],
          ],
        }),
      ),
    ).toBeNull();
  });

  it("needs color_options, and none of them empty", () => {
    expect(stepperSlots(ability({ produced: "{U|R}{U|R}" }))).toBeNull();
    expect(stepperSlots(ability({ color_options: [UR, []] }))).toBeNull();
  });
});

describe("splitColors: the rows, in the server's order", () => {
  it("is the union of the lists by first appearance (identity first, #843)", () => {
    expect(splitColors([["G", "W"], ["W", "U", "G"], ["B"]])).toEqual(["G", "W", "U", "B"]);
  });
});

describe("Hall's condition", () => {
  it("is trivially met for equal lists", () => {
    const lists = times(5, FIVE);
    expect(splitFeasible(lists, { W: 2, U: 1, G: 2 })).toBe(true);
    expect(splitComplete(lists, { W: 2, U: 1, G: 2 })).toBe(true);
    expect(splitFeasible(lists, { R: 5 })).toBe(true);
    expect(splitFeasible(lists, { R: 6 })).toBe(false);
  });

  it("holds a fixed slot beside a wide one to its colour", () => {
    // Command Tower narrowed to green (CR 903.4f) beside any colour.
    const lists = [["G"], FIVE];
    expect(splitComplete(lists, { G: 1, R: 1 })).toBe(true);
    expect(splitComplete(lists, { G: 2 })).toBe(true);
    // Both on red: the fixed slot can't make red.
    expect(splitFeasible(lists, { R: 2 })).toBe(false);
    // A partial vector that can still be completed is feasible.
    expect(splitFeasible(lists, { R: 1 })).toBe(true);
  });

  it("refuses a vector two different lists cannot meet", () => {
    const lists = [
      ["W", "U"],
      ["W", "B"],
    ];
    expect(splitComplete(lists, { U: 1, B: 1 })).toBe(true);
    expect(splitComplete(lists, { W: 2 })).toBe(true);
    // Only one slot offers U.
    expect(splitFeasible(lists, { U: 2 })).toBe(false);
    // Three lists, two of which offer only {U, B}: three of U and B
    // together would need three slots that offer either.
    const three = [
      ["U", "B"],
      ["U", "B"],
      ["W", "R"],
    ];
    expect(splitFeasible(three, { U: 2, B: 1 })).toBe(false);
    expect(splitFeasible(three, { U: 1, B: 1, W: 1 })).toBe(true);
  });

  it("refuses a count on a colour no list offers, and a negative count", () => {
    expect(splitFeasible(times(2, UR), { G: 1 })).toBe(false);
    expect(splitFeasible(times(2, UR), { U: -1 })).toBe(false);
    expect(splitFeasible(times(2, UR), { G: 0, U: 2 })).toBe(true);
  });
});

describe("+ and − enablement", () => {
  it("disables + once the total is reached", () => {
    const lists = times(3, UR);
    expect(canAdd(lists, { U: 2 }, "R")).toBe(true);
    expect(canAdd(lists, { U: 2, R: 1 }, "R")).toBe(false);
    expect(canAdd(lists, { U: 2, R: 1 }, "U")).toBe(false);
  });

  it("disables + when one more would break the condition", () => {
    const lists = [["G"], FIVE];
    expect(canAdd(lists, { R: 1 }, "R")).toBe(false);
    expect(canAdd(lists, { R: 1 }, "G")).toBe(true);
    // A colour only the fixed slot offers, beside a list without it.
    const fixedBlack = [["B"], UR];
    expect(canAdd(fixedBlack, { B: 1 }, "B")).toBe(false);
  });

  it("disables − at 0 and on a fixed slot's share", () => {
    const lists = [["G"], ["G"], FIVE];
    expect(fixedShare(lists, "G")).toBe(2);
    expect(canRemove(lists, { G: 3 }, "G")).toBe(true);
    expect(canRemove(lists, { G: 2, R: 1 }, "G")).toBe(false);
    expect(canRemove(lists, { G: 2, R: 1 }, "R")).toBe(true);
    expect(canRemove(lists, { G: 2, R: 1 }, "U")).toBe(false);
  });
});

describe("splitAssignment: colors[] in slot order", () => {
  it("names one colour per slot for equal lists", () => {
    const lists = times(3, UR);
    const out = splitAssignment(lists, { U: 1, R: 2 });
    expect(respects(lists, out)).toBe(true);
    expect(tally(out!)).toEqual({ U: 1, R: 2 });
  });

  it("puts the fixed slot's colour on the fixed slot", () => {
    const lists = [FIVE, ["G"], FIVE];
    const out = splitAssignment(lists, { G: 1, R: 1, W: 1 });
    expect(respects(lists, out)).toBe(true);
    expect(out![1]).toBe("G");
    expect(tally(out!)).toEqual({ G: 1, R: 1, W: 1 });
  });

  it("re-routes an earlier slot when a later one needs its colour (an augmenting path)", () => {
    // Slot 0 would greedily take U; slot 1 offers only U, so slot 0
    // must move to W.
    const lists = [["U", "W"], ["U"]];
    expect(splitAssignment(lists, { U: 1, W: 1 })).toEqual(["W", "U"]);
    const chain = [["U", "B"], ["B", "R"], ["U"]];
    const out = splitAssignment(chain, { U: 1, B: 1, R: 1 });
    expect(out).toEqual(["B", "R", "U"]);
  });

  it("is null for an incomplete or infeasible vector", () => {
    expect(splitAssignment(times(3, UR), { U: 2 })).toBeNull();
    expect(
      splitAssignment(
        [
          ["W", "U"],
          ["W", "B"],
        ],
        { U: 2 },
      ),
    ).toBeNull();
  });

  it("respects every list over many slots (Cascading Cataracts' five)", () => {
    const lists = times(5, FIVE);
    const out = splitAssignment(lists, { W: 1, U: 1, B: 1, R: 1, G: 1 });
    expect(respects(lists, out)).toBe(true);
    expect(new Set(out).size).toBe(5);
  });
});

describe("splitStart", () => {
  it("is the split last confirmed for this card when it still fits", () => {
    const key = splitMemoryKey("g1", "vivi", 0);
    rememberSplit(key, { U: 1, R: 2 });
    expect(splitStart(times(3, UR), rememberedSplit(key))).toEqual({ U: 1, R: 2 });
  });

  it("ignores a remembered split that no longer fits (Vivi's power changed)", () => {
    const key = splitMemoryKey("g1", "vivi", 0);
    rememberSplit(key, { U: 1, R: 2 });
    expect(splitStart(times(4, UR), rememberedSplit(key))).toEqual({ U: 4 });
  });

  it("is all of it on the first colour every list offers", () => {
    expect(splitStart(times(3, ["R", "G"]))).toEqual({ R: 3 });
    expect(splitStart([FIVE, ["G"]])).toEqual({ G: 2 });
  });

  it("falls back to each slot on its own first option", () => {
    expect(
      splitStart([
        ["W", "U"],
        ["B", "R"],
        ["W", "G"],
      ]),
    ).toEqual({ W: 2, B: 1 });
  });

  it("always fills the total", () => {
    for (const lists of [times(6, UR), [["G"], FIVE], [["W"], ["U", "B"]]]) {
      expect(splitComplete(lists, splitStart(lists))).toBe(true);
    }
  });

  it("keeps memory per game and per card", () => {
    rememberSplit(splitMemoryKey("g1", "vivi", 0), { R: 3 });
    expect(rememberedSplit(splitMemoryKey("g2", "vivi", 0))).toBeNull();
    expect(rememberedSplit(splitMemoryKey("g1", "other", 0))).toBeNull();
    expect(rememberedSplit(splitMemoryKey("g1", "vivi", 0))).toEqual({ R: 3 });
  });
});
