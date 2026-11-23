package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bastion Inventor — improvise (CR 702.126) with hexproof.
//
// Improvise is Spec.TapCost (Improvise()); the keywords are declared as
// printed and the engine does the rest. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "900ae045-0160-4e85-b283-e35399c2acec",
		Name:            "Bastion Inventor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		TapCost:         effects.Improvise(),
	})
}
