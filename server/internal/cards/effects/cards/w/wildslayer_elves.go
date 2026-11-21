package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildslayer Elves — Creature — Elf Warrior {3}{G}, 3/3 (Shadowmoor,
// #5591):
//
//	"Wither (This deals damage to creatures in the form of -1/-1
//	 counters.)"
//
// Wither rides PrintedKeywords (ADR 0056, CR 702.80a): its damage to a
// creature is dealt as -1/-1 counters, as on Stigma Lasher.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c2f0b795-9f73-43db-aa8c-3b06963ba2b6",
		Name:            "Wildslayer Elves",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"wither"},
	})
}
