package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian of Ghirapur — Creature — Angel {2}{W}, 3/3 (March of the
// Machine, #5477):
//
//	"Flying
//	 When this creature enters, exile up to one other target creature or
//	 artifact you control. Return it to the battlefield under its owner's
//	 control at the beginning of the next end step."
//
// Skyskipper Duo's delayed blink (exileTargetsThenScheduleReturn) with
// Abuelo, Ancestral Echo's target, "creature or artifact you control".
// The card returns as a new object at the next end step (CR 400.7,
// CR 603.7); a token exiled this way does not return. Choosing no target
// exiles nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2553ab57-ceb6-4654-a59d-bc61735da908",
		Name:            "Guardian of Ghirapur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Guardian of Ghirapur — exile up to one other target creature or artifact you control; return it at the beginning of the next end step",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ExileTargetsThenScheduleReturn(effects.NewContext(g, item), "Guardian of Ghirapur — return the exiled card")
					}),
				effects.Another(effects.TargetPermanent("up to one other target creature or artifact you control", effects.Or(effects.Creature(), effects.Artifact()), effects.YouControl()).WithCount(0, 1)),
			),
		},
	})
}
