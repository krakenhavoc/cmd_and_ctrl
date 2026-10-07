import { describe, expect, it } from "vitest";
import type { CardView } from "./protocol";
import { applyCastChoices, castRevealCost } from "./targeting";

// revealCost.test.ts — ADR 0100 amendment 2026-10-07: a reveal or
// behold branch on the client. The chosen branch drives the one-card
// picker, and the pick rides cast_spell as reveal_ids.

function vanquisher(): CardView {
  return {
    instance_id: "wr",
    name: "Wren's Run Vanquisher",
    owner: "p0",
    controller: "p0",
    additional_cost: {
      label: "Reveal an Elf card from your hand or pay {3}",
      branches: [
        {
          key: "reveal",
          label: "Reveal an Elf card from your hand",
          reveal: true,
          reveal_options: { cards: ["elf"], min: 1, max: 1 },
          payable: true,
        },
        { key: "mana", label: "Pay {3}", mana_cost: "{3}", payable: true },
      ],
    },
  } as CardView;
}

describe("reveal and behold branches", () => {
  it("asks for a card only when the reveal branch is chosen", () => {
    expect(castRevealCost(vanquisher(), { costBranch: 0 })).toEqual({
      label: "Reveal an Elf card from your hand",
      behold: false,
      options: ["elf"],
    });
    expect(castRevealCost(vanquisher(), { costBranch: 1 })).toBeUndefined();
    expect(castRevealCost(vanquisher(), undefined)).toBeUndefined();
  });

  it("marks a behold branch", () => {
    const card = vanquisher();
    card.additional_cost!.branches![0].behold = true;
    expect(castRevealCost(card, { costBranch: 0 })?.behold).toBe(true);
  });

  it("sends reveal_ids on cast_spell only when a card was named", () => {
    const sent: Record<string, unknown> = {};
    applyCastChoices(sent, { costBranch: 0, revealIDs: ["elf"] });
    expect(sent.cost_branch).toBe(0);
    expect(sent.reveal_ids).toEqual(["elf"]);
    const none: Record<string, unknown> = {};
    applyCastChoices(none, { costBranch: 1 });
    expect(none.reveal_ids).toBeUndefined();
  });
});
