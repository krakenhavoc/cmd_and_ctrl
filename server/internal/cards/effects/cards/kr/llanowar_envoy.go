package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Envoy — Creature — Elf Scout {2}{G}, 3/2 (Dominaria, #5530):
//
//	"{1}{G}: Add one mana of any color."
//
// Orochi Leafcaller's filter at {1}{G}: a mana ability, so it does not
// use the stack (CR 605.3b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6f096e98-d351-4050-90ed-4c6f12e129d6",
		Name:         "Llanowar Envoy",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Mana: "{1}{G}"}, "{1}{G}: Add one mana of any color."),
		},
	})
}
