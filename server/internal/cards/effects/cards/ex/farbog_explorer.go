package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farbog Explorer — {2}{W} Creature — Human Scout 2/3 (Avacyn Restored, #5570):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)"
//
// Swampwalk, a landwalk keyword (CR 702.14).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f474c1f8-43bb-4e56-8131-27c71b3193d2",
		Name:            "Farbog Explorer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
