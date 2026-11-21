package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vizier of the Scorpion — Creature — Zombie Wizard {2}{B}, 1/1 (War of the Spark, #5520):
//
//	"When this creature enters, amass Zombies 1. (Put a +1/+1 counter on an Army you control. It's also a Zombie. If you don't control an Army, create a 0/0 black Zombie Army creature token first.)
//	 Zombie tokens you control have deathtouch."
//
// War of the Spark printed "amass N"; the errata reads "amass Zombies N"
// (CR 701.47d), and the Amass primitive finds or makes the Zombie Army.
// The deathtouch grant is Eternal Skylord's static with another keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "61897e1b-dea0-44c2-9e38-6e2627647e17",
		Name:         "Vizier of the Scorpion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Amass: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Vizier of the Scorpion — amass Zombies 1", effects.Do(effects.Amass{Subtype: "Zombie", N: 1})),
		},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.ZombieTokenYouControl, "deathtouch"),
		},
	})
}
