package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oona's Gatewarden — Creature — Faerie Soldier {U/B}, 2/1 (Shadowmoor,
// #5591):
//
//	"Defender, flying
//	 Wither (This deals damage to creatures in the form of -1/-1
//	 counters.)"
//
// Defender and flying are printed keywords. Wither rides PrintedKeywords
// (ADR 0056, CR 702.80a): its damage to a creature is dealt as -1/-1
// counters, as on Stigma Lasher.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "439f7ac8-19f0-4227-82a7-48294adc2114",
		Name:            "Oona's Gatewarden",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "flying", "wither"},
	})
}
