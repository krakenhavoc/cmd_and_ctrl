package effects

// Despise — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose a creature or
//	 planeswalker card from it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a creature or planeswalker card may be chosen (CR 701.9b).
// A hand with no creature or planeswalker card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b13b40f1-1b1f-40e8-989b-ef24dc05007a",
		Name:         "Despise",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    TargetRevealsYouChooseDiscard(Or(Creature(), Planeswalker()), "creature or planeswalker card"),
	})
}
