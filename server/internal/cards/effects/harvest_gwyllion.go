package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harvest Gwyllion — Creature — Hag {2}{W/B}{W/B}, 2/4 (Eventide, #5590):
//
//	"Wither (This deals damage to creatures in the form of -1/-1 counters.)"
//
// The printed keyword (CR 702.80): its damage to a creature is dealt as
// -1/-1 counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bcdd64d7-8414-488b-9b13-d5678d06625f",
		Name:            "Harvest Gwyllion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
	})
}
