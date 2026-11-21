package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shivan Raptor — Creature — Dinosaur, {2}{R}, 3/1:
//
//	"First strike, haste
//	 Echo {2}{R} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bc01d36b-dea3-47d4-afdf-0dbbba7ca7b2",
		Name:            "Shivan Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.Echo("Shivan Raptor", "{2}{R}"),
		},
	})
}
