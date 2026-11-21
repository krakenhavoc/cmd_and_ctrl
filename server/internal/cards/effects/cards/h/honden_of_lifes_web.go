package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Honden of Life's Web — Legendary Enchantment — Shrine {4}{G}
// (Champions of Kamigawa, #5603):
//
//	"At the beginning of your upkeep, create a 1/1 colorless Spirit
//	 creature token for each Shrine you control."
//
// The Shrines are counted as the trigger resolves (CR 608.2h); the
// tokens are the shared 1/1 colorless Spirit row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a1fdfcc6-4b3c-40f8-a9e4-dff9ace74a59",
		Name:         "Honden of Life's Web",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChkB1Honden("Honden of Life's Web — create a 1/1 colorless Spirit creature token for each Shrine you control",
				func(g *game.Game, item *game.StackItem, shrines int) error {
					return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Spirit"), N: shrines}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
