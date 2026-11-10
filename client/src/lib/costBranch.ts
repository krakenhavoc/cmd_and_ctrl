// costBranch.ts — #4566 (ADR 0100 amendment 2026-10-30): an activated
// ability whose cost is an either/or ("{3}, {T}, Sacrifice an artifact or
// discard a nonland card"), and since #5397 a mana ability whose cost
// forages (Thornvault Forager). The row ships `cost_branches`; the client asks
// which branch first, then folds the chosen branch into the row and runs
// the ordinary announce chain on it, so the sacrifice or discard picker is
// the one a row printing that clause alone opens. The index rides the
// activation as `cost_branch`.

import type { AbilityCostBranchView, ModeSpecView } from "./protocol";

// BranchedRow is what the fold reads and writes: an activated row, or
// (#5397) a mana row, both of which carry `cost_branches` and the
// sacrifice / discard / graveyard-exile clause fields a branch folds into.
type BranchedRow = {
  cost_branches?: AbilityCostBranchView[];
  sacrifice_label?: string;
  sacrifice_options?: AbilityCostBranchView["sacrifice_options"];
  discard_cost_n?: number;
  discard_cost_label?: string;
  discard_cost_options?: string[];
  exile_cost_n?: number;
  exile_cost_label?: string;
  exile_cost_options?: string[];
  exile_cost_zone?: AbilityCostBranchView["exile_cost_zone"];
};

// abilityCostBranches is the row's branches, or none.
export function abilityCostBranches(a: BranchedRow | null | undefined) {
  return a?.cost_branches ?? [];
}

// withCostBranch is the row with branch `k` folded in: the branch's
// sacrifice, discard or graveyard-exile (forage, #5184) clause on the
// row's own fields, and no branches
// left, so the rest of the chain cannot tell it from a row that prints
// that clause alone. The row unchanged when `k` is out of range.
export function withCostBranch<T extends BranchedRow>(a: T, k: number): T {
  const b = abilityCostBranches(a)[k];
  if (!b) return a;
  const out: T = { ...a, cost_branches: undefined };
  if (b.sacrifice_options) {
    out.sacrifice_label = b.sacrifice_label;
    out.sacrifice_options = b.sacrifice_options;
  }
  if (b.discard_cost_n) {
    out.discard_cost_n = b.discard_cost_n;
    out.discard_cost_label = b.discard_cost_label;
    out.discard_cost_options = b.discard_cost_options;
  }
  if (b.exile_cost_n) {
    out.exile_cost_n = b.exile_cost_n;
    out.exile_cost_label = b.exile_cost_label;
    out.exile_cost_options = b.exile_cost_options;
    out.exile_cost_zone = b.exile_cost_zone;
  }
  return out;
}

// costBranchModeSpec is the branch question in the shape the mode picker
// reads: choose exactly one, a branch the activator can't pay disabled.
export function costBranchModeSpec(a: BranchedRow): ModeSpecView {
  return {
    prompt: "Choose how to pay",
    min: 1,
    max: 1,
    options: abilityCostBranches(a).map((b) => ({ label: b.label, used: !b.payable })),
  };
}

// payableCostBranches is the indexes of the branches the activator can
// pay now.
export function payableCostBranches(a: BranchedRow): number[] {
  const out: number[] = [];
  abilityCostBranches(a).forEach((b, i) => {
    if (b.payable) out.push(i);
  });
  return out;
}
