package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coliseum Behemoth — Creature — Beast {5}{G}{G}, 7/7 (Final Fantasy,
// #4462):
//
//	"Trample
//	 When this creature enters, choose one —
//	 • Destroy target artifact or enchantment.
//	 • Draw a card."
//
// A modal enters trigger (CR 603.3c): the mode is picked as the trigger
// goes on the stack, and the destroy bullet can't be picked with no
// artifact or enchantment to target. A target gone in response leaves
// the destroy with nothing to do (CR 608.2b). Trample rides
// PrintedKeywords.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Coliseum Behemoth — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeDoing("Destroy target artifact or enchantment.",
			TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			DestroyTheModesTarget),
		ModeWithPurpose(ModeSteps("Draw a card.", nil, DrawCards{N: 1}), game.Purpose{Draws: 1}),
	)
	Register(Spec{
		OracleID:        "076cf4f1-b3d4-4175-b976-938ad851cc0e",
		Name:            "Coliseum Behemoth",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
