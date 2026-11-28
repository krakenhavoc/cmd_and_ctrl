package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rime Dryad — Creature — Dryad {G}, 1/2 (Ice Age, S82 OLD, #5636):
//
//	"Snow forestwalk (This creature can't be blocked as long as defending player controls a snow Forest.)"
//
// The snow landwalk keyword (CR 702.14): only a snow Forest turns it on, a plain Forest does not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0464176d-75d2-403a-aa9d-40e85f106d01",
		Name:            "Rime Dryad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"snow forestwalk"},
	})
}
