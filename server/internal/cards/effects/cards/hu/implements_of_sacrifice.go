package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Implements of Sacrifice — Artifact {2} (Fallen Empires, #5637):
//
//	"{1}, {T}, Sacrifice this artifact: Add two mana of any one color."
//
// A mana ability (CR 605.3b) with a mana component in its cost, paid
// before the artifact taps and is sacrificed. The two mana are one
// colour, picked once (OneColorOfAmount, Goldspan Dragon's Treasure).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "74558981-4226-4961-be33-0af867d0bdf2",
		Name:         "Implements of Sacrifice",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Mana: "{1}", Tap: true, Sacrifice: true},
			Produced: effects.OneColorOfAmount(2),
			Label:    "{1}, {T}, Sacrifice this artifact: Add two mana of any one color",
		}},
	})
}
