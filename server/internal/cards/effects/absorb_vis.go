package effects

// Absorb Vis — Sorcery {6}{B} (Conflux, #5587):
//
//	"Target player loses 4 life and you gain 4 life.
//	 Basic landcycling {1}{B} ({1}{B}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// Soul Feast's drain at 4, through the shared
// targetPlayerLosesLifeYouGain body, with the engine's basic landcycling
// row (CR 702.29e), which works only while the card is in its owner's
// hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04ea7e88-7ea5-4b4c-9a7c-2a6ac0bc0682",
		Name:         "Absorb Vis",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve:    targetPlayerLosesLifeYouGain(4),
		Activated:    []ActivatedAbility{BasicLandcycling("{1}{B}")},
	})
}
