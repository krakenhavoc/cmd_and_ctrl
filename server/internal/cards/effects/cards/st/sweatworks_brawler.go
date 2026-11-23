package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sweatworks Brawler — improvise (CR 702.126) with menace.
//
// Improvise is Spec.TapCost (Improvise()); the keywords are declared as
// printed and the engine does the rest. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9bb1de3f-b6af-4ee2-b896-a05aae50e238",
		Name:            "Sweatworks Brawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		TapCost:         effects.Improvise(),
	})
}
