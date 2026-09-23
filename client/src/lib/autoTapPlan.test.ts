import { describe, expect, it } from "vitest";

import type { AutoTapPreview } from "./api";
import { paymentOf, planRows, planSummary } from "./autoTapPlan";

// #1285: the preview must show every source the payment spends — a
// Spirit Guide out of the hand included — and say what paying with
// each one costs.

const board = new Map([
  ["m1", "Mountain"],
  ["m2", "Mountain"],
]);
const nameOf = (id: string) => board.get(id);

describe("planRows", () => {
  it("names a hand source the battlefield lookup cannot find", () => {
    const preview: AutoTapPreview = {
      ok: true,
      cost: "{R}{R}{R}",
      plan: ["m1", "gold", "ssg"],
      sources: [
        { card_id: "m1", name: "Mountain", zone: "battlefield", tap: true },
        { card_id: "gold", name: "Gold", zone: "battlefield", sacrifice: true },
        { card_id: "ssg", name: "Simian Spirit Guide", zone: "hand", exile: true },
      ],
    };
    const rows = planRows(preview, nameOf);
    expect(rows.map((r) => r.name)).toEqual(["Mountain", "Gold", "Simian Spirit Guide"]);
    expect(rows.map((r) => r.payment)).toEqual(["tap", "sacrifice", "exile"]);
    expect(rows[2].fromHand).toBe(true);
    expect(rows.map((r) => r.gone)).toEqual([false, true, true]);
  });

  it("falls back to the board name and a tap for an undescribed plan", () => {
    const rows = planRows({ ok: true, cost: "{2}", plan: ["m1", "m2"] }, nameOf);
    expect(rows.map((r) => [r.name, r.payment])).toEqual([
      ["Mountain", "tap"],
      ["Mountain", "tap"],
    ]);
  });

  it("reads a Treasure as tap + sacrifice", () => {
    expect(paymentOf({ card_id: "t", tap: true, sacrifice: true })).toBe("tap_sacrifice");
  });
});

describe("planSummary", () => {
  it("counts taps and names every card that leaves the hand", () => {
    const rows = planRows(
      {
        ok: true,
        cost: "{2}{R}",
        plan: ["m1", "m2", "ssg"],
        sources: [
          { card_id: "m1", name: "Mountain", zone: "battlefield", tap: true },
          { card_id: "m2", name: "Mountain", zone: "battlefield", tap: true },
          { card_id: "ssg", name: "Simian Spirit Guide", zone: "hand", exile: true },
        ],
      },
      nameOf,
    );
    expect(planSummary(rows)).toBe(
      "Taps 2 permanents and exiles Simian Spirit Guide from your hand.",
    );
  });

  it("names what is sacrificed", () => {
    const rows = planRows(
      {
        ok: true,
        cost: "{2}",
        plan: ["gold", "spawn"],
        sources: [
          { card_id: "gold", name: "Gold", sacrifice: true },
          { card_id: "spawn", name: "Eldrazi Spawn", sacrifice: true },
        ],
      },
      nameOf,
    );
    expect(planSummary(rows)).toBe("Sacrifices Gold and Eldrazi Spawn.");
  });
});
