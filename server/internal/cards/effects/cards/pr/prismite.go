package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prismite — Artifact Creature — Golem {2}, 2/1 (War of the Spark, #5520):
//
//	"{2}: Add one mana of any color."
//
// Crossroads Candleguide's filter: {2} for one mana of any color.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "feb4fcec-6f3a-4643-9ffb-69d3f5953322",
		Name:         "Prismite",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			effects.SamiAnyColorMana(effects.ManaAbilityCost{Mana: "{2}"}, "{2}: Add one mana of any color."),
		},
	})
}
