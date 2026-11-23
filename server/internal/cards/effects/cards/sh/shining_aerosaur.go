package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shining Aerosaur — Creature — Dinosaur {4}{W}, 3/4 (Ixalan, #5532):
//
//	"Flying"
//
// Flying is an engine keyword carried as printed data: it can be
// blocked only by creatures with flying or reach (CR 702.9b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7520962-46e0-4abd-bb19-569334913a05",
		Name:            "Shining Aerosaur",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
