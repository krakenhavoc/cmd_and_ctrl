package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agatha's Champion — Creature {4}{G}, 4/3 Human Knight (Wilds of
// Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Trample
//	 When this creature enters, if it was bargained, it fights up to one
//	 target creature you don't control."
//
// Diminisher Witch's shape: bargain is the shared optional cost, and the
// intervening "if" (CR 603.4) reads the permanent's own carried record of
// what was paid, so an unbargained Champion, a reanimated one and a token
// copy put nothing on the stack and nobody is asked for a target. The
// fight is Affectionate Indrik's (CR 701.14): up to one target, and with
// none chosen, or the Champion gone, nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "debdfab7-454f-4fec-9ee8-aa5b570ea7fb",
		Name:            "Agatha's Champion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		OptionalCosts:   []game.AdditionalCost{g2Bargain()},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, diminisherWitchWasBargained),
					"Agatha's Champion — it fights up to one target creature you don't control",
					b30SourceFightsFirstLegalTarget),
				TargetCreature("up to one target creature you don't control", OpponentControls()).WithCount(0, 1)),
		},
	})
}
