package effects

// Inquisition of Kozilek — Sorcery {B}:
//
//	"Target player reveals their hand. You choose a nonland card from
//	 it with mana value 3 or less. That player discards that card."
//
// The revealed-hand pick (ADR 0116): the whole table sees the hand
// (CR 701.20a), and only a nonland card with mana value 3 or less may be chosen (CR 701.9b).
// Mana value is read as the card is in the hand: X is 0 (CR 202.3e)
// and a split card is its combined cost (CR 709.4b).
// A hand with no nonland card with mana value 3 or less is revealed and nothing is discarded
// (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "395611fa-aec2-4c9e-92c3-70cf95cc00f4",
		Name:         "Inquisition of Kozilek",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    TargetRevealsYouChooseDiscard(And(Nonland(), ManaValueLE(3)), "nonland card with mana value 3 or less"),
	})
}
