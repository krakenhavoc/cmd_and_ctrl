// costBranch.ts — #4566 (ADR 0100 amendment 2026-10-30): an activated
// ability whose cost is an either/or ("{3}, {T}, Sacrifice an artifact or
// discard a nonland card"). The row ships `cost_branches`; the client asks
// which branch first, then folds the chosen branch into the row and runs
// the ordinary announce chain on it, so the sacrifice or discard picker is
// the one a row printing that clause alone opens. The index rides the
// activation as `cost_branch`.

import type { ActivatedAbilityView, ModeSpecView } from "./protocol";

// abilityCostBranches is the row's branches, or none.
export function abilityCostBranches(a: ActivatedAbilityView | null | undefined) {
  return a?.cost_branches ?? [];
}

// withCostBranch is the row with branch `k` folded in: the branch's
// sacrifice or discard clause on the row's own fields, and no branches
// left, so the rest of the chain cannot tell it from a row that prints
// that clause alone. The row unchanged when `k` is out of range.
export function withCostBranch(a: ActivatedAbilityView, k: number): ActivatedAbilityView {
  const b = abilityCostBranches(a)[k];
  if (!b) return a;
  const out: ActivatedAbilityView = { ...a, cost_branches: undefined };
  if (b.sacrifice_options) {
    out.sacrifice_label = b.sacrifice_label;
    out.sacrifice_options = b.sacrifice_options;
  }
  if (b.discard_cost_n) {
    out.discard_cost_n = b.discard_cost_n;
    out.discard_cost_label = b.discard_cost_label;
    out.discard_cost_options = b.discard_cost_options;
  }
  return out;
}

// costBranchModeSpec is the branch question in the shape the mode picker
// reads: choose exactly one, a branch the activator can't pay disabled.
export function costBranchModeSpec(a: ActivatedAbilityView): ModeSpecView {
  return {
    prompt: "Choose how to pay",
    min: 1,
    max: 1,
    options: abilityCostBranches(a).map((b) => ({ label: b.label, used: !b.payable })),
  };
}

// payableCostBranches is the indexes of the branches the activator can
// pay now.
export function payableCostBranches(a: ActivatedAbilityView): number[] {
  const out: number[] = [];
  abilityCostBranches(a).forEach((b, i) => {
    if (b.payable) out.push(i);
  });
  return out;
}
