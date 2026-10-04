package effects

// Dark Inquiry — Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a nonland card from
//	 it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a nonland card may be chosen (CR 701.9b).
// A modal double-faced card with a spell front is a nonland card in a
// hand (CR 712.8a).
// A hand with no nonland card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fa04f161-4141-491b-ad6f-a16f69c862f3",
		Name:         "Dark Inquiry",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    TargetRevealsYouChooseDiscard(Nonland(), "nonland card"),
	})
}
