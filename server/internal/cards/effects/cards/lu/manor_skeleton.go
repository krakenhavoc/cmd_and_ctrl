package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Manor Skeleton — Creature — Skeleton {1}{B}, 1/1 (Innistrad, #5572):
//
//	"Haste
//	 {1}{B}: Regenerate this creature."
//
// Printed haste and Carrion Wall's regeneration row at {1}{B}
// (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aa3f4938-ac4b-4f52-a8d3-481c1b387467",
		Name:            "Manor Skeleton",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{B}")},
	})
}
