package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gluttonous Slime — Creature — Ooze {2}{G}, 2/2 (Conflux, #5587):
//
//	"Flash
//	 Devour 1 (As this creature enters, you may sacrifice any number
//	 of creatures. It enters with that many +1/+1 counters on it.)"
//
// Flash is a printed keyword (CR 702.8a). Devour 1 is the shared as-
// enters replacement (devour.go, CR 702.82a): the sacrifice happens as
// it enters, zero is a legal answer, and the counters ride the entry
// event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5797684a-a022-45d2-aa2e-4d375b33017e",
		Name:            "Gluttonous Slime",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements:    []game.ReplacementEffect{effects.Devour("Gluttonous Slime", 1)},
	})
}
