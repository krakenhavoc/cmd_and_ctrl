import { describe, expect, it } from "vitest";
import { blockTaxFor, blockTaxNote, finishBlocksParams } from "./blockTax";
import type { GameView } from "./protocol";

function view(rows: unknown): GameView {
  return {
    seats: [{ id: "a" }, { id: "b" }],
    turn: { block_taxes: rows },
  } as unknown as GameView;
}

describe("block tax (#5375)", () => {
  const row = { seat: 1, cost: "{1}{1}", each: "{1}", labels: ["Archangel of Tithes"] };

  it("finds the viewer's row by seat index", () => {
    expect(blockTaxFor(view([row]), "b")).toEqual(row);
    expect(blockTaxFor(view([row]), "a")).toBeNull();
    expect(blockTaxFor(view(undefined), "b")).toBeNull();
    expect(blockTaxFor(null, "b")).toBeNull();
  });

  it("auto-taps only when the staged blocks are taxed", () => {
    expect(finishBlocksParams(row)).toEqual({ auto_tap: true });
    expect(finishBlocksParams({ ...row, cost: undefined })).toBeUndefined();
    expect(finishBlocksParams(null)).toBeUndefined();
  });

  it("says what finishing costs, or what each blocker will", () => {
    expect(blockTaxNote(row)).toBe("blocking costs {1}{1}, paid with your lands when you finish");
    expect(blockTaxNote({ ...row, cost: undefined })).toBe("each blocker costs {1}");
    expect(blockTaxNote(null)).toBe("");
  });
});

describe("the block staging dock under a block tax (#5375)", () => {
  it("puts the price in the dock's detail", async () => {
    const { blockRequest } = await import("./combatDock");
    const r = blockRequest(
      2,
      () => {},
      undefined,
      "blocking costs {1}{1}, paid with your lands when you finish",
    );
    expect(r.detail).toContain("blocking costs {1}{1}");
    expect(blockRequest(2, () => {}).detail).toBe(
      "click your creature, then the attacker it blocks",
    );
  });
});
