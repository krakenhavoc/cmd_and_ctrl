package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Utopia Tree (Invasion, S82 OLD batch 646, #5616):
//
//	Creature — Plant {1}{G}, 0/2:
//	"{T}: Add one mana of any color."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73a9122c-60d2-4fdc-bc7b-3576bcbca241",
		Name:         "Utopia Tree",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Tap: true}, "{T}: Add one mana of any color."),
		},
	})
}
