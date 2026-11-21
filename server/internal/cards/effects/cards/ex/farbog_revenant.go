package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farbog Revenant — Creature — Spirit {2}{B}, 1/3
// (Shadows over Innistrad, #5543):
//
//	"Skulk (This creature can't be blocked by creatures with greater
//	 power.)
//	 Lifelink (Damage dealt by this creature also causes you to gain that
//	 much life.)"
//
// Skulk and lifelink are engine keywords carried as printed data (CR
// 702.118, CR 702.15).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6d32139f-839f-441b-930f-151816143142",
		Name:            "Farbog Revenant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"skulk", "lifelink"},
	})
}
