package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Adaptive Shimmerer — Creature — Insect {5}, 0/0 (Ikoria: Lair of Behemoths, #5512):
//
//	"Flash
//	 This creature enters with three +1/+1 counters on it."
//
// Flash, and the counters as a replacement on its entry (CR 614.1c), so
// it never sits at 0/0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8003ca2d-0c1d-47c1-bcd1-4609615c7a82",
		Name:            "Adaptive Shimmerer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 3, "Adaptive Shimmerer: enters with three +1/+1 counters"),
		},
	})
}
