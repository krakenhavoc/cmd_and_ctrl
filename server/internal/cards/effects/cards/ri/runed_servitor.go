package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runed Servitor — Artifact Creature — Construct {2}, 2/2 (Rise of the
// Eldrazi, #5581):
//
//	"When this creature dies, each player draws a card."
//
// Friendly Teddy's dies trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e5e81e1-ab47-4f08-9c42-dba0adc5cf9a",
		Name:         "Runed Servitor",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Runed Servitor — each player draws a card", func(g *game.Game, item *game.StackItem) error {
				return effects.B05EachPlayerDraws(g, item, 1)
			}),
		},
	})
}
