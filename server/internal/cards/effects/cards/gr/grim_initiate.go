package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Initiate — Creature — Zombie Warrior {R}, 1/1 (War of the Spark, #5520):
//
//	"First strike
//	 When this creature dies, amass Zombies 1. (Put a +1/+1 counter on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "96264dde-c0b6-4912-afc4-18757f938a4d",
		Name:            "Grim Initiate",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Amass: 1},
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Grim Initiate — amass Zombies 1", effects.Do(effects.Amass{Subtype: "Zombie", N: 1})),
		},
	})
}
