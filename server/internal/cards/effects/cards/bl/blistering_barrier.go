package cardsbl

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Blistering Barrier — Creature — Wall {2}{R}, 5/2 (Mirage, S82 OLD b806):
//
//	"Defender"
//
// Defender is an engine keyword carried as printed data (CR 702.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8783533d-dd98-4b80-b348-29051e1b5c28",
		Name:            "Blistering Barrier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
