package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Earth Tremor — Instant {3}{R} (Commander Legends: Battle for Baldur's Gate,
// S82 GAP #5948):
//
//	"Earth Tremor deals damage to target creature or planeswalker equal to the
//	 number of lands you control."
//
// The lands are counted as the spell resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1dd7d11b-e3f3-4886-b412-0c085187869c",
		Name:         "Earth Tremor",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			n := effects.B02CountLandsControlledBy(ctx.Game, item.Controller)
			if n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: n}.Apply(ctx)
		},
	})
}
