package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f7716dff-5de7-4b52-966e-72ee3adc8d9a",
		Name:         "Gleaming Overseer",
		Purpose:      game.Purpose{Amass: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gleaming Overseer — amass Zombies 1",
				Do(Amass{Subtype: "Zombie", N: 1})),
		},
		Static: []game.StaticAbility{
			KeywordGrant(zombieTokenYouControl, "hexproof"),
			KeywordGrant(zombieTokenYouControl, "menace"),
		},
	})
}
