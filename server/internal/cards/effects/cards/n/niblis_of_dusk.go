package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niblis of Dusk — Creature — Spirit {2}{U}, 2/1
// (Shadows over Innistrad, #5543):
//
//	"Flying
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Flying and prowess are engine keywords carried as printed data (CR
// 702.9, CR 702.108).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d14d41a5-6f38-43c7-b95c-b4a723002c20",
		Name:            "Niblis of Dusk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
	})
}
