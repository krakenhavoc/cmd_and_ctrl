// attackSelect.test.ts — #4326: several attackers selected, one target.

import { describe, expect, it } from "vitest";

import { dropFromSelection, joinsSelection, selectionName, stageSelectionAt } from "./attackSelect";

const sel = (cardID: string, exert?: boolean | null) =>
  exert === undefined
    ? { kind: "attacker" as const, cardID }
    : { kind: "attacker" as const, cardID, exert };
const none = () => false;

describe("joinsSelection", () => {
  it("adds a second undeclared creature that asks nothing", () => {
    expect(joinsSelection(sel("a"), sel("b"), none)).toBe(true);
  });

  it("starts over with nothing selected, or the same creature", () => {
    expect(joinsSelection(null, sel("b"), none)).toBe(false);
    expect(joinsSelection(sel("a"), sel("a"), none)).toBe(false);
  });

  it("keeps a creature with an exert question alone", () => {
    expect(joinsSelection(sel("a", null), sel("b"), none)).toBe(false);
    expect(joinsSelection(sel("a"), sel("b", null), none)).toBe(false);
    expect(joinsSelection(sel("a", false), sel("b"), none)).toBe(false);
  });

  it("keeps a declared attacker's re-point alone", () => {
    expect(joinsSelection(sel("a"), sel("b"), (id) => id === "a")).toBe(false);
    expect(joinsSelection(sel("a"), sel("b"), (id) => id === "b")).toBe(false);
  });
});

describe("dropFromSelection", () => {
  it("takes out one of the others", () => {
    expect(dropFromSelection(sel("c"), ["a", "b"], "a")).toEqual({
      primary: sel("c"),
      more: ["b"],
    });
  });

  it("promotes the last of the others when the primary goes", () => {
    expect(dropFromSelection(sel("c"), ["a", "b"], "c")).toEqual({
      primary: sel("b"),
      more: ["a"],
    });
    expect(dropFromSelection(sel("c"), [], "c")).toEqual({ primary: null, more: [] });
  });
});

describe("stageSelectionAt", () => {
  const may = (id: string, target: string) => !(id === "b" && target === "bob");

  it("stages every one that may attack the target, oldest first, and leaves the rest", () => {
    const r = stageSelectionAt([{ attacker: "z", target: "cy" }], sel("c"), ["a", "b"], "bob", may);
    expect(r).toEqual({
      stage: [
        { attacker: "z", target: "cy" },
        { attacker: "a", target: "bob" },
        { attacker: "c", target: "bob" },
      ],
      left: ["b"],
    });
  });

  it("re-aims one already staged instead of staging it twice", () => {
    const r = stageSelectionAt([{ attacker: "a", target: "cy" }], sel("a"), [], "bob", may);
    expect(r?.stage).toEqual([{ attacker: "a", target: "bob" }]);
  });

  it("keeps the primary's exert answer, and stages nothing while it is open", () => {
    expect(stageSelectionAt([], sel("a", true), [], "bob", may)?.stage).toEqual([
      { attacker: "a", target: "bob", exert: true },
    ]);
    expect(stageSelectionAt([], sel("a", null), [], "bob", may)).toBeNull();
  });
});

describe("selectionName", () => {
  it("names the last creature and counts the others", () => {
    expect(selectionName("Wolf", 0)).toBe("Wolf");
    expect(selectionName("Wolf", 2)).toBe("Wolf and 2 more");
  });
});
