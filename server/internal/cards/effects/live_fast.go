package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Live Fast — Sorcery {2}{B}:
//
//	"You draw two cards, lose 2 life, and get {E}{E} (two energy
//	 counters)."
//
// ADR 0129 PR 1.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04d02394-e17f-43cd-932f-f58e8cc56629",
		Name:         "Live Fast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Energy: 2},
		// Each sentence waits for the prompt the one before asks (a dredge offer, a CR 616
		// order on the life loss; CR 608.2c, #4155).
		OnResolve: ResolveSteps(
			DrawCards{N: 2},
			Instruction(func(ctx *Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
			}),
			GetEnergy{N: 2},
		),
	})
}
