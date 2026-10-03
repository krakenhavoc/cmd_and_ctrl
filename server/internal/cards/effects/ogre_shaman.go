package effects

// Ogre Shaman — Creature — Ogre Shaman {3}{R}{R}, 3/3:
//
//	"{2}, Discard a card at random: This creature deals 2 damage to any
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
		OracleID:     "04c7bb20-6e40-4e79-a0ec-ced920d3491e",
		Name:         "Ogre Shaman",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Discard a card at random: This creature deals 2 damage to any target.",
			Cost:    Plus(ManaCost("{2}"), DiscardAtRandom(1, "a card at random")),
			Targets: TargetAny(),
			Effect:  sourceDealsDamageToEachLegalTarget(2),
		}},
	})
}
