package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mox Ruby — Artifact {0}: "{T}: Add {R}." (Alpha, #5948). In the catalog so
// the card a collector conjures (ADR 0099 amendment 2026-11-24) does
// what it prints; also playable from a decklist.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed85fa82-e4fa-434b-92a8-36b6075708d1",
		Name:         "Mox Ruby",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
	})
}
