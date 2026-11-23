package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plated Crusher — Creature — Beast {4}{G}{G}{G}, 7/6 (Battle for
// Zendikar, #5547):
//
//	"Trample
//	 Hexproof (This creature can't be the target of spells or
//	 abilities your opponents control.)"
//
// Trample and hexproof are engine keywords carried as printed data (CR
// 702.19b, CR 702.11b): an opponent's spell or ability can't target it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "411f3ea6-36ad-4bbc-8cde-e79662a08ee9",
		Name:            "Plated Crusher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "hexproof"},
	})
}
