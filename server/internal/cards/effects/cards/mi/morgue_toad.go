package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Morgue Toad — Creature — Frog {2}{B}, 2/2 (Planeshift, S82 OLD b644, #5615):
//
//	"Sacrifice this creature: Add {U}{R}."
//
// Basal Thrull's sacrifice mana ability: a CR 605 mana ability (no target, adds mana,
// not a loyalty ability), resolved at once without the stack, producing one blue and
// one red.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c615540b-bf85-4e16-8ac2-9d541cba732f",
		Name:         "Morgue Toad",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Sacrifice: true},
			Produced: "{U}{R}",
			Label:    "Sacrifice this creature: Add {U}{R}",
		}},
	})
}
