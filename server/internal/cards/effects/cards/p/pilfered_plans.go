package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pilfered Plans — Sorcery {1}{U}{B}:
//
//	"Target player mills two cards. Draw two cards."
//
// The shared TargetPlayerMills body for two cards, then the draw. The draw
// waits for any prompt the mill asks (CR 608.2c, #4155).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b734cfd9-afef-41f1-8377-68c885fb3a69",
		Name:         "Pilfered Plans",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.TargetPlayerMills(2)(ctx.Game, ctx.Item) }),
			effects.DrawCards{N: 2},
		),
	})
}
