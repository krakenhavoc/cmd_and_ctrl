package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Lens — Artifact {3} (Invasion, #5616):
//
//	"{T}, Pay 1 life: Add one mana of any color."
//
// A mana ability (CR 605.1a) whose cost taps the Lens and pays 1 life
// (Blood Celebrant's life payment), producing one mana of the player's choice.
// With 0 life the payment cannot be made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2ce3832-a12d-4c7a-9bc2-51dc92764994",
		Name:         "Phyrexian Lens",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}, Pay 1 life: Add one mana of any color",
		}},
	})
}
