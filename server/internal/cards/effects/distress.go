package effects

// Distress — Sorcery {B}{B}:
//
//	"Target player reveals their hand. You choose a nonland card from
//	 it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a nonland card may be chosen (CR 701.9b).
// Targeting yourself reveals your hand to everyone else too (the
// 2014-02-01 ruling).
// A hand with no nonland card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7bb1b48b-aa6b-4064-9355-c43b0f509388",
		Name:         "Distress",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    TargetRevealsYouChooseDiscard(Nonland(), "nonland card"),
	})
}
