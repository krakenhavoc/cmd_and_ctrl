package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mana Prism — Artifact {3} (Mirage, #5633):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color."
//
// Two mana abilities (CR 605.1a); the second is Arcum's Astrolabe's row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f4669822-716b-45c7-a7f2-040062509fe9",
		Name:         "Mana Prism",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.ColorlessTapRow(),
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "{1}, {T}: Add one mana of any color.",
			},
		},
	})
}
