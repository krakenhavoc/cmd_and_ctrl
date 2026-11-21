package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glint-Sleeve Artisan — Creature — Dwarf Artificer {2}{W}, 2/2 (Kaladesh, #5539):
//
//	"Fabricate 1 (When this creature enters, put a +1/+1 counter on it
//	 or create a 1/1 colorless Servo artifact creature token.)"
//
// Fabricate is the shared keyword constructor (fabricate_keyword.go,
// CR 702.123a): the enters trigger asks counters or Servos as it
// resolves, and makes the Servos without asking once this creature has
// left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5f35577-3d94-4835-bfca-3090cb0c63ff",
		Name:         "Glint-Sleeve Artisan",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Fabricate("Glint-Sleeve Artisan", 1)},
	})
}
