package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Tatters — Creature — Wraith {4}{B}, 4/2 (Zendikar, #5583):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)"
//
// Swampwalk is landwalk, an engine keyword (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e1e67a2f-61b6-49c5-8bcf-5d2b3f0df35d",
		Name:            "Bog Tatters",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
