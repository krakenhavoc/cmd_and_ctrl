package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Funeral Rites — Sorcery {2}{B} (Theros Beyond Death, #5514):
//
//	"You draw two cards, lose 2 life, then mill two cards."
//
// Live Fast's draw and life loss, then the mill, each step waiting for
// any prompt the one before asks (CR 608.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "771210f7-0f00-4438-a886-a1f716e3d80e",
		Name:         "Funeral Rites",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
			}),
			effects.MillCards{N: 2},
		),
	})
}
