package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gerrard's Wisdom — Sorcery {2}{W}{W} (Weatherlight, #5629):
//
//	"You gain 2 life for each card in your hand."
//
// The hand is counted once, as the spell resolves (CR 608.2h); the
// spell itself is on the stack then, not in the hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3e30e8f2-f437-4620-a3bc-dd9c29ae6570",
		Name:         "Gerrard's Wisdom",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: 2 * effects.B14HandSize(ctx.Game, ctx.Controller())}.Apply(ctx)
		},
	})
}
