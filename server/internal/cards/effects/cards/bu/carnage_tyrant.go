package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carnage Tyrant — Creature — Dinosaur {4}{G}{G}, 7/6:
//
//	"This spell can't be countered.
//	 Trample, hexproof"
//
// The uncounterable half is Spec.CantBeCountered, the same gate
// Allosaurus Shepherd and Thrun use; trample and hexproof ride
// PrintedKeywords, whose consumers are the combat engine and
// CanBeTargetedBy.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8c411f4e-a091-447c-9450-d895b10b4985",
		Name:            "Carnage Tyrant",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"trample", "hexproof"},
	})
}
