package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mox Emerald — Artifact {0}: "{T}: Add {G}." (Alpha, #5948). In the catalog so
// the card a collector conjures (ADR 0099 amendment 2026-11-24) does
// what it prints; also playable from a decklist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "376ee366-e082-402f-b4db-6592fcfcacd2",
		Name:         "Mox Emerald",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
