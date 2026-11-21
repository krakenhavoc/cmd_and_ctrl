package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ornithopter of Paradise — Artifact Creature — Thopter {2}, 0/2
// (EDHREC rank 221):
//
//	"Flying
//	 {T}: Add one mana of any color."
//
// Birds of Paradise as a colourless artifact: the same five-colour
// pipe (all five, the commander's identity listed first, like
// the Birds and every Treasure), a 0/2 body, and a mana cost any
// deck can pay. Flying is printed, so it rides PrintedKeywords and
// the combat engine honours it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3a940dfa-a026-4969-8981-ef90cdfbe9ef",
		Name:            "Ornithopter of Paradise",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
