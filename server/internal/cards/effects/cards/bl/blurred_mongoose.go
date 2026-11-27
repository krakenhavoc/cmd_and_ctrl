package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blurred Mongoose — Creature — Mongoose {1}{G}, 2/1 (Invasion, S82 #5616):
//
//	"This spell can't be countered.
//	 Shroud"
//
// The first line is Abrupt Decay's CantBeCountered, a rule about the spell on the stack; shroud is a printed keyword (CR 702.18).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "588a5bac-dcad-410e-8878-0230500c2e3a",
		Name:            "Blurred Mongoose",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"shroud"},
	})
}
