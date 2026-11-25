package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Lotus — Artifact {0}: "{T}, Sacrifice this artifact: Add three
// mana of any one color." (Alpha, #5948). One of the Power Nine Oracle
// of the Alpha conjures (ADR 0099 amendment 2026-11-24); Lotus Bloom's
// mana ability without the suspend.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5089ec1a-f881-4d55-af14-5d996171203b",
		Name:         "Black Lotus",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Sacrifice: true},
			Produced: effects.OneColorOfAmount(3),
			Label:    "{T}, Sacrifice this artifact: Add three mana of any one color",
		}},
	})
}
