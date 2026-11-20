package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreadhorde Twins — Creature — Zombie Jackal Warrior {3}{R}, 2/2 (War of the Spark, #5520):
//
//	"When this creature enters, amass Zombies 2. (Put two +1/+1 counters on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)
//	 Zombie tokens you control have trample."
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
// The trample grant is Eternal Skylord's static with another keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4fbb8e03-22f1-4171-88e9-877244e25613",
		Name:         "Dreadhorde Twins",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Amass: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dreadhorde Twins — amass Zombies 2", Do(Amass{Subtype: "Zombie", N: 2})),
		},
		Static: []game.StaticAbility{
			KeywordGrant(zombieTokenYouControl, "trample"),
		},
	})
}
