package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Exiled Boggart — Creature — Goblin Rogue {1}{B}, 2/2 (Lorwyn):
//
//	"When this creature dies, discard a card."
//
// Mandatory, and the controller chooses the card (CR 701.9a); with an empty
// hand nothing is discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c07878f-348c-4fd9-955f-87e975ed61fa",
		Name:         "Exiled Boggart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Exiled Boggart — discard a card", func(g *game.Game, item *game.StackItem) error {
				return effects.DiscardOne(g, item, item.Controller)
			}),
		},
	})
}
