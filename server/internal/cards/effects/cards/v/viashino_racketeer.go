package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Racketeer — Creature — Lizard Rogue {2}{R}, 2/1 (Return to
// Ravnica, #5566):
//
//	"When this creature enters, you may discard a card. If you do, draw a
//	 card."
//
// Fissure Wizard's enters loot: the shared discard-then-draw prompt, where
// the draw follows only a discard that happened. Declining, or an empty
// hand, draws nothing.
//
// No simplification.
func init() {
	const racketeerLabel = "Viashino Racketeer — you may discard a card. If you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "18b09c2d-20cb-40e9-8209-8a1ec2edde97",
		Name:         "Viashino Racketeer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(racketeerLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, racketeerLabel, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
