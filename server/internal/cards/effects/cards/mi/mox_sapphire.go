package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mox Sapphire — Artifact {0}: "{T}: Add {U}." (Alpha, #5948). In the catalog so
// the card a collector conjures (ADR 0099 amendment 2026-11-24) does
// what it prints; also playable from a decklist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d5ed1233-df87-4b90-8918-13922ec95249",
		Name:         "Mox Sapphire",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
