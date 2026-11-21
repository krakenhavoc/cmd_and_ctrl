package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gleaming Overseer — Creature — Zombie Wizard {1}{U}{B}, 1/4:
//
//	"When this creature enters, amass Zombies 1.
//	 Zombie tokens you control have hexproof and menace."
//
// Eternal Skylord's two halves with two keywords: the amass makes (or
// grows) a Zombie Army, and the static reads the effective subtypes, so
// the Army is a Zombie token the moment it exists. The two keywords are
// two layer 6 grants over the same filter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7716dff-5de7-4b52-966e-72ee3adc8d9a",
		Name:         "Gleaming Overseer",
		Purpose:      game.Purpose{Amass: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gleaming Overseer — amass Zombies 1",
				effects.Do(effects.Amass{Subtype: "Zombie", N: 1})),
		},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.ZombieTokenYouControl, "hexproof"),
			effects.KeywordGrant(effects.ZombieTokenYouControl, "menace"),
		},
	})
}
