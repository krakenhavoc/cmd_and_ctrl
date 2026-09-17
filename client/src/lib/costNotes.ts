import type { CardView } from "./protocol";

// costNotes.ts — #746 (ADR 0048 addendum, open question 2): what the
// X picker says about a spell whose price depends on its targets.
//
// The X picker opens before targeting (ADR 0021 §3), so for Fireball
// ("This spell costs {1} more to cast for each target beyond the
// first") the number of targets is not known yet. The server's
// auto-tap preview is asked without targets and so prices the spell
// at one target. Rather than guess, the picker quotes the printed
// clause under its readout: honest, cheap, and it keeps one cast
// flow. The clauses come off the wire (`target_cost_notes`), so the
// client never parses oracle text.

// xPickerCostNotes returns the clauses to show under the X readout.
// Only a CAST prices by the spell's own targets: an activated
// ability's X picker (abilityIndex set) prices the ability's cost,
// which no self cost modifier touches, so it shows nothing.
export function xPickerCostNotes(card: CardView | null, abilityIndex?: number): string[] {
  if (!card || abilityIndex !== undefined) return [];
  return (card.target_cost_notes ?? []).filter((n) => n.trim() !== "");
}
