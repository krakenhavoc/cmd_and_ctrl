import { describe, expect, it } from "vitest";

import { xPickerCostNotes } from "./costNotes";
import type { CardView } from "./protocol";

// costNotes.test.ts — the X picker's note for a target-priced spell
// (#746, ADR 0048 addendum open question 2, option a).

function card(over: Partial<CardView> = {}): CardView {
  return {
    instance_id: "fireball",
    name: "Fireball",
    type_line: "Sorcery",
    owner: "me",
    controller: "me",
    ...over,
  } as CardView;
}

const clause = "This spell costs {1} more to cast for each target beyond the first.";

describe("xPickerCostNotes", () => {
  it("quotes the printed clause for a spell cast", () => {
    expect(xPickerCostNotes(card({ target_cost_notes: [clause] }))).toEqual([clause]);
  });

  it("is empty for a spell whose price does not depend on targets", () => {
    expect(xPickerCostNotes(card())).toEqual([]);
  });

  it("is empty for an activated ability's X, which no self modifier prices", () => {
    expect(xPickerCostNotes(card({ target_cost_notes: [clause] }), 0)).toEqual([]);
  });

  it("is empty with no card and drops blank entries", () => {
    expect(xPickerCostNotes(null)).toEqual([]);
    expect(xPickerCostNotes(card({ target_cost_notes: ["", " "] }))).toEqual([]);
  });
});
