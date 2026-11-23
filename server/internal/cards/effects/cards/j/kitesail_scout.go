package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kitesail Scout — Creature — Kor Scout {W}, 1/1 (Battle for Zendikar,
// #5547):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b365a7dd-fed1-40f8-9544-2b5bc1af761c",
		Name:            "Kitesail Scout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
