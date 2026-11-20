package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ukud Cobra — Creature — Snake {3}{B}, 2/5 (Foundations Jumpstart):
//
//	"Deathtouch"
//
// Deathtouch is an engine keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d38cbb37-0ddd-4768-bb05-f35b38730e44",
		Name:            "Ukud Cobra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
