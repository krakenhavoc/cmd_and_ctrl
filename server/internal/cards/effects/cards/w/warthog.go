package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warthog — Creature — Boar {1}{G}{G}, 3/2 (Visions, #5631):
//
//	"Swampwalk (This creature can't be blocked as long as defending player
//	 controls a Swamp.)"
//
// Swampwalk is landwalk, an engine keyword carried as printed data: the
// block check reads it off the defending player's lands (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ae1f22a5-89cf-440c-b90d-fed194b2dee5",
		Name:            "Warthog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
