package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Ram-Gang — Creature — Goblin Warrior {R/G}{R/G}{R/G}, 3/3
// (Shadowmoor, #5591):
//
//	"Haste
//	 Wither (This deals damage to creatures in the form of -1/-1
//	 counters.)"
//
// Haste is a printed keyword. Wither rides PrintedKeywords (ADR 0056, CR
// 702.80a): its damage to a creature is dealt as -1/-1 counters, as on
// Stigma Lasher.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "30d2437a-87c9-4f88-8fb8-b686d6522677",
		Name:            "Boggart Ram-Gang",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste", "wither"},
	})
}
