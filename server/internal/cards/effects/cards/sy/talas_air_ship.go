package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Talas Air Ship — Creature — Human Pirate {3}{U}, 3/2 (Portal Second Age,
// #5625):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be blocked
// only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8cb3f047-bd8b-4a34-8a2b-ced2cc0487eb",
		Name:            "Talas Air Ship",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
