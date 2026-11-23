package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scale the Heights — Sorcery {2}{G} (Jumpstart 2022, #5948):
//
//	"Put a +1/+1 counter on up to one target creature. You gain 2 life.
//	 You may play an additional land this turn.
//	 Draw a card."
//
// With no target the counter is skipped and the rest still happens; the
// extra land play is a permission for this turn (Azusa's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "13d5b664-db13-406e-9956-3d9fe4177d33",
		Name:         "Scale the Heights",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, LifeGain: 2},
		Targets:      effects.TargetCreature("up to one target creature").WithCount(0, 1),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			effects.GainLife{Amount: 2},
			effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.GrantAdditionalLandPlayForEffect(ctx.Controller(), 1)
				return nil
			}),
			effects.DrawCards{N: 1},
		),
	})
}
