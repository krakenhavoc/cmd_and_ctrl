package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nobody — {1}{U/R}{U/R} Artifact Creature — Human Hero 3/2 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"When this creature enters, return up to one other target artifact you
//	 control to its owner's hand. Scry 1."
//
// Rescue, Pepper Potts's "up to one other target" enters trigger, narrowed
// to artifacts, followed by scry 1. With no target chosen the trigger still
// resolves and scries; a chosen target that is no longer legal is not
// returned (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eddab60f-2eac-4f20-934c-3d9307f885c3",
		Name:         "Nobody",
		Purpose:      game.Purpose{Scry: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Nobody — return up to one other target artifact you control to its owner's hand; scry 1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := effects.B35BounceChosenPermanents(ctx); err != nil {
						return err
					}
					return effects.Scry{N: 1}.Apply(ctx)
				}),
				effects.Another(effects.TargetPermanent("up to one other target artifact you control", effects.Artifact(), effects.YouControl()).WithCount(0, 1))),
		},
	})
}
