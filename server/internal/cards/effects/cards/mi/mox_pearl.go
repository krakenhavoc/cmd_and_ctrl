package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mox Pearl — Artifact {0}: "{T}: Add {W}." (Alpha, #5948). In the catalog so
// the card a collector conjures (ADR 0099 amendment 2026-11-24) does
// what it prints; also playable from a decklist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "824597b8-c89a-47ec-8526-7efc6e24ef0e",
		Name:         "Mox Pearl",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
