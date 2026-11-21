package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elfhame Wurm — Creature — Wurm {4}{G}, 5/4 (Dominaria United, S82):
//
//	"Vigilance, trample"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9a74a1dd-c36f-4900-b835-fc39dd725bde",
		Name:            "Elfhame Wurm",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
	})
}
