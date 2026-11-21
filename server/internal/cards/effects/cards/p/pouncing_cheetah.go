package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pouncing Cheetah — Creature — Cat {2}{G}, 3/2
// (Amonkhet, #5535):
//
//	"Flash"
//
// A printed keyword the engine runs off the card; the Spec declares it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1d5e32b1-1a2d-4b3d-80f2-f8c688b90587",
		Name:            "Pouncing Cheetah",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
	})
}
