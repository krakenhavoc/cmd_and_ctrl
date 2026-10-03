package effects

// Pardic Swordsmith — Creature — Dwarf {2}{R}, 1/1:
//
//	"{R}, Discard a card at random: This creature gets +2/+0 until end of
//	 turn."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "946d770f-0a59-415b-b334-8f7353b96046",
		Name:         "Pardic Swordsmith",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "{R}, Discard a card at random: This creature gets +2/+0 until end of turn.",
			Cost:   Plus(ManaCost("{R}"), DiscardAtRandom(1, "a card at random")),
			Effect: thisGetsUntilEndOfTurn(2, 0, "Pardic Swordsmith — +2/+0"),
		}},
	})
}
