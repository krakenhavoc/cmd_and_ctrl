package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravaging Horde — Creature — Human Soldier {3}{R}{R}, 3/3 (Portal Three Kingdoms, S82 OLD b713, #5622):
//
//	"When this creature enters, destroy target land."
//
// Avalanche Riders' trigger: one target clause, chosen as the trigger is
// put on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b775b83c-0c3d-437d-9d3c-1d5c9cd418a8",
		Name:         "Ravaging Horde",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Ravaging Horde — destroy target land", effects.DestroyFirstLegalCardTarget),
				effects.TargetPermanent("target land", effects.Land())),
		},
	})
}
