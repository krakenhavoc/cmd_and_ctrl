package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lazotep Reaver — Creature — Zombie Beast {1}{B}, 1/2 (War of the Spark, #5520):
//
//	"When this creature enters, amass Zombies 1. (Put a +1/+1 counter on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85643f9b-49ea-434b-9f27-c1431064345e",
		Name:         "Lazotep Reaver",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Lazotep Reaver — amass Zombies 1", effects.Do(effects.Amass{Subtype: "Zombie", N: 1})),
		},
	})
}
