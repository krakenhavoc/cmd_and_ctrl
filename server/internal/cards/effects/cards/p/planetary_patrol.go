package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planetary Patrol — Creature — Scout {1}{R} (Star Trek, #4973), 2/2:
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +1/+1 until end of turn."
//
// Attercop's landfall pump without its keywords: one land, one trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fba54c2d-a664-4e4c-b9d4-e24cf2523ab9",
		Name:         "Planetary Patrol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Planetary Patrol — +1/+1 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Planetary Patrol — +1/+1 until end of turn")),
		},
	})
}
