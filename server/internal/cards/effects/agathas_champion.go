package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agatha's Champion — Creature — Human Knight {4}{G}, 4/4 (Wilds of
// Eldraine, #4471):
//
//	"Bargain (You may sacrifice an artifact, enchantment, or token as you
//	 cast this spell.)
//	 Trample
//	 When this creature enters, if it was bargained, it fights up to one
//	 target creature you don't control. (Each deals damage equal to its
//	 power to the other.)"
//
// Bargain is the shared optional cost (g2Bargain), and the intervening if
// (CR 603.4) is Diminisher Witch's: it reads the permanent's own record of
// what was paid, so an unbargained Champion, a reanimated one and a token
// copy put nothing on the stack. The fight is Surly Badgersaur's "up to
// one target creature you don't control": with no target chosen, or a
// target gone by resolution, or a Champion that has left, no damage is
// dealt (CR 701.14b).
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
