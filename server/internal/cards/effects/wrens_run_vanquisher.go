package effects

// Wren's Run Vanquisher — Creature — Elf Warrior {1}{G}, 3/3:
//
//	"As an additional cost to cast this spell, reveal an Elf card from your hand or pay {3}.
//	 Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)"
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
		OracleID:        "fe9cdd15-a390-4a1e-bae9-474ce8d355b8",
		Name:            "Wren's Run Vanquisher",
		Completeness:    CompletenessFull,
		AdditionalCost:  RevealOrPay("an", "Elf", "{3}"),
		PrintedKeywords: []string{"deathtouch"},
	})
}
