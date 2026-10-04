package effects

// Ostracize — Sorcery {B}:
//
//	"Target opponent reveals their hand. You choose a creature card from
//	 it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a creature card may be chosen (CR 701.9b).
// A hand with no creature card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43975c07-103c-4664-b7e6-55bb29d85183",
		Name:         "Ostracize",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    TargetRevealsYouChooseDiscard(Creature(), "creature card"),
	})
}
