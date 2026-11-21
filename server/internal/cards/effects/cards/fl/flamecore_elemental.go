package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flamecore Elemental — Creature — Elemental, {2}{R}{R}, 5/4:
//
//	"Echo {2}{R}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad3a649f-1a79-45c7-b761-9d7a26cdd0c1",
		Name:         "Flamecore Elemental",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Echo("Flamecore Elemental", "{2}{R}{R}"),
		},
	})
}
