package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hedron Rover — Artifact Creature — Construct {4}, 2/2 (Worldwake, #5582):
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// Kazandu Mammoth's landfall pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b2826ea-a3d5-4b8a-b362-0b0880d1cc48",
		Name:         "Hedron Rover",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Hedron Rover — +2/+2 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Hedron Rover — +2/+2 until end of turn")),
		},
	})
}
