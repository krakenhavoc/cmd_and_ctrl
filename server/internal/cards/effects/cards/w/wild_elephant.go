package cardsw

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Wild Elephant — Creature — Elephant {3}{G}, 3/3 (Mirage, S82 OLD b806):
//
//	"Trample"
//
// Trample is an engine keyword carried as printed data (CR 702.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3b2ce431-7101-4256-827d-a14de9f867fd",
		Name:            "Wild Elephant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
