package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geyser Glider — Creature — Elemental Beast {3}{R}{R}, 4/4 (Zendikar, #5583):
//
//	"Landfall — Whenever a land you control enters, this creature gains
//	 flying until end of turn."
//
// Landfall grants flying to the creature itself until end of turn, a
// layer-6 effect pinned to the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "840d1002-abf4-4d05-945a-05c3fe9a2ebd",
		Name:         "Geyser Glider",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Geyser Glider — gains flying until end of turn (landfall)",
				effects.ThisGainsKeywordUntilEndOfTurn("flying", "Geyser Glider — gains flying until end of turn")),
		},
	})
}
