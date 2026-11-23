package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Live Fast — Sorcery {2}{B}:
//
//	"You draw two cards, lose 2 life, and get {E}{E} (two energy
//	 counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04d02394-e17f-43cd-932f-f58e8cc56629",
		Name:         "Live Fast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Energy: 2},
		// Each sentence waits for the prompt the one before asks (a dredge offer, a CR 616
		// order on the life loss; CR 608.2c, #4155).
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
			}),
			effects.GetEnergy{N: 2},
		),
	})
}
