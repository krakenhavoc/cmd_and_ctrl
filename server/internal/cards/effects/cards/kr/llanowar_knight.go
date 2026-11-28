package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Llanowar Knight — Creature — Elf Knight {G}{W}, 2/2 (Invasion, S82 #5616):
//
//	"Protection from black"
//
// Protection from black is a printed keyword (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6b26951f-57be-4524-ae0e-a102e84e81ef",
		Name:            "Llanowar Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
	})
}
