package effects

// Pardic Lancer — Creature — Human Barbarian {4}{R}, 3/2:
//
//	"Discard a card at random: This creature gets +1/+0 and gains first
//	 strike until end of turn."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1bb2b090-19a4-40d9-823a-671f6eaca9f9",
		Name:         "Pardic Lancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:  "Discard a card at random: This creature gets +1/+0 and gains first strike until end of turn.",
			Cost:   DiscardAtRandom(1, "a card at random"),
			Effect: thisCreatureUntilEOT("Pardic Lancer — +1/+0 and first strike", 1, 0, "first strike"),
		}},
	})
}
