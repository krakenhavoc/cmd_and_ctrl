package effects

// Canyon Drake — Creature — Drake {2}{R}{R}, 1/2:
//
//	"Flying
//	 {1}, Discard a card at random: This creature gets +2/+0 until end
//	 of turn."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5b46c99-1cc1-465a-a831-3fd6665ef560",
		Name:            "Canyon Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:  "{1}, Discard a card at random: This creature gets +2/+0 until end of turn.",
			Cost:   Plus(ManaCost("{1}"), DiscardAtRandom(1, "a card at random")),
			Effect: thisGetsUntilEndOfTurn(2, 0, "Canyon Drake — +2/+0"),
		}},
	})
}
