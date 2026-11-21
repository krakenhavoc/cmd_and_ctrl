package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	enters := effects.WhenThisEnters("Coliseum Behemoth — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("Destroy target artifact or enchantment.",
			effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			effects.DestroyTheModesTarget),
		effects.ModeWithPurpose(effects.ModeSteps("Draw a card.", nil, effects.DrawCards{N: 1}), game.Purpose{Draws: 1}),
	)
	effects.Register(effects.Spec{
		OracleID:        "076cf4f1-b3d4-4175-b976-938ad851cc0e",
		Name:            "Coliseum Behemoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
