package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whispering Shade — Creature — Shade {3}{B}, 1/1 (Odyssey, #5613):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.)
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Nantuko Shade's row (selfPumpRow) on a printed landwalker
// (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0036d062-10dc-4267-98b4-1d2f6b190a61",
		Name:            "Whispering Shade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated:       []effects.ActivatedAbility{effects.SelfPumpRow("Whispering Shade", "{B}", 1, 1)},
	})
}
