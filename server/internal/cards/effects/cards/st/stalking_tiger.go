package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stalking Tiger — Creature — Cat {3}{G}, 3/3 (Portal, #5630):
//
//	"This creature can't be blocked by more than one creature."
//
// A block restriction on itself, as Outland Colossus has: a declaration
// putting a second blocker on it is illegal (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6af1429-276f-41a8-88d0-062661dc0cd4",
		Name:         "Stalking Tiger",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.MaxBlockers(effects.OnSelf(), 1)},
	})
}
