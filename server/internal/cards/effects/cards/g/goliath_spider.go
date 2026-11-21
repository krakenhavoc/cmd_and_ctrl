package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goliath Spider — Creature — Spider {6}{G}{G}, 7/6 (Ravnica: City of
// Guilds, #5600):
//
//	"Reach (This creature can block creatures with flying.)"
//
// Reach is an engine keyword carried as printed data: it can block
// creatures with flying (CR 702.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "17415789-159e-43cf-93f8-9926405ee8f1",
		Name:            "Goliath Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
