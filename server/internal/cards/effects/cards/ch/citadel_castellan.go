package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Citadel Castellan — Creature — Human Knight {1}{G}{W}, 2/3 (Magic Origins, #5549):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	 Renown 2 (When this creature deals combat damage to a player, if it isn't renowned, put two +1/+1 counters on it and it becomes renowned.)"
//
// Keywords only: vigilance, and renown 2, the engine's keyword trigger
// (game/renown.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b9e93832-93b7-409d-91aa-13757225824d",
		Name:            "Citadel Castellan",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "renown 2"},
	})
}
