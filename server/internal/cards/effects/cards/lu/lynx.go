package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lynx — Creature — Cat {1}{G}, 2/1 (Portal Second Age,
// #5625):
//
//	"Forestwalk (This creature can't be blocked as long as defending player controls a Forest.)"
//
// Forestwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Forest (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "61652a2d-dfe6-4d4f-8fad-9de1960208cb",
		Name:            "Lynx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
