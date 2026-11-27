package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nomadic Elf (Invasion, S82 OLD batch 645, #5616):
//
//	Creature — Elf Nomad {1}{G}, 2/2:
//	"{1}{G}: Add one mana of any color."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "025bbfc8-bccd-4062-b758-7f9e1f8ad420",
		Name:         "Nomadic Elf",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Mana: "{1}{G}"}, "{1}{G}: Add one mana of any color."),
		},
	})
}
