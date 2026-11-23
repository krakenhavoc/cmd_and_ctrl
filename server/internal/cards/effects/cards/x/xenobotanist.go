package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Xenobotanist — Creature — Vulcan Scientist {1}{G}, 1/3 (Star Trek,
// #4973):
//
//	"{T}: Add one mana of any color."
//
// A plain mana ability (CR 605.1a); summoning sickness gates the {T}
// (CR 302.6) through the engine's own check.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9df60bf-f327-4d50-ab4d-99324fc2dcfe",
		Name:         "Xenobotanist",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
