package effects

// Pilfer — Sorcery {1}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a nonland card may be chosen (CR 701.9b).
// Dark Inquiry's text at a lower cost.
// A hand with no nonland card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d13f0907-de39-4a90-940a-831740d7aa9b",
		Name:         "Pilfer",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    TargetRevealsYouChooseDiscard(Nonland(), "nonland card"),
	})
}
