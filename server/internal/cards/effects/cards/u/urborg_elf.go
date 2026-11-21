package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Urborg Elf — Creature — Elf Druid {1}{G}, 1/1 (Apocalypse, #5614):
//
//	"{T}: Add {B}, {G}, or {U}."
//
// A tap mana ability with a choice of three colours (Abzan Devotee's
// "{W|B|G}" form), the colour picked as it is activated. Summoning
// sickness applies to the {T} cost (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9c382817-5c62-44ff-8f29-d283e65712a1",
		Name:         "Urborg Elf",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|G|U}",
			Label:    "{T}: Add {B}, {G}, or {U}",
		}},
	})
}
