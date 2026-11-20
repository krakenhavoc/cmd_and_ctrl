package effects

// Mosquito Guard — Creature — Kithkin Soldier {W}, 1/1 (Morningtide,
// #5592):
//
//	"First strike
//	 Reinforce 1—{1}{W} ({1}{W}, Discard this card: Put a +1/+1 counter on
//	 target creature.)"
//
// Printed first strike, and the shared reinforce row (morReinforce), which
// works only while the card is in its owner's hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6beaeec2-b4e9-4904-b817-c0118549d4a9",
		Name:            "Mosquito Guard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated: []ActivatedAbility{
			morReinforce(1, "{1}{W}", "Reinforce 1—{1}{W} ({1}{W}, Discard this card: Put a +1/+1 counter on target creature.)"),
		},
	})
}
