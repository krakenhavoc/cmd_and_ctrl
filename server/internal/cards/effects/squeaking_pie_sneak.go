package effects

// Squeaking Pie Sneak — Creature — Goblin Rogue {1}{B}, 2/2:
//
//	"As an additional cost to cast this spell, reveal a Goblin card from your hand or pay {3}.
//	 Fear (This creature can't be blocked except by artifact creatures and/or black creatures.)"
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
		OracleID:        "7243b935-73fc-47d2-ac65-ab390ed94f3f",
		Name:            "Squeaking Pie Sneak",
		Completeness:    CompletenessFull,
		AdditionalCost:  RevealOrPay("a", "Goblin", "{3}"),
		PrintedKeywords: []string{"fear"},
	})
}
