package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Veteran Cavalier — Creature — Human Knight {W}{W}, 2/2
// (Starter 1999, #5620):
//
//	"Vigilance (Attacking doesn't cause this creature to tap.)"
//
// Vigilance is printed data: attacking doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2d02cfce-a142-4b3b-a01f-4e156e75164e",
		Name:            "Veteran Cavalier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
