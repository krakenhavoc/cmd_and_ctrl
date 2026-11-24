package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mox Jet — Artifact {0}: "{T}: Add {B}." (Alpha, #5948). In the catalog so
// the card a collector conjures (ADR 0099 amendment 2026-11-24) does
// what it prints; also playable from a decklist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0677f49e-f8bf-4349-af52-2ccde9287c2e",
		Name:         "Mox Jet",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
