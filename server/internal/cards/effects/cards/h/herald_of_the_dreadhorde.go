package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herald of the Dreadhorde — Creature — Zombie Warrior {3}{B}, 3/2 (War of the Spark, #5520):
//
//	"When this creature dies, amass Zombies 2. (Put two +1/+1 counters on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9cdbe247-9dd3-4894-8e65-43d879c19321",
		Name:         "Herald of the Dreadhorde",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Herald of the Dreadhorde — amass Zombies 2", effects.Do(effects.Amass{Subtype: "Zombie", N: 2})),
		},
	})
}
