package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoldering Crater — Land (Dominaria Remastered, #5948):
//
//	"This land enters tapped.
//	 {T}: Add {R}.
//	 Cycling {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "be09e83f-6486-4f42-8d2e-416cb95173f9",
		Name:         "Smoldering Crater",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
