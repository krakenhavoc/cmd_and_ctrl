package effects

// Pelakka Predation // Pelakka Caverns — modal double-faced card. This
// file is the FRONT face, Sorcery {2}{B}:
//
//	"Target opponent reveals their hand. You choose a card from it with
//	 mana value 3 or greater. That player discards that card."
//
// The back face, Pelakka Caverns, is registered with the MDFC land
// cycle in mdfc_lands.go under "<oracle>#1".
//
// The revealed-hand pick (ADR 0116) filtered by ManaValueGE(3), which
// reads each card as it is in the hand: a modal double-faced card by
// its front face (CR 712.8a, and this card's own 2020-09-25 ruling), a
// split card by its combined cost (CR 709.4b), and X as 0 (CR 202.3e,
// the same ruling). A land is mana value 0 and can never be chosen. A
// hand with nothing of mana value 3 or greater is revealed and nothing
// is discarded (CR 609.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b0fd6889-20b4-439b-aa97-2e90aca1675a",
		Name:         "Pelakka Predation",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve:    TargetRevealsYouChooseDiscard(ManaValueGE(3), "card with mana value 3 or greater"),
	})
}
