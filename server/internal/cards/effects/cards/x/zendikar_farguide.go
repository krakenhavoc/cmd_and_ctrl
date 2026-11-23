package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zendikar Farguide — Creature — Elemental {4}{G}, 3/3 (Zendikar, #5583):
//
//	"Forestwalk (This creature can't be blocked as long as defending
//	 player controls a Forest.)"
//
// Forestwalk is landwalk, an engine keyword (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "01ab37a4-21d2-49f4-9b85-cb522a4d04ad",
		Name:            "Zendikar Farguide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
	})
}
