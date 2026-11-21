package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Darksteel Myr — Artifact Creature — Myr {3}, 0/1 (Scars of Mirrodin,
// #5577):
//
//	"Indestructible"
//
// A keyword body: indestructible is the engine's (CR 702.12b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f90fe86c-5cca-483b-93c2-6e856fa01c88",
		Name:            "Darksteel Myr",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
	})
}
