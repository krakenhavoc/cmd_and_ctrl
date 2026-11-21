package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul's Might — Sorcery {4}{G} (Jumpstart 2022, #5948):
//
//	"Put X +1/+1 counters on target creature, where X is that
//	 creature's power."
//
// X is the creature's power as the spell resolves, through the layered view
// (a creature already pumped counts the pump).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3538a5e-3d08-4c61-bcb4-ef1f5e5b659d",
		Name:         "Soul's Might",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			ctx.Game.RecomputeLayersIfStaleLocked()
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found || c.CurrentPower() <= 0 {
				return nil
			}
			return effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: c.CurrentPower()}.Apply(ctx)
		},
	})
}
