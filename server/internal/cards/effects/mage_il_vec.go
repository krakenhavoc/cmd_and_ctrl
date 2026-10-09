package effects

// Mage il-Vec — Creature — Human Wizard {2}{R}, 2/2:
//
//	"{T}, Discard a card at random: This creature deals 1 damage to any
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
		OracleID:     "adea7db6-634c-4c5d-be40-264b4acffc53",
		Name:         "Mage il-Vec",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Discard a card at random: This creature deals 1 damage to any target.",
			Cost:    Plus(TapCost(), DiscardAtRandom(1, "a card at random")),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  sourceDealsDamageToEachLegalTarget(1),
		}},
	})
}
