package effects

// Duress — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose a noncreature,
//	 nonland card from it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a noncreature, nonland card may be chosen (CR 701.9b).
// A hand with no noncreature, nonland card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "33d405ea-7a9a-4970-b70f-9c05d90dd6f0",
		Name:         "Duress",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    TargetRevealsYouChooseDiscard(And(Noncreature(), Nonland()), "noncreature, nonland card"),
	})
}
