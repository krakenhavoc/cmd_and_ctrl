package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keldon Raider — Creature — Human Warrior {2}{R}{R}, 4/3:
//
//	"When this creature enters, you may discard a card. If you do, draw a
//	 card."
//
// Battlefield Scavenger's optional discard whose draw follows only from
// the discard (b39MayDiscardThenDraw).
//
// No simplification.
func init() {
	const label = "Keldon Raider — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "004f82f0-74cc-4bc7-ab22-ab1a5247bc29",
		Name:         "Keldon Raider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawAsManyAsDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
