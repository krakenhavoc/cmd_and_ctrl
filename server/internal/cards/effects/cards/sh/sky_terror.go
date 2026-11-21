package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sky Terror — Creature — Dinosaur {R}{W}, 2/2 (Ixalan, #5532):
//
//	"Flying, menace"
//
// Flying and menace are engine keywords carried as printed data: it can
// be blocked only by creatures with flying or reach (CR 702.9b), and it
// can't be blocked except by two or more creatures (CR 702.111b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5fe2ed67-55a6-4816-910f-0e3c38ea06a6",
		Name:            "Sky Terror",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "menace"},
	})
}
