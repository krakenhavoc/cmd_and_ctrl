package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pharika's Chosen — Creature — Snake {B}, 1/1 (Foundations Jumpstart):
//
//	"Deathtouch"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5a26a1b-d2be-406e-a6a9-1993045313b5",
		Name:            "Pharika's Chosen",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
