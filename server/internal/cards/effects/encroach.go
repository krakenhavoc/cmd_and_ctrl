package effects

// Encroach — Sorcery {B}:
//
//	"Target player reveals their hand. You choose a nonbasic land card
//	 from it. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a nonbasic land card may be chosen (CR 701.9b).
// A land card without the basic supertype (CR 205.4c, and the
// 2004-10-04 ruling): a Snow-Covered Swamp is basic and can't be
// chosen, a shock land with basic land types can.
// A hand with no nonbasic land card is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9959d63-5489-4ed1-a04e-9d6ba4a5f68e",
		Name:         "Encroach",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    TargetRevealsYouChooseDiscard(NonbasicLand(), "nonbasic land card"),
	})
}
