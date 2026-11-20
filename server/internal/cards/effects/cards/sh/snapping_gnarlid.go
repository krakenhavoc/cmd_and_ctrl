package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Snapping Gnarlid — Creature — Beast {1}{G}, 2/2 (Foundations Jumpstart):
//
//	"Landfall — Whenever a land you control enters, this creature gets
//	 +1/+1 until end of turn."
//
// Attercop's landfall pump, one land one trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eeba70f5-336c-46e6-90db-50818f88e8a5",
		Name:         "Snapping Gnarlid",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Snapping Gnarlid — +1/+1 until end of turn (landfall)",
				effects.ThisGetsUntilEndOfTurn(1, 1, "Snapping Gnarlid — +1/+1 until end of turn")),
		},
	})
}
