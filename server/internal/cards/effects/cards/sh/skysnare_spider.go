package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skysnare Spider — Creature — Spider {4}{G}{G}, 6/6 (Magic Origins, #5549):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)
//	 Reach (This creature can block creatures with flying.)"
//
// Keywords only: vigilance and reach.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b820c29e-2428-4074-a024-9502848484d6",
		Name:            "Skysnare Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "reach"},
	})
}
