package effects

// Daring Buccaneer — Creature — Human Pirate {R}, 2/2:
//
//	"As an additional cost to cast this spell, reveal a Pirate card from your hand or pay {2}."
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the reveal branch added by its 2026-10-07 amendment
// (RevealOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {2}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "1e6f2997-a84b-4302-9d8b-ca4928961d24",
		Name:           "Daring Buccaneer",
		Completeness:   CompletenessFull,
		AdditionalCost: RevealOrPay("a", "Pirate", "{2}"),
	})
}
