package effects

// On the Job — Instant {2}{W}{W} (Murders at Karlov Manor, #4469):
//
//	"Creatures you control get +2/+1 until end of turn. Investigate.
//	 (Create a Clue token. It's an artifact with "{2}, Sacrifice this
//	 token: Draw a card.")"
//
// Overrun's team pump, the set of creatures locked as it resolves
// (CR 611.2c), so the Clue and any creature that arrives later are not
// pumped; then one Clue (CR 701.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4733889-4d4c-4bf7-8a9b-fb5ad7a38395",
		Name:         "On the Job",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Toughness: 1, Label: "On the Job — +2/+1"},
			Investigate{N: 1},
		),
	})
}
