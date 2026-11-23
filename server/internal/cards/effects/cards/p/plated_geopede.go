package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plated Geopede — Creature — Insect {1}{R}, 1/1 (Zendikar, #5583):
//
//	"First strike
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// First strike is the printed keyword. Canopy Baloth's landfall pump: a
// land entering under its controller's control, the boost on the
// creature while it is still the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d1efc4fd-2978-4434-8fdf-ddc05484ebb1",
		Name:            "Plated Geopede",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Plated Geopede — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Plated Geopede — +2/+2 until end of turn")),
		},
	})
}
