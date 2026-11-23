package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Striker — Creature — Goblin Berserker {1}{R}, 1/1 (Mirrodin,
// #5606):
//
//	"First strike, haste"
//
// Two printed keywords the engine already honours: first strike
// (CR 702.7b) and haste (CR 702.10b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c991d1b8-3adb-4854-9fd3-83f06aeb3941",
		Name:            "Goblin Striker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "haste"},
	})
}
