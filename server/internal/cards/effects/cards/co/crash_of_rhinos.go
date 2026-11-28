package cardsco

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Crash of Rhinos — Creature — Rhino {6}{G}{G}, 8/4 (Mirage, S82 OLD b807):
//
//	"Trample"
//
// Trample is an engine keyword carried as printed data (CR 702.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b2d7251e-2848-4b85-8d99-7c43be163dfb",
		Name:            "Crash of Rhinos",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
