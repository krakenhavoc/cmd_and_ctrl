package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrapwork Mutt — Artifact Creature — Dog {2}, 2/1:
//
//	"When this creature enters, you may discard a card. If you do, draw a
//	 card.
//	 Unearth {1}{R}"
func init() {
	const label = "Scrapwork Mutt — you may discard a card; if you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:     "5906b378-c7a7-44fd-a63b-55a347d97854",
		Name:         "Scrapwork Mutt",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{effects.Unearth("{1}{R}")},
	})
}
