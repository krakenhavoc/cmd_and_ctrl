package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Subversion — Enchantment {3}{B}{B} (Urza's Legacy, S82 OLD b716, #5623):
//
//	"At the beginning of your upkeep, each opponent loses 1 life. You gain life equal to the life lost this way."
//
// Ob Nixilis' drain (B21DrainEachOpponentAndGainTheTotal): every opponent
// loses 1, and the controller gains the life actually lost, summed over the
// table.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0772c5d2-9189-47e2-b0ee-13babd94ea93",
		Name:         "Subversion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Subversion — each opponent loses 1 life; you gain life equal to the life lost this way",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B21DrainEachOpponentAndGainTheTotal(g, item, 1)
				}),
		},
	})
}
