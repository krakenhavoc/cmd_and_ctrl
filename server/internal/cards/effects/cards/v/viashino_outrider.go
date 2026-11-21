package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Outrider — Creature — Lizard, {2}{R}, 4/3:
//
//	"Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e81c22e-52ec-4142-9a98-88681d04ce60",
		Name:         "Viashino Outrider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Viashino Outrider", "{2}{R}"),
		},
	})
}
