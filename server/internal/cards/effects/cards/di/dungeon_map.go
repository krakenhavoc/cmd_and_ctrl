package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dungeon Map — Artifact {3}:
//
//	"{T}: Add {C}.
//	 {3}, {T}: Venture into the dungeon. Activate only as a sorcery."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20f1bba5-d398-4d6e-a76a-75e29df42aec",
		Name:         "Dungeon Map",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}, {T}: Venture into the dungeon. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect:       effects.Venture,
		}},
	})
}
