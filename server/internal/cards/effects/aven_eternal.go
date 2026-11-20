package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aven Eternal — Creature — Zombie Bird Warrior {2}{U}, 2/2 (War of the Spark, #5520):
//
//	"Flying
//	 When this creature enters, amass Zombies 1. (Put a +1/+1 counter on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)"
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ba7aad10-6897-4e85-a976-1bf9282de5de",
		Name:            "Aven Eternal",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Amass: 1},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Aven Eternal — amass Zombies 1", Do(Amass{Subtype: "Zombie", N: 1})),
		},
	})
}
