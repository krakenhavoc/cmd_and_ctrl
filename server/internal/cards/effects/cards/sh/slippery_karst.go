package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slippery Karst — Land (Dominaria Remastered, #5948):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 Cycling {2}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "58bfd9a1-67ce-41d0-be38-f05addd1dd9e",
		Name:         "Slippery Karst",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
