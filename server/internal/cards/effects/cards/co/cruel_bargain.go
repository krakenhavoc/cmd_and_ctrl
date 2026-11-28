package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Bargain — Sorcery {B}{B}{B} (Portal, S82 OLD batch 783, #5630):
//
//	"Draw four cards. You lose half your life, rounded up."
//
// The draw, then the loss as the next instruction (CR 608.2c): half the life
// total as it stands after the draw, rounded up (CR 119.3), and nothing at a
// life total of zero or less (Doomsday's rule).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcf105c1-d37b-4e92-a420-9872a2187764",
		Name:         "Cruel Bargain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4},
		OnResolve: effects.ResolveSteps(effects.DrawCards{N: 4}, effects.Instruction(func(ctx *effects.Context) error {
			return effects.R35Msc18DoomsdayLoseHalfOfYourLife(ctx, ctx.Controller())
		})),
	})
}
