package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleet Swallower — Creature — Fish {5}{U}{U}, 6/6 (Secret Lair Drop, #5462):
//
//	"Whenever this creature attacks, target player mills half their library,
//	 rounded up."
//
// Terisian Mindbreaker's half (S82Bro13MillHalfRoundedUp) read as the ability
// resolves, aimed at a target player of any kind: the controller's own
// library included. A target that has become illegal mills nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52d349e7-4bd2-468a-ab26-df8cd262f6f6",
		Name:         "Fleet Swallower",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Fleet Swallower — target player mills half their library, rounded up",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						player, ok := effects.FirstLegalPlayerTarget(ctx)
						if !ok {
							return nil
						}
						return effects.S82Bro13MillHalfRoundedUp(ctx, player)
					}),
				effects.TargetPlayer("target player")),
		},
	})
}
