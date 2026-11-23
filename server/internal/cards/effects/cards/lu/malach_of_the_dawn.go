package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Malach of the Dawn — Creature — Angel {2}{W}{W}, 2/4 (Planar Chaos, #5595):
//
//	"Flying
//	 {W}{W}{W}: Regenerate this creature."
//
// Flying is printed data (CR 702.9b); regeneration is on the ability's own
// source (CR 701.19), lrwB1RegenerateThisRow.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c351a325-42cb-43ad-8648-c204e2317b50",
		Name:            "Malach of the Dawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.LrwB1RegenerateThisRow("{W}{W}{W}")},
	})
}
