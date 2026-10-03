package effects

// Amok — Enchantment {1}{R}:
//
//	"{1}, Discard a card at random: Put a +1/+1 counter on target creature."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ddbbe65-f928-4b8e-8c4c-a9ce82e2594d",
		Name:         "Amok",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Discard a card at random: Put a +1/+1 counter on target creature.",
			Cost:    Plus(ManaCost("{1}"), DiscardAtRandom(1, "a card at random")),
			Targets: TargetCreature("target creature"),
			Effect:  putPlusOneCounterOnEachLegalTarget,
		}},
	})
}
