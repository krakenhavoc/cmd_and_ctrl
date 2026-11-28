package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Omen — Sorcery {1}{U} (Portal, S82 OLD batch 783, #5630):
//
//	"Look at the top three cards of your library, then put them back in any
//	 order. You may shuffle.
//	 Draw a card."
//
// Pondering Mage's chain (OldB159LookThreeMayShuffleThenDraw): the order is
// asked first, the shuffle second, and the draw follows on both answers (CR
// 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9b8c44a0-82ef-4beb-a762-c2a2a21380f3",
		Name:         "Omen",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.S82OldB783OmenLook(ctx)
		},
	})
}
