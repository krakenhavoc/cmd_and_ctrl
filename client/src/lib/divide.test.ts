// #1563, CR 601.2d — "divided as you choose". The pure half of the
// picker: the amount a clause divides (the client twin of
// game.(*DivideSpec).TotalFor), the even split it opens with (the bot's
// game.EvenDistribution), the rules Confirm is gated on (the engine's
// settleDistribution), and the walk plumbing that carries the division
// to the action.

import { describe, expect, it } from "vitest";

import type { CardView, LegalTargetsView } from "./protocol";
import {
  advance,
  distributionOf,
  divideTotal,
  divisionProblem,
  evenSplit,
  needsDivision,
  openWalk,
  stepsFor,
  togglePick,
  withDivision,
} from "./targeting";

describe("divideTotal", () => {
  it("reads a fixed amount", () => {
    expect(divideTotal({ total: 4 }, 9)).toBe(4);
  });
  it("reads X, doubling from the threshold (Shatterskull Smashing)", () => {
    const d = { from_x: true, double_from_x: 6 };
    expect(divideTotal(d, 5)).toBe(5);
    expect(divideTotal(d, 6)).toBe(12);
    expect(divideTotal({ from_x: true }, undefined)).toBe(0);
  });
});

describe("evenSplit", () => {
  it("gives the remainder to the earliest picks", () => {
    expect(evenSplit(["a", "b", "c"], 5)).toEqual({ a: 2, b: 2, c: 1 });
    expect(evenSplit(["a", "b"], 4)).toEqual({ a: 2, b: 2 });
  });
  it("is always a division the gate accepts when it can be", () => {
    const ids = ["a", "b", "c"];
    expect(divisionProblem(ids, 3, evenSplit(ids, 3))).toBeNull();
  });
});

describe("divisionProblem", () => {
  const ids = ["a", "b"];
  it("refuses a zero share", () => {
    expect(divisionProblem(ids, 4, { a: 4, b: 0 })).toMatch(/at least 1/);
  });
  it("refuses a sum that is short or over", () => {
    expect(divisionProblem(ids, 4, { a: 1, b: 2 })).toMatch(/assign 4/);
    expect(divisionProblem(ids, 4, { a: 3, b: 2 })).toMatch(/assign 4/);
  });
  it("refuses more targets than the amount", () => {
    expect(divisionProblem(["a", "b", "c"], 2, { a: 1, b: 1, c: 0 })).toMatch(/can't share/);
  });
  it("accepts an uneven split that adds up", () => {
    expect(divisionProblem(ids, 4, { a: 3, b: 1 })).toBeNull();
  });
});

const card = { instance_id: "fury", name: "Fury" } as CardView;
const fury: LegalTargetsView = {
  cards: ["a", "b", "c", "d", "e"],
  min: 0,
  max: 0,
  divide: { total: 4 },
};

describe("the target walk", () => {
  it("carries the step's amount, resolved against X", () => {
    const smash: LegalTargetsView = {
      cards: ["a", "b"],
      min: 0,
      max: 2,
      divide: { from_x: true, double_from_x: 6 },
    };
    expect(stepsFor("creature", smash, undefined, 0, { xValue: 7 })[0].divide).toBe(14);
    expect(
      stepsFor("creature", { cards: ["a"], min: 1, max: 1 }, undefined, 0)[0].divide,
    ).toBeUndefined();
  });

  it("refuses a pick past the amount, even under an unbounded max", () => {
    let t = openWalk(card, stepsFor("creature", fury, undefined, 0), {});
    for (const id of ["a", "b", "c", "d", "e"]) t = togglePick(t, { kind: "card", id });
    expect(t.picked.map((p) => p.id)).toEqual(["a", "b", "c", "d"]);
  });

  it("asks for a division only with two or more picks", () => {
    let t = openWalk(card, stepsFor("creature", fury, undefined, 0), {});
    t = togglePick(t, { kind: "card", id: "a" });
    expect(needsDivision(t)).toBe(false);
    t = togglePick(t, { kind: "card", id: "b" });
    expect(needsDivision(t)).toBe(true);
  });

  it("keeps the division through advance and sends it", () => {
    let t = openWalk(card, stepsFor("creature", fury, undefined, 0), {});
    t = togglePick(togglePick(t, { kind: "card", id: "a" }), { kind: "card", id: "b" });
    t = withDivision(t, { a: 3, b: 1 });
    expect(advance(t)).toBeNull(); // one step: the walk is done
    expect(distributionOf(t)).toEqual({ a: 3, b: 1 });
  });

  it("sends nothing when nothing was divided", () => {
    const t = openWalk(card, stepsFor("creature", fury, undefined, 0), {});
    expect(distributionOf(t)).toBeUndefined();
  });
});
