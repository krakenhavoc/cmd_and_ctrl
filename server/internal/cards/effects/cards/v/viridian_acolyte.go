package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viridian Acolyte — Creature — Elf Shaman {G} (Darksteel, #5605):
//
//	"{1}, {T}: Add one mana of any color."
//
// Crystal Grotto's filtered mana ability on a creature: it doesn't use
// the stack (CR 605.3a), and a summoning-sick Acolyte can't pay {T} (CR
// 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9db7d8f0-0712-4e32-825c-1339e7a8d768",
		Name:         "Viridian Acolyte",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{W|U|B|R|G}",
			Label:    "{1}, {T}: Add one mana of any color",
		}},
	})
}
