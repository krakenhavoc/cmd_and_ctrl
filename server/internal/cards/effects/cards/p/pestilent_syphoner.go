package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pestilent Syphoner — Creature — Phyrexian Insect {1}{B}, 1/1 (Phyrexia: All Will Be One, #5479):
//
//	"Flying
//	 Toxic 1"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b27303b-756e-45b9-956f-6bbbd83d31ab",
		Name:            "Pestilent Syphoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "toxic 1"},
	})
}
