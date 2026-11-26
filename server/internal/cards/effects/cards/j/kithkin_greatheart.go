package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kithkin Greatheart — Creature — Kithkin Soldier {1}{W}, 2/1 (Lorwyn, S82
// OLD b545, #5593):
//
//	"As long as you control a Giant, this creature gets +1/+1 and has first
//	 strike."
//
// Two statics with one condition (Battle Brawler's shape): +1/+1 in layer
// 7c (CR 613.4c) and first strike in layer 6 (CR 613.1f), both re-read on
// every recompute. A changeling you control is a Giant (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6034e7c1-b3ec-4927-b517-f8e3ca87252e",
		Name:         "Kithkin Greatheart",
		Completeness: effects.CompletenessFull,
		Static:       effects.S82OldB545SelfWhileYouControlSubtype("Giant", 1, 1, "first strike", "As long as you control a Giant, this creature gets +1/+1 and has first strike."),
	})
}
