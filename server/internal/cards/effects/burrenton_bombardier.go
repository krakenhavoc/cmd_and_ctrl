package effects

// Burrenton Bombardier — Creature — Kithkin Soldier {2}{W}, 2/2
// (Morningtide, #5592):
//
//	"Flying
//	 Reinforce 2—{2}{W} ({2}{W}, Discard this card: Put two +1/+1 counters
//	 on target creature.)"
//
// Printed flying, and the shared reinforce row (morReinforce), which works
// only while the card is in its owner's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "13ec2390-e08a-4645-99e9-bac7e41917bb",
		Name:            "Burrenton Bombardier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{
			morReinforce(2, "{2}{W}", "Reinforce 2—{2}{W} ({2}{W}, Discard this card: Put two +1/+1 counters on target creature.)"),
		},
	})
}
