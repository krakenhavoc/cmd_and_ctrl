package effects

// Flamekin Bladewhirl — Creature — Elemental Warrior {R}, 2/1:
//
//	"As an additional cost to cast this spell, reveal an Elemental card from your hand or pay {3}."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the reveal branch added by its 2026-10-07 amendment
// (RevealOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {3}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "c37770aa-4770-4198-b6a6-f9ff86dbb06b",
		Name:           "Flamekin Bladewhirl",
		Completeness:   CompletenessFull,
		AdditionalCost: RevealOrPay("an", "Elemental", "{3}"),
	})
}
