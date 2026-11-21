package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "73f808f0-3d56-46c9-967b-d452bb3611af",
		Name:            "Gargadon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		SpecialActions: []game.SpecialAction{
			Suspend(4, "{1}{R}"),
		},
	})
}
