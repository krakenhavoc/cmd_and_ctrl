package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charging Rhino — Creature — Rhino {3}{G}{G}, 4/4 (Portal, #5630):
//
//	"This creature can't be blocked by more than one creature."
//
// A block restriction on itself, as Outland Colossus has: a declaration
// putting a second blocker on it is illegal (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26966ecb-15d3-47e5-ab63-e38510c87ecc",
		Name:         "Charging Rhino",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
	})
}
