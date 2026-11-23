package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zodiac Pig — Creature — Boar {3}{B}, 3/3 (Portal Three Kingdoms, #5622):
//
//	"Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)"
//
// Swampwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "eb0a0515-7fa4-4319-9fc3-7a31c608ce9b",
		Name:            "Zodiac Pig",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
