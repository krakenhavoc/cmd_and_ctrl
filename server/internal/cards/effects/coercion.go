package effects

// Coercion — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a card from it.
//	 That player discards that card."
//
// The revealed-hand pick (ADR 0116) with no filter: the whole table
// sees the hand (CR 701.20a) and you choose any card in it, a land
// included (CR 701.9b). An empty hand reveals nothing and nothing is
// discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68413337-ddb9-46f8-8c8e-f3d2d672c652",
		Name:         "Coercion",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    TargetRevealsYouChooseDiscard(nil, "card"),
	})
}
