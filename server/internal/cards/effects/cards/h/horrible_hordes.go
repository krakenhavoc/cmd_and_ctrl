package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Horrible Hordes — Artifact Creature — Spirit {3}, 2/2 (Mirage, #5633):
//
//	"Rampage 1 (Whenever this creature becomes blocked, it gets +1/+1 until end of
//	 turn for each creature blocking it beyond the first.)"
//
// Rampage is the engine's own keyword (CR 702.23).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79bfef02-3586-424a-aff9-a6ec0d4f1135",
		Name:            "Horrible Hordes",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"rampage 1"},
	})
}
