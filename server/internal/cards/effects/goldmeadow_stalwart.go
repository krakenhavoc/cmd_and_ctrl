package effects

// Goldmeadow Stalwart — Creature — Kithkin Soldier {W}, 2/2:
//
//	"As an additional cost to cast this spell, reveal a Kithkin card from your hand or pay {3}."
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
		OracleID:       "1aff8653-12c5-4549-915f-ac2fd6dfd43b",
		Name:           "Goldmeadow Stalwart",
		Completeness:   CompletenessFull,
		AdditionalCost: RevealOrPay("a", "Kithkin", "{3}"),
	})
}
