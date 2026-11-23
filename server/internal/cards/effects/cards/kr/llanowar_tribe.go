package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Tribe — Creature — Elf Druid {G}{G}{G}, 3/3 (EDHREC rank
// 2244):
//
//	"{T}: Add {G}{G}{G}."
//
// Three Llanowar Elves in one body. A creature mana ability with a
// tap cost, so summoning sickness applies (CR 302.6) — the engine
// enforces it inside ActivateManaAbility. The mana is three green
// slots, no pick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cfd0f0e6-1bbf-4450-97d2-c54907abb7e4",
		Name:         "Llanowar Tribe",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}{G}{G}",
			Label:    "Add {G}{G}{G}",
		}},
	})
}
