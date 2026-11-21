package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manalith — Artifact {3} (Magic 2012, #5573):
//
//	"{T}: Add one mana of any color."
//
// Alloy Myr's mana ability on an artifact: {T} for one mana of any color.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd9e416a-89b3-4912-be9b-49fce6a93dc9",
		Name:         "Manalith",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
