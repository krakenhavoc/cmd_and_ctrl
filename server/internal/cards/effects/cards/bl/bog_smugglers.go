package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Smugglers — Creature — Human Mercenary {1}{B}{B}, 2/2 (Mercadian
// Masques, S82 OLD b683, #5619):
//
//	"Swampwalk (This creature can't be blocked as long as defending player
//	 controls a Swamp.)"
//
// Swampwalk is a printed keyword (CR 702.14a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5041dbc4-d01d-4f0a-b5c5-6e1df0976340",
		Name:            "Bog Smugglers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
