package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul's Grace — Instant {1}{W} (Shards of Alara, S82 OLD batch 512, #5588):
//
//	"You gain life equal to target creature's power."
//
// The power is read as the spell resolves; a target that is no longer legal
// counters the spell (CR 608.2b), so nothing is gained.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "42946216-59e2-46dc-b770-5e604512f949",
		Name:         "Soul's Grace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(ts[0].ID)
			if !ok {
				return nil
			}
			return effects.GainLife{Player: item.Controller, Amount: max(c.CurrentPower(), 0)}.Apply(ctx)
		},
	})
}
