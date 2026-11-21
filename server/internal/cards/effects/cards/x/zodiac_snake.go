package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zodiac Snake — Creature — Snake {2}{B}, 2/2 (Portal Three Kingdoms, #5622):
//
//	"Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)"
//
// Swampwalk is an engine keyword carried as printed data: it can't be
// blocked while the defending player controls a Swamp (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b97c4d9a-8ea0-40d0-839a-afcd785e807e",
		Name:            "Zodiac Snake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
