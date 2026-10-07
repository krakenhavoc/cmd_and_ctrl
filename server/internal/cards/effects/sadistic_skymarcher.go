package effects

// Sadistic Skymarcher — Creature — Vampire Soldier {2}{B}, 2/2:
//
//	"As an additional cost to cast this spell, reveal a Vampire card from your hand or pay {1}.
//	 Flying, lifelink"
//
// The additional cost is the either/or branch cost of ADR 0100 §2
// with the reveal branch added by its 2026-10-07 amendment
// (RevealOrPay): the caster announces the branch, names the
// card on reveal_ids, and either shows it to the table or pays {1}
// more.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e1fe5ec-1669-4587-8e42-b189fd6d39e0",
		Name:            "Sadistic Skymarcher",
		Completeness:    CompletenessFull,
		AdditionalCost:  RevealOrPay("a", "Vampire", "{1}"),
		PrintedKeywords: []string{"flying", "lifelink"},
	})
}
