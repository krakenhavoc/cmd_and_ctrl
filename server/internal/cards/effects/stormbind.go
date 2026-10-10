package effects

// Stormbind — Enchantment {1}{R}{G}:
//
//	"{2}, Discard a card at random: This enchantment deals 2 damage to any
//	 target."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "78f50668-36fc-4911-84f7-93667436b0c7",
		Name:         "Stormbind",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Discard a card at random: This enchantment deals 2 damage to any target.",
			Cost:    Plus(ManaCost("{2}"), DiscardAtRandom(1, "a card at random")),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
