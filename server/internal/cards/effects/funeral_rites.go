package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Funeral Rites — Sorcery {2}{B} (Theros Beyond Death, #5514):
//
//	"You draw two cards, lose 2 life, then mill two cards."
//
// Live Fast's draw and life loss, then the mill, each step waiting for
// any prompt the one before asks (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "771210f7-0f00-4438-a886-a1f716e3d80e",
		Name:         "Funeral Rites",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: ResolveSteps(
			DrawCards{N: 2},
			Instruction(func(ctx *Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
			}),
			MillCards{N: 2},
		),
	})
}
