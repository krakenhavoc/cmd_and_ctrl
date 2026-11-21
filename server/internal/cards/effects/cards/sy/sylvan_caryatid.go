package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvan Caryatid — Creature — Plant {1}{G}, 0/3:
//
//	"Defender, hexproof
//	 {T}: Add one mana of any color."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13d4c46b-c2d6-44cc-a252-4a991d471854",
		Name:            "Sylvan Caryatid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "hexproof"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "{T}: Add one mana of any color",
		}},
	})
}
