package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardians of Meletis — Artifact Creature — Golem {3}, 0/6 (Theros, #5561):
//
//	"Defender (This creature can't attack.)"
//
// Printed keywords only; the engine reads them off the spec.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7604aac7-82ae-4f5b-888d-a008754f716a",
		Name:            "Guardians of Meletis",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
	})
}
