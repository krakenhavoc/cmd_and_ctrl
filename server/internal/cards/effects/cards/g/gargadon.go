package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gargadon — Creature — Beast {5}{R}{R}, 7/5 (Time Spiral):
//
//	"Trample
//	 Suspend 4—{1}{R}"
//
// Trample is a printed keyword; suspend is the special action Deep-Sea
// Kraken uses (CR 702.62).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73f808f0-3d56-46c9-967b-d452bb3611af",
		Name:            "Gargadon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SpecialActions: []game.SpecialAction{
			effects.Suspend(4, "{1}{R}"),
		},
	})
}
