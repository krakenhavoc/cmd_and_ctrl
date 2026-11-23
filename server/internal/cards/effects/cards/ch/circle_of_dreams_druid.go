package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Circle of Dreams Druid — Creature — Elf Druid {G}{G}{G}, 2/1:
//
//	"{T}: Add {G} for each creature you control."
//
// Gaea's Cradle's mana ability on a creature: the count is read as the
// ability resolves (CR 605.3b), the Druid itself included, and a creature
// with summoning sickness can't pay the {T} (CR 302.6), which the engine
// supplies.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c857fa4-e823-494d-9ea4-7eb9e719c9b6",
		Name:         "Circle of Dreams Druid",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			ProducedAmount: effects.ProducedPerPermanent("G", effects.MatchCreature),
			Label:          "Add {G} for each creature you control",
		}},
	})
}
