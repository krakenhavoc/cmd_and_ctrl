package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Omen Hawker — Creature — Octopus Advisor {U}, 1/1 (March of the
// Machine, S82 MOM batch 13):
//
//	"{T}: Add {C}{U}. Spend this mana only to activate abilities."
//
// Pit Automaton's activate-only restriction (ManaRestrictActivate): the
// mana pays for any activated ability, a mana ability with a mana cost
// included, and refuses a spell. It is one ability that adds both
// symbols at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "215fd0b4-5817-4335-a0ed-1694e44bde41",
		Name:         "Omen Hawker",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}{U}",
			Label:        "{T}: Add {C}{U}. Spend this mana only to activate abilities.",
			Restrictions: []string{game.ManaRestrictActivate},
		}},
	})
}
