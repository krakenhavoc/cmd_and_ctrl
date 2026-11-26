package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Perilous Research — Instant {1}{U} (Commander 2011, #5574):
//
//	"Draw two cards, then sacrifice a permanent."
//
// The sacrifice is the controller's choice among every permanent they control,
// lands and the spell's own board included, asked after the draws have settled
// (CR 701.21a). With nothing they can sacrifice nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9243b539-5f9e-4bb3-9973-3d48366aa7cc",
		Name:         "Perilous Research",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				candidates := effects.PermanentsControlledBy(ctx.Game, ctx.Controller())
				if len(candidates) == 0 {
					return nil
				}
				return effects.SacrificeChoice{
					Player:     ctx.Controller(),
					Candidates: candidates,
					Question:   "Perilous Research — sacrifice a permanent",
				}.Apply(ctx)
			}),
		),
	})
}
