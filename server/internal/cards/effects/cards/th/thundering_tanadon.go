package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Tanadon — Artifact Creature — Phyrexian Beast {4}{G/P}{G/P}, 5/4 (New Phyrexia, #5575):
//
//	"({G/P} can be paid with either {G} or 2 life.)
//	 Trample"
//
// Both Phyrexian symbols are paid by the engine at cast (CR 107.4f);
// trample is the engine's (CR 702.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "487382ea-0050-4a3f-bcf6-0ce41d6f4ebf",
		Name:            "Thundering Tanadon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
