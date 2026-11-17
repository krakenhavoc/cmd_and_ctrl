package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "215fd0b4-5817-4335-a0ed-1694e44bde41",
		Name:         "Omen Hawker",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{C}{U}",
			Label:        "{T}: Add {C}{U}. Spend this mana only to activate abilities.",
			Restrictions: []string{game.ManaRestrictActivate},
		}},
	})
}
