package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hedron Scrabbler — Artifact Creature — Construct {2}, 1/1 (Zendikar, #5583):
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +1/+1 until end of turn."
//
// Canopy Baloth's landfall pump: a land entering under its controller's
// control, the boost on the creature while it is still the same object
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "180d9c48-6d6c-4d20-965a-0ee2ad57f99b",
		Name:         "Hedron Scrabbler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Hedron Scrabbler — +1/+1 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Hedron Scrabbler — +1/+1 until end of turn")),
		},
	})
}
