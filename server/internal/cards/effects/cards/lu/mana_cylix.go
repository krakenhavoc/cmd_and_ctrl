package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mana Cylix — Artifact {1} (Conflux):
//
//	"{1}, {T}: Add one mana of any color."
//
// Capital City's filter ability on an artifact: a mana ability (CR 605.1a)
// whose cost is {1} and {T}, and whose output is one mana of any of the five
// colours, chosen as it is produced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47f20da0-cd31-4877-8eca-0c8389a66e89",
		Name:         "Mana Cylix",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{W|U|B|R|G}",
			Label:    "{1}, {T}: Add one mana of any color",
		}},
	})
}
