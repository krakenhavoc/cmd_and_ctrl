package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snapping Creeper — Creature — Plant {2}{G}, 2/3 (Worldwake, #5582):
//
//	"Landfall — Whenever a land you control enters, this creature gains
//	 vigilance until end of turn."
//
// Landfall with a layer-6 grant pinned to the creature until the
// turn ends (thisGainsKeywordUntilEndOfTurn).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bb4e1340-da63-46b8-a5e3-79dfedec3e78",
		Name:         "Snapping Creeper",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Snapping Creeper — gains vigilance until end of turn (landfall)",
				effects.ThisGainsKeywordUntilEndOfTurn("vigilance", "Snapping Creeper — vigilance until end of turn")),
		},
	})
}
